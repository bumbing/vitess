package knox

import (
	"testing"
)

func TestparseKnoxCreds(t *testing.T) {
	testcases := []struct {
		in           string
		wantUser     string
		wantPassword string
		wantHost     string
	}{
		{in: "vt_app@localhost|",
			wantUser:     "vt_app",
			wantPassword: "",
			wantHost:     "localhost"},
		{in: "vt_dba@localhost|test",
			wantUser:     "vt_dba",
			wantPassword: "test",
			wantHost:     "localhost"},
		{in: "vt_filtered@%|",
			wantUser:     "vt_filtered",
			wantPassword: "",
			wantHost:     "%"},
		{in: "vt_repl@%|testing",
			wantUser:     "vt_repl",
			wantPassword: "testing",
			wantHost:     "%"},
	}
	for _, c := range testcases {
		gotUser, gotPass, gotHost, err := parseKnoxCreds(c.in, "vtapp")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotUser != c.wantUser || gotPass != c.wantPassword || gotHost != c.wantHost {
			t.Errorf("parseKnoxCreds(%#v, vtapp) = (%#v, %#v, %#v), want (%#v, %#v, %#v)", c.in, gotUser, gotPass, gotHost, c.wantUser, c.wantPassword, c.wantHost)
		}
	}
}
