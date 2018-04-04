// Package opentsdb is copied without modification from magnus:
// https://phabricator.pinadmin.com/diffusion/M/browse/master/src/pinterest.com/logging/opentsdb/opentsdb.go
//
// NOTE(dweitzman): For the moment I don't want to modify this file at all from Magnus, but it does seem to
// have room for some improvement. Some observations:
// - It fails to record any stats if one of the fields (like GitSha) is an empty string.
// - tcollector already knows and sends your hostname, so there's no need to pass it explicitly.
// - What's the story with using the strange and brittle telent-style protocol instead of the OpenTSDB http API?
package opentsdb

import (
	"bytes"
	"encoding/json"
	"expvar"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"unicode"
)

// Metadata is the associated data to be included in openTSDB calls. All attributes are required.
type Metadata struct {
	GitSha    string
	Hostname  string
	Service   string
	GoVersion string
	Timestamp int64
}

// Metric is a specific openTSDB metric with a key, value, and associated tags.
type Metric struct {
	Key   string
	Value string
	Tags  map[string]string
}

func (m *Metric) format(md *Metadata) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("put expvar.%s %d %s host=%s service=%s version=%s go=%s",
		Sanitize(m.Key),
		md.Timestamp,
		Sanitize(m.Value),
		Sanitize(md.Hostname),
		Sanitize(md.Service),
		Sanitize(md.GitSha),
		Sanitize(md.GoVersion),
	))
	for k, v := range m.Tags {
		buffer.WriteString(fmt.Sprintf(" %s=%s", Sanitize(k), Sanitize(v)))
	}
	return buffer.Bytes()
}

// Send sends a list of metrics to an openTSDB host with associated metadata.
func Send(host string, ms []Metric, md *Metadata) error {
	addr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	for _, m := range ms {
		buffer.Write(m.format(md))
		buffer.WriteByte('\n')
	}

	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	sent := 0
	for sent < buffer.Len() {
		// Keep trying to send until an error happens or finished with data
		i, err := conn.Write(buffer.Bytes()[sent:])
		if err != nil {
			return err
		}
		sent += i
	}
	return nil
}

// Sanitize replaces all invalid characters with underscores.
// Per OpenTSDB docs, Only the following characters are allowed:
// a to z, A to Z, 0 to 9, -, _, ., / or Unicode letters (as per the specification)
func Sanitize(in string) string {
	var b bytes.Buffer
	for _, r := range in {
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '-' || r == '_' || r == '/' || r == '.' {
			b.WriteRune(r)
		} else {
			// For characters that would cause errors, write underscore instead
			b.WriteRune('_')
		}
	}
	return strings.ToLower(b.String())
}

// SanitizeForTag is the same as sanitize, but it also replaces '.' with underscore.
func SanitizeForTag(in string) string {
	var b bytes.Buffer
	for _, r := range in {
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '-' || r == '_' || r == '/' {
			b.WriteRune(r)
		} else {
			// For characters that would cause errors, write underscore instead
			b.WriteRune('_')
		}
	}
	return strings.ToLower(b.String())
}

// GetHTTPExpVars will return a list of Metrics from the ExpVars hosted at a given http server.
func GetHTTPExpVars(host string) ([]Metric, error) {

	url := "http://" + host + "/debug/vars"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error doing http GET: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("error closing response body: %v", err)
	}

	return fromJSON(body)
}

// GetExpVars will return a list of Metrics from the current ExpVars.
func GetExpVars() ([]Metric, error) {

	// This code is essentially duplicating the http handler for expvar.
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(&buffer, ",\n")
		}
		first = false
		fmt.Fprintf(&buffer, "%q: %s", kv.Key, kv.Value)
	})
	buffer.WriteByte('}')

	return fromJSON(buffer.Bytes())
}

func fromJSON(s []byte) ([]Metric, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(s, &obj); err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %v", err)
	}

	return walk("", obj), nil
}

func walk(prefix string, obj map[string]interface{}) []Metric {
	var ms []Metric
	for k, v := range obj {
		if prefix != "" {
			k = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]interface{}:
			ms = append(ms, walk(k, val)...)
		case float64:
			var nk string
			ts := map[string]string{}

			// split the key into parts and separate tags
			for _, s := range strings.Split(k, ".") {
				split := strings.SplitN(s, "#", 2)
				switch len(split) {
				case 1:
					if nk != "" {
						nk += "."
					}
					nk += s
				case 2:
					ts[split[0]] = split[1]
				}
			}
			ms = append(ms, Metric{
				Key:   nk,
				Value: strings.Replace(fmt.Sprintf("%g", val), "+", "", -1),
				Tags:  ts,
			})
		}
	}
	return ms
}
