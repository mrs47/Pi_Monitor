// Author： mrs47
// Since:   2019/11/30
// Version: 1.0
// Describe:用于获取Linux 网络状态信息
package info

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NET_PATH=/proc/net/dev,NET_PATH保存网络信息的文件位置
const NET_PATH string = "/proc/net/dev"
// NetInfo 封装了网络信息，其中包含
// NetInterface map[string]int     网络接口信息
// NetIn        map[string][10]int 网络接口的下载（下行）速度
// NetOut       map[string][10]int 网络接口的上传（上行）速度
// Time         map[string][10]string 对应时间戳
type NetInfo struct {
	NetInterface map[string]int `json:"net_interface"`
	NetIn        map[string][10]int `json:"netIn"`
	NetOut       map[string][10]int `json:"netOut"`
	Time         [10]string `json:"time"`
}
// 用于初始化结构体内的map集合（map集合需要被初始化才能使用）
func (netinfo *NetInfo)init()  {
	netinfo.NetInterface = make(map[string]int)
	netinfo.NetIn = make(map[string][10]int)
	netinfo.NetOut = make(map[string][10]int)
	//netinfo.Time = make([10]string)
}
// updateNetInfo() 主入口函数
func (netinfo *NetInfo)updateNetInfo(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("➤ \tNet: 采集线程开始 \t🔥")
	netinfo.init()
	netinfo.getNetInterface()
	for i := 0;i < 10 ; i++ {
		receive,transmit,infoTime := netinfo.getNetInfo()
		for k, v := range transmit {
			netinfo.Time[i] = infoTime

			temp := netinfo.NetOut[k]
			temp[i]=v
			netinfo.NetOut[k] = temp
		}
		for k, v := range receive {
			temp := netinfo.NetIn[k]
			temp[i] = v
			netinfo.NetIn[k] = temp
		}
		time.Sleep(18*time.Second)
	}
	log.Println("➤ \tNet: ok \t\t✔")
}

// getNetInfo()函数
// 获取网络状态信息 包括：上行、下行
// netMap1、netMap2相隔1秒采集，计算每秒的数据量。
// 单位：bytes
func (netinfo *NetInfo)getNetInfo()(receive map[string]int, transmit map[string]int,infoTime string){
	receive = make(map[string]int)
	transmit = make(map[string]int)
	netMap1 := readNowNetInfo()
	time.Sleep(1*time.Second)
	netMap2 := readNowNetInfo()
	for k, v := range netMap1 {
		value,_ := strconv.Atoi(v[0])
		receive[k] = value
	}
	for k, v := range netMap1 {
		value,_ := strconv.Atoi(v[8])
		transmit[k ]= value
	}
	for k, v := range netMap2 {
		value,_ := strconv.Atoi(v[0])
		receive[k] = value-receive[k]
	}
	for k, v := range netMap2 {
		value,_:= strconv.Atoi(v[8])
		transmit[k] = value-transmit[k]
	}
	infoTime = getTimeNow()
	return
}
// getNetInterface()
// 获取设备网络接口包括蓝牙接口
func (netinfo *NetInfo)getNetInterface(){
	file,err := os.Open(NET_PATH)
	if err != nil{
		log.Println("❌ \t错误:",err)
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for{
		line,_,err := reader.ReadLine()
		if err != nil{
			//log.Println("❌ \t错误:",err)
			break
		}
		if strings.Contains(string(line), ":") {
			temp := strings.ReplaceAll( strings.Split(string(line),":")[0]," ","")
			netinfo.NetInterface[temp] = 1
		}
	}
}
// readNowNetInfo()
// 实现从文件”NET_PATH“ 读取网络接口信息
// NET_PATH = /proc/net/dev
func readNowNetInfo()map[string][16]string{
	MEM_MAP := make(map[string][16]string)
	file,err := os.Open(NET_PATH)
	if err != nil{
		log.Println("❌ \t错误:",err)
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for{
		line,_,err := reader.ReadLine()
		if err != nil{
			//log.Println("❌ \t错误:",err)
			break
		}
		if strings.Contains(string(line), ":") {
			var value [16]string
			i := 0
			temp := strings.Split( strings.Split(string(line),":")[1]," ")
			for _,temp2 := range temp{
				if temp2 != "" {
					value[i]=temp2
					i++
				}
			}
			MEM_MAP[strings.ReplaceAll(strings.Split(string(line),":")[0]," ","")] = value
		}
	}
	return MEM_MAP
}