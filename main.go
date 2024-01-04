package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeirimpan/tcpprobe/probe"
)

func main() {
	nodesFlag := flag.String("nodes", "", "Comma-separated list of node addresses")
	probeDurFlag := flag.Duration("probe-interval", 1*time.Second, "Duration between probes")
	timeoutFlag := flag.Duration("timeout", 5*time.Second, "Timeout")

	flag.Parse()

	addresses := strings.Split(*nodesFlag, ",")

	// wait for SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	m := probe.NewManager(len(addresses), *probeDurFlag)
	for _, addr := range addresses {
		m.Add(&probe.Conn{Address: addr})
	}

	ctx, cancel = context.WithTimeout(ctx, *timeoutFlag)
	defer cancel()

	conn, err := m.GetHealthy(ctx)
	if err != nil {
		log.Printf("%s\n", err)
	} else {
		log.Printf("%s\n", conn.Address)
	}
}
