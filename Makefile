.PHONY: proto

build:
	@go build -o bin/ffaas cmd/ffaas/main.go 

wasmserver:
	@go build -o bin/wasmserver cmd/wasmserver/main.go
	@./bin/wasmserver

run: build
	@./bin/ffaas --seed

test:
	@go test ./pkg/* -v

proto:
	protoc --go_out=. --go_opt=paths=source_relative --proto_path=. proto/types.proto

clean:
	@rm -rf bin/ffaas

goex:
	GOOS=wasip1 GOARCH=wasm go build -o examples/go/app.wasm examples/go/main.go 