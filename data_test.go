package config_test

import (
	"testing"

	"github.com/kayac/go-config"
)

func TestData(t *testing.T) {
	loader := config.New()
	data := make(map[string]string)
	data["foo"] = "DataFoo"
	loader.Data = data

	src := []byte(`foo: '{{ .foo }}'`)
	c := make(map[string]string)
	if err := loader.LoadWithEnvBytes(&c, src); err != nil {
		t.Error(err)
	}
	if c["foo"] != "DataFoo" {
		t.Errorf("failed to inject foo: %#v", c)
	}
}
