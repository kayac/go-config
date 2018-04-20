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
        "secret_key": "{{ env `SOME_SECRET_KEY` }}"
    }
}
```

```
$ SOME_SECRET_KEY=some_secret_key merge-env-config -json function.prod.json.tmpl
{
    "name": "some function name",
    "description": "some description",
    "environment": {
        "secret_key": "some_secret_key"
    }
}
```

## Author

Copyright (c) 2017 KAYAC Inc.

## LICENSE

MIT
