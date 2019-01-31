package main

import (
	"strings"
	"testing"
)

func TestPinschemaRemoveAutoinc(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	got, err := removeAutoInc(ddls, pinschemaConfig{})
	if err != nil {
		t.Error(err)
	}
	want := strings.Join(
		[]string{
			"alter table advertisers modify id bigint(20) not null;",
			"alter table accepted_tos modify id bigint(20) not null;",
			"alter table campaigns modify id bigint(20) not null;",
			"alter table ad_groups modify id bigint(20) not null;",
		}, "\n") + "\n"
	if got != want {
		t.Errorf("buildSequenceDDLs: \"%s\", want \"%s\"", got, want)
	}
}
