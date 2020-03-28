package tfstate_test

import (
	"fmt"

	"github.com/kayac/go-config"
	"github.com/kayac/go-config/tfstate"
)

func ExampleMustLoad() {
	loader := config.New()
	loader.Funcs(tfstate.MustLoad("file://./testdata/terraform.tfstate"))
	var c map[string]string
	if err := loader.LoadWithEnv(&c, "./testdata/config.yaml"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(c["aws_account_id"])
	fmt.Println(c["log_group"])
	// Output:
	//123456789012
	///main/app
}
