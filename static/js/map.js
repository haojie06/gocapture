// 从服务器获取数据，测试用
// 创建map 城市或者国家（查询不到具体城市的时候作为键，值为一系列的ipstruct）
let chart = echarts.init(document.getElementById('map'))
let coordPointData = []
let linesData = []
let locationMap = new HashMap()
let ipMap = new HashMap()
//起点应该都是固定的...但是考虑到获取到的可能是局域网ip，所以暂时通过手动设置经纬度来设置
// 流量转为更加易读的方式
const simplifyBandwidthOutput = (bytes) => {
  let formatBandwidth
  if (bytes < 1024) {
    formatBandwidth = bytes + 'B'
  } else if (bytes < 1048576) {
    formatBandwidth = (bytes / 1024).toFixed(2) + 'KB'
  } else if (bytes < 1073741824) {
    formatBandwidth = (bytes / 1048576).toFixed(2) + 'MB'
  } else if (bytes < Math.pow(1024, 4)) {
    formatBandwidth = (bytes / Math.pow(1024, 3)).toFixed(2) + 'GB'
  } else {
    formatBandwidth = (bytes / Math.pow(1024, 4)).toFixed(2) + 'TB'
  }
  return formatBandwidth
}
const judgeIfActive = (timeStamp, difTime) => {
  let now = Date.now()
  // 10s内活跃连接
  let cmpTime = now - difTime
  let lastActiveTime = Date.parse(timeStamp)
  return lastActiveTime > cmpTime
}

// 获取数据并处理 改为websocket方式
const getData = () => {
  let ws
  //如何动态修改这个url？
  if (ws == null) {
    ws = new WebSocket(wsUrl)
  }

  ws.onopen = function (evt) {
    console.log('WS OPEN')
  }
  ws.onclose = function (evt) {
    console.log('WS CLOSE')
    ws = null
  }
  ws.onmessage = function (evt) {
    // 清空之前的
    coordPointData = [{ name: startName, value: startPos }]
    linesData = []
    // 对元素进行更新
    console.log('WS RECV DATA')
    let recvData = JSON.parse(evt.data)
    let textArea = document.getElementById('stats')
    textArea.innerHTML = recvData.bandwidthstr

    for (let data of recvData.bandwidthlist) {
      ipMap.set(data.value.ip, data)
      // 私网ip不绘图 （但是之后统计需要使用）
      if (data.value.longitude == 0 && data.value.latitude == 0) {
        continue
      }
      if (data.value.city == '') {
        //locationMap里面存的是同一个地点的一系列的ip:data map
        let locationIPMap = locationMap.get(data.value.country)
        if (locationIPMap === undefined) {
          locationIPMap = new HashMap()
        }
        locationIPMap.set(data.key, data)
        locationMap.set(data.value.country, locationIPMap)
      } else {
        let locationIPMap = locationMap.get(data.value.city)
        if (locationIPMap === undefined) {
          locationIPMap = new HashMap()
        }
        locationIPMap.set(data.key, data)
        locationMap.set(data.value.city, locationIPMap)
      }
    }

    // 处理需要显示的数据 排除过期的
    for (let locationPair of locationMap) {
      let location = locationPair.key
      let country
      // 如果一个location所有的ip 连接数据都过期了，那么不显示 （只要有一个没过期的就可以显示
      // 10s内活跃连接
      let locationValid = false
      let locationLongitude, locationLatitude
      for (let ipPair of locationPair.value) {
        if (judgeIfActive(ipPair.value.value.lastactive, 60 * 1000)) {
          locationValid = true
          // 是否考虑建立城市-国家map？
          country = ipPair.value.value.country
          locationLongitude = ipPair.value.value.longitude
          locationLatitude = ipPair.value.value.latitude
          // 放入显示列表中
        } else {
          // console.log('发现过期')
        }
      }
      // 只有当前地点至少有一条一分钟内活跃过的连接，才通过图形进行显示
      if (locationValid) {
        coordPointData.push({
          name: location,
          value: [locationLongitude, locationLatitude],
        })
        // GeoLite中的经纬度精度高了，导致同一城市可能出现不同的经纬度，造成重新画线
        linesData.push({
          fromName: startName,
          toName: location,
          coords: [startPos, [locationLongitude, locationLatitude]],
        })
      } else {
        // 出现过期城市 清除
      }
    }

    // 排序一次能否减少dataIndex变化-但是无法彻底解决问题 （或者画线但是不显示？
    linesData.sort((m, n) => m.toName - n.toName)

    let option = chart.getOption()
    option.series[0].data = coordPointData
    option.series[1].data = linesData
    chart.setOption(option)
  }

  ws.onerror = function (evt) {
    console.log('ERROR: ' + evt.data)
  }
  // let response = await fetch(fetchUrl + '/json/')
  // let jsonData
  // if (response.ok) {
  //   jsonData = await response.json()
  // } else {
  //   console.log(response.status + '失败')
  //   jsonData = Promise.reject('failed')
  // }
  // 坐标点 起点手动设置
  // 之后需要考虑时间戳，以及创建一个map，记录到同一个地点的多个连接（多个ip归属于一个地方）
  // 原始数据处理、分类、去除过期的数据
}
// 绘图
const draw = () => {
  let option = {
    // backgroundColor: '#1b1b1b',
    // color: ['gold', 'aqua', 'lime'],
    title: {
      show: true,
      text: '实时流量可视化 60s内活跃连接',
      x: 'center',
      // textStyle: {
      //   color: '#fff',
      // },
    },
    // 国家的tooltip
    tooltip: {
      trigger: 'item',
      formatter: function (params, ticket, callback) {},
    },
    toolbox: {
      show: true,
      orient: 'vertical',
      x: 'right',
      y: 'center',
      feature: {
        mark: { show: true },
        dataView: { show: true, readOnly: false },
        restore: { show: true },
        saveAsImage: { show: true },
        dataZoom: { show: true },
      },
    },
    geo: {
      roam: true,
      zoom: 1.25,
      map: 'world',
      hoverable: false,
      silent: true,
    },
    series: [
      // {
      //   type: 'map',
      //   map: 'world',
      //   data: [],
      //   markLine: {
      //     smooth: true,
      //     symbol: ['none', 'circle'],
      //     symbolSize: 1,
      //     data: [],
      //   },
      //   geoCoord: coordinatePoint,
      //   nameMap: nameMap,
      // },
      {
        type: 'scatter', //  指明图表类型：带涟漪效果的散点图
        zlevel: 5,
        coordinateSystem: 'geo', //  指明绘制在geo坐标系上
        animationDelay: 500,
        symbolSize: 7,
        data: [],
        // markPoint: {
        //   zlevel: 6,
        //   symbolSize: 20,
        //   itemStyle: {
        //     color: 'azure',
        //   },
        //   data: [
        //     {
        //       name: startName,
        //       coord: startPos,
        //     },
        //   ],
        // },
        tooltip: {
          trigger: 'item',
          position: 'inside',
          formatter: (params, ticket, callback) => {
            // 展示一个地区的所有连接 以及流量 1min内活跃
            // 获取一个地区所有的ip-data map
            let lIPMap = locationMap.get(params.name)
            let tips = `<p>${params.name}</p>`
            // tips += '<ul style="list-style: none">'
            for (let pair of lIPMap) {
              if ((judgeIfActive(pair.value.value.lastactive), 60 * 1000)) {
                tips += `<p>ip:${pair.key} totalbytes:${simplifyBandwidthOutput(
                  pair.value.value.totalbytes
                )} upload:${simplifyBandwidthOutput(
                  pair.value.value.outbytes
                )} download:${simplifyBandwidthOutput(
                  pair.value.value.inbytes
                )}</p>`
              }
            }
            // tips += '</ul>'
            return tips
          },
        },
      },
      {
        type: 'lines',
        coordinateSystem: 'geo',
        zlevel: 2,
        symbolSize: 110,
        // animationDurationUpdate: ,
        effect: {
          // 模拟效果路线特效
          show: true,
          period: 6,
          trailLength: 0,
          symbolSize: 3,
        },
        tooltip: {
          formatter: (params, ticket, callback) => {
            // console.log(params)
            let tip = `<p>${params.data.fromName}->${params.data.toName} ${params.dataIndex}</p>`
            let ipMap = locationMap.get(params.data.toName)
            let totalIn = 0,
              totalOut = 0,
              totalSum = 0

            for (let pair of ipMap) {
              // 全时间段的流量
              totalIn += Number(pair.value.value.inbytes)
              totalOut += Number(pair.value.value.outbytes)
            }
            totalSum = totalIn + totalOut
            totalIn = simplifyBandwidthOutput(totalIn)
            totalOut = simplifyBandwidthOutput(totalOut)
            totalSum = simplifyBandwidthOutput(totalSum)
            tip += `<p>upload:${totalOut} download:${totalIn} total:${totalSum}</p>`
            // 连线显示到该地的流量统计(1分钟内的活跃ip 以及总共统计)
            return tip
          },
        },

        lineStyle: {
          normal: {
            color: '#3f73a8',
            width: 1,
            opacity: 0.6,
            curveness: 0.2,
          },
          emphasis: {
            width: 2,
            color: 'red',
          },
        },
        // label: label,   // 将上面let的label注入
        data: [],
      },
    ],
  }
  chart.setOption(option)
}

// 保证缩放的时候，散点图和线不会错位
window.onresize = () => {
  chart.setOption(chart.getOption())
  chart.resize()
}

// 最初绘图
draw()
getData()
// 运行（后期数据更新
// setInterval(async () => {
//   try {
//     await updateMap()
//   } catch (err) {
//     console.log('error')
//   }
// }, 5000)
