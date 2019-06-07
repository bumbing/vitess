package grpcclient

import (
	"flag"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	knoxRole = flag.String("grpc_auth_knox_role", "", "sets a knox role to authenticate with server as")
	// KnoxAuthClientCreds implements client interface to be able to WithPerRPCCredentials
	_ credentials.PerRPCCredentials = (*KnoxAuthClientCreds)(nil)
)

// KnoxAuthClientCreds holder for client credentials
type KnoxAuthClientCreds struct {
	role string
}

// GetRequestMetadata  gets the request metadata as a map from KnoxAuthClientCreds
func (c *KnoxAuthClientCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.role,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
// We don't want these credentials sent unencrypted, so this is set to true.
// The normal mysql protocol doesn't verify the server's identity in a secure
// way, so arguably we could require encryption without server identity and
// have equivalent security to mysql protocol.
func (c *KnoxAuthClientCreds) RequireTransportSecurity() bool {
	return true
}

// AppendKnoxAuth optionally appends static auth credentials if provided.
func AppendKnoxAuth(opts []grpc.DialOption) ([]grpc.DialOption, error) {
	if *knoxRole == "" {
		return opts, nil
	}
	clientCreds := KnoxAuthClientCreds{
		role: *knoxRole,
	}
	creds := grpc.WithPerRPCCredentials(&clientCreds)
	opts = append(opts, creds)
	return opts, nil
}

func init() {
	RegisterGRPCDialOptions(AppendKnoxAuth)
}
