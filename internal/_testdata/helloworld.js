// This should come from the official SDK.
// But there is no official SDK yet, so we keep it here.
function respond(res, status) {
    var buffer = new ArrayBuffer(8);
    var view = new DataView(buffer);
    view.setUint32(0, status, true);
    view.setUint32(4, res.length, true);

    putstr(res);
    writebytes(view)
}

console.log("USER LOGS");
console.log("USER LOGS");
console.log("USER LOGS");
console.log("USER LOGS");
console.log("USER LOGS");

respond("Hello world!", 200)