package config_test

import (
	"os"
	"testing"

	"github.com/kayac/go-config"
)

func TestDelims(t *testing.T) {
	os.Setenv("FOO", "test_foo")
	config.Delims("<%", "%>")
	defer config.Delims("{{", "}}")

	src := []byte(`foo: '<% env "FOO" %>'`)
	c := make(map[string]string)
	if err := config.LoadWithEnvBytes(&c, src); err != nil {
		t.Error(err)
	}
	if c["foo"] != "test_foo" {
		t.Errorf("failed to inject FOO: %#v", c)
	}
}
