// var myChart = echarts.init(document.getElementById('map'))
// 从服务器获取数据，测试用
// 创建map 城市或者国家（查询不到具体城市的时候作为键，值为一系列的ipstruct）
let coordPointData = []
let linesData = []
let locationMap = new HashMap()
let ipMap = new HashMap()
//起点应该都是固定的...但是考虑到获取到的可能是局域网ip，所以暂时通过手动设置经纬度来设置
let startPos = [114.2662, 30.5851]
let startName = 'Wuhan'
// 流量转为更加易读的方式
const simplifyBandwidthOutput = (bytes) => {}
const findAndDelInactive = () => {}
const getData = async () => {
  coordPointData = []
  linesData = []
  coordPointData = [{ name: startName, value: startPos }]
  linesData = []
  //如何动态修改这个url？
  let response = await fetch('http://localhost:8080/json')
  let jsonData
  if (response.ok) {
    // let jsonResult = response
    jsonData = await response.json()
  } else {
    console.log(response.status + '失败')
    jsonData = Promise.reject('failed')
  }
  // 坐标点 起点手动设置
  // 之后需要考虑时间戳，以及创建一个map，记录到同一个地点的多个连接（多个ip归属于一个地方）

  // 原始数据处理、分类、去除过期的数据
  for (let data of jsonData) {
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
    // 如果一个location所有的ip 连接数据都过期了，那么不显示
    let now = Date.now()
    // 10s内活跃连接
    let cmpTime = now - 60 * 1000
    let locationValid = false
    let locationLongitude, locationLatitude
    for (let ipPair of locationPair.value) {
      let activeTime = Date.parse(ipPair.value.value.lastactive)
      if (activeTime > cmpTime) {
        locationValid = true
        // 是否考虑建立城市-国家map？
        country = ipPair.value.value.country
        locationLongitude = ipPair.value.value.longitude
        locationLatitude = ipPair.value.value.latitude
        // 放入显示列表中
      }
    }
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
      // 出现过期城市
    }
  }
  // 排序一次能否减少dataIndex变化-但是无法彻底解决问题 （或者画线但是不显示？
  linesData.sort((m, n) => m.toName - n.toName)
}

const start = async () => {
  let chart = echarts.init(document.getElementById('map'))

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
      formatter: function (params, ticket, callback) {
        /*$.get('detail?name=' + params.name, function (content) {
              callback(ticket, toHTML(content));
          });*/
        var tips = '<ul style="list-style: none">'
        tips += '<li>行政区划：' + params.name + '</li>'
        tips += '<li>历史已报件：1000个</li>'
        tips += '<li>最近上报时间：' + new Date().toLocaleTimeString() + '</li>'
        tips += '</ul>'
        return tips
      },
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
        tooltip: {
          trigger: 'item',
          formatter: (params, ticket, callback) => {
            /*$.get('detail?name=' + params.name, function (content) {
              callback(ticket, toHTML(content));
          });*/
            // 展示一个地区的所有连接 以及流量 显示所有的ip，而非活跃ip
            let ipMap = locationMap.get(params.name)
            let tips = `<p>${params.name}</p>`
            // tips += '<ul style="list-style: none">'
            for (let pair of ipMap) {
              tips += `<p>ip:${pair.key} totalbytes:${pair.value.value.totalbytes} upload:${pair.value.value.outbytes} download:${pair.value.value.inbytes}</p>`
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
            let tip = `<p>${params.data.fromName}-${params.data.toName} ${params.dataIndex}</p>`
            // 获取同一地点下的所有(活跃)ip
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
  // 定时更新
  const instantRun = async () => {
    await getData()
    option.series[1].data = linesData
    option.series[0].data = coordPointData
    linesData.forEach((item, index) => {
      console.log(index + ':' + item.toName)
    })
    // console.log('update', option.series[0].data)
    chart.setOption(option)
  }
  instantRun()
  setInterval(async () => {
    await instantRun()
  }, 5000)
}

// 保证缩放的时候，散点图和线不会错位
window.onresize = () => {
  chart.resize()
}

start()
