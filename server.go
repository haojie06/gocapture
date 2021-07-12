package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/websocket"
)

var wsConnList []*websocket.Conn
var sigChan, jsonChan, strChan chan string
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 解决跨域问题 (允许) 除了域名和端口,跨域和协议也相关
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	var listenPort string
	startMenu()
	fmt.Print("请输入web服务器监听端口:")
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
	http.HandleFunc("/ws", wsHandler)
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

//websocket处理
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// allowCORS(w)
	conn, err := upgrader.Upgrade(w, r, nil)
	// handleErr(err, "升级ws连接")
	//开一个协程和客户端通信
	if err == nil {
		// 添加到列表中,之后数据更新的时候进行通知
		wsConnList = append(wsConnList, conn)
		sigChan <- "initData"
		jsonData := <-jsonChan
		// 建立连接后,立即推送初始数据
		conn.WriteMessage(websocket.TextMessage, []byte(jsonData))
		// 暂时不需要从连接中读取输入,所以先注释掉了
		// go func() {
		// 	//升级连接后，持续的读写
		// 	defer conn.Close()
		// 	// 注册断开处理
		// 	conn.SetCloseHandler(func(code int, text string) error {
		// 		// 是否应该用更好的方式清除列表中已断开的连接?
		// 		log.Println("客户端断开连接", code, "->", text)
		// 		return nil
		// 	})
		// 	for {
		// 		messageType, message, err := conn.ReadMessage()
		// 		logErr(err, "从websocket中读取数据")
		// 		err = conn.WriteMessage(messageType, message)
		// 		logErr(err, "向websocket中写入数据")
		// 	}
		// }()
	} else {
		// 无效ws请求
		fmt.Println("无效请求", err.Error())
		fmt.Fprintf(w, "无效请求")
	}
}

// 主协程监听web端口，所以需要第三个协程与抓包协程通信并计算,以及暂时保存数据
func getData() {
	// 用于和web服务器通信 (测试时传递文本而不是List)
	bandwidthListChan := make(chan BandwidthData, 10)
	wsDataChan := make(chan IPStruct)
	var bandwidthData BandwidthData
	go gocapture(bandwidthListChan, wsDataChan)
	// 需要使用select 关键字  不断从bandwidthChan中获取信息 存入变量中，从dataChan收到信号则传送数据
	for {
		select {
		case signal := <-sigChan:
			{
				// 接收到信号返回本地变量 (被动推送)
				if signal == "getStr" {
					strChan <- bandwidthData.BandwidthStatisticStr
				} else if signal == "getData" {
					// JSON形式返回
					// 对时间参数进行筛选(仅对JSON请求有效)
					jsonData, err := json.Marshal(bandwidthData.BandwidthList)
					handleErr(err, "序列化BandwidthList为JSON")
					jsonChan <- string(jsonData)
				} else if signal == "initData" {
					// 建立ws连接后立即推送一次数据
					jsonData, err := json.Marshal(bandwidthData)
					handleErr(err, "序列化BandwidthData为JSON")
					jsonChan <- string(jsonData)
				}
			}
		case data := <-bandwidthListChan:
			{
				// 改成主动通过websocket推送?
				// 不断从抓包协程里获得流量统计信息，并更新本地变量
				bandwidthData = data
				jsonData, err := json.Marshal(bandwidthData)
				handleErr(err, "序列化BandwidthList为JSON")
				// 需要发送两个消息,怎么区分呢
				writeMessageThroughWS(jsonData)
			}
			// case _ = <-wsDataChan:
			// 	{
			// 		// 立即向当前开放的ws conn列表推送
			// 		// jsonData, err := json.Marshal(wsData)
			// 		// handleErr(err, "ws传输数据转为JSON")
			// 		// writeMessageThroughWS(jsonData)
			// 	}
		}
	}
}
func writeMessageThroughWS(msg []byte) {
	for index, conn := range wsConnList {
		//如果向关闭的连接写数据,会有异常,移除该连接
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			conn.Close()
			wsConnList = append(wsConnList[:index], wsConnList[index+1:]...)
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
