// A vtgate user authenticator that compares credentials to what's stored in knox.

package knoxauth

import (
	"bytes"
	"net"

	"vitess.io/vitess/go/knox"
	"vitess.io/vitess/go/mysql"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

// Init registers a knox-based authenticator for vtgate.
func Init() {
	knoxClient := knox.CreateFromFlags()
	mysql.RegisterAuthServerImpl("knox", newAuthServerKnox(knoxClient))
}

// authServerKnox can authenticate against credentials from knox.
type authServerKnox struct {
	knoxClient knox.Client
}

// newAuthServerKnox returns a new authServerKnox that authenticates with the provided
// username -> knox.Client pairs.
func newAuthServerKnox(knoxClient knox.Client) *authServerKnox {
	return &authServerKnox{knoxClient: knoxClient}
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
		groups := a.knoxClient.GetGroupsByRole(role)
		return &knoxUserData{user: user, groups: groups}, nil
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
