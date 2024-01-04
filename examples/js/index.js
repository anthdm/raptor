function respond(res, status) {
	console.log(res+"|"+status)
}

function foo() {
	return "hello from my application"
}

const result = foo()

console.log("foo")
console.log("bar")
console.log("baz")

respond(result, 200)

