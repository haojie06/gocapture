package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var dataChan chan string

func main() {
	var listenPort string
	fmt.Println("请输入web服务器监听端口")
	fmt.Scanln(&listenPort)
	dataChan = make(chan string)
	// ！！！注意，如果模板中用了本地的js、css文件，需要先指明文件在哪
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	go getData(dataChan)
	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/str", strHandler)
	http.HandleFunc("/json", jsonHandler)
	err := http.ListenAndServe(":"+listenPort, nil)
	handleErr(err)
}
func pageHandler(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	//加载模板
	tpl, err := template.ParseFiles("templates/index.tmpl")
	handleErr(err)
	err = tpl.Execute(w, "")
	handleErr(err)
}
func strHandler(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	dataChan <- "getStr"
	data := <-dataChan
	fmt.Fprintf(w, data)
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	dataChan <- "getData"
	data := <-dataChan
	fmt.Fprintf(w, data)
}

// 主协程监听web端口，所以需要第三个协程与抓包协程通信并计算,以及暂时保存数据
func getData(dataChan chan string) {
	// 用于和web服务器通信 (测试时传递文本而不是List)
	bandwidthListChan := make(chan BandwidthData)
	var bandwidthData BandwidthData
	go gocapture(bandwidthListChan)
	// 需要使用select 关键字  不断从bandwidthChan中获取信息 存入变量中，从dataChan收到信号则传送数据
	for {
		select {
		case signal := <-dataChan:
			{
				// 接收到信号返回本地变量
				if signal == "getStr" {
					dataChan <- bandwidthData.bandwidthStatisticStr
				} else if signal == "getData" {
					// JSON形式返回
					jsonData, err := json.Marshal(bandwidthData.bandwidthList)
					handleErr(err)
					dataChan <- string(jsonData)
				} else {
					dataChan <- "无效信号"
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
