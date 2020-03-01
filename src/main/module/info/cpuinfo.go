// Author： mrs47
// Since:   2019/11/25
// Version: 1.0
// Describe:用于获取Linux CPU信息
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

// CPU状态信息，用于统计CPU使用率
const CPU_PATH string = "/proc/stat"
// 用于获取CPU 架构、名称
const CPUINFO_PATH string = "/proc/cpuinfo"
// 用于获取CPU 温度
const CPUTEMP_PATH string = "/sys/class/thermal/thermal_zone0/temp"

// 封装CPU 信息
// Usage    [10]int 使用率
// Temp     [10]string CPU 温度
// Time     [10]string 时间戳
// CpuModel string 架构名
type CPUInfo struct {
	 Usage    [10]float64 `json:"usage"`
	 Temp     [10]string `json:"temp"`
	 Time     [10]string `json:"time"`
	 CpuModel string `json:"cpu_model"`
}

// (cpuinfo *CPUInfo)updateCPUInfo()  函数：
// 入口函数
// 约每30秒获取一组CPU 状态信息
// 一次数据更新 共5分钟
func (cpuinfo *CPUInfo) updateCPUInfo(wg *sync.WaitGroup ) {
	defer wg.Done()
	log.Println("➤ \tCPU: 采集线程开始 \t🔥")
	cpuinfo.getCpuModel()
	for count := 0; count<10 ; count++ {
		time.Sleep(18 *time.Second)
		var cpuUsages [5]float64
		for i:=0;i < 5;i++{
			info1 := getCpuInfo()
			time.Sleep(20*time.Millisecond)
			info2 := getCpuInfo()
			cpuUsages[i] = changeInfo(info1,info2)
			time.Sleep(20*time.Millisecond)
		}
		var sum float64
		for _,value := range cpuUsages{
			sum += value
		}
		avg := sum/5
		cpuinfo.Usage[count] = avg
		cpuinfo.Temp[count] =getCPUTemp()
		cpuinfo.Time[count] = getTimeNow()
	}
	log.Println("➤ \tCPU: ok \t\t✔")
}
// 从文件中获取CPU 使用信息
func getCpuInfo() [10]float64{
	file,err := os.Open(CPU_PATH)
	isSuccess := checkErr(err)
	if !isSuccess {
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	info,_,err := reader.ReadLine()
	data := string(info)
	datas := (strings.Split(data," "))[2:]
	var temp [10]float64
	for index,value := range datas {
		temp[index],_ = strconv.ParseFloat(value,64)
	}
	return temp
}
// 从文件中获取CPU 架构信息
func (cpuinfo *CPUInfo) getCpuModel(){
	file,err := os.Open(CPUINFO_PATH)
	isSuccess := checkErr(err)
	if !isSuccess {
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	for   {
		info,_,err := reader.ReadLine()
		if err != nil {
			//log.Println("❌ \t错误:",err)
			break
		}
		if strings.Contains(string(info),":") {
			data:=strings.Split(string(info),":")
			if strings.Contains(data[0],"model name") {
				cpuinfo.CpuModel = data[1]
				return
			}
		}
	}
	cpuinfo.CpuModel="未知"
}
// 获取CPU 温度
// 单位：摄氏度
// 精度：小数点后一位
func getCPUTemp()string{
	file,err:=os.Open(CPUTEMP_PATH)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return "-999999"
	}
	defer file.Close()
	reader:=bufio.NewReader(file)
	line,_,_:=reader.ReadLine()
	temp,_:=strconv.ParseFloat(string(line),32)
	temp=temp/1000
	return strconv.FormatFloat(temp,'f',1,32)
}

// 将数据进行计算
func changeInfo(info1 [10]float64, info2 [10]float64) float64{
	var s1 float64
	var s2 float64
	for _,value := range info1{
		s1+=value
	}
	for _,value := range info2{
		s2+=value
	}
	idle := info2[3]-info1[3]
	total := s2-s1
	cpuUsage := (total-idle)/total*100
	return cpuUsage
}

