package knox

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"regexp"
	"testing"

	querypb "vitess.io/vitess/go/vt/proto/query"
)

func TestVerifyTLS(t *testing.T) {
	pepsiLatest := &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{
				Subject: pkix.Name{
					CommonName: "pepsi-latest-12345",
				},
			},
		},
	}

	testCases := []struct {
		callerid        *querypb.VTGateCallerID
		connectionState *tls.ConnectionState
		groupTLSRegexes map[string]*regexp.Regexp
		wantErr         string
	}{
		// no TLS, no required TLS for that user
		{
			callerid: &querypb.VTGateCallerID{
				Username: "scriptro",
				Groups:   []string{"readers"},
			},
			connectionState: nil,
			groupTLSRegexes: map[string]*regexp.Regexp{
				"writers": regexp.MustCompile("pepsi-latest-.*"),
			},
			wantErr: "",
		},
		// user with no TLS but requires TLS
		{
			callerid: &querypb.VTGateCallerID{
				Username: "scriptrw",
				Groups:   []string{"readers", "writers"},
			},
			connectionState: nil,
			groupTLSRegexes: map[string]*regexp.Regexp{
				"writers": regexp.MustCompile("pepsi-latest-.*"),
			},
			wantErr: `VerifyTLS: username:"scriptrw" groups:"readers" groups:"writers"  requires TLS for groups map[writers:pepsi-latest-.*]`,
		},
		// TLS but no valid CN
		{
			callerid: &querypb.VTGateCallerID{
				Username: "scriptrw",
				Groups:   []string{"readers", "writers"},
			},
			connectionState: pepsiLatest,
			groupTLSRegexes: map[string]*regexp.Regexp{
				"writers": regexp.MustCompile("pepsi-prod-.*"),
			},
			wantErr: `VerifyTLS: username:"scriptrw" groups:"readers" groups:"writers"  requires a TLS CN matching pepsi-prod-.* but only found []string{"pepsi-latest-12345"}`,
		},
		// TLS with valid CN
		{
			callerid: &querypb.VTGateCallerID{
				Username: "scriptrw",
				Groups:   []string{"readers", "writers"},
			},
			connectionState: pepsiLatest,
			groupTLSRegexes: map[string]*regexp.Regexp{
				"writers": regexp.MustCompile("pepsi-latest-.*"),
			},
			wantErr: "",
		},
	}

	for _, testCase := range testCases {
		gotErr := verifyTLSInternal(testCase.callerid, testCase.connectionState, testCase.groupTLSRegexes)

		var gotErrStr string
		if gotErr != nil {
			gotErrStr = gotErr.Error()
		}

		if gotErrStr != testCase.wantErr {
			t.Errorf("VerifyTLS(): Want: %v.  Got: %v", testCase.wantErr, gotErrStr)
		}
	}
}
