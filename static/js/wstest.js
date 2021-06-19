let ws = new WebSocket('ws://127.0.0.1:8080/ws')
ws.onopen = function (evt) {
  console.log('OPEN')
}
ws.onclose = function (evt) {
  console.log('CLOSE')
  ws = null
}
ws.onmessage = function (evt) {
  console.log('RESPONSE: ' + evt.data)
}
ws.onerror = function (evt) {
  console.log('ERROR: ' + evt.data)
}

// setInterval(() => {
//   console.log('ping')
//   ws.send('pong')
// }, 1000)
