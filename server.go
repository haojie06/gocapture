package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var sigChan, jsonChan, strChan chan string

func main() {
	var listenPort string
	fmt.Print("请输入web服务器监听端口: ")
	fmt.Scanln(&listenPort)
	// 不要使用同一个管道传输str和json，不然在遇见同时请求的时候会出问题（顺序不符合预期）
	sigChan = make(chan string)
	jsonChan = make(chan string)
	strChan = make(chan string)
	// ！！！注意，如果模板中用了本地的js、css文件，需要先指明文件在哪
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	go getData()
	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/str/", strHandler)
	http.HandleFunc("/json/", jsonHandler)
	err := http.ListenAndServe(":"+listenPort, nil)
	handleErr(err, "开启服务器监听")
}
func pageHandler(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	//加载模板
	tpl, err := template.ParseFiles("templates/index.tmpl")
	handleErr(err, "解析模板")
	err = tpl.Execute(w, "")
	handleErr(err, "加载模板")
}
func strHandler(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	sigChan <- "getStr"
	data := <-strChan
	fmt.Fprintf(w, data)
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	// param := req.URL.Path[1:]
	sigChan <- "getData"
	data := <-jsonChan
	fmt.Fprintf(w, data)
}

// 主协程监听web端口，所以需要第三个协程与抓包协程通信并计算,以及暂时保存数据
func getData() {
	// 用于和web服务器通信 (测试时传递文本而不是List)
	bandwidthListChan := make(chan BandwidthData)
	var bandwidthData BandwidthData
	go gocapture(bandwidthListChan)
	// 需要使用select 关键字  不断从bandwidthChan中获取信息 存入变量中，从dataChan收到信号则传送数据
	for {
		select {
		case signal := <-sigChan:
			{
				// 接收到信号返回本地变量
				if signal == "getStr" {
					strChan <- bandwidthData.bandwidthStatisticStr
				} else if signal == "getData" {
					// JSON形式返回
					// 对时间参数进行筛选(仅对JSON请求有效)
					jsonData, err := json.Marshal(bandwidthData.bandwidthList)
					handleErr(err, "序列化BandwidthList为JSON")
					jsonChan <- string(jsonData)
				}
			}
		case data := <-bandwidthListChan:
			{
				// 不断从抓包协程里获得流量统计信息，并更新本地变量
				bandwidthData = data
			}
		}
	}
}

func allowCORS(w http.ResponseWriter) {
	// 允许跨域
	w.Header().Set("Access-Control-Allow-Origin", "*")                                                            // 允许访问所有域，可以换成具体url，注意仅具体url才能带cookie信息
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token") //header的类型
	w.Header().Add("Access-Control-Allow-Credentials", "true")                                                    //设置为true，允许ajax异步请求带cookie信息
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")                             //允许请求方法
}

// 将bandwidthList 转为JSON
// func dataToJson(bandwidthList PairList) {

// }
