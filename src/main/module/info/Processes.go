// Author： mrs47
// Since:   2019/12/02
// Version: 1.0
// Describe:用于获取Linux 进程信息
package info

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)
// 进程存储路径
const UPTIME         string = "/proc/uptime"
const MEMINFO_PATH   string = "/proc/meminfo"
const UID_USERNAME   string = "/etc/passwd"
const PROCESSES_PATH string = "/proc/"
// 进程结构体
type Process struct {
	PID 	 string `json:"pid"`
	PName 	 string `json:"pName"`
	User  	 string `json:"user"`
	Status   string `json:"status"`
	CPUUsage float64 `json:"cpuUsage"`
	MemUsage float64 `json:"memUsage"`
	Time	 string `json:"time"`
}
type Processes struct {
	ProcessList [10]Process `json:"processList"`
}
// updateProcInfo() 入口函数
func (processList *Processes) updateProcInfo(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("➤ \tProc: 采集线程开始 \t🔥")
	time.Sleep(25*time.Second)
	list:=setUsage(getPIDList(PROCESSES_PATH))
	processList.sort(list)
	processList.setUser()
	log.Println("➤ \tProc: ok \t\t✔")
}
// 按CPU使用率排序 并取头10个数据
func (processList *Processes)sort(list map[string]Process){
	len := len(list)
	pList :=make([]Process,len,len)
	n:=0
	for _, v := range list {
		pList[n]=v
		n++
	}
	for i := 0; i < len; i++ {
		for j := i + 1; j < len-i; j++ {
			if pList[j].CPUUsage< pList[j-1].CPUUsage {
				temp := pList[j]
				pList[j] = pList[j-1]
				pList[j-1] = temp
			}
		}
		if i > 9 {
			break
		}
	}
	top := pList[(len-10):]
	for index, v := range top {
		processList.ProcessList[9-index] = v
	}
}
// 根据uid从/etc/passwd 中获取username
func (processList *Processes) setUser() {
	userMap := make(map[string]string)
	file,err := os.Open(UID_USERNAME)
	if err != nil {
		log.Println("❌ \t错误:",err)
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line,_,err := reader.ReadLine()
		if err != nil{
			//log.Println("❌ \t错误:",err)
			break
		}
		userInfo := strings.Split(string(line),":")
		userMap[userInfo[2]] = userInfo[0]
	}
	for index, v := range processList.ProcessList {
		processList.ProcessList[index].User = userMap[v.User]
	}
}
// setUsage(PIDList map[string]int)
// 设置 进程信息
func setUsage(PIDList map[string]int)map[string]Process{
	infoList := make(map[string]Process)
	for k, _ := range PIDList {
		usageInfo := readStat(k)
		infoMap := readStatus(k)
		if usageInfo == nil || infoMap == nil{
			break
		}
		process := new(Process)
		process.Time = getTimeNow()
		process.PID = infoMap["Pid"]
		process.PName = infoMap["Name"]
		process.User = infoMap["Uid"]
		process.Status = infoMap["State"]
		process.CPUUsage = getProcCPUUsage(k)
		v,_ := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(infoMap["VmRSS"]," ",""),"kB",""),64)
		total := getMemTotal()
		process.MemUsage = v/total*100
		infoList[process.PID] = *process
	}
	return infoList
}

// getMemTotal()
// 从“/proc/meminfo”获取内存总容量
// 单位：KB
func getMemTotal()float64{
 	file,err:= os.Open(MEMINFO_PATH)
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
		key := strings.Split(string(line),":")
		if key[0] == "MemTotal" {
			value,_ := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(key[1]," ",""),"kB",""),64)
			return value
		}
	}
	return 0
}
// getProcCPUUsage(pid string)float64  函数:
// 获取线程CPU使用率 计算方法：
// 要计算特定进程的CPU使用率，您需要以下内容：
// /proc/uptime
// #1 系统正常运行时间（秒）
// /proc/[PID]/stat
// #14 utime- 用户代码中花费的CPU时间，以时钟周期计算
// #15 stime- 在内核代码中花费的CPU时间，以时钟周期计算
// #16 cutime- 等待孩子在用户代码中花费的 CPU时间（以时钟周期为单位）
// #17 cstime- 等待孩子在内核代码中花费的 CPU时间（以时钟周期为单位）
// #22 starttime- 过程开始的时间，以时钟滴答为单位
// 赫兹（系统的每秒时钟周期数）。树莓派默认为：100
// 在大多数情况下，getconf CLK_TCK可用于返回时钟周期数。
// 在sysconf(_SC_CLK_TCK)C函数调用也可以用来返回赫兹值。
// 计算
// 首先，我们确定该过程花费的总时间：
// total_time = utime + stime
// 我们还必须决定是否要包括”chil“流程的时间。如果我们这样做，那么我们将这些值添加到total_time：
// total_time = total_time + cutime + cstime
// 接下来，我们获取自进程启动以来的总耗用时间（以秒为单位）：
// seconds = uptime - (starttime / Hertz)
// 最后我们计算CPU使用百分比：
// cpu_usage = 100 * ((total_time / Hertz) / seconds
func getProcCPUUsage(pid string)float64{
	info := readStat(pid)
	if info == nil {
		return 0
	}
	var sum float64
	for i := 13; i<=16;i++  {
		value,_ := strconv.ParseFloat(info[i],64)
		sum += value
	}
	file,err := os.Open(UPTIME)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return 0
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	line,_,err := reader.ReadLine()
	uptime,_ := strconv.ParseFloat(strings.Split(string(line)," ")[0],64)
	startTime,_ := strconv.ParseFloat(info[21],64)
	seconds := uptime - ( startTime/ 100)
	cpuUsage := 100 *(sum / 100)/seconds
	return cpuUsage
}
// readStat(pid string)[]string 函数：
// 读/proc/(pid)/stat 进程状态信息 存入“info”数组
func readStat(pid string)[]string{
	path := PROCESSES_PATH+pid+"/stat"
	file,err := os.Open(path)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return nil
	}
	reader := bufio.NewReader(file)
	line,_,err := reader.ReadLine()
	info := strings.Split(string(line)," ")
	return info
}
// readStatus(pid string)map[string]string 函数：
// 读/proc/(pid)/status 并存入map集合
func readStatus(pid string)map[string]string{
	path := PROCESSES_PATH+pid+"/status"
	data := make(map[string]string)
	file,err := os.Open(path)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return nil
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	reg := regexp.MustCompile(`[/]*`)
	for{
		line,_,err := reader.ReadLine()
		if err != nil{
			//log.Println("❌ \t错误:",err)
			break
		}
		temp := strings.Split(string(line),":")
		if  temp[0] == "Uid"{
			data[temp[0]] = strings.Split(temp[1],"\t")[1]
		}else {
			data[temp[0]] = reg.ReplaceAllString(strings.ReplaceAll(temp[1],"\t",""), ``)
		}
	}
	return data
}
// getPIDList(path string) map[string]int 函数：
// 获取进程 生成Mmap集合
func getPIDList(path string) map[string]int{
	proc := make(map[string]int)
	fs,_ := ioutil.ReadDir(path)
	for _,file := range fs{
		if file.IsDir()&&isContains(file.Name()){
			proc[file.Name()] = 0
		}
	}
	return proc
}
// isContains(str string )bool 函数：
// 判断文件夹是否符合进程文件夹命名规则
// 符合返回：true
// 不符合返回：false
func isContains(str string )bool{
	isCon,_ := regexp.MatchString("[0-9]",str)
	return isCon
}


