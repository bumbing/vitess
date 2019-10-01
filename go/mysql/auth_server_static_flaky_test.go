package mysql

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestStaticConfigHUPWithRotation(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "mysql_auth_server_static_file.json")
	if err != nil {
		t.Fatalf("couldn't create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	*mysqlAuthServerStaticFile = tmpFile.Name()

	savedReloadInterval := *mysqlAuthServerStaticReloadInterval
	defer func() { *mysqlAuthServerStaticReloadInterval = savedReloadInterval }()
	*mysqlAuthServerStaticReloadInterval = 10 * time.Millisecond

	oldStr := "str1"
	jsonConfig := fmt.Sprintf("{\"%s\":[{\"Password\":\"%s\"}]}", oldStr, oldStr)
	if err := ioutil.WriteFile(tmpFile.Name(), []byte(jsonConfig), 0600); err != nil {
		t.Fatalf("couldn't write temp file: %v", err)
	}

	InitAuthServerStatic()
	defer func() {
		// delete registered Auth server
		for auth := range authServers {
			delete(authServers, auth)
		}
	}()
	aStatic := GetAuthServer("static").(*AuthServerStatic)

	if aStatic.Entries[oldStr][0].Password != oldStr {
		t.Fatalf("%s's Password should still be '%s'", oldStr, oldStr)
	}

	hupTestWithRotation(t, tmpFile, oldStr, "str4")
	hupTestWithRotation(t, tmpFile, "str4", "str5")
}
