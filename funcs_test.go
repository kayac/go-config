package config_test

import (
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/kayac/go-config"
)

func TestFuncs(t *testing.T) {
	os.Setenv("PREFIX", "test_")
	config.Funcs(template.FuncMap{
		"word": func(keys ...string) string {
			return strings.Join(keys, "_")
		},
	})

	src := []byte(`foo: '{{ env "PREFIX" }}{{ word "foo" "bar" }}'`)
	c := make(map[string]string)
	if err := config.LoadWithEnvBytes(&c, src); err != nil {
		t.Error(err)
	}
	if c["foo"] != "test_foo_bar" {
		t.Errorf("failed to inject FOO: %#v", c)
	}
}
