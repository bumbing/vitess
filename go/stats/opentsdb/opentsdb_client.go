// Package opentsdb is copied from magnus:
// https://phabricator.pinadmin.com/diffusion/M/browse/master/src/pinterest.com/logging/opentsdb/opentsdb.go
//
// Unused parts of the code have been removed.
//
// NOTE(dweitzman): For the moment I don't want to modify this file at all from Magnus, but it does seem to
// have room for some improvement. Some observations:
// - It fails to record any stats if one of the fields (like GitSha) is an empty string.
// - tcollector already knows and sends your hostname, so there's no need to pass it explicitly.
// - We should switch to an open-source REST-based opentsdb client instead of this code.
package opentsdb

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"unicode"
)

// Metadata is the associated data to be included in openTSDB calls. All attributes are required.
type Metadata struct {
	GitSha    string
	Hostname  string
	Service   string
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
	buffer.WriteString(fmt.Sprintf("put %s %d %s host=%s service=%s version=%s",
		Sanitize(m.Key),
		md.Timestamp,
		Sanitize(m.Value),
		Sanitize(md.Hostname),
		Sanitize(md.Service),
		Sanitize(md.GitSha),
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
