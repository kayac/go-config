GIT_VER := $(shell git describe --tags)

cmd/merge-env-config/merge-env-config: *.go cmd/merge-env-config/*.go
	cd cmd/merge-env-config && go build -ldflags "-s -w -X main.Version=${GIT_VER}" -gcflags="-trimpath=${PWD}"

install: cmd/merge-env-config/merge-env-config
	install cmd/merge-env-config/merge-env-config ${GOPATH}/bin

packages:
	cd cmd/merge-env-config && gox -os="linux darwin" -arch="amd64" -output "../../pkg/{{.Dir}}-${GIT_VER}-{{.OS}}-{{.Arch}}" -ldflags "-w -s -X main.Version=${GIT_VER}"
	cd pkg && find . -name "*${GIT_VER}*" -type f -exec zip {}.zip {} \;

clean:
	rm -f cmd/merge-env-config/merge-env-config
	rm -f pkg/*
