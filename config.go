// Yaml Config Loader
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var defaultCustomizer = &customizer{}

type customFunc func(data []byte) ([]byte, error)

type unmarshaler func([]byte, interface{}) error

// Load loads YAML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func Load(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, nil, yaml.Unmarshal)
}

// Load loads JSON files from `configPaths`.
// and assigns decoded values into the `conf` value.
func LoadJSON(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, nil, json.Unmarshal)
}

// Load loads TOML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func LoadTOML(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, nil, toml.Unmarshal)
}

// LoadBytes loads YAML bytes
func LoadBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, nil, yaml.Unmarshal)
}

// LoadJSONBytes loads JSON bytes
func LoadJSONBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, nil, json.Unmarshal)
}

// LoadTOMLBytes loads TOML bytes
func LoadTOMLBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, nil, toml.Unmarshal)
}

// LoadWithEnv loads YAML files with Env
// replace {{ env "ENV" }} to os.Getenv("ENV")
// if you set default value then {{ env "ENV" "default" }}
func LoadWithEnv(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, defaultCustomizer.Customize, yaml.Unmarshal)
}

// LoadWithEnvJSON loads JSON files with Env
func LoadWithEnvJSON(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, defaultCustomizer.Customize, json.Unmarshal)
}

// LoadWithEnvTOML loads TOML files with Env
func LoadWithEnvTOML(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, defaultCustomizer.Customize, toml.Unmarshal)
}

// LoadWithEnvBytes loads YAML bytes with Env
func LoadWithEnvBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, defaultCustomizer.Customize, yaml.Unmarshal)
}

// LoadWithEnvJSONBytes loads JSON bytes with Env
func LoadWithEnvJSONBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, defaultCustomizer.Customize, json.Unmarshal)
}

// LoadWithEnvTOMLBytes loads TOML bytes with Env
func LoadWithEnvTOMLBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, defaultCustomizer.Customize, toml.Unmarshal)
}

// Marshal serializes the value provided into a YAML document.
var Marshal = yaml.Marshal

// MarshalJSON returns the JSON encoding of v with indent by 2 white spaces.
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func loadWithFunc(conf interface{}, configPaths []string, custom customFunc, unmarshal unmarshaler) error {
	for _, configPath := range configPaths {
		err := loadConfig(conf, configPath, custom, unmarshal)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadConfig(conf interface{}, configPath string, custom customFunc, unmarshal unmarshaler) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.Wrapf(err, "%s read failed", configPath)
	}
	if err := loadConfigBytes(conf, data, custom, unmarshal); err != nil {
		return errors.Wrapf(err, "%s load failed", configPath)
	}
	return nil
}

func loadConfigBytes(conf interface{}, data []byte, custom customFunc, unmarshal unmarshaler) error {
	var err error
	if custom != nil {
		data, err = custom(data)
		if err != nil {
			// Go 1.12 text/template catches a panic raised in user-defined function.
			// https://golang.org/doc/go1.12#text/template
			if strings.Index(err.Error(), "must_env: environment variable") != -1 {
				panic(err)
			}
			return errors.Wrap(err, "custom failed")
		}
	}
	if err := unmarshal(data, conf); err != nil {
		return errors.Wrap(err, "parse failed")
	}
	return nil
}

// Delims sets the action delimiters to the specified strings.
func Delims(left, right string) {
	defaultCustomizer.Delims(left, right)
}

type Replacer interface {
	Keyword() string
	Replace(...string) string
}

type customizer struct {
	mu          sync.Mutex
	left, right string
	replacers   []Replacer
	repTpl      *template.Template
}

func (c *customizer) clearRepTpl() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.repTpl = nil
}

func (c *customizer) getRepTpl() *template.Template {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.repTpl == nil {
		c.buildRepTpl()
	}
	return c.repTpl
}

func (c *customizer) buildRepTpl() {
	funcMap := template.FuncMap{
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
	}
	for _, replacer := range c.replacers {
		funcMap[replacer.Keyword()] = replacer.Replace
	}
	c.repTpl = template.New("conf").Funcs(funcMap)
	if c.left != "" || c.right != "" {
		c.repTpl.Delims(c.left, c.right)
	}
}

func (c *customizer) Replacers(replacers ...Replacer) {
	c.replacers = replacers
	c.clearRepTpl()
}

func (c *customizer) Delims(left, right string) {
	c.left = left
	c.right = right
	c.clearRepTpl()
}

func (c *customizer) Customize(data []byte) ([]byte, error) {
	t, err := c.getRepTpl().Parse(string(data))
	if err != nil {
		return nil, errors.Wrap(err, "config parse by template failed")
	}
	buf := &bytes.Buffer{}
	if err = t.Execute(buf, nil); err != nil {
		return nil, errors.Wrap(err, "template attach failed")
	}
	return buf.Bytes(), nil
}
