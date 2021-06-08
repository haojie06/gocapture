const getStr = async () => {
  console.log('update')
  let response = await fetch(fetchUrl + '/str')
  let strData
  if (response.ok) {
    // let jsonResult = response
    strData = await response.text()
  } else {
    console.log(response.status + '失败')
    strData = Promise.reject('failed')
  }
  let textArea = document.getElementById('stats')
  textArea.innerHTML = strData
}
getStr()
setInterval(async () => {
  await getStr()
}, 1500)
