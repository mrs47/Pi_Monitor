// Author： mrs47
// Since:   2019/11/37
// Version: 1.0
// Describe:用于获取Linux 主存设备信息
package info

import (
	"log"
	"sync"
	"syscall"
)
// 用于封装主存信息的结构体
type HD struct {
	Total  uint64 `json:"total"`
	Used   uint64 `json:"used"`
	Free   uint64 `json:"free"`
	Time   string `json:"time"`
}
// (hd *HD)updateHdInfo() 函数:
// 入口函数
func (hd *HD) updateHdInfo(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("➤ \thd: 采集线程开始 \t🔥")
	total,free,used := diskUsage("/")
	hd.Total = total
	hd.Used = used
	hd.Free = free
	hd.Time = getTimeNow()
	log.Println("➤ \thd: ok \t\t\t✔")
}
// diskUsage(path string) (Total uint64,Free uint64,Used uint64) 函数：
// Linux系统函数调用，获取指定路径的文件/文件夹信息
// Total: 总容量大小
// Free:  未使用大小
func diskUsage(path string) (Total uint64,Free uint64,Used uint64){
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return
	}
	Total = fs.Blocks * uint64(fs.Bsize)
	Free  = fs.Bfree * uint64(fs.Bsize)
	Used  = Total - Free
	return
}