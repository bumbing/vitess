package mysql

import (
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"testing"

	"golang.org/x/net/context"
	"vitess.io/vitess/go/vt/tlstest"
	"vitess.io/vitess/go/vt/vttls"
)

// TestTLSServer creates a Server with TLS support, then uses mysql
// client to connect to it.
func TestTLSServer(t *testing.T) {
	th := &testHandler{}

	authServer := NewAuthServerStatic("", "", 0)
	authServer.entries["user1"] = []*AuthServerStaticEntry{{
		Password: "password1",
	}}
	defer authServer.close()

	// Create the listener, so we can get its host.
	// Below, we are enabling --ssl-verify-server-cert, which adds
	// a check that the common name of the certificate matches the
	// server host name we connect to.
	l, err := NewListener("tcp", ":0", authServer, th, 0, 0, false)
	if err != nil {
		t.Fatalf("NewListener failed: %v", err)
	}
	defer l.Close()

	// Make sure hostname is added as an entry to /etc/hosts, otherwise ssl handshake will fail
	host, err := os.Hostname()
	if err != nil {
		t.Fatalf("Failed to get os Hostname: %v", err)
	}

	port := l.Addr().(*net.TCPAddr).Port

	// Create the certs.
	root, err := ioutil.TempDir("", "TestTLSServer")
	if err != nil {
		t.Fatalf("TempDir failed: %v", err)
	}
	defer os.RemoveAll(root)
	tlstest.CreateCA(root)
	tlstest.CreateSignedCert(root, tlstest.CA, "01", "server", host)
	tlstest.CreateSignedCert(root, tlstest.CA, "02", "client", "Client Cert")

	// Create the server with TLS config.
	serverConfig, err := vttls.ServerConfig(
		path.Join(root, "server-cert.pem"),
		path.Join(root, "server-key.pem"),
		path.Join(root, "ca-cert.pem"))
	if err != nil {
		t.Fatalf("TLSServerConfig failed: %v", err)
	}
	l.TLSConfig.Store(serverConfig)
	go l.Accept()

	// Setup the right parameters.
	params := &ConnParams{
		Host:  host,
		Port:  port,
		Uname: "user1",
		Pass:  "password1",
		// SSL flags.
		Flags:   CapabilityClientSSL,
		SslCa:   path.Join(root, "ca-cert.pem"),
		SslCert: path.Join(root, "client-cert.pem"),
		SslKey:  path.Join(root, "client-key.pem"),
	}

	// Run a 'select rows' command with results.
	conn, err := Connect(context.Background(), params)
	//output, ok := runMysql(t, params, "select rows")
	if err != nil {
		t.Fatalf("mysql failed: %v", err)
	}
	results, err := conn.ExecuteFetch("select rows", 1000, true)
	if err != nil {
		t.Fatalf("mysql fetch failed: %v", err)
	}
	output := ""
	for _, row := range results.Rows {
		r := make([]string, 0)
		for _, col := range row {
			r = append(r, col.String())
		}
		output = output + strings.Join(r, ",") + "\n"
	}

	if results.Rows[0][1].ToString() != "nice name" ||
		results.Rows[1][1].ToString() != "nicer name" ||
		len(results.Rows) != 2 {
		t.Errorf("Unexpected output for 'select rows': %v", output)
	}

	// make sure this went through SSL
	results, err = conn.ExecuteFetch("ssl echo", 1000, true)
	if err != nil {
		t.Fatalf("mysql fetch failed: %v", err)
	}
	if results.Rows[0][0].ToString() != "ON" {
		t.Errorf("Unexpected output for 'ssl echo': %v", results)
	}

	checkCountForTLSVer(t, versionTLS12, 1)
	checkCountForTLSVer(t, versionNoTLS, 0)
	conn.Close()

}
