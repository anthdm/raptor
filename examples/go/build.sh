echo "building go application with tinygo..."
tinygo build -o examples/go/app.wasm --no-debug -target wasi examples/go/main.go
echo "done!"