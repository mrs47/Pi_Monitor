// Author： mrs47
// Since:   2019/12/03
// Version: 1.0
package info

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"
)

type Info struct {
	State     int `json:"state"`
	CPU       *CPUInfo `json:"cpu"`
	Hd        *HD `json:"hd"`
	Mem       *MEMInfo `json:"mem"`
	Net       *NetInfo `json:"net"`
	Processes *Processes `json:"processes"`
}

// 起始入口 golang 除基本类型使用时不需要自己手动初始化，
// -->其他类型在使用时切记 要手动初始化<--
func (info *Info)Update(){
	log.Println("➤\t☭ 数据采集启动 ☭")
	info.Hd  = new(HD)
	info.Mem = new(MEMInfo)
	info.CPU = new(CPUInfo)
	info.Net = new(NetInfo)
	info.Processes = new(Processes)
	wg := new(sync.WaitGroup)
	wg.Add(5)
	go info.Hd.updateHdInfo(wg)
	go info.Mem.updateMemInfo(wg)
	go info.CPU.updateCPUInfo(wg)
	go info.Net.updateNetInfo(wg)
	go info.Processes.updateProcInfo(wg)
	wg.Wait()
	info.State = 1
	log.Println("➤\t☺ 数据采集完成 ☺")
}
// 获取当前时间
func getTimeNow() string {
	return strings.Split(time.Now().String(),".")[0]
}
// 检查“error”是否为空
func checkErr(err error) bool {
	if err != nil{
		log.Println("❌ \t错误:",err)
		return false
	}
	return true
}
// infoToJson() 转换成Json字符串
func (info *Info)infoToJson() string{
	json,err:=json.Marshal(info)
	if !checkErr(err){
		log.Println("❌ \t错误:",err)
		panic(err)
	}
	return string(json)
}