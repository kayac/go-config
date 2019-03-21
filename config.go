// Yaml Config Loader
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func init() {
	initEnvReplacer()
}

type customFunc func(data []byte) ([]byte, error)

type unmarshaler func([]byte, interface{}) error

// Load loads YAML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func Load(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, nil, configPaths, yaml.Unmarshal)
}

// Load loads JSON files from `configPaths`.
// and assigns decoded values into the `conf` value.
func LoadJSON(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, nil, configPaths, json.Unmarshal)
}

// Load loads TOML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func LoadTOML(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, nil, configPaths, toml.Unmarshal)
}

// LoadWithEnv loads YAML files with Env
// replace {{ env "ENV" }} to os.Getenv("ENV")
// if you set default value then {{ env "ENV" "default" }}
func LoadWithEnv(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, envReplacer, configPaths, yaml.Unmarshal)
}

// LoadWithEnvJSON loads JSON files with Env
func LoadWithEnvJSON(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, envReplacer, configPaths, json.Unmarshal)
}

// LoadWithEnvTOML loads TOML files with Env
func LoadWithEnvTOML(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, envReplacer, configPaths, toml.Unmarshal)
}

// Marshal serializes the value provided into a YAML document.
var Marshal = yaml.Marshal

// MarshalJSON returns the JSON encoding of v with indent by 2 white spaces.
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func loadWithFunc(conf interface{}, custom customFunc, configPaths []string, unmarshal unmarshaler) error {
	for _, configPath := range configPaths {
		err := loadConfig(configPath, conf, custom, unmarshal)
		if err != nil {
			// Go 1.12 text/template catches a panic raised in user-defined function.
			// https://golang.org/doc/go1.12#text/template
			if strings.Index(err.Error(), "must_env: environment variable") != -1 {
				panic(err)
			}
			return err
		}
	}
	return nil
}

func loadConfig(configPath string, conf interface{}, custom customFunc, unmarshal unmarshaler) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.Wrapf(err, "%s read failed", configPath)
	}
	if custom != nil {
		data, err = custom(data)
		if err != nil {
			return errors.Wrapf(err, "%s custom failed", configPath)
		}
	}
	if err := unmarshal(data, conf); err != nil {
		return errors.Wrapf(err, "%s parse failed", configPath)
	}
	return nil
}

var envRepTpl *template.Template

func initEnvReplacer() {
	envRepTpl = template.New("conf").Funcs(template.FuncMap{
		"env": func(keys ...string) string {
			v := ""
			for _, k := range keys {
				v = os.Getenv(k)
				if v != "" {
					return v
				}
				v = k
			}
			return v
		},
		"must_env": func(key string) string {
			if v, ok := os.LookupEnv(key); ok {
				return v
			}
			panic(fmt.Sprintf("environment variable %s is not defined", key))
		},
	})
}

func envReplacer(data []byte) ([]byte, error) {
	t, err := envRepTpl.Parse(string(data))
	if err != nil {
		return nil, errors.Wrap(err, "config parse by template failed")
	}
	buf := &bytes.Buffer{}
	if err = t.Execute(buf, nil); err != nil {
		return nil, errors.Wrap(err, "template attach failed")
	}
	return buf.Bytes(), nil
}
