// A vtgate user authenticator that compares credentials to what's stored in knox.

package knoxauth

import (
	"bytes"
	"flag"
	"net"
	"strings"

	"vitess.io/vitess/go/flagutil"
	"vitess.io/vitess/go/knox"
	"vitess.io/vitess/go/mysql"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

var (
	knoxRoleMapping flagutil.StringMapValue
)

func init() {
	flag.Var(&knoxRoleMapping, "knox_role_mapping", "comma separated list of role1:group1:group2:...,role2:group1:... mappings from knox to table acl roles")
}

// Init registers a knox-based authenticator for vtgate.
func Init() {
	knoxClient := knox.CreateFromFlags()
	parsedKnoxRoleMapping := make(map[string][]string)
	for knoxRole, unparsedTableACLGroups := range knoxRoleMapping {
		groups := strings.Split(unparsedTableACLGroups, ":")
		// Make sure the group includes the role name itself, if it wasn't explicitly provided on the command line.
		shouldAddKnoxRole := true
		for _, group := range groups {
			if group == knoxRole {
				shouldAddKnoxRole = false
			}
		}
		if shouldAddKnoxRole {
			groups = append(groups, knoxRole)
		}
		parsedKnoxRoleMapping[knoxRole] = groups
	}
	mysql.RegisterAuthServerImpl("knox", newAuthServerKnox(knoxClient, parsedKnoxRoleMapping))
}

// authServerKnox can authenticate against credentials from knox.
type authServerKnox struct {
	knoxClient  knox.Client
	roleMapping map[string][]string
}

// newAuthServerKnox returns a new authServerKnox that authenticates with the provided
// username -> knox.Client pairs.
func newAuthServerKnox(knoxClient knox.Client, roleMapping map[string][]string) *authServerKnox {
	return &authServerKnox{
		knoxClient:  knoxClient,
		roleMapping: roleMapping,
	}
}

// AuthMethod is part of the AuthServer interface.
func (a *authServerKnox) AuthMethod(user string) (string, error) {
	return mysql.MysqlNativePassword, nil
}

// Salt is part of the AuthServer interface.
func (a *authServerKnox) Salt() ([]byte, error) {
	return mysql.NewSalt()
}

// ValidateHash is part of the AuthServer interface.
func (a *authServerKnox) ValidateHash(salt []byte, user string, authResponse []byte, remoteAddr net.Addr) (mysql.Getter, error) {
	role, password, err := a.knoxClient.GetActivePassword(user)
	if err != nil {
		return &knoxUserData{user: "", groups: nil}, mysql.NewSQLError(
			mysql.ERAccessDeniedError, mysql.SSAccessDeniedError, "Access denied: %s", err.Error())
	}

	computedAuthResponse := mysql.ScramblePassword(salt, []byte(password))
	if bytes.Compare(authResponse, computedAuthResponse) == 0 {
		return &knoxUserData{user: user, groups: a.roleMapping[role]}, nil
	}

	// None of the active credentials matched.
	return &knoxUserData{user: "", groups: nil}, mysql.NewSQLError(
		mysql.ERAccessDeniedError, mysql.SSAccessDeniedError,
		"Access denied for user '%v' (credentials don't match knox)", user)
}

// Negotiate is part of the AuthServer interface.
// It will never be called.
func (a *authServerKnox) Negotiate(c *mysql.Conn, user string, remotAddr net.Addr) (mysql.Getter, error) {
	panic("Negotiate should not be called as AuthMethod returned mysql_native_password")
}

// knoxUserData holds the username
type knoxUserData struct {
	user   string
	groups []string
}

// Get returns the wrapped username
func (kud *knoxUserData) Get() *querypb.VTGateCallerID {
	return &querypb.VTGateCallerID{Username: kud.user, Groups: kud.groups}
}
