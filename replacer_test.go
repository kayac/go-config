package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/kayac/go-config"
)

type wordReplacer struct{}

func (r wordReplacer) Keyword() string { return "word" }
func (r wordReplacer) Replace(keys ...string) string {
	return strings.Join(keys, "_")
}

func TestReplacer(t *testing.T) {
	os.Setenv("PREFIX", "test_")
	config.Replacers(wordReplacer{})
	defer config.Replacers()

	src := []byte(`foo: '{{ env "PREFIX" }}{{ word "foo" "bar" }}'`)
	c := make(map[string]string)
	if err := config.LoadWithEnvBytes(&c, src); err != nil {
		t.Error(err)
	}
	if c["foo"] != "test_foo_bar" {
		t.Errorf("failed to inject FOO: %#v", c)
	}
}
