package knoxauth

import (
	"net"
	"reflect"
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
	return "knoxUserRole", "knoxActivePassword", nil
}

func (fkc *fakeKnoxClient) GetPrimaryCredentials(role string) (username string, password string, err error) {
	return "knoxUserName", "knoxActivePassword", nil
}

func (fkc *fakeKnoxClient) GetGroupsByRole(role string) []string {
	result, _ := fkc.roleMapping[role]
	return result
}

func TestPopulateCallerID(t *testing.T) {
	fakeKnoxClient := &fakeKnoxClient{
		roleMapping: map[string][]string{"knoxUserRole": {"group1", "group2"}},
	}

	auth := newAuthServerKnox(fakeKnoxClient)
	salt := []byte{}
	addr := net.IPAddr{}
	authResponse := mysql.ScramblePassword(salt, []byte("knoxActivePassword"))
	got, err := auth.ValidateHash(salt, "knoxUserName", authResponse, &addr)

	if err != nil {
		t.Errorf("Validating password failed: %v", err)
		return
	}

	want := &knoxUserData{
		user:   "knoxUserName",
		groups: []string{"group1", "group2"},
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
	_, err := auth.ValidateHash(salt, "wrongUserName", authResponse, &addr)

	wantErr := "Access denied for user 'wrongUserName' (credentials don't match knox) (errno 1045) (sqlstate 28000)"

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
	_, err := auth.ValidateHash(salt, "knoxUserName", authResponse, &addr)

	wantErr := "Access denied for user 'knoxUserName' (credentials don't match knox) (errno 1045) (sqlstate 28000)"

	if err == nil {
		t.Fatalf("should have rejected a bad password")
	}

	if err.Error() != wantErr {
		t.Errorf("Wrong failure. Want: %v. Got: %v", wantErr, err)
	}
}
