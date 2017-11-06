// Yaml Config Loader
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func init() {
	initEnvReplacer()
}

type customFunc func(data []byte) ([]byte, error)

// Yaml files load from `configPaths`.
// and assigns decoded values into the `conf` value.
func Load(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, nil, configPaths)
}

// Load with Env
// replace {{ env "ENV" }} to os.Getenv("ENV")
// if you set default value then {{ env "ENV" "default" }}
func LoadWithEnv(conf interface{}, configPaths ...string) error {
	return loadWithFunc(conf, envReplacer, configPaths)
}

func loadWithFunc(conf interface{}, custom customFunc, configPaths []string) error {
	for _, configPath := range configPaths {
		err := loadConfig(configPath, conf, custom)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadConfig(configPath string, conf interface{}, custom customFunc) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.Wrapf(err, "%s read failed", configPath)
	}
	if custom != nil {
		data, err = custom(data)
		if err != nil {
			return errors.Wrapf(err, "%s yaml custom failed", configPath)
		}
	}
	if err := yaml.Unmarshal(data, conf); err != nil {
		return errors.Wrapf(err, "%s yaml parse failed", configPath)
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
			if v := os.Getenv(key); v != "" {
				return v
			}
			for _, env := range os.Environ() {
				if strings.HasPrefix(env, key+"=") {
					// defined but empty string
					return ""
				}
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
