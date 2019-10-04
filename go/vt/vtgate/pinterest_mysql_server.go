package vtgate

import (
	"crypto/x509"
	"flag"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"vitess.io/vitess/go/flagutil"
	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/vt/callerid"
	"vitess.io/vitess/go/vt/log"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vtrpcpb "vitess.io/vitess/go/vt/proto/vtrpc"
	"vitess.io/vitess/go/vt/sqlparser"
	"vitess.io/vitess/go/vt/topo/topoproto"
	"vitess.io/vitess/go/vt/vttls"
)

var (
	mysqlSslReloadFrequency = flag.Duration("mysql_server_ssl_reload_frequency", 0, "how frequently to poll for TLS cert/key/CA changes on disk")
	// User-specific timeouts take precedence over mysqlQueryTimeout for ComQuery
	mysqlUserQueryTimeouts flagutil.StringDurationMapValue
)

func init() {
	flag.Var(&mysqlUserQueryTimeouts, "mysql_user_query_timeouts", "per-user query timeouts. comma separated list of username:duration pairs. Takes precedence over -mysql_server_query_timeout")
}

const (
	QueryOptTarget          string = "VitessTarget"
	QueryOptTimeout         string = "TimeoutMS"
	QueryOptAdvertiserID    string = "AdvertiserID"
	QueryOptUserID          string = "UserID"
	QueryOptApplicationName string = "ApplicationName"
)

var pinterestDirectiveComments = regexp.MustCompile(`\b(?P<key>[a-zA-Z0-9]+)=(?P<value>[^,\s]+),?\b`)

func parsePinterestOptionsFromQuery(query string) map[string]string {
	_, marginComments := sqlparser.SplitMarginComments(query)
	result := map[string]string{}

	allIndexes := pinterestDirectiveComments.FindAllStringSubmatchIndex(marginComments.Leading, -1)
	for _, loc := range allIndexes {
		key := marginComments.Leading[loc[2]:loc[3]]
		value := marginComments.Leading[loc[4]:loc[5]]
		result[key] = value
	}

	allIndexes = pinterestDirectiveComments.FindAllStringSubmatchIndex(marginComments.Trailing, -1)
	for _, loc := range allIndexes {
		key := marginComments.Trailing[loc[2]:loc[3]]
		value := marginComments.Trailing[loc[4]:loc[5]]
		result[key] = value
	}

	return result
}

func queryTimeout(im *querypb.VTGateCallerID, pinterestOpts map[string]string) time.Duration {
	if timeoutOverride, ok := pinterestOpts[QueryOptTimeout]; ok {
		millis, err := strconv.ParseUint(timeoutOverride, 10, 32)
		if err != nil {
			warnings.Add("IgnoringBadQueryTimeout", 1)
			log.Infof("Can't parse timeout in query comment: %s, %s", timeoutOverride, err)
		} else {
			return time.Duration(millis) * time.Millisecond
		}
	}

	if im != nil {
		if userSpecificTimeout := mysqlUserQueryTimeouts[im.Username]; userSpecificTimeout > 0 {
			return userSpecificTimeout
		}
	}

	return *mysqlQueryTimeout
}

func getPinterestEffectiveCallerId(c *mysql.Conn, pinterestOpts map[string]string) *vtrpcpb.CallerID {
	// The effective caller ID is made to give us options for configuring the transaction limiter.
	// We can default to limiting the number of concurrent transactions by
	// (service, advertiser ID, endpoint) as a default. If an advertiser is causing issues across
	// multiple endpoints or an endpoint is causing issue across multiple advertisers, we can
	// reconfigure the tx limiter for vttablet during an incident.

	// Probably something like pepsirw
	principle := c.User

	// Advertiser ID if available, User ID otherwise
	component := "unknown"
	if advertiser, ok := pinterestOpts[QueryOptAdvertiserID]; ok {
		component = "advertiser:" + advertiser
	} else if user, ok := pinterestOpts[QueryOptUserID]; ok {
		component = "user:" + user
	}

	// Endpoint
	subcomponent := "VTGate MySQL Connector"
	if endpoint, ok := pinterestOpts[QueryOptApplicationName]; ok {
		subcomponent = endpoint
	}

	return callerid.NewEffectiveCallerID(principle, component, subcomponent)
}

// maybeTargetOverrideForQuery is a Pinterest-specific feature that can look in a comment
// like /* ApplicationName=Pepsi.Service.GetPinPromotionsByAdGroupId, VitessTarget=foo, AdvertiserID=1234 */
// and pull out the VitessTarget to use for a single query.
// The choice to parse this format for leading comments is because the primary user of these comments will be
// pepsi, which is a Java service using the connector-j jdbc driver for mysql. This is the format of commments
// adding by that driver when setClientInfo() is called on a connection.
func maybeTargetOverrideForQuery(query string, pinterestOpts map[string]string) string {
	stmtType := sqlparser.Preview(query)

	removeKeyspaceIdForInserts := func(target string) string {
		if stmtType != sqlparser.StmtInsert {
			return target
		}

		// NOTE(dweitzman): v2-targeting is disabled for insert statements
		// because v2 execution mode doesn't respect sequences and can
		// silently do the wrong thing. The vitess sharding model requires
		// insert statements to have a column that can be used to determine
		// the keyspace ID anyway, so v2-targeting an insert statement
		// has no benefits anyway.
		//
		// Remove anything from ":" or "[" until the end of the string
		destKeyspace, destTabletType, _, err := topoproto.ParseDestination(target, defaultTabletType)
		if err != nil {
			// Target is badly formatted. It'll generate an error later in the executor.
			return target
		}
		result := destKeyspace
		if destTabletType != defaultTabletType {
			// This case shouldn't really ever happen because inserts are always against master.
			result = result + "@" + strings.ToLower(destTabletType.String())
		}
		return result
	}

	if target, ok := pinterestOpts[QueryOptTarget]; ok {
		return removeKeyspaceIdForInserts(target)
	}
	return ""
}

// periodicallyReloadTLSCertificate is a Pinterest-specific function to make sure we can
// reload TLS certificates from disk every few minutes. Normandie certificates expire every 12
// hours. New certificates become available 2 hours before the old ones expire.
func periodicallyReloadTLSCertificate(tlsConfig *atomic.Value) {
	if *mysqlSslReloadFrequency > 0 {
		ticker := time.NewTicker(*mysqlSslReloadFrequency)
		go func() {
			var lastSerialNumber *big.Int

			for range ticker.C {
				newTLSConfig, err := vttls.ServerConfig(*mysqlSslCert, *mysqlSslKey, *mysqlSslCa)
				if err != nil {
					log.Errorf("Error refreshing TLS config: %v", err)
					warnings.Add("TlsReloadFailed", 1)
					continue
				}

				if len(newTLSConfig.Certificates) == 0 {
					log.Warningf("Refreshing TLS failed: certificate list is empty")
					warnings.Add("TlsReloadFailed", 1)
					continue
				}

				for _, cert := range newTLSConfig.Certificates {
					if len(cert.Certificate) == 0 {
						continue
					}

					parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
					if err != nil {
						log.Warningf("Failed to parse new certificate as x509: %v", err)
					} else {
						if lastSerialNumber == nil || parsedCert.SerialNumber == nil || lastSerialNumber.Cmp(parsedCert.SerialNumber) != 0 {
							log.Infof("Refreshed TLS cert Serial: %v. Subject: %v, Expires: %v", parsedCert.SerialNumber, parsedCert.Subject, parsedCert.NotAfter)
						}
						lastSerialNumber = parsedCert.SerialNumber
					}
				}
				tlsConfig.Store(newTLSConfig)
			}
		}()
	}
}
