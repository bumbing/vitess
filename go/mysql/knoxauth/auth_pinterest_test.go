package knoxauth

import (
	"crypto/tls"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"

	"vitess.io/vitess/go/knox"
	"vitess.io/vitess/go/mysql"
)

var (
	_ knox.Client = (*fakeKnoxClient)(nil)
)

type fakeKnoxClient struct {
	roleMapping map[string][]string
}

func (fkc *fakeKnoxClient) GetActivePassword(user string) (role string, password string, err error) {
	if user == "knoxUserName" {
		return "knoxUserRole", "knoxActivePassword", nil
	}
	return "", "", fmt.Errorf("Unknown user: %v", user)
}

func (fkc *fakeKnoxClient) GetPrimaryCredentials(role string) (username string, password string, err error) {
	if role == "knoxUserName" {
		return "knoxUserName", "knoxActivePassword", nil
	}
	return "", "", fmt.Errorf("Unknown role: %v", role)
}

func (fkc *fakeKnoxClient) VerifyTLS(role string, tlsState *tls.ConnectionState, knoxAuthenticated bool) ([]string, error) {
	groups, ok := fkc.roleMapping[role]
	if !ok {
		return nil, fmt.Errorf("Bad role: %s", role)
	}
	return append(groups, "role="+role+",knoxAuth="+strconv.FormatBool(knoxAuthenticated)), nil
}

func TestAuthUsingKnox(t *testing.T) {
	fakeKnoxClient := &fakeKnoxClient{
		roleMapping: map[string][]string{"knoxUserRole": {"group1", "group2"}},
	}

	auth := newAuthServerKnox(fakeKnoxClient)
	salt := []byte{}
	addr := net.IPAddr{}
	authResponse := mysql.ScramblePassword(salt, []byte("knoxActivePassword"))
	got, err := auth.ValidateHash(salt, "knoxUserName", authResponse, &addr, nil)

	if err != nil {
		t.Errorf("Validating password failed: %v", err)
		return
	}

	want := &knoxUserData{
		user:   "knoxUserName",
		groups: []string{"group1", "group2", "role=knoxUserRole,knoxAuth=true"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrong user data. Expected %v, got %v", want, got)
	}
}

func TestAuthUsingTLS(t *testing.T) {
	fakeKnoxClient := &fakeKnoxClient{
		roleMapping: map[string][]string{"knoxUserRole": {"group1", "group2"}},
	}

	auth := newAuthServerKnox(fakeKnoxClient)
	salt := []byte{}
	addr := net.IPAddr{}
	authResponse := mysql.ScramblePassword(salt, []byte("randomPassword"))
	got, err := auth.ValidateHash(salt, "knoxUserRole", authResponse, &addr, nil)

	if err != nil {
		t.Errorf("Validating password failed: %v", err)
		return
	}

	want := &knoxUserData{
		user:   "knoxUserRole",
		groups: []string{"group1", "group2", "role=knoxUserRole,knoxAuth=false"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrong user data. Expected %v, got %v", want, got)
	}
}

func TestBadUser(t *testing.T) {
	fakeKnoxClient := &fakeKnoxClient{
		roleMapping: map[string][]string{"knoxUserRole": {"group1", "group2"}},
	}

	auth := newAuthServerKnox(fakeKnoxClient)
	salt := []byte{}
	addr := net.IPAddr{}
	authResponse := mysql.ScramblePassword(salt, []byte("wrongPassword"))
	_, err := auth.ValidateHash(salt, "wrongUserName", authResponse, &addr, nil)

	wantErr := "Access denied: Bad role: wrongUserName (errno 1045) (sqlstate 28000)"

	if err == nil {
		t.Fatalf("should have rejected a bad password")
	}

	if err.Error() != wantErr {
		t.Errorf("Wrong failure. Want: %v. Got: %v", wantErr, err)
	}
}

func TestBadPassword(t *testing.T) {
	fakeKnoxClient := &fakeKnoxClient{
		roleMapping: map[string][]string{"knoxUserRole": {"group1", "group2"}},
	}

	auth := newAuthServerKnox(fakeKnoxClient)
	salt := []byte{}
	addr := net.IPAddr{}
	authResponse := mysql.ScramblePassword(salt, []byte("wrongPassword"))
	_, err := auth.ValidateHash(salt, "knoxUserName", authResponse, &addr, nil)

	wantErr := "Access denied for user 'knoxUserName' (credentials don't match knox) (errno 1045) (sqlstate 28000)"

	if err == nil {
		t.Fatalf("should have rejected a bad password")
	}

	if err.Error() != wantErr {
		t.Errorf("Wrong failure. Want: %v. Got: %v", wantErr, err)
	}
}
