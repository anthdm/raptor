function generateEndHex(length, status) {
  var buffer = new ArrayBuffer(8);
  var view = new DataView(buffer);

  view.setUint32(0, status, true);
  view.setUint32(4, length, true);

  var hexString = "";
  for (var i = 0; i < buffer.byteLength; i++) {
    var hex = view.getUint8(i).toString(16);
    hexString += hex.padStart(2, "0");
  }

  return hexString;
}

// This should come from the official SDK. 
// But there is no official SDK yet, so we keep it here.
function respond(res, status) {
	putstr(res+generateEndHex(res.length, status))
}

respond("<h1>From my Raptor application</h1></br>some other stuff here</br>", 200)