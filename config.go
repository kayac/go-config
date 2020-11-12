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
	defaultLoader = New()
}

type customFunc func(data []byte) ([]byte, error)

type unmarshaler func([]byte, interface{}) error

// Load loads YAML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func Load(conf interface{}, configPaths ...string) error {
	return defaultLoader.Load(conf, configPaths...)
}

// Load loads JSON files from `configPaths`.
// and assigns decoded values into the `conf` value.
func LoadJSON(conf interface{}, configPaths ...string) error {
	return defaultLoader.LoadJSON(conf, configPaths...)
}

// Load loads TOML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func LoadTOML(conf interface{}, configPaths ...string) error {
	return defaultLoader.LoadTOML(conf, configPaths...)
}

// LoadBytes loads YAML bytes
func LoadBytes(conf interface{}, src []byte) error {
	return defaultLoader.LoadBytes(conf, src)
}

// LoadJSONBytes loads JSON bytes
func LoadJSONBytes(conf interface{}, src []byte) error {
	return defaultLoader.LoadJSONBytes(conf, src)
}

// LoadTOMLBytes loads TOML bytes
func LoadTOMLBytes(conf interface{}, src []byte) error {
	return defaultLoader.LoadTOMLBytes(conf, src)
}

// LoadWithEnv loads YAML files with Env
// replace {{ env "ENV" }} to os.Getenv("ENV")
// if you set default value then {{ env "ENV" "default" }}
func LoadWithEnv(conf interface{}, configPaths ...string) error {
	return defaultLoader.LoadWithEnv(conf, configPaths...)
}

// LoadWithEnvJSON loads JSON files with Env
func LoadWithEnvJSON(conf interface{}, configPaths ...string) error {
	return defaultLoader.LoadWithEnvJSON(conf, configPaths...)
}

// LoadWithEnvTOML loads TOML files with Env
func LoadWithEnvTOML(conf interface{}, configPaths ...string) error {
	return defaultLoader.LoadWithEnvTOML(conf, configPaths...)
}

// LoadWithEnvBytes loads YAML bytes with Env
func LoadWithEnvBytes(conf interface{}, src []byte) error {
	return defaultLoader.LoadWithEnvBytes(conf, src)
}

// LoadWithEnvJSONBytes loads JSON bytes with Env
func LoadWithEnvJSONBytes(conf interface{}, src []byte) error {
	return defaultLoader.LoadWithEnvJSONBytes(conf, src)
}

// LoadWithEnvTOMLBytes loads TOML bytes with Env
func LoadWithEnvTOMLBytes(conf interface{}, src []byte) error {
	return defaultLoader.LoadWithEnvTOMLBytes(conf, src)
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
	defaultLoader.Delims(left, right)
}

// Funcs adds the elements of the argument map.
// Caution: global settings are overwritten. can't go back.
func Funcs(funcMap template.FuncMap) {
	defaultLoader.Funcs(funcMap)
}

var defaultLoader *Loader

// Loader represents config loader.
type Loader struct {
	Data interface{}

	leftDelim  string
	rightDelim string
	funcMap    template.FuncMap
}

// DefaultFuncMap defines built-in template functions.
var DefaultFuncMap = template.FuncMap{
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
	"json_escape": func(s string) string {
		b, _ := json.Marshal(s)        // marshal as JSON string
		return string(b[1 : len(b)-1]) // remove " on head and tail
	},
}

// New creates a Loader instance.
func New() *Loader {
	l := &Loader{
		funcMap: make(template.FuncMap, len(DefaultFuncMap)),
	}
	l.Funcs(DefaultFuncMap)
	return l
}

func (l *Loader) newTemplate() *template.Template {
	tmpl := template.New("conf").Funcs(l.funcMap)
	if l.leftDelim != "" && l.rightDelim != "" {
		tmpl.Delims(l.leftDelim, l.rightDelim)
	}
	return tmpl
}

func (l *Loader) replacer(data []byte) ([]byte, error) {
	t, err := l.newTemplate().Parse(string(data))
	if err != nil {
		return nil, errors.Wrap(err, "config parse by template failed")
	}
	buf := &bytes.Buffer{}
	if err = t.Execute(buf, l.Data); err != nil {
		return nil, errors.Wrap(err, "template attach failed")
	}
	return buf.Bytes(), nil
}

// Load loads YAML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func (l *Loader) Load(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, nil, yaml.Unmarshal)
}

// LoadJSON loads JSON files from `configPaths`.
// and assigns decoded values into the `conf` value.
func (l *Loader) LoadJSON(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, nil, json.Unmarshal)
}

// LoadTOML loads TOML files from `configPaths`.
// and assigns decoded values into the `conf` value.
func (l *Loader) LoadTOML(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, nil, toml.Unmarshal)
}

// LoadBytes loads YAML bytes
func (l *Loader) LoadBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, nil, yaml.Unmarshal)
}

// LoadJSONBytes loads JSON bytes
func (l *Loader) LoadJSONBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, nil, json.Unmarshal)
}

// LoadTOMLBytes loads TOML bytes
func (l *Loader) LoadTOMLBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, nil, toml.Unmarshal)
}

// LoadWithEnv loads YAML files with Env
// replace {{ env "ENV" }} to os.Getenv("ENV")
// if you set default value then {{ env "ENV" "default" }}
func (l *Loader) LoadWithEnv(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, l.replacer, yaml.Unmarshal)
}

// LoadWithEnvJSON loads JSON files with Env
func (l *Loader) LoadWithEnvJSON(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, l.replacer, json.Unmarshal)
}

// LoadWithEnvTOML loads TOML files with Env
func (l *Loader) LoadWithEnvTOML(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, configPaths, l.replacer, toml.Unmarshal)
}

// LoadWithEnvBytes loads YAML bytes with Env
func (l *Loader) LoadWithEnvBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, l.replacer, yaml.Unmarshal)
}

// LoadWithEnvJSONBytes loads JSON bytes with Env
func (l *Loader) LoadWithEnvJSONBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, l.replacer, json.Unmarshal)
}

// LoadWithEnvTOMLBytes loads TOML bytes with Env
func (l *Loader) LoadWithEnvTOMLBytes(conf interface{}, src []byte) error {
	return loadConfigBytes(conf, src, l.replacer, toml.Unmarshal)
}

// Delims sets the action delimiters to the specified strings.
func (l *Loader) Delims(left, right string) {
	l.leftDelim = left
	l.rightDelim = right
}

// Funcs adds the elements of the argument map.
func (l *Loader) Funcs(funcMap template.FuncMap) {
	for name, fn := range funcMap {
		l.funcMap[name] = fn
	}
}
