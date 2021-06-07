package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var dataChan chan string

func captureHandler(w http.ResponseWriter, req *http.Request) {
	// 向getData协程发送信号
	param := req.URL.Path[1:]
	if param == "str" {
		dataChan <- "getStr"
	} else if param == "json" {
		dataChan <- "getData"
	} else {
		dataChan <- " "
	}
	data := <-dataChan
	fmt.Fprintf(w, data)
}

func main() {
	dataChan = make(chan string)
	go getData(dataChan)
	http.HandleFunc("/", captureHandler)
	err := http.ListenAndServe("localhost:8080", nil)
	handleErr(err)
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

// 将bandwidthList 转为JSON
// func dataToJson(bandwidthList PairList) {

// }
