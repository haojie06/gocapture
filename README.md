# gocapture

抓包以及流量统计,并在 web 进行可视化 [demo](http://con.ifine.eu:8080/)
运行前安装 winpcap（windows）或者 libpcap-dev（linux）
当前可视化默认支持本地的 8080 端口(如果使用其他端口以及修改城市经纬度，可以编辑 config.js)，
![gif](https://github.com/aoyouer/gocapture/raw/main/gif/CPT2106080056-800x385.gif)

通过 http://localhost:8080 访问流量地图

web 接口 /str 或者 /json 可以获取文本数据。
