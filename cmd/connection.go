package cmd

import (
	"fmt"

	"google.golang.org/grpc"
)

// dial opens a connection to the configured todo service. The caller is
// responsible for closing the returned connection.
func dial() (*grpc.ClientConn, error) {
	secure := requireSecureConnection(rootOpts.serviceURI) && !rootOpts.allowInsecureConnection

	conn, err := grpc.NewClient(
		rootOpts.serviceURI,
		grpc.WithTransportCredentials(getConnectionCredentials(secure)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service: %w", err)
	}

	return conn, nil
}
