package dbconfigs

import (
	"sync"

	"vitess.io/vitess/go/knox"
)

// knoxCredentialsServer is an implementation of CredentialsServer that takes credentials
// from knox using the golang knox client.
type knoxCredentialsServer struct {
	once       sync.Once
	knoxClient knox.Client
}

// GetUserAndPassword is part of the CredentialsServer interface
func (kcs *knoxCredentialsServer) GetUserAndPassword(role string) (string, string, error) {
	// Lazy-create the knox client to make sure command line flags have been parsed.
	kcs.once.Do(func() {
		kcs.knoxClient = knox.CreateFromFlags()
	})

	user, password, err := kcs.knoxClient.GetPrimaryCredentials(role)
	if err != nil {
		return "", "", err
	}
	return user, password, nil
}
