package main

import (
	"fmt"
	"log"
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

type IPStruct struct {
	inBytes    int
	outBytes   int
	totalBytes int
}

// 排序输出
type Pair struct {
	Key   string
	Value *IPStruct
}

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value.totalBytes < p[j].Value.totalBytes }

type PairList []Pair

// 需要以管理员权限运行 以及安装 winpcap或者libpcap
func main() {
	//流量统计 ip map 注意是一个指针map，可以直接修改其中元素
	bandwidthMap := make(map[string]*IPStruct)
	// Find all devices
	devices, err := pcap.FindAllDevs()
	handleErr(err)
	// Print device information
	// 不知道为什么, 没有获取到WLAN的ipv4,只得到了v6地址
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
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~")
	}
	fmt.Print("请选择一张网卡进行抓包:")
	var selectIndex int
	fmt.Scanln(&selectIndex)
	clearScreen()
	fmt.Println("开始进行抓包")
	// timeout表示多久刷新一次数据包，负数表示立即刷新
	handle, err := pcap.OpenLive(devices[selectIndex].Name, 1024, false, -1)
	handleErr(err)
	defer handle.Close()
	f, _ := os.Create("test.pcap")
	w := pcapgo.NewWriter(f)
	w.WriteFileHeader(1024, layers.LinkTypeEthernet)
	defer f.Close()
	packetCount := 0
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		// Process packet here
		packetCount++
		// 是否要写到文件中去
		w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		// 是否实时打印包
		// log.Println(packet.NetworkLayer().NetworkFlow().Dst().String())
		// 考虑到流量统计...不开混杂模式的时候只抓得到本地的包
		// 首先判断src部分
		//!! 注意 ARP的包没有网络层...所以会出现空指针错误
		if packet.NetworkLayer() != nil {
			// log.Println(packet.NetworkLayer().NetworkFlow().String())
			if ipBandwithInfo, ok := bandwidthMap[packet.NetworkLayer().NetworkFlow().Src().String()]; ok {
				// 已经有记录时
				ipBandwithInfo.outBytes += packet.Metadata().Length
				ipBandwithInfo.totalBytes = ipBandwithInfo.inBytes + ipBandwithInfo.outBytes
			} else {
				// 还没有对应ip的记录时
				bandwidthMap[packet.NetworkLayer().NetworkFlow().Src().String()] = &IPStruct{outBytes: packet.Metadata().Length, inBytes: 0, totalBytes: packet.Metadata().Length}
			}

			// 然后是 dst部分
			if ipBandwithInfo, exist := bandwidthMap[packet.NetworkLayer().NetworkFlow().Dst().String()]; exist {
				// 已经有记录时
				ipBandwithInfo.inBytes += packet.Metadata().Length
				ipBandwithInfo.totalBytes = ipBandwithInfo.inBytes + ipBandwithInfo.outBytes
			} else {
				// 还没有对应ip的记录时
				bandwidthMap[packet.NetworkLayer().NetworkFlow().Dst().String()] = &IPStruct{outBytes: 0, inBytes: packet.Metadata().Length, totalBytes: packet.Metadata().Length}
			}
			// 每十个包打印一次统计
			if packetCount >= 100 {
				clearScreen()
				fmt.Println("MAP LENGTH:", len(bandwidthMap))
				bandwidthList := sortIPs(bandwidthMap)
				for _, ips := range bandwidthList {
					record := geoIP(ips.Key)
					fmt.Printf("ip: %-16s output: %-2.2fMB input: %-2.2fMB total: %-2.2fMB country: %-8s\n", ips.Key, float32(ips.Value.outBytes)/8388608, float32(ips.Value.inBytes)/8388608, float32(ips.Value.totalBytes)/8388608, record.Country.Names["en"])
					// fmt.Println("IP:", ips.Key, "OUT BYTES:", ips.Value.outBytes, "IN BYTES:", ips.Value.inBytes, "TOTAL BYTES:", ips.Value.totalBytes, "COUNTRY:", record.Country.Names["en"])
				}
				packetCount = 0
			}
			// fmt.Printf("\b\b\b\b\b\b")
			// fmt.Printf("%d/100", packetCount)
			if packetCount%10 == 0 {
				fmt.Print(".")
			}
		}
	}
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
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
	return pl
}

func geoIP(ipStr string) *geoip2.Country {
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	handleErr(err)
	// defer db.Close()
	ip := net.ParseIP(ipStr)
	record, err := db.Country(ip)
	handleErr(err)
	return record
}
