// This should come from the official SDK. 
// But there is no official SDK yet, so we keep it here.
function respond(res, status) {
  var buffer = new ArrayBuffer(8);
  var view = new DataView(buffer);

  view.setUint32(0, status, true);
  view.setUint32(4, res.length, true);

  var bytes = new Uint8Array(buffer);
  print(res)
  print(bytes)
}

console.log("user log here")
console.log("user log here")
console.log("user log here")
console.log("user log here")
console.log("user log here")

respond("<h1>From my Raptor application</h1></br>some other stuff here</br>", 200)