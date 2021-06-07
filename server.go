package main

import (
	"fmt"
	"log"
	"net/http"
)

var dataChan chan string

func HelloServer(w http.ResponseWriter, req *http.Request) {
	// 向getData协程发送信号
	dataChan <- "get"
	data := <-dataChan
	fmt.Println("Inside HelloServer handler")
	fmt.Fprintf(w, data)
}

func main() {
	dataChan = make(chan string)
	go getData(dataChan)
	http.HandleFunc("/", HelloServer)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

// 主协程监听web端口，所以需要第三个协程与抓包协程通信并计算,以及暂时保存数据
func getData(dataChan chan string) {
	// 用于和web服务器通信 (测试时传递文本而不是List)
	bandwidthListChan := make(chan string)
	var bandwidthData string
	go gocapture(bandwidthListChan)
	// 需要使用select 关键字  不断从bandwidthChan中获取信息 存入变量中，从dataChan收到信号则传送数据
	for {
		select {
		case signal := <-dataChan:
			{
				// 接收到信号返回本地变量
				if signal == "get" {
					dataChan <- bandwidthData
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
