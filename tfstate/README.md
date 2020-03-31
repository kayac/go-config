# tfstate add-on package

This package extends config.Loader and provides functions to reference Terform tfstate values

## Usage

tfstate pakcage include `MustLoad` and `Load`.  
for example:
config.yaml has the following content
```yaml
aws_account_id: {{ tfstate "data.aws_caller_identity.current.account_id" }}
remote_aws_account_id: {{ s3tfstate "data.aws_caller_identity.current.account_id" }}
```

The code to load this configuration is as follows:
```go
package main

import (
    "fmt"

	"github.com/kayac/go-config"
	"github.com/kayac/go-config/tfstate"
)

func main() {
	loader := config.New()
	loader.Funcs(tfstate.MustLoad("./testdata/terraform.tfstate"))
	loader.Funcs(tfstate.MustLoadWithName("s3tfstate", "s3://bucket.example.com/terraform.tfstate"))
	var c map[string]string
	if err := loader.LoadWithEnv(&c, "./config.yaml"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(c["aws_account_id"])
}
```

Load tfstate URL support s3, fiile schemes

## Note

./testdata/terraform.tfstate is the same data as testdata used in https://github.com/fujiwara/tfstate-lookup
