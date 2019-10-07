package vtgate

import (
	"reflect"
	"testing"
	"time"
)

func TestParsePinterestComments(t *testing.T) {
	testcases := []struct {
		in   string
		want map[string]string
	}{
		{
			in: "/* ApplicationName=Pepsi.Service.GetPinPromotionsByAdGroupId, VitessTarget=foo, AdvertiserID=1234 */ select 1",
			want: map[string]string{
				"ApplicationName": "Pepsi.Service.GetPinPromotionsByAdGroupId",
				"VitessTarget":    "foo",
				"AdvertiserID":    "1234",
			},
		},
		{
			in: "/* ApplicationName=Pepsi.Service.GetPinPromotionsByAdGroupId, VitessTarget=patio[-80]@master */ select 1",
			want: map[string]string{
				"ApplicationName": "Pepsi.Service.GetPinPromotionsByAdGroupId",
				"VitessTarget":    "patio[-80]@master",
			},
		},
		{
			in: "/* VitessTarget=foo[bar] */ select 1",
			want: map[string]string{
				"VitessTarget": "foo[bar]",
			},
		},
		{
			in:   "/* VitessTarget= */ select 1",
			want: map[string]string{},
		},
		{
			in: "/* VitessTarget=, Foo=Bar */ select 1",
			want: map[string]string{
				"Foo": "Bar",
			},
		},
		{
			in: "select 1 /* VitessTarget=, Foo=Bar */",
			want: map[string]string{
				"Foo": "Bar",
			},
		},
	}

	for _, c := range testcases {
		got := parsePinterestOptionsFromQuery(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parsePinterestOptionsFromQuery(%#v). Want: %v. Got: %v", c.in, c.want, got)
		}
	}
}

func TestParsePinterestTargetOverride(t *testing.T) {
	testcases := []struct {
		query string
		opts  map[string]string
		want  string
	}{
		{
			query: "/* some comment */ select 1 /* another comment */",
			opts:  map[string]string{"VitessTarget": "foo"},
			want:  "foo",
		},
		{
			query: "/* some comment */ select 1 /* another comment */",
			opts:  map[string]string{"VitessTarget": "patio[-80]@master"},
			want:  "patio[-80]@master",
		},
		{
			query: "/* some comment */ select 1 /* another comment */",
			opts:  map[string]string{"VitessTarget": "patio:80-@rdonly"},
			want:  "patio:80-@rdonly",
		},
		{
			query: "/* some comment */ set autocommit=1 /* another comment */",
			opts:  map[string]string{"VitessTarget": "patio:80-@rdonly"},
			want:  "patio:80-@rdonly",
		},
		// Insert statements don't get v2 routing
		{
			query: "/* some comment */ insert into foo (a) values (1) /* some comment */",
			opts:  map[string]string{"VitessTarget": "patio:80-@rdonly"},
			want:  "patio@rdonly",
		},
		{
			query: "/* some comment */ insert into foo (a) values (1) /* some comment */",
			opts:  map[string]string{"VitessTarget": "patio:80-@master"},
			want:  "patio",
		},
		{
			query: "/* some comment */ insert into foo (a) values (1) /* some comment */",
			opts:  map[string]string{"VitessTarget": "patio[abcd]"},
			want:  "patio",
		},
		{
			query: "/* some comment */ insert into foo (a) values (1) /* some comment */",
			opts:  map[string]string{"VitessTarget": "patio[no_closing_bracket"},
			want:  "patio[no_closing_bracket",
		},
		{
			query: "/* some comment */ select 1 /* another comment */",
			opts:  map[string]string{"VitessTarget": ""},
			want:  "",
		},
	}

	for _, c := range testcases {
		got := maybeTargetOverrideForQuery(c.query, c.opts)
		if got != c.want {
			t.Errorf("maybeTargetOverrideForQuery(%#v, %#v). Want: %v. Got: %v", c.query, c.opts, c.want, got)
		}
	}
}

func TestParseQueryTimeout(t *testing.T) {
	queryTimeout(nil, map[string]string{"Timeout": "2s"})
	testcases := []struct {
		opts map[string]string
		want time.Duration
	}{
		{
			opts: map[string]string{"TimeoutMS": "2000"},
			want: time.Second * 2,
		},
		{
			opts: map[string]string{"TimeoutMS": "bad"},
			want: 0,
		},
		{
			opts: map[string]string{"TimeoutMS": "-10"},
			want: 0,
		},
	}

	for _, c := range testcases {
		got := queryTimeout(nil, c.opts)
		if got != c.want {
			t.Errorf("queryTimeout(nil, %#v). Want: %v. Got: %v", c.opts, c.want, got)
		}
	}
}
