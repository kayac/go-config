package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kayac/go-config"
)

var testsJSON = []string{
	`{"foo":"bar"}`,
	`{"foo":"b\nar"}`,
	`{"foo":"b\"ar"}`,
	`{"foo":"b\u1234ar\t"}`,
	`{"foo":"\u2029"}`,
	`["A", "B", "C"]`,
	`"string"`,
}

var templateTestJSON = []byte(`{
	"json": "{{ env "JSON" | json_escape }}"
}`)

func TestJSONEncode(t *testing.T) {
	defer os.Unsetenv("JSON")
	for _, s := range testsJSON {
		os.Setenv("JSON", s)

		var before interface{}
		if err := json.Unmarshal([]byte(s), &before); err != nil {
			t.Error("failed to unmarshal before", err)
		}
		conf := make(map[string]string, 0)
		if err := config.LoadWithEnvJSONBytes(&conf, templateTestJSON); err != nil {
			t.Error("failed to LoadWithEnvJSONBytes", err)
		}
		t.Logf("%#v", conf)
		var after interface{}
		if err := json.Unmarshal([]byte(conf["json"]), &after); err != nil {
			t.Error("failed to unmarshal after", err)
		}
		if cmp.Diff(before, after) != "" {
			t.Errorf("%v != %v", before, after)
		}
	}
}
