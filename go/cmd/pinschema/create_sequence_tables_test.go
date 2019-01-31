package main

import (
	"strings"
	"testing"
)

func TestPinschemaSequenceDDLs(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	got, err := buildSequenceDDLs(ddls, pinschemaConfig{})
	if err != nil {
		t.Error(err)
	}
	want := strings.Join(
		[]string{
			"create table if not exists `advertisers_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
			"create table if not exists `accepted_tos_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
			"create table if not exists `campaigns_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
			"create table if not exists `ad_groups_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
		}, "\n") + "\n"
	if got != want {
		t.Errorf("buildSequenceDDLs: \"%s\", want \"%s\"", got, want)
	}
}
