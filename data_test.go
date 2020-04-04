package config_test

import (
	"testing"

	"github.com/kayac/go-config"
)

func TestData(t *testing.T) {
	person := struct {
		Name string
		Age  int
	}{
		Name: "Bob",
		Age:  25,
	}

	loader := config.New()
	loader.Data(config.DataMap{
		"Person": person,
	})

	src := []byte(`profile: '{{ .Person.Name }}({{ .Person.Age }})'`)
	c := make(map[string]string)
	if err := loader.LoadWithEnvBytes(&c, src); err != nil {
		t.Error(err)
	}
	if c["profile"] != "Bob(25)" {
		t.Errorf("failed to inject data: %#v", c)
	}
}
