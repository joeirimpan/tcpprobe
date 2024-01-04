# tcpprobe
Return the first active connection in concurrent fashion.

## Prerequisites

- Go 1.16 or higher

## Installation

1. Clone the repository:

```shell
git clone https://github.com/joeirimpan/tcpprobe.git
```

2. Build the application:

```shell
cd tcpprobe
make dist
```

## Usage

Run the application with the following command:
```
./tcpprobe.bin --nodes www.example.com:443,www.google1.com:443 --probe-interval 5s --timeout 15s
```