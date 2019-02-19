package knox

import (
	"crypto/tls"
	"flag"
	"fmt"
	"regexp"

	"vitess.io/vitess/go/flagutil"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

var (
	groupTLSRegexesFlag flagutil.StringMapValue

	// Regular expressions that much match the some TLS common name
	// if the authenticated user has a group listed in this map.
	groupTLSRegexes map[string]*regexp.Regexp
)

func init() {
	flag.Var(&groupTLSRegexesFlag, "group_tls_regexes", "user1:regex1,group2:regex2 requires user1 or users in group2 to have a TLS common name matching the regex")
}

func initTLS() {
	compiledRegexes := map[string]*regexp.Regexp{}
	for role, uncompiledRegex := range groupTLSRegexesFlag {
		compiledRegexes[role] = regexp.MustCompile(uncompiledRegex)
	}
}

// VerifyTLS makes sure the TLS certificate allows acting as the specified user.
func VerifyTLS(callerid *querypb.VTGateCallerID, tlsState *tls.ConnectionState) error {
	return verifyTLSInternal(callerid, tlsState, groupTLSRegexes)
}

func verifyTLSInternal(
	callerid *querypb.VTGateCallerID,
	tlsState *tls.ConnectionState,
	groupTLSRegexes map[string]*regexp.Regexp) error {

	// Collect a list of regexes that need to match the certificate CNs
	groupsToVerify := make(map[string]*regexp.Regexp)

	for _, role := range callerid.Groups {
		regex, ok := groupTLSRegexes[role]
		if ok {
			groupsToVerify[role] = regex
		}
	}

	if len(groupsToVerify) == 0 {
		return nil
	}

	if tlsState == nil {
		return fmt.Errorf("VerifyTLS: %v requires TLS for groups %v", callerid, groupsToVerify)
	}

	// Collect the CN values from the certificates.
	commonNames := []string{}
	peerCerts := tlsState.PeerCertificates
	if peerCerts != nil {
		for _, c := range peerCerts {
			commonName := c.Subject.CommonName
			commonNames = append(commonNames, commonName)
		}
	}

	// Make sure each regex matches at least one commonName
OUTER:
	for _, regex := range groupsToVerify {
		for _, commonName := range commonNames {
			matches := regex.MatchString(commonName)
			if matches {
				continue OUTER
			}
		}
		return fmt.Errorf("VerifyTLS: %v requires a TLS CN matching %v but only found %#v", callerid, regex, commonNames)
	}

	// All regexes matched some commonName
	return nil
}
