package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kayac/go-config"
)

var dir string

type DBConfig struct {
	Master  string        `yaml:"master"`
	Slave   string        `yaml:"slave"`
	Timeout time.Duration `yml:"timeout"`
}
type Conf struct {
	Domain  string        `yaml:"domain"`
	IsDev   bool          `yaml:"is_dev"`
	Timeout time.Duration `yml:"timeout"`
	DB      DBConfig      `yaml:"db"`
}

func TestMain(m *testing.M) {
	runner := func() int {
		var err error
		dir, err = ioutil.TempDir("", "go-config")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(dir)
		return m.Run()
	}
	os.Exit(runner())
}

func ExampleLoad() {
	type DBConfig struct {
		Master  string        `yaml:"master"`
		Slave   string        `yaml:"slave"`
		Timeout time.Duration `yml:"timeout"`
	}
	type Conf struct {
		Domain  string        `yaml:"domain"`
		IsDev   bool          `yaml:"is_dev"`
		Timeout time.Duration `yml:"timeout"`
		DB      DBConfig      `yaml:"db"`
	}

	baseConfig := `
# config.yml
domain: example.com
db:
  master:  rw@/example
  slave:   ro@/example
  timeout: 0.5s
`
	localConfig := `
# config_local.yml
domain: dev.example.com
is_dev: true
`
	conf := &Conf{}
	baseConfigYaml, _ := genConfigFile("config.yml", baseConfig)         // /path/to/config.yml
	localConfigYaml, _ := genConfigFile("config_local.yml", localConfig) // /path/to/config_local.yml

	err := config.Load(conf, baseConfigYaml, localConfigYaml)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", conf)
	// Output:
	// &{Domain:dev.example.com IsDev:true Timeout:0s DB:{Master:rw@/example Slave:ro@/example Timeout:500ms}}
}

func ExampleLoadWithEnv() {
	type DBConfig struct {
		Master  string        `yaml:"master"`
		Slave   string        `yaml:"slave"`
		Timeout time.Duration `yml:"timeout"`
	}
	type Conf struct {
		Domain  string        `yaml:"domain"`
		IsDev   bool          `yaml:"is_dev"`
		Timeout time.Duration `yml:"timeout"`
		DB      DBConfig      `yaml:"db"`
	}

	baseConfig := `
# config.yml
domain: {{ env "DOMAIN"}}
db:
  master:  rw@/example
  slave:   ro@/example
  timeout: 0.5s
`
	localConfig := `
# config_local.yml
is_dev: true
db:
  master:  {{ env "RW_USER" "rw" }}@/{{ env "DB_NAME" "example_dev" }}
  slave:   {{ env "RO_USER" "ro" }}@/{{ env "DB_NAME" "example_dev" }}
`
	os.Setenv("DOMAIN", "dev.example.com")
	os.Setenv("DB_NAME", "example_local")

	conf := &Conf{}
	baseConfigYaml, _ := genConfigFile("config.yml", baseConfig)         // /path/to/config.yml
	localConfigYaml, _ := genConfigFile("config_local.yml", localConfig) // /path/to/config_local.yml

	err := config.LoadWithEnv(conf, baseConfigYaml, localConfigYaml)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", conf)
	// Output:
	// &{Domain:dev.example.com IsDev:true Timeout:0s DB:{Master:rw@/example_local Slave:ro@/example_local Timeout:500ms}}
}

func TestLoad(t *testing.T) {
	a, err := genConfigFile("a.yml", `## a.yml
domain: example.com
db:
  master:  rw@/example
  slave:   ro@/example
  timeout: 0.5s
`)
	if err != nil {
		t.Error(err)
		return
	}
	b, err := genConfigFile("b.yml", `## b.yml
is_dev: true
`)
	if err != nil {
		t.Error(err)
		return
	}
	c, err := genConfigFile("c.yml", `## c.yml
db:
  master:  rw@/example2
  slave:   ro@/example2
  timeout: 200ms
`)
	if err != nil {
		t.Error(err)
		return
	}
	er, err := genConfigFile("err.yml", `## err.yml
db:
  master  rw@/example2
`)
	if err != nil {
		t.Error(err)
		return
	}

	e := &Conf{
		Domain:  "example.com",
		IsDev:   false,
		Timeout: time.Duration(0),
		DB: DBConfig{
			Master:  "rw@/example",
			Slave:   "ro@/example",
			Timeout: time.Duration(500) * time.Millisecond,
		},
	}
	t1 := &Conf{}
	err = config.Load(t1, a)
	if err != nil {
		t.Errorf("a.yml load error %s", err)
	}
	if !reflect.DeepEqual(t1, e) {
		t.Errorf("a.yml not match. got: %#v, expect: %#v", t1, e)
	}

	t2 := &Conf{}
	e.IsDev = true
	err = config.Load(t2, a, b)
	if err != nil {
		t.Errorf("a.yml, b.yml load error %s", err)
	}
	if !reflect.DeepEqual(t2, e) {
		t.Errorf("a.yml, b.yml not match. got: %#v, expect: %#v", t2, e)
	}

	t3 := &Conf{}
	e.DB.Master = "rw@/example2"
	e.DB.Slave = "ro@/example2"
	e.DB.Timeout = time.Duration(200) * time.Millisecond
	err = config.Load(t3, a, b, c)
	if err != nil {
		t.Errorf("a.yml, b.yml, c.yml load error %s", err)
	}
	if !reflect.DeepEqual(t3, e) {
		t.Errorf("a.yml, b.yml, c.yml not match. got: %#v, expect: %#v", t3, e)
	}

	t4 := &Conf{}
	err = config.Load(t4, filepath.Join(dir, "nothing.yml"))
	if err == nil {
		t.Errorf("nothing.yml is not found.")
	}
	t.Log(err)

	t5 := &Conf{}
	err = config.Load(t5, er)
	if err == nil {
		t.Errorf("err.yml is format err.")
	}
	t.Log(err)
}

func genConfigFile(name string, config string) (string, error) {
	path := filepath.Join(dir, name)
	io, err := os.Create(path)
	if err != nil {
		return "", err
	}
	if _, err := io.WriteString(config); err != nil {
		return "", err
	}
	return path, nil
}

func TestLoadMustEnvPanic(t *testing.T) {
	f, err := genConfigFile("must-panic.yml", `## must.yml
domain: '{{ must_env "MUST_DOMAIN_PANIC" }}'
`)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("must_env must raise panic")
		} else {
			t.Logf("must_env raise panic:%s", r)
		}
	}()

	c := &Conf{}
	config.LoadWithEnv(c, f)
}

func TestLoadMustEnv(t *testing.T) {
	f, err := genConfigFile("must.yml", `## must.yml
domain: '{{ must_env "MUST_DOMAIN" }}'
`)
	if err != nil {
		t.Error(err)
		return
	}
	mustDomain := "must.example.com"
	os.Setenv("MUST_DOMAIN", mustDomain)
	c := &Conf{}
	if err := config.LoadWithEnv(c, f); err != nil {
		t.Error(err)
	}
	if c.Domain != mustDomain {
		t.Errorf("domain expected %s got %s", mustDomain, c.Domain)
	}

	os.Setenv("MUST_DOMAIN", "")
	c2 := &Conf{}
	if err := config.LoadWithEnv(c2, f); err != nil {
		t.Error(err)
	}
	if c2.Domain != "" {
		t.Errorf("domain expected \"\" got %s", c2.Domain)
	}
}

func TestLoadJSON(t *testing.T) {
	a, err := genConfigFile("a.json", `{
  "foo": "bar",
  "env_foo": "{{ env "FOO" }}"
}`)
	if err != nil {
		t.Error(err)
	}
	b, err := genConfigFile("b.json", `{
  "bar": "baz"
}`)
	if err != nil {
		t.Error(err)
	}
	os.Setenv("FOO", "BOO")
	c := make(map[string]string)
	err = config.LoadWithEnvJSON(&c, a, b)
	if err != nil {
		t.Error(err)
	}
	if c["foo"] != "bar" {
		t.Errorf("foo expected bar got %s", c["foo"])
	}
	if c["env_foo"] != "BOO" {
		t.Errorf("env_foo expected BOO got %s", c["env_foo"])
	}
	if c["bar"] != "baz" {
		t.Errorf("bar expected baz got %s", c["bar"])
	}
}

func TestLoadTOML(t *testing.T) {
	a, err := genConfigFile("a.toml", `
foo = "bar"
env_foo = '{{ env "FOO" }}'
`)
	if err != nil {
		t.Error(err)
	}
	b, err := genConfigFile("b.toml", `
bar = "baz"
`)
	if err != nil {
		t.Error(err)
	}
	os.Setenv("FOO", "BOO")
	c := make(map[string]string)
	err = config.LoadWithEnvTOML(&c, a, b)
	if err != nil {
		t.Error(err)
	}
	if c["foo"] != "bar" {
		t.Errorf("foo expected bar got %s", c["foo"])
	}
	if c["env_foo"] != "BOO" {
		t.Errorf("env_foo expected BOO got %s", c["env_foo"])
	}
	if c["bar"] != "baz" {
		t.Errorf("bar expected baz got %s", c["bar"])
	}
}
