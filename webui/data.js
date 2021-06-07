//上报数据的城市，其name对应geo地图数据中的properties数组元素的name
var dataSource = [{
  dataSourceId: 2018,
  data: [
      {
          name: '文山壮族苗族自治州',
          value: 10
      },
      {
          name: '怒江傈僳族自治州',
          value: 20
      },
      {
          name: '西双版纳傣族自治州',
          value: 20
      },
      {
          name: '大理白族自治州',
          value: 10
      },
      {
          name: '楚雄彝族自治州',
          value: 20
      },
      {
          name: '曲靖市',
          value: 1
      }]
},
  {
      dataSourceId: 2017,
      data: [
          {
              name: '文山壮族苗族自治州',
              value: 90
          }
      ]
  }];
//中心汇入点城市
var center = '昆明市';
