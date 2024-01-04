package probe

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

var (
	ErrorNoHealthyConn = errors.New("no healthy connection")
)

// Conn is a connection
type Conn struct {
	Address string
}

// Manager manages connections
type Manager struct {
	conns []*Conn

	probeDuration time.Duration
}

// NewManager returns a new Manager
func NewManager(size int, probeDur time.Duration) *Manager {
	return &Manager{
		conns:         make([]*Conn, 0, size),
		probeDuration: probeDur,
	}
}

// Add adds a connection to the manager
func (m *Manager) Add(conn *Conn) {
	m.conns = append(m.conns, conn)
}

// GetHealthy returns a healthy connection from the manager
func (m *Manager) GetHealthy(parentCtx context.Context) (*Conn, error) {
	ticker := time.NewTicker(m.probeDuration)
	defer ticker.Stop()

	// single item buffer
	responseCh := make(chan *Conn, 1)

	// wait group to wait for all probes to finish
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(parentCtx)

	// start probes parallelly and wait for a response
	go func() {
	loop:
		for {
			startProbes(ctx, wg, m.conns, responseCh)

			select {
			case <-ctx.Done():
				break loop
			case <-ticker.C:
			}
		}

		// push a nil response to the channel to unblock the receiver
		select {
		case responseCh <- nil:
		default:
		}
	}()

	// wait for a response
	node := <-responseCh
	// cancel all probes
	cancel()

	// wait for all probes to finish
	wg.Wait()

	// if node is nil, return error
	if node == nil {
		return nil, ErrorNoHealthyConn
	}

	return node, nil
}

// startProbe starts a tcp probe for the given connection
func startProbe(ctx context.Context, wg *sync.WaitGroup, c *Conn, dialer *net.Dialer, responseCh chan *Conn) {
	defer wg.Done()

	// exit if context is already cancelled
	select {
	case <-ctx.Done():
		return
	default:
	}

	// dial the connection
	conn, err := dialer.DialContext(ctx, "tcp", c.Address)
	if err != nil {
		return
	}
	defer conn.Close()

	// push the connection to the response channel
	select {
	case responseCh <- c:
	default:
	}

}

// startProbes starts tcp probes for all connections in the connection manager
func startProbes(ctx context.Context, wg *sync.WaitGroup, conns []*Conn, responseCh chan *Conn) {
	dialer := &net.Dialer{Timeout: 1 * time.Second}

	for i := 0; i < len(conns); i++ {
		conn := conns[i]

		if conn == nil {
			continue
		}

		wg.Add(1)
		go startProbe(ctx, wg, conn, dialer, responseCh)
	}
}
