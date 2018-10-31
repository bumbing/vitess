package servenv

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"vitess.io/vitess/go/knox"
	"vitess.io/vitess/go/vt/log"
)

var (
	// KnoxAuthPlugin implements AuthPlugin interface
	_ Authenticator = (*KnoxAuthPlugin)(nil)
)

// internal type and value
type key int

var roleInfoKey key

// KnoxAuthPlugin implements knox-based username/password authentication for grpc.
type KnoxAuthPlugin struct {
	knoxClient knox.Client
}

// GetKnoxAuthenticatedRole pulls out the role from a context if it's been previously authenticated.
func GetKnoxAuthenticatedRole(ctx context.Context) (string, bool) {
	role, _ := ctx.Value(roleInfoKey).(string)
	return role, role != ""
}

// Authenticate implements AuthPlugin interface. This method will be used inside a middleware in grpc_server to authenticate
// incoming requests.
func (sa *KnoxAuthPlugin) Authenticate(ctx context.Context, fullMethod string) (context.Context, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["username"]) == 0 || len(md["password"]) == 0 {
			return nil, grpc.Errorf(codes.Unauthenticated, "username and password must be provided")
		}
		username := md["username"][0]
		password := md["password"][0]

		role, knoxPassword, err := sa.knoxClient.GetActivePassword(username)
		if err != nil {
			return nil, grpc.Errorf(codes.PermissionDenied, "auth failure: caller %q not registered with -knox_supported_usernames", username)
		}

		if password == knoxPassword {
			return context.WithValue(ctx, roleInfoKey, role), nil
		}
		return nil, grpc.Errorf(codes.PermissionDenied, "auth failure: caller %q provided invalid credentials", username)
	}
	return nil, grpc.Errorf(codes.Unauthenticated, "username and password must be provided")
}

func knoxAuthPluginInitializer() (Authenticator, error) {
	knoxAuthPlugin := &KnoxAuthPlugin{
		knoxClient: knox.CreateFromFlags(),
	}
	log.Info("knox auth plugin has initialized successfully")
	return knoxAuthPlugin, nil
}

func init() {
	RegisterAuthPlugin("knox", knoxAuthPluginInitializer)
}
