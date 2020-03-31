package tfstate

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/fujiwara/tfstate-lookup/tfstate"
	"github.com/pkg/errors"
)

const (
	defaultFuncName = "tfstate"
)

// Load tfstate based on URL and provide tamplate.FuncMap
func Load(stateFile string) (template.FuncMap, error) {
	return LoadWithName(defaultFuncName, stateFile)
}

// LoadWithName provides tamplate.FuncMap. can lockup values from tfstate.
func LoadWithName(name string, stateFile string) (template.FuncMap, error) {
	state, err := tfstate.ReadFile(stateFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tfstate: %s", stateFile)
	}
	return template.FuncMap{
		name: func(addrs string) string {
			if strings.Contains(addrs, "'") {
				addrs = strings.ReplaceAll(addrs, "'", "\"")
			}
			attrs, err := state.Lookup(addrs)
			if err != nil {
				panic(fmt.Sprintf("failed to lookup %s in tfstate: %s", addrs, err))
			}
			if attrs.Value == nil {
				panic(fmt.Sprintf("%s is not found in tfstate", addrs))
			}
			return attrs.String()
		},
	}, nil
}

// MustLoad is similar to Load, but panics if it cannot get and parse tfstate. Simplifies registration with config.Loader
func MustLoad(stateURL string) template.FuncMap {
	return MustLoadWithName(defaultFuncName, stateURL)
}

// MustLoadWithName is similar to LoadWithName, but panics if it cannot get and parse tfstate. Simplifies registration with config.Loader
func MustLoadWithName(name string, stateURL string) template.FuncMap {
	funcMap, err := LoadWithName(name, stateURL)
	if err != nil {
		panic(err)
	}
	return funcMap
}
