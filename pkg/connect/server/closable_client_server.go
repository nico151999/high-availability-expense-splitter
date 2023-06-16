package server

import (
	"context"

	"github.com/nico151999/high-availability-expense-splitter/pkg/closable"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

type ClosableClientsServer interface {
	// closes the server
	Close() error
	// returns the entire list of clients which are closable
	GetClosableClients() []closable.Closer
}

func CloseClients(ctx context.Context, client ClosableClientsServer) error {
	gr, _ := errgroup.WithContext(ctx)
	// as further connections to servers are added they all need to be closed here
	for _, c := range client.GetClosableClients() {
		closeable := c
		gr.Go(func() error {
			if err := closeable.Close(); err != nil {
				return eris.Wrap(err, "failed closing client")
			}
			return nil
		})
	}
	if err := gr.Wait(); err != nil {
		return eris.Wrap(err, "failed closing at least one client")
	}
	return nil
}
