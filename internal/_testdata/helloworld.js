// This should come from the official SDK.
// But there is no official SDK yet, so we keep it here.
function hexLog(s) {
    for (let i = 0; i < s.length; i++) {
        putstr(s.charCodeAt(i).toString(16).padStart(2, "0"))
    }
    putstr("0a")
}

console.log = hexLog

function respond(res, status) {
    var buffer = new ArrayBuffer(8);
    var view = new DataView(buffer);
    view.setUint32(0, status, true);
    view.setUint32(4, res.length, true);

    for (let i = 0; i < res.length; i++) {
        putstr(res.charCodeAt(i).toString(16).padStart(2, "0"))
    }
    for (let i = 0; i < view.buffer.byteLength; i++) {
        putstr(view.getUint8(i).toString(16).padStart(2, "0"));
    }
}

respond("Hello world!", 200)