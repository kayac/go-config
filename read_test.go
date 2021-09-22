package config_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kayac/go-config"
)

var templateTestRead = []byte(`xxx{{ env "FOO" }}xxx`)

func TestRead(t *testing.T) {
	os.Setenv("FOO", "foobar")
	defer os.Unsetenv("FOO")

	loader := config.New()
	b, err := loader.ReadWithEnvBytes(templateTestRead)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(b, []byte("xxxfoobarxxx")) {
		t.Error("unexpected read result", string(b))
	}
}
