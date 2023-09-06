package testing

import (
	"net"

	natsserver "github.com/nats-io/nats-server/v2/server"
	natstestserver "github.com/nats-io/nats-server/v2/test"
)

// RunMQServer creates a new MQ server meant for testing.
// Passing port -1 causes a random free port to be chosen.
// The actual port used will be returned along the server.
func RunMQServer(port int) (*natsserver.Server, int) {
	opts := natstestserver.DefaultTestOptions
	server := natstestserver.RunServer(&opts)
	return server, server.Addr().(*net.TCPAddr).Port
}
