/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vtgate

import (
	"crypto/x509"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"vitess.io/vitess/go/flagutil"
	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/callerid"
	"vitess.io/vitess/go/vt/callinfo"
	"vitess.io/vitess/go/vt/log"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vtgatepb "vitess.io/vitess/go/vt/proto/vtgate"
	"vitess.io/vitess/go/vt/servenv"
	"vitess.io/vitess/go/vt/vttls"
)

var (
	mysqlServerPort               = flag.Int("mysql_server_port", -1, "If set, also listen for MySQL binary protocol connections on this port.")
	mysqlServerBindAddress        = flag.String("mysql_server_bind_address", "", "Binds on this address when listening to MySQL binary protocol. Useful to restrict listening to 'localhost' only for instance.")
	mysqlServerSocketPath         = flag.String("mysql_server_socket_path", "", "This option specifies the Unix socket file to use when listening for local connections. By default it will be empty and it won't listen to a unix socket")
	mysqlTCPVersion               = flag.String("mysql_tcp_version", "tcp", "Select tcp, tcp4, or tcp6 to control the socket type.")
	mysqlAuthServerImpl           = flag.String("mysql_auth_server_impl", "static", "Which auth server implementation to use.")
	mysqlAllowClearTextWithoutTLS = flag.Bool("mysql_allow_clear_text_without_tls", false, "If set, the server will allow the use of a clear text password over non-SSL connections.")
	mysqlServerVersion            = flag.String("mysql_server_version", mysql.DefaultServerVersion, "MySQL server version to advertise.")

	mysqlServerRequireSecureTransport = flag.Bool("mysql_server_require_secure_transport", false, "Reject insecure connections but only if mysql_server_ssl_cert and mysql_server_ssl_key are provided")

	mysqlSslCert = flag.String("mysql_server_ssl_cert", "", "Path to the ssl cert for mysql server plugin SSL")
	mysqlSslKey  = flag.String("mysql_server_ssl_key", "", "Path to ssl key for mysql server plugin SSL")
	mysqlSslCa   = flag.String("mysql_server_ssl_ca", "", "Path to ssl CA for mysql server plugin SSL. If specified, server will require and validate client certs.")

	mysqlSslReloadFrequency = flag.Duration("mysql_server_ssl_reload_frequency", 0, "how frequently to poll for TLS cert/key/CA changes on disk")

	mysqlSlowConnectWarnThreshold = flag.Duration("mysql_slow_connect_warn_threshold", 0, "Warn if it takes more than the given threshold for a mysql connection to establish")

	mysqlConnReadTimeout  = flag.Duration("mysql_server_read_timeout", 0, "connection read timeout")
	mysqlConnWriteTimeout = flag.Duration("mysql_server_write_timeout", 0, "connection write timeout")
	mysqlQueryTimeout     = flag.Duration("mysql_server_query_timeout", 0, "mysql query timeout")
	// User-specific timeouts take precedence over mysqlQueryTimeout for ComQuery
	mysqlUserQueryTimeouts flagutil.StringDurationMapValue

	busyConnections int32
)

// vtgateHandler implements the Listener interface.
// It stores the Session in the ClientData of a Connection, if a transaction
// is in progress.
type vtgateHandler struct {
	vtg *VTGate
}

func newVtgateHandler(vtg *VTGate) *vtgateHandler {
	return &vtgateHandler{
		vtg: vtg,
	}
}

func (vh *vtgateHandler) NewConnection(c *mysql.Conn) {
}

func (vh *vtgateHandler) ConnectionClosed(c *mysql.Conn) {
	// Rollback if there is an ongoing transaction. Ignore error.
	var ctx context.Context
	var cancel context.CancelFunc
	if *mysqlQueryTimeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), *mysqlQueryTimeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}
	session, _ := c.ClientData.(*vtgatepb.Session)
	if session != nil {
		if session.InTransaction {
			defer atomic.AddInt32(&busyConnections, -1)
		}
		_, _, _ = vh.vtg.Execute(ctx, session, "rollback", make(map[string]*querypb.BindVariable))
	}
}

func (vh *vtgateHandler) ComQuery(c *mysql.Conn, query string, callback func(*sqltypes.Result) error) error {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx = callinfo.MysqlCallInfo(ctx, c)

	// Fill in the ImmediateCallerID with the UserData returned by
	// the AuthServer plugin for that user. If nothing was
	// returned, use the User. This lets the plugin map a MySQL
	// user used for authentication to a Vitess User used for
	// Table ACLs and Vitess authentication in general.
	im := c.UserData.Get()
	ef := callerid.NewEffectiveCallerID(
		c.User,                  /* principal: who */
		c.RemoteAddr().String(), /* component: running client process */
		"VTGate MySQL Connector" /* subcomponent: part of the client */)

	ctx = context.Background()

	if queryTimeout := vh.queryTimeout(im, query); queryTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, queryTimeout)
		defer cancel()
	}

	ctx = callerid.NewContext(ctx, ef, im)

	session, _ := c.ClientData.(*vtgatepb.Session)
	if session == nil {
		session = &vtgatepb.Session{
			Options: &querypb.ExecuteOptions{
				IncludedFields: querypb.ExecuteOptions_ALL,
			},
			Autocommit: true,
		}
		if c.Capabilities&mysql.CapabilityClientFoundRows != 0 {
			session.Options.ClientFoundRows = true
		}
	}

	if !session.InTransaction {
		atomic.AddInt32(&busyConnections, 1)
	}
	defer func() {
		if !session.InTransaction {
			atomic.AddInt32(&busyConnections, -1)
		}
	}()

	if c.SchemaName != "" {
		session.TargetString = c.SchemaName
	}
	if session.Options.Workload == querypb.ExecuteOptions_OLAP {
		err := vh.vtg.StreamExecute(ctx, session, query, make(map[string]*querypb.BindVariable), callback)
		return mysql.NewSQLErrorFromError(err)
	}
	session, result, err := vh.vtg.Execute(ctx, session, query, make(map[string]*querypb.BindVariable))
	c.ClientData = session
	err = mysql.NewSQLErrorFromError(err)
	if err != nil {
		return err
	}
	return callback(result)
}

func (vh *vtgateHandler) queryTimeout(im *querypb.VTGateCallerID, query string) time.Duration {
	// The reason for taking a query argument even though it's not used is to remind us that
	// we might consider supporting SQL comment annotations to set a per-query individual timeout.
	// sqlparser.ExtractCommentDirectives() could be useful for pulling timeout directives out of
	// a query.

	if im != nil {
		if userSpecificTimeout := mysqlUserQueryTimeouts[im.Username]; userSpecificTimeout > 0 {
			return userSpecificTimeout
		}
	}

	return *mysqlQueryTimeout
}

func (vh *vtgateHandler) WarningCount(c *mysql.Conn) uint16 {
	session, _ := c.ClientData.(*vtgatepb.Session)
	if session != nil {
		return uint16(len(session.GetWarnings()))
	}
	return 0
}

var mysqlListener *mysql.Listener
var mysqlUnixListener *mysql.Listener

// initiMySQLProtocol starts the mysql protocol.
// It should be called only once in a process.
func initMySQLProtocol() {
	// Flag is not set, just return.
	if *mysqlServerPort < 0 && *mysqlServerSocketPath == "" {
		return
	}

	// If no VTGate was created, just return.
	if rpcVTGate == nil {
		return
	}

	// Initialize registered AuthServer implementations (or other plugins)
	for _, initFn := range pluginInitializers {
		initFn()
	}
	authServer := mysql.GetAuthServer(*mysqlAuthServerImpl)

	switch *mysqlTCPVersion {
	case "tcp", "tcp4", "tcp6":
		// Valid flag value.
	default:
		log.Exitf("-mysql_tcp_version must be one of [tcp, tcp4, tcp6]")
	}

	// Create a Listener.
	var err error
	vh := newVtgateHandler(rpcVTGate)
	if *mysqlServerPort >= 0 {
		mysqlListener, err = mysql.NewListener(*mysqlTCPVersion, net.JoinHostPort(*mysqlServerBindAddress, fmt.Sprintf("%v", *mysqlServerPort)), authServer, vh, *mysqlConnReadTimeout, *mysqlConnWriteTimeout)
		if err != nil {
			log.Exitf("mysql.NewListener failed: %v", err)
		}
		if *mysqlServerVersion != "" {
			mysqlListener.ServerVersion = *mysqlServerVersion
		}
		if *mysqlSslCert != "" && *mysqlSslKey != "" {
			originalTLSConfig, err := vttls.ServerConfig(*mysqlSslCert, *mysqlSslKey, *mysqlSslCa)
			if err != nil {
				log.Exitf("grpcutils.TLSServerConfig failed: %v", err)
				return
			}
			mysqlListener.RequireSecureTransport = *mysqlServerRequireSecureTransport
			mysqlListener.TLSConfig.Store(originalTLSConfig)
			periodicallyReloadTLSCertificate(&mysqlListener.TLSConfig)
		}
		mysqlListener.AllowClearTextWithoutTLS = *mysqlAllowClearTextWithoutTLS
		// Check for the connection threshold
		if *mysqlSlowConnectWarnThreshold != 0 {
			log.Infof("setting mysql slow connection threshold to %v", mysqlSlowConnectWarnThreshold)
			mysqlListener.SlowConnectWarnThreshold = *mysqlSlowConnectWarnThreshold
		}
		// Start listening for tcp
		go mysqlListener.Accept()
	}

	if *mysqlServerSocketPath != "" {
		// Let's create this unix socket with permissions to all users. In this way,
		// clients can connect to vtgate mysql server without being vtgate user
		oldMask := syscall.Umask(000)
		mysqlUnixListener, err = newMysqlUnixSocket(*mysqlServerSocketPath, authServer, vh)
		_ = syscall.Umask(oldMask)
		if err != nil {
			log.Exitf("mysql.NewListener failed: %v", err)
			return
		}
		// Listen for unix socket
		go mysqlUnixListener.Accept()
	}
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

// newMysqlUnixSocket creates a new unix socket mysql listener. If a socket file already exists, attempts
// to clean it up.
func newMysqlUnixSocket(address string, authServer mysql.AuthServer, handler mysql.Handler) (*mysql.Listener, error) {
	listener, err := mysql.NewListener("unix", address, authServer, handler, *mysqlConnReadTimeout, *mysqlConnWriteTimeout)
	switch err := err.(type) {
	case nil:
		return listener, nil
	case *net.OpError:
		log.Warningf("Found existent socket when trying to create new unix mysql listener: %s, attempting to clean up", address)
		// err.Op should never be different from listen, just being extra careful
		// in case in the future other errors are returned here
		if err.Op != "listen" {
			return nil, err
		}
		_, dialErr := net.Dial("unix", address)
		if dialErr == nil {
			log.Errorf("Existent socket '%s' is still accepting connections, aborting", address)
			return nil, err
		}
		removeFileErr := os.Remove(address)
		if removeFileErr != nil {
			log.Errorf("Couldn't remove existent socket file: %s", address)
			return nil, err
		}
		listener, listenerErr := mysql.NewListener("unix", address, authServer, handler, *mysqlConnReadTimeout, *mysqlConnWriteTimeout)
		return listener, listenerErr
	default:
		return nil, err
	}
}

func shutdownMysqlProtocolAndDrain() {
	if mysqlListener != nil {
		mysqlListener.Close()
		mysqlListener = nil
	}
	if mysqlUnixListener != nil {
		mysqlUnixListener.Close()
		mysqlUnixListener = nil
	}

	if atomic.LoadInt32(&busyConnections) > 0 {
		log.Infof("Waiting for all client connections to be idle (%d active)...", atomic.LoadInt32(&busyConnections))
		start := time.Now()
		reported := start
		for atomic.LoadInt32(&busyConnections) != 0 {
			if time.Since(reported) > 2*time.Second {
				log.Infof("Still waiting for client connections to be idle (%d active)...", atomic.LoadInt32(&busyConnections))
				reported = time.Now()
			}

			time.Sleep(1 * time.Millisecond)
		}
	}
}

func init() {
	flag.Var(&mysqlUserQueryTimeouts, "mysql_user_query_timeouts", "per-user query timeouts. comma separated list of username:duration pairs. Takes precedence over -mysql_server_query_timeout")

	servenv.OnRun(initMySQLProtocol)
	servenv.OnTermSync(shutdownMysqlProtocolAndDrain)
}

var pluginInitializers []func()

// RegisterPluginInitializer lets plugins register themselves to be init'ed at servenv.OnRun-time
func RegisterPluginInitializer(initializer func()) {
	pluginInitializers = append(pluginInitializers, initializer)
}
