// Maintains a collection of knox clients (you need one per username) plus logic for
// parsing passwords out of the unusal storage format and workaround around what
// seems like a bug in the go knox client.

package knox

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	log "github.com/golang/glog"

	"github.com/pinterest/knox"
	"vitess.io/vitess/go/flagutil"
)

var (
	knoxSupportedRoles flagutil.StringListValue
	errParsingCreds    = errors.New("delimiter '@%|' is missing, which should separate username from password")
)

// Client fetches passwords for a pre-determined set of users from knox.
type Client struct {
	clientsByRole map[string]knox.Client
}

// CreateFromFlags creates Client for the set of users configured with -knox_supported_usernames.
func CreateFromFlags() *Client {
	clientsByRole := make(map[string]knox.Client)
	for _, username := range knoxSupportedRoles {
		knoxKey := fmt.Sprintf("mysql:rbac:%s:credentials", username)
		clientsByRole[username] = requireFileClient(knoxKey)
	}

	return &Client{
		clientsByRole: clientsByRole,
	}
}

// GetActivePassword the role and active password for the given user.
// Assumes that every user has only one password at any given time, and that password rotation also
// involves user rotation.
func (c *Client) GetActivePassword(user string) (role string, password string, err error) {
	for role, knoxClient := range c.clientsByRole {
		for _, unparsedActiveCredentials := range knoxClient.GetActive() {
			if unparsedActiveCredentials == "" {
				// TODO(dweitzman): Looks like there's a bug in the knox client that can return
				// empty entries in the list of active credentials. We should fix this in the client.
				continue
			}

			candidateUsername, candidatePassword, err := parseKnoxCreds(unparsedActiveCredentials, user)
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

// GetPrimaryCredentials returns the primary credentials for the given user, or an error.
func (c *Client) GetPrimaryCredentials(role string) (username string, password string, err error) {
	knoxClient, ok := c.clientsByRole[role]
	if !ok {
		return "", "", fmt.Errorf("Role %s was not whitelisted with -knox_supported_roles", role)
	}

	return parseKnoxCreds(knoxClient.GetPrimary(), role)
}

// Knox mashes usernames and credentials in a non-standard format (sadness) so we need custom code
// to parse it.
//
// The format is "<username>@%|<password>"
func parseKnoxCreds(rawCredentials string, role string) (username string, password string, err error) {
	splitCreds := strings.Split(rawCredentials, "@%|")
	if len(splitCreds) != 2 {
		return "", "", errParsingCreds
	}

	return splitCreds[0], splitCreds[1], nil
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

func init() {
	flag.Var(&knoxSupportedRoles, "knox_supported_roles", "comma separated list of roles to support for knox authentication")
}
