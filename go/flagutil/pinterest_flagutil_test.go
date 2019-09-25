package flagutil

import (
	"flag"
	"testing"
	"time"
)

type durationPair struct {
	in      string
	out     map[string]time.Duration
	wantErr string
}

func TestStringDurationMap(t *testing.T) {
	v := StringDurationMapValue(nil)
	var _ flag.Value = &v
	wanted := []durationPair{
		{
			in:  "tag1:2s,tag2:3m,tag3:0",
			out: map[string]time.Duration{"tag1": 2 * time.Second, "tag2": 3 * time.Minute, "tag3": 0},
		},
		{
			in:      `tag1:-1`,
			out:     map[string]time.Duration{"tag1": 0 * time.Second, "tag2": -1},
			wantErr: "time: missing unit in duration -1",
		},
	}
	for _, want := range wanted {
		if err := v.Set(want.in); err != nil {
			if want.wantErr == "" {
				t.Errorf("v.Set(%v): Unexpected error: %v", want.in, err)
			} else if want.wantErr != err.Error() {
				t.Errorf("v.Set(%v): Expected error: %v. Received error: %v", want.in, want.wantErr, err)
			}
			continue
		}
		if len(want.out) != len(v) {
			t.Errorf("want %#v, got %#v", want.out, v)
			continue
		}
		for key, value := range want.out {
			if v[key] != value {
				t.Errorf("want %#v, got %#v", want.out, v)
				continue
			}
		}
	}
}
