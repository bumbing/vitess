package knox

import (
	"crypto/tls"
	"fmt"
	"strings"
)

// VerifyTLS makes sure the TLS certificate allows acting as the specified user.
func VerifyTLS(username string, tlsState *tls.ConnectionState, knoxAuthenticated bool) ([]string, error) {
	return verifyTLSInternal(username, tlsState, knoxAuthenticated, pinterestAuthConfig)
}

func verifyTLSInternal(
	username string,
	tlsState *tls.ConnectionState,
	knoxAuthenticated bool,
	config authConfig) ([]string, error) {

	// Make sure the name is legitimate.
	groups, ok := config.UserGroups[username]
	if !ok {
		return nil, fmt.Errorf("VerifyTLS: %s is not a recognized role with -pinterest_auth_config", username)
	}

	subjectNames := []string{}
	if tlsState != nil {
		// Collect the CN values from the certificates.
		peerCerts := tlsState.PeerCertificates
		if peerCerts != nil {
			for _, c := range peerCerts {
				subjectNames = append(subjectNames, c.Subject.CommonName)
				for _, dnsName := range c.DNSNames {
					subjectNames = append(subjectNames, dnsName)
				}
				for _, uri := range c.URIs {
					subjectNames = append(subjectNames, uri.String())
				}
			}
		}
	}

GROUPS:
	for _, group := range groups {
		groupAuthz := config.GroupAuthz[group]

		if knoxAuthenticated {
			for _, knoxRole := range groupAuthz.KnoxRoles {
				if knoxRole == username {
					continue GROUPS
				}
			}
		}

		for _, tlsSubject := range groupAuthz.TLSSubjects {
			// Allow one "*" wildcard in a tls subject for pre-spiffe host pattern matching
			parts := strings.SplitN(tlsSubject, "*", 2)
			for _, peerSubject := range subjectNames {
				if (len(parts) < 2 && tlsSubject == peerSubject) ||
					(len(parts) == 2 && strings.HasPrefix(peerSubject, parts[0]) &&
						strings.HasSuffix(peerSubject, parts[1])) {
					continue GROUPS
				}
			}
		}

		return nil, fmt.Errorf(
			"VerifyTLS: user %s in group %s requires auth %v but only found %#v (knox auth? %v)",
			username, groups, groupAuthz, subjectNames, knoxAuthenticated)
	}

	// All regexes matched some commonName
	return groups, nil
}
