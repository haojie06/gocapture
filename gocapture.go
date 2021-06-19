package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/oschwald/geoip2-golang"
)

// 需要以管理员权限运行以及安装 winpcap或者libpcap
func gocapture(bandwidthDataChan chan BandwidthData, wsDataChan chan IPStruct) {
	var option Option
	// 流量统计 ip map 注意是一个指针map，可以直接修改其中元素
	bandwidthMap := make(map[string]*IPStruct)
	// 选择进行抓包的网卡、刷新频率等
	setOption(&option)
	// 抓包并打印
	capturePackets(bandwidthMap, option, bandwidthDataChan, wsDataChan)
}

func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func sortIPs(bandwidthMap map[string]*IPStruct) PairList {
	pl := make(PairList, len(bandwidthMap))
	i := 0
	for k, v := range bandwidthMap {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	// sort.Sort(pl)
	return pl
}
func getGeoDb(dbType string) *geoip2.Reader {
	var db *geoip2.Reader
	var err error
	if dbType == "city" {
		db, err = geoip2.Open("GeoLite2-City.mmdb")
		handleErr(err, "打开GEO City数据库")
	} else if dbType == "country" {
		db, err = geoip2.Open("GeoLite2-Country.mmdb")
		handleErr(err, "打开GEO Country数据库")
	}
	return db
}
func geoIPCountry(ipStr string, geoDB *geoip2.Reader) *geoip2.Country {
	// 不要重复打开！使用闭包！
	// defer db.Close()
	ip := net.ParseIP(ipStr)
	record, err := geoDB.Country(ip)
	handleErr(err, "解析国家信息")
	return record
}

func geoIPCity(ipStr string, geoDB *geoip2.Reader) *geoip2.City {
	// defer db.Close()
	ip := net.ParseIP(ipStr)
	record, err := geoDB.City(ip)
	handleErr(err, "解析城市信息")
	return record
}

// 开始前配置
func setOption(option *Option) {
	var selectIndex, flushInterval, ifWritePcap int
	devices, err := pcap.FindAllDevs()
	handleErr(err, "获取设备")
	// Print device information
	fmt.Println("Devices found:")
	for index, device := range devices {
		fmt.Println("第" + strconv.Itoa(index) + "张网卡")
		fmt.Println("Name: ", device.Name)
		fmt.Println("Description: ", device.Description)
		fmt.Println("Devices addresses: ", device.Description)
		for _, address := range device.Addresses {
			fmt.Println("- IP address: ", address.IP)
			fmt.Println("- Subnet mask: ", address.Netmask)
		}
		fmt.Println("-----------------------------------------------------------------")
	}
	fmt.Print("请选择一张网卡进行抓包: ")
	_, err = fmt.Scanln(&selectIndex)
	handleErr(err, "非法输入")
	fmt.Print("请选择多少个包刷新一次流量统计(大流量请设置高一些): ")
	_, err = fmt.Scanln(&flushInterval)
	handleErr(err, "非法输入")
	if flushInterval == 0 {
		flushInterval = 500
	}
	fmt.Print("是否写入pcap文件packet.pcap 如果只是想统计流量, 请选择否 1.是 2.否: ")
	_, err = fmt.Scanln(&ifWritePcap)
	handleErr(err, "非法输入")
	option.deviceName = devices[selectIndex].Name
	option.flushInterval = flushInterval
	clearScreen()
	fmt.Println("开始进行抓包")
}

func capturePackets(bandwidthMap map[string]*IPStruct, option Option, bandwidthDataChan chan BandwidthData, wsDataChan chan IPStruct) {
	deviceName := option.deviceName
	flushInterval := option.flushInterval
	// timeout表示多久刷新一次数据包，负数表示立即刷新
	handle, err := pcap.OpenLive(deviceName, 1024, false, -1)
	handleErr(err, "打开设备的流")
	defer handle.Close()
	var w *pcapgo.Writer
	if option.ifWritePcap == 1 {
		f, _ := os.Create("packet.pcap")
		w = pcapgo.NewWriter(f)
		w.WriteFileHeader(1024, layers.LinkTypeEthernet)
		defer f.Close()
	}
	packetCount := 0
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	// 打开GEO数据库
	// 根据不同版本开启不同的geo数据库
	geoDB := getGeoDb("city")
	defer geoDB.Close()
	for packet := range packetSource.Packets() {
		// Process packet here
		packetCount++
		// 是否要写到文件中去
		if option.ifWritePcap == 1 {
			w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		}
		// 是否实时打印包
		// log.Println(packet.NetworkLayer().NetworkFlow().Dst().String())
		// 考虑到流量统计...不开混杂模式的时候只抓得到本地的包
		// 首先判断src部分
		//!! 注意 ARP的包没有网络层...所以会出现空指针错误
		// var pushIPInfo IPStruct
		if packet.NetworkLayer() != nil {
			// log.Println(packet.NetworkLayer().NetworkFlow().String())
			if ipBandwithInfo, ok := bandwidthMap[packet.NetworkLayer().NetworkFlow().Src().String()]; ok {
				// 已经有记录时
				ipBandwithInfo.OutBytes += packet.Metadata().Length
				ipBandwithInfo.TotalBytes = ipBandwithInfo.InBytes + ipBandwithInfo.OutBytes
				ipBandwithInfo.LastActive = packet.Metadata().Timestamp
				// pushIPInfo = *ipBandwithInfo
			} else {
				// 还没有对应ip的记录时
				bandwidthMap[packet.NetworkLayer().NetworkFlow().Src().String()] = &IPStruct{OutBytes: packet.Metadata().Length, InBytes: 0, TotalBytes: packet.Metadata().Length, LastActive: packet.Metadata().Timestamp}
				// pushIPInfo = *bandwidthMap[packet.NetworkLayer().NetworkFlow().Src().String()]
			}
			// 然后是 dst部分
			if ipBandwithInfo, exist := bandwidthMap[packet.NetworkLayer().NetworkFlow().Dst().String()]; exist {
				// 已经有记录时
				ipBandwithInfo.InBytes += packet.Metadata().Length
				ipBandwithInfo.TotalBytes = ipBandwithInfo.InBytes + ipBandwithInfo.OutBytes
				ipBandwithInfo.LastActive = packet.Metadata().Timestamp
				// pushIPInfo = *ipBandwithInfo
			} else {
				// 还没有对应ip的记录时
				bandwidthMap[packet.NetworkLayer().NetworkFlow().Dst().String()] = &IPStruct{OutBytes: 0, InBytes: packet.Metadata().Length, TotalBytes: packet.Metadata().Length, LastActive: packet.Metadata().Timestamp}
				// pushIPInfo = *bandwidthMap[packet.NetworkLayer().NetworkFlow().Dst().String()]
			}
			// [临时放置]推送给ws频道
			// wsDataChan <- pushIPInfo
			wsDataChan = nil

			fmt.Printf("\r[%d/%d]", packetCount, flushInterval)
			// 每flushInterval个包打印一次统计
			if packetCount >= flushInterval {
				printStatistic(bandwidthMap, "city", bandwidthDataChan, geoDB)
				packetCount = 0
			}

		}
	}
}

// 数据量转化为可读性较高的方式
func dataTransfer(byteCount int) string {
	var formatBandwidth string
	if byteCount < 1024 {
		formatBandwidth = strconv.Itoa(byteCount) + "B"
	} else if byteCount < 1048576 {
		formatBandwidth = strconv.Itoa(byteCount/1024) + "KB"
	} else if byteCount < 1073741824 {
		formatBandwidth = strconv.Itoa(byteCount/1048576) + "MB"
	} else if byteCount < int(math.Pow(1024, 4)) {
		//较大单位保留2位小数
		formatBandwidth = fmt.Sprintf("%.2fGB", (float64(byteCount) / math.Pow(1024, 3)))
		// formatBandwidth = strconv.Itoa(byteCount/int(math.Pow(1024, 3))) + "GB"
	} else {
		formatBandwidth = fmt.Sprintf("%.3fTB", (float64(byteCount) / math.Pow(1024, 4)))
	}
	return formatBandwidth
}

// 打印统计信息 (这个传参嵌套太多层了)
func printStatistic(bandwidthMap map[string]*IPStruct, geoType string, bandwidthDataChan chan BandwidthData, geoDB *geoip2.Reader) {
	var bandwidthData BandwidthData
	drawStr := fmt.Sprintf("记录IP数: %d", len(bandwidthMap))
	// 通过Slice对Map进行排序
	bandwidthList := sortIPs(bandwidthMap)
	// bandwidthListChan <- bandwidthList
	// listLen := len(bandwidthList)
	for index, ips := range bandwidthList {
		//当前使用城市IP库 (影响Location字段)
		var IPLocation string
		if geoType == "city" {
			record := geoIPCity(ips.Key, geoDB)
			if record.Country.Names["en"] == "" {
				IPLocation = "PrivateIP"
			} else {
				IPLocation = fmt.Sprintf("%s - %s (%f,%f)", record.Country.Names["en"], record.City.Names["en"], record.Location.Longitude, record.Location.Latitude)
			}
			// 添加geoip信息到List中
			ips.Value.City = record.City.Names["en"]
			ips.Value.Country = record.Country.Names["en"]
			ips.Value.Longitude = record.Location.Longitude
			ips.Value.Latitude = record.Location.Latitude
		} else if geoType == "country" {
			record := geoIPCountry(ips.Key, geoDB)
			if record.Country.Names["en"] == "" {
				IPLocation = "PrivateIP"
			} else {
				IPLocation = fmt.Sprintf("%s", record.Country.Names["en"])
			}
		}
		if index == 0 {
			drawStr = fmt.Sprintf("%s\nip: %-16s output: %-6s input: %-6s total: %-7s location: %-8s(Local)", drawStr, ips.Key, dataTransfer(ips.Value.OutBytes), dataTransfer(ips.Value.InBytes), dataTransfer(ips.Value.TotalBytes), IPLocation)
		} else {
			drawStr = fmt.Sprintf("%s\nip: %-16s output: %-6s input: %-6s total: %-7s location: %-8s", drawStr, ips.Key, dataTransfer(ips.Value.OutBytes), dataTransfer(ips.Value.InBytes), dataTransfer(ips.Value.TotalBytes), IPLocation)
		}
	}
	clearScreen()
	fmt.Println(drawStr)
	// 写在这不合适，该函数应该专职打印
	bandwidthData.BandwidthStatisticStr = drawStr
	bandwidthData.BandwidthList = bandwidthList
	bandwidthDataChan <- bandwidthData
}
