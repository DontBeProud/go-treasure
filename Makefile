HOST_OS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
GIT_VERSION=$(shell git describe --tags --always)
ifeq ($(HOST_OS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\git.exe,bin\bash.exe,$(shell where git)))
	PROTO_FILES=$(shell $(Git_Bash) -c "find . -name '*.proto' ! -path './third_party/*'")
else
	PROTO_FILES=$(shell find -name *.proto)
endif

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go get -u github.com/gogo/protobuf/protoc-gen-gogofast/plugin/protoc-gen-go-errors
	go get -u google.golang.org/grpc
	go mod tidy

.PHONY: build-proto
build-proto:
	protoc --proto_path=. \
 	       --go_out=paths=source_relative:pb \
 	       --go-http_out=paths=source_relative:pb \
 	       --go-grpc_out=paths=source_relative:pb \
 	       --go-errors_out=paths=source_relative:pb \
	       $(PROTO_FILES)

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: pre-commit
pre-commit:
	make init
	make build-proto
	make lint
	go test ./...