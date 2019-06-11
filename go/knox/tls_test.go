package knox

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
	"reflect"
	"testing"
)

func makeCert(t *testing.T, commonName string, spiffe []string) *tls.ConnectionState {
	var uris []*url.URL
	for _, uri := range spiffe {
		parsedUri, err := url.Parse(uri)
		if err != nil {
			t.Error("Error setting up test with spiffe URI cert", err)
		}
		uris = append(uris, parsedUri)
	}

	return &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{
				Subject: pkix.Name{
					CommonName: commonName,
				},
				URIs: uris,
			},
		},
	}
}

func TestVerifyTLS(t *testing.T) {
	typicalConfig := authConfig{
		UserGroups: map[string][]string{
			"scriptro": []string{"reader"},
			"scriptrw": []string{"reader", "writer", "admin"},
		},
		GroupAuthz: map[string]subjectList{
			"admin": subjectList{
				TLSSubjects: []string{
					"spiffe://pin220.com/teletraan/pepsi/latest",
					"foo-*.ec2.pin220.com",
				},
			},
			"writer": subjectList{
				TLSSubjects: []string{
					"spiffe://pin220.com/teletraan/pepsi/latest",
					"foo-*.ec2.pin220.com",
				},
			},
			"reader": subjectList{
				TLSSubjects: []string{
					"spiffe://pin220.com/teletraan/pepsi/latest",
					"foo-*.ec2.pin220.com",
				},
				KnoxRoles: []string{"scriptro"},
			},
		},
	}

	testCases := []struct {
		desc            string
		username        string
		connectionState *tls.ConnectionState
		hasKnox         bool
		config          authConfig
		wantGroups      []string
		wantErr         string
	}{
		{
			desc:     "bad role name",
			username: "bad_user",
			config:   typicalConfig,
			wantErr:  "VerifyTLS: bad_user is not a recognized role with -pinterest_auth_config",
		},
		{
			desc:     "no TLS, no knox",
			username: "scriptro",
			config:   typicalConfig,
			wantErr:  "VerifyTLS: user scriptro in group [reader] requires auth {[spiffe://pin220.com/teletraan/pepsi/latest foo-*.ec2.pin220.com] [scriptro]} but only found []string{} (knox auth? false)",
		},
		{
			desc:       "no TLS, no required TLS for that user",
			username:   "scriptro",
			hasKnox:    true,
			config:     typicalConfig,
			wantGroups: []string{"reader"},
			wantErr:    "",
		},
		{
			desc:     "user with no TLS but requires TLS",
			username: "scriptrw",
			config:   typicalConfig,
			wantErr:  `VerifyTLS: user scriptrw in group [reader writer admin] requires auth {[spiffe://pin220.com/teletraan/pepsi/latest foo-*.ec2.pin220.com] [scriptro]} but only found []string{} (knox auth? false)`,
		},
		{
			desc:            "TLS but no valid CN or SPIFFE",
			username:        "scriptrw",
			connectionState: makeCert(t, "random-cn", []string{"spiffe://pin220.com/teletraan/pepsi/wrong"}),
			config:          typicalConfig,
			wantErr:         `VerifyTLS: user scriptrw in group [reader writer admin] requires auth {[spiffe://pin220.com/teletraan/pepsi/latest foo-*.ec2.pin220.com] [scriptro]} but only found []string{"random-cn", "spiffe://pin220.com/teletraan/pepsi/wrong"} (knox auth? false)`,
		},
		{
			desc:            "TLS with valid CN",
			username:        "scriptrw",
			connectionState: makeCert(t, "foo-1234.ec2.pin220.com", nil),
			config:          typicalConfig,
			wantGroups:      []string{"reader", "writer", "admin"},
			wantErr:         "",
		},
		{
			desc:            "TLS with valid SPIFFE",
			username:        "scriptrw",
			connectionState: makeCert(t, "random-cn", []string{"spiffe://pin220.com/teletraan/pepsi/latest"}),
			config:          typicalConfig,
			wantGroups:      []string{"reader", "writer", "admin"},
			wantErr:         "",
		},
	}

	for _, testCase := range testCases {
		gotGroups, gotErr := verifyTLSInternal(testCase.username, testCase.connectionState, testCase.hasKnox, testCase.config)

		var gotErrStr string
		if gotErr != nil {
			gotErrStr = gotErr.Error()
		}

		if gotErrStr != testCase.wantErr {
			t.Errorf("VerifyTLS(%v): Want: %v.  Got: %v", testCase.desc, testCase.wantErr, gotErrStr)
		}

		if !reflect.DeepEqual(gotGroups, testCase.wantGroups) {
			t.Errorf("VerifyTLS(%v): Want groups: %v. Got: %v", testCase.desc, testCase.wantGroups, gotGroups)
		}
	}
}
