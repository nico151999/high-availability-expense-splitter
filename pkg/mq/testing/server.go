package testing

import (
	"net"
	"testing"

	natsserver "github.com/nats-io/nats-server/v2/server"
	natstestserver "github.com/nats-io/nats-server/v2/test"
)

func RunMQServer(t *testing.T) (*natsserver.Server, int) {
	opts := natstestserver.DefaultTestOptions
	// we cannot use port 0 since this causes nats to fall back to 4222
	// but we might want to have multiple instances running for parallel tests
	opts.Port = getFreePort(t)
	server := natstestserver.RunServer(&opts)
	return server, server.Addr().(*net.TCPAddr).Port
}

func getFreePort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
