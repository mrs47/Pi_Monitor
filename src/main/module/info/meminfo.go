// Author： mrs47
// Since:   2019/11/27
// Version: 1.0
// Describe:用于获取Linux 内存信息
package info

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// 获取内存信息路径
const MEM_PATH string = "/proc/meminfo"
// 封装内存信息
type MEMInfo  struct {
	MemTotal  string `json:"memTotal"`
	SwapTotal string `json:"swapTotal"`
	MemFree   [10]string `json:"memFree"`
	SwapFree  [10]string `json:"swapFree"`
	Time      [10]string `json:"time"`
}

// (meminfo *MEMInfo)updateMemInfo() 函数：
// 每30秒调用readMemInfo()函数获取一次内存信息
// 完成更新共计5分钟
func (meminfo *MEMInfo) updateMemInfo(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("➤ \tMem: 采集线程开始 \t🔥")
	infoMap := readMemInfo()
	meminfo.MemTotal = infoMap["MemTotal"]
	meminfo.SwapTotal = infoMap["SwapTotal"]
	meminfo.MemFree[0] = infoMap["MemFree"]
	meminfo.SwapFree[0] = infoMap["SwapFree"]
	meminfo.Time[0] = getTimeNow()
	for i:=1; i<10 ;i++ {
		infoMap = readMemInfo()
		meminfo.MemFree[i] = infoMap["MemFree"]
		meminfo.SwapFree[i] = infoMap["SwapFree"]
		meminfo.Time[i] = getTimeNow()
		time.Sleep(18*time.Second)
	}
	log.Println("➤ \tMem: ok \t\t✔")
}
// readMemInfo()map[string]string 函数：
// 从文件中读取内存信息
// 返回map集合
func readMemInfo()map[string]string{
	memMap := make(map[string]string)
	file,err := os.Open(MEM_PATH)
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
		memMap[strings.Split(string(line),":")[0]] = strings.ReplaceAll(strings.Split(string(line),":")[1]," ","")
	}
	return memMap
}
