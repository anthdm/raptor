build:
	@go build -o bin/ffaas cmd/ffaas/main.go 

run: build
	@./bin/ffaas --seed

test:
	@go test ./pkg/* -v

clean:
	@rm -rf bin/ffaas

goex:
	GOOS=wasip1 GOARCH=wasm go build -o examples/go/app.wasm examples/go/main.go 