package tfstate

import (
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/fujiwara/tfstate-lookup/tfstate"
	"github.com/mashiike/urlio"
	"github.com/pkg/errors"
)

// Load tfstate based on URL and provide tamplate.FuncMap
func Load(stateURL string) (template.FuncMap, error) {
	u, err := url.Parse(stateURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tfstate: %s", stateURL)
	}
	reader, err := urlio.NewReader(u)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tfstate: %s", stateURL)
	}
	defer reader.Close()
	state, err := tfstate.Read(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tfstate: %s", stateURL)
	}
	return template.FuncMap{
		"tfstate": func(addrs string) string {
			if strings.Contains(addrs, "'") {
				addrs = strings.ReplaceAll(addrs, "'", "\"")
			}
			attrs, err := state.Lookup(addrs)
			if err != nil {
				return ""
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
	funcMap, err := Load(stateURL)
	if err != nil {
		panic(err)
	}
	return funcMap
}
