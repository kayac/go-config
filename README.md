# go-config

go-config is a configuration loader package.

See [godoc.org/github.com/kayac/go-config](https://godoc.org/github.com/kayac/go-config).

## merge-env-config

merge-env-config is the cli tool to deal with template files.

for example:
```
{
    "name": "some function name",
    "description": "some description",
    "environment": {
        "account_id": "{{ must_env `SOME_MUST_ACCOUNT_ID` }}",
        "secret_key": "{{ env `SOME_SECRET_KEY` }}"
    }
}
```

```
$ SOME_MUST_ACCOUNT_ID=must_account_id SOME_SECRET_KEY=some_secret_key merge-env-config -json function.prod.json.tmpl
{
    "name": "some function name",
    "description": "some description",
    "environment": {
        "account_id": "must_account_id",
        "secret_key": "some_secret_key"
    }
}
```

## Author

Copyright (c) 2017 KAYAC Inc.

## LICENSE

MIT
