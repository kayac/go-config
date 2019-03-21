package main

import (
	"flag"
	"fmt"
	"os"

	config "github.com/kayac/go-config"
)

var Version = "current"

type Loader func(interface{}, ...string) error

type Marshaler func(interface{}) ([]byte, error)

func main() {
	os.Exit(_main())
}

func _main() int {
	var isJSON, showVersion bool

	flag.BoolVar(&isJSON, "json", false, "file(s) is JSON")
	flag.BoolVar(&showVersion, "v", false, "show version number")
	flag.BoolVar(&showVersion, "version", false, "show version number")
	flag.Parse()

	if showVersion {
		fmt.Println("merge-env-config ", Version)
		return 0
	}

	args := flag.Args()

	if len(args) == 0 {
		printUsage()
		return 1
	}

	var (
		load    Loader
		marshal Marshaler
		conf    map[string]interface{}
	)

	if isJSON {
		load = config.LoadWithEnvJSON
		marshal = config.MarshalJSON
	} else {
		load = config.LoadWithEnv
		marshal = config.Marshal
	}

	err := load(&conf, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	b, err := marshal(&conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Stdout.Write(b)
	return 0
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage of merge-env-config:

  merge-env-config [-json] config1.yaml [config2.yaml ...]`)
	flag.PrintDefaults()
}
