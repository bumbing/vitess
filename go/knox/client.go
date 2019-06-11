// Maintains a collection of knox clients (you need one per username) plus logic for
// parsing passwords out of the unusal storage format and workaround around what
// seems like a bug in the go knox client.

package knox

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"sync"

	"github.com/pinterest/knox"
	"vitess.io/vitess/go/vt/log"
)

var (
	knoxRe          = regexp.MustCompile(`^([^@|]+)@([^@|]*)\|([^@|]*)$`)
	errParsingCreds = fmt.Errorf("failed to parse knox creds. Should match %v", knoxRe)
)

func init() {
}

// Client provides access to username/role/password data from knox.
type Client interface {
	GetActivePassword(user string) (role string, password string, err error)
	GetPrimaryCredentials(role string) (username string, password string, err error)
	// VerifyTLS typically just forwards to knox.VerifyTLS, but can be mocked by tests.
	VerifyTLS(username string, tlsState *tls.ConnectionState, knoxAuthenticated bool) ([]string, error)
}

// clientImpl fetches passwords for a pre-determined set of users from knox.
type clientImpl struct {
	clientsByRole map[string]knox.Client
}

var (
	knoxClientOnce        sync.Once
	knoxClient            Client
	parsedKnoxRoleMapping map[string][]string
)

func initKnoxClient() {
	clientsByRole := make(map[string]knox.Client)

	for _, authz := range pinterestAuthConfig.GroupAuthz {
		for _, knoxRole := range authz.KnoxRoles {
			_, alreadyCreated := clientsByRole[knoxRole]
			if !alreadyCreated {
				knoxKey := fmt.Sprintf("mysql:rbac:%s:credentials", knoxRole)
				clientsByRole[knoxRole] = requireFileClient(knoxKey)
			}
		}
	}

	knoxClient = &clientImpl{
		clientsByRole: clientsByRole,
	}
}

// CreateFromFlags creates Client for the set of users configured with -knox_supported_usernames.
func CreateFromFlags() Client {
	knoxClientOnce.Do(initKnoxClient)

	return knoxClient
}

// GetActivePassword the role and active password for the given user.
// Assumes that every user has only one password at any given time, and that password rotation also
// involves user rotation.
func (c *clientImpl) GetActivePassword(user string) (role string, password string, err error) {
	for role, knoxClient := range c.clientsByRole {
		for _, unparsedActiveCredentials := range knoxClient.GetActive() {
			if unparsedActiveCredentials == "" {
				// TODO(dweitzman): Looks like there's a bug in the knox client that can return
				// empty entries in the list of active credentials. We should fix this in the client.
				continue
			}

			candidateUsername, candidatePassword, _, err := parseKnoxCreds(unparsedActiveCredentials, user)
			if err != nil {
				log.Errorf("Problems parsing creds for role %s: %v", role, err)
				continue
			}

			if candidateUsername == user {
				return role, candidatePassword, nil
			}
		}
	}

	return "", "", fmt.Errorf("User %s not found for any of the whitelisted knox roles", user)
}

// VerifyTLS forwards to knox.VerifyTLS() which uses command line parameters to enforce authz.
func (c *clientImpl) VerifyTLS(username string, tlsState *tls.ConnectionState, knoxAuthenticated bool) ([]string, error) {
	return VerifyTLS(username, tlsState, knoxAuthenticated)
}

// GetPrimaryCredentials returns the primary credentials for the given user, or an error.
func (c *clientImpl) GetPrimaryCredentials(role string) (username string, password string, err error) {
	knoxClient, ok := c.clientsByRole[role]
	if !ok {
		return "", "", fmt.Errorf("Role %s was not whitelisted with -knox_supported_roles", role)
	}
	user, pass, _, err := parseKnoxCreds(knoxClient.GetPrimary(), role)
	return user, pass, err
}

// Knox mashes usernames and credentials in a non-standard format (sadness) so we need custom code
// to parse it.
//
// The format is "<username>@<host pattern>|<password>"
//
// Typically the host pattern is '%', but for credentials that should only be used
// by vttablet authenticating to mysqld it might be "localhost".
func parseKnoxCreds(rawCredentials string, role string) (username string, password string, host string, err error) {
	if match := knoxRe.FindStringSubmatch(rawCredentials); match != nil {
		user, host, pass := match[1], match[2], match[3]
		return user, pass, host, nil
	}

	return "", "", "", errParsingCreds
}

// requireFileClient is the same as NewFileClient, but panics if there is an
// error.
func requireFileClient(keyID string) knox.Client {
	c, err := knox.NewFileClient(keyID)
	if err != nil {
		log.Fatalf("Error making knox client for key %v: %v", keyID, err)
	}
	return c
}
