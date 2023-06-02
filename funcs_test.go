package config_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/kayac/go-config"
)

func TestFuncs(t *testing.T) {
	t.Setenv("PREFIX", "test_")
	loader := config.New()
	loader.Funcs(template.FuncMap{
		"word": func(keys ...string) string {
			return strings.Join(keys, "_")
		},
	})

	src := []byte(`foo: '{{ env "PREFIX" }}{{ word "foo" "bar" }}'`)
	c := make(map[string]string)
	if err := loader.LoadWithEnvBytes(&c, src); err != nil {
		t.Error(err)
	}
	if c["foo"] != "test_foo_bar" {
		t.Errorf("failed to inject FOO: %#v", c)
	}
}
