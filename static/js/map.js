// var myChart = echarts.init(document.getElementById('map'))
// 从服务器获取数据，测试用
// 创建map 城市或者国家（查询不到具体城市的时候作为键，值为一系列的ipstruct）
let coordPointData = []
let linesData = []
let locationMap = new HashMap()
let ipMap = new HashMap()
//起点应该都是固定的...但是考虑到获取到的可能是局域网ip，所以暂时通过手动设置经纬度来设置
let startPos = [114.2662, 30.5851]
let startName = 'China Wuhan'

const findAndDelInactive = () => {}
const getData = async () => {
  // 考虑不要清空而是删掉超时的？
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
      locationIPMap.set(data.value.ip, data)
      locationMap.set(data.value.country, locationIPMap)
    } else {
      let locationIPMap = locationMap.get(data.value.city)
      if (locationIPMap === undefined) {
        locationIPMap = new HashMap()
      }
      locationIPMap.set(data.value.ip, data)
      locationMap.set(data.value.city, locationIPMap)
    }
  }

  // 处理需要显示的数据 排除过期的
  for (let locationPair of locationMap) {
    let location = locationPair.key
    // 如果一个location所有的ip 连接数据都过期了，那么不显示
    let now = Date.now()
    // 10s内活跃连接
    let cmpTime = now - 10 * 1000
    let locationValid = false
    let locationLongitude, locationLatitude
    for (let ipPair of locationPair.value) {
      let activeTime = Date.parse(ipPair.value.value.lastactive)
      if (activeTime > cmpTime) {
        locationValid = true
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
      linesData.push({
        fromName: startName,
        toName: location,
        coords: [startPos, [locationLongitude, locationLatitude]],
      })
    }
    // 需要排除掉过期的点以及经纬度无法查询到的点 (还没有做完)
    // coordinatePoint.push({
    //   name: startName,
    //   value: startPos,
    // })
    // if (
    //   activeTime > cmpTime &&
    //   (data.value.longitude !== 0 || data.value.latitude !== 0)
    // ) {
    // coordinatePoint.push({
    //   name: data.value.city,
    //   value: [data.value.longitude, data.value.latitude],
    // })
    // linesData.push({
    //   fromName: startName,
    //   toName: data.value.country + ' ' + data.value.city,
    //   coords: [startPos, [data.value.longitude, data.value.latitude]],
    // })
    // }
  }
}

const start = async () => {
  let chart = echarts.init(document.getElementById('map'))
  let nameMap = {
    Afghanistan: '阿富汗',
    Albania: '阿尔巴尼亚',
    Algeria: '阿尔及利亚',
    Angola: '安哥拉',
    Argentina: '阿根廷',
    Armenia: '亚美尼亚',
    Australia: '澳大利亚',
    Austria: '奥地利',
    Azerbaijan: '阿塞拜疆',
    Bahamas: '巴哈马',
    Bahrain: '巴林',
    Bangladesh: '孟加拉国',
    Belarus: '白俄罗斯',
    Belgium: '比利时',
    Belize: '伯利兹',
    Benin: '贝宁',
    Bhutan: '不丹',
    Bolivia: '玻利维亚',
    'Bosnia and Herz.': '波斯尼亚和黑塞哥维那',
    Botswana: '博茨瓦纳',
    Brazil: '巴西',
    'British Virgin Islands': '英属维京群岛',
    Brunei: '文莱',
    Bulgaria: '保加利亚',
    'Burkina Faso': '布基纳法索',
    Burundi: '布隆迪',
    Cambodia: '柬埔寨',
    Cameroon: '喀麦隆',
    Canada: '加拿大',
    'Cape Verde': '佛得角',
    'Cayman Islands': '开曼群岛',
    'Central African Rep.': '中非共和国',
    Chad: '乍得',
    Chile: '智利',
    China: '中国',
    Colombia: '哥伦比亚',
    Comoros: '科摩罗',
    Congo: '刚果',
    'Costa Rica': '哥斯达黎加',
    Croatia: '克罗地亚',
    Cuba: '古巴',
    Cyprus: '塞浦路斯',
    'Czech Rep.': '捷克共和国',
    "Côte d'Ivoire": '科特迪瓦',
    'Dem. Rep. Congo': '刚果民主共和国',
    'Dem. Rep. Korea': '朝鲜',
    Denmark: '丹麦',
    Djibouti: '吉布提',
    'Dominican Rep.': '多米尼加共和国',
    Ecuador: '厄瓜多尔',
    Egypt: '埃及',
    'El Salvador': '萨尔瓦多',
    'Equatorial Guinea': '赤道几内亚',
    Eritrea: '厄立特里亚',
    Estonia: '爱沙尼亚',
    Ethiopia: '埃塞俄比亚',
    'Falkland Is.': '福克兰群岛',
    Fiji: '斐济',
    Finland: '芬兰',
    'Fr. S. Antarctic Lands': '所罗门群岛',
    France: '法国',
    Gabon: '加蓬',
    Gambia: '冈比亚',
    Georgia: '格鲁吉亚',
    Germany: '德国',
    Ghana: '加纳',
    Greece: '希腊',
    Greenland: '格陵兰',
    Guatemala: '危地马拉',
    Guinea: '几内亚',
    'Guinea-Bissau': '几内亚比绍',
    Guyana: '圭亚那',
    Haiti: '海地',
    Honduras: '洪都拉斯',
    Hungary: '匈牙利',
    Iceland: '冰岛',
    India: '印度',
    Indonesia: '印度尼西亚',
    Iran: '伊朗',
    Iraq: '伊拉克',
    Ireland: '爱尔兰',
    'Isle of Man': '马恩岛',
    Israel: '以色列',
    Italy: '意大利',
    Jamaica: '牙买加',
    Japan: '日本',
    Jordan: '约旦',
    Kazakhstan: '哈萨克斯坦',
    Kenya: '肯尼亚',
    Korea: '韩国',
    Kuwait: '科威特',
    Kyrgyzstan: '吉尔吉斯斯坦',
    'Lao PDR': '老挝',
    Latvia: '拉脱维亚',
    Lebanon: '黎巴嫩',
    Lesotho: '莱索托',
    Liberia: '利比里亚',
    Libya: '利比亚',
    Lithuania: '立陶宛',
    Luxembourg: '卢森堡',
    Macedonia: '马其顿',
    Madagascar: '马达加斯加',
    Malawi: '马拉维',
    Malaysia: '马来西亚',
    Maldives: '马尔代夫',
    Mali: '马里',
    Malta: '马耳他',
    Mauritania: '毛利塔尼亚',
    Mauritius: '毛里求斯',
    Mexico: '墨西哥',
    Moldova: '摩尔多瓦',
    Monaco: '摩纳哥',
    Mongolia: '蒙古',
    Montenegro: '黑山共和国',
    Morocco: '摩洛哥',
    Mozambique: '莫桑比克',
    Myanmar: '缅甸',
    Namibia: '纳米比亚',
    Nepal: '尼泊尔',
    Netherlands: '荷兰',
    'New Caledonia': '新喀里多尼亚',
    'New Zealand': '新西兰',
    Nicaragua: '尼加拉瓜',
    Niger: '尼日尔',
    Nigeria: '尼日利亚',
    Norway: '挪威',
    Oman: '阿曼',
    Pakistan: '巴基斯坦',
    Panama: '巴拿马',
    'Papua New Guinea': '巴布亚新几内亚',
    Paraguay: '巴拉圭',
    Peru: '秘鲁',
    Philippines: '菲律宾',
    Poland: '波兰',
    Portugal: '葡萄牙',
    'Puerto Rico': '波多黎各',
    Qatar: '卡塔尔',
    Reunion: '留尼旺',
    Romania: '罗马尼亚',
    Russia: '俄罗斯',
    Rwanda: '卢旺达',
    'S. Geo. and S. Sandw. Is.': '南乔治亚和南桑威奇群岛',
    'S. Sudan': '南苏丹',
    'San Marino': '圣马力诺',
    'Saudi Arabia': '沙特阿拉伯',
    Senegal: '塞内加尔',
    Serbia: '塞尔维亚',
    'Sierra Leone': '塞拉利昂',
    Singapore: '新加坡',
    Slovakia: '斯洛伐克',
    Slovenia: '斯洛文尼亚',
    'Solomon Is.': '所罗门群岛',
    Somalia: '索马里',
    'South Africa': '南非',
    Spain: '西班牙',
    'Sri Lanka': '斯里兰卡',
    Sudan: '苏丹',
    Suriname: '苏里南',
    Swaziland: '斯威士兰',
    Sweden: '瑞典',
    Switzerland: '瑞士',
    Syria: '叙利亚',
    Tajikistan: '塔吉克斯坦',
    Tanzania: '坦桑尼亚',
    Thailand: '泰国',
    Togo: '多哥',
    Tonga: '汤加',
    'Trinidad and Tobago': '特立尼达和多巴哥',
    Tunisia: '突尼斯',
    Turkey: '土耳其',
    Turkmenistan: '土库曼斯坦',
    'U.S. Virgin Islands': '美属维尔京群岛',
    Uganda: '乌干达',
    Ukraine: '乌克兰',
    'United Arab Emirates': '阿拉伯联合酋长国',
    'United Kingdom': '英国',
    'United States': '美国',
    Uruguay: '乌拉圭',
    Uzbekistan: '乌兹别克斯坦',
    Vanuatu: '瓦努阿图',
    'Vatican City': '梵蒂冈城',
    Venezuela: '委内瑞拉',
    Vietnam: '越南',
    'W. Sahara': '西撒哈拉',
    Yemen: '也门',
    Yugoslavia: '南斯拉夫',
    Zaire: '扎伊尔',
    Zambia: '赞比亚',
    Zimbabwe: '津巴布韦',
  }

  let option = {
    // backgroundColor: '#1b1b1b',
    // color: ['gold', 'aqua', 'lime'],
    title: {
      show: true,
      text: '实时流量可视化 10s内活跃连接',
      x: 'center',
      // textStyle: {
      //   color: '#fff',
      // },
    },
    tooltip: {
      // trigger: 'item',
      // formatter: function (params, ticket, callback) {
      //   /*$.get('detail?name=' + params.name, function (content) {
      //         callback(ticket, toHTML(content));
      //     });*/
      //   var tips = '<ul style="list-style: none">'
      //   tips += '<li>行政区划：' + params.name + '</li>'
      //   tips += '<li>历史已报件：1000个</li>'
      //   tips += '<li>最近上报时间：' + new Date().toLocaleTimeString() + '</li>'
      //   tips += '</ul>'
      //   return tips
      // },
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
      },
    },
    geo: {
      map: 'world',
      hoverable: false,
      nameMap: nameMap,
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
        zlevel: 1,
        coordinateSystem: 'geo', //  指明绘制在geo坐标系上
        animationDelay: 500,
        data: [],
        tooltip: {
          trigger: 'item',
          formatter: (params, ticket, callback) => {
            /*$.get('detail?name=' + params.name, function (content) {
              callback(ticket, toHTML(content));
          });*/
            let tips = `<p>${params.name}</p>`
            return tips
          },
        },
      },
      {
        type: 'lines',
        coordinateSystem: 'geo',
        zlevel: 2,
        symbolSize: 110,
        effect: {
          // 模拟效果路线特效
          show: true,
          period: 6,
          trailLength: 0,
          symbolSize: 3,
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
    // console.log('update', option.series[0].data)
    chart.setOption(option)
  }
  instantRun()
  setInterval(async () => {
    await instantRun()
  }, 5000)
}

start()
