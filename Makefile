.PHONY: proto

PROTO_PATH := "../../go/pkg/mod/github.com/anthdm/hollywood@v0.0.0-20231230110106-87b55a8811e9/actor"

build:
	@go build -o bin/api cmd/api/main.go 
	@go build -o bin/ingress cmd/ingress/main.go 
	@go build -o bin/raptor cmd/cli/main.go 
	@go build -o bin/runtime cmd/runtime/main.go 

ingress: build
	@./bin/ingress

runtime: build
	@./bin/runtime

api: build
	@./bin/api --seed

test:
	@./internal/_testdata/build.sh
	@go test -v ./internal/*

proto:
	protoc --go_out=. --go_opt=paths=source_relative --proto_path=$(PROTO_PATH) --proto_path=. proto/types.proto

clean:
	@rm -rf bin/api
	@rm -rf bin/wasmserver
	@rm -rf bin/raptor

goex:
	GOOS=wasip1 GOARCH=wasm go build -o examples/go/app.wasm examples/go/main.go 

jsex:
	javy compile examples/js/index.js -o examples/js/index.wasm

postgres:
	docker run --name raptordb -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres