// Maintains a collection of knox clients (you need one per username) plus logic for
// parsing passwords out of the unusal storage format and workaround around what
// seems like a bug in the go knox client.

package knox

import (
	"flag"
	"fmt"
	"strings"

	log "github.com/golang/glog"

	"github.com/pinterest/knox"
	"vitess.io/vitess/go/flagutil"
)

var (
	knoxSupportedUsernames flagutil.StringListValue
)

// Client fetches passwords for a pre-determined set of users from knox.
type Client struct {
	clientsByUsername map[string]knox.Client
}

// CreateFromFlags creates Client for the set of users configured with -knox_supported_usernames.
func CreateFromFlags() *Client {
	clientsByUsername := make(map[string]knox.Client)
	for _, username := range knoxSupportedUsernames {
		knoxKey := fmt.Sprintf("mysql:rbac:%s:credentials", username)
		clientsByUsername[username] = requireFileClient(knoxKey)
	}

	return &Client{
		clientsByUsername: clientsByUsername,
	}
}

// GetActivePasswords returns a list of all valid passwords for the given user, or an error. This is only for
// validating passwords. For sending passwords, use GetPrimaryPassword.
func (c *Client) GetActivePasswords(user string) ([]string, error) {
	var result []string

	knoxClient, ok := c.clientsByUsername[user]
	if !ok {
		return nil, fmt.Errorf("User %s was not whitelisted with -knox_supported_usernames", user)
	}

	for _, unparsedActiveCredentials := range knoxClient.GetActive() {
		if unparsedActiveCredentials == "" {
			// TODO(dweitzman): Looks like there's a bug in the knox client that can return
			// empty entries in the list of active credentials. We should fix this in the client.
			continue
		}

		password, err := parseKnoxPassword(unparsedActiveCredentials, user)
		if err != nil {
			return nil, err
		}

		result = append(result, password)
	}
	return result, nil
}

// GetPrimaryPassword returns the primary passwords for the given user, or an error.
func (c *Client) GetPrimaryPassword(user string) (string, error) {
	knoxClient, ok := c.clientsByUsername[user]
	if !ok {
		return "", fmt.Errorf("User %s was not whitelisted with -knox_supported_usernames", user)
	}

	return parseKnoxPassword(knoxClient.GetPrimary(), user)
}

// Knox mashes usernames and credentials in a non-standard format (sadness) so we need custom code
// to parse it.
//
// The format is "<username>@%|<password>", so we ignore everything before the '|'.
func parseKnoxPassword(rawCredentials string, user string) (string, error) {
	splitCreds := strings.Split(rawCredentials, "|")

	if len(splitCreds) != 2 {
		return "", fmt.Errorf("Knox client returned unparsable credentials for user %s", user)
	}
	return splitCreds[1], nil
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
	flag.Var(&knoxSupportedUsernames, "knox_supported_usernames", "comma separated list of usernames to support for knox authentication")
}
