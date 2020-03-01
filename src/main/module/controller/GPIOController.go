package controller

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type GPIO struct {
	Pin       int `json:"pin"`
	InOrOut   int `json:"inOrOut"`
	HighOrLow int `json:"highOrLow"`
}

type GPIOList struct {
	List [17]GPIO `json:"list"`
}
// 限定可控制口
var CAN_CONTROL = [17]int{4,17,27,22,5,6,13,19,26,18,23,24,25,12,16,20,21}
// 互斥锁
var mutex sync.Mutex

// 操作指令集
const   (
	IN   string = "in"
	OUT  string = "out"
	HIGH string = "1"
	LOW  string = "0"
)

func (gpio *GPIO)GPIOControl(){
	gpio.doControl()
}
// 返回当前已被内核调出的I/O口
func (gpioList *GPIOList)SelectAll(){
	log.Println("➤ \tGPIO: 采集线程开始 \t🔥")
	count := 0
	activity := new([17]GPIO)
	for _, v := range CAN_CONTROL {
		// 判断文件是否存在
		b := isExists("/sys/class/gpio/gpio"+strconv.Itoa(v))
		if !b {
			continue
		}
		gpio := new(GPIO)
		gpio.Pin = v
		file,err := os.Open("/sys/class/gpio/gpio"+strconv.Itoa(v)+"/direction")
		if err != nil {
			log.Println("❌ \t错误:",err)
			continue
		}
		reader := bufio.NewReader(file)
		line,_,err := reader.ReadLine()
		if err != nil {
			log.Println("❌ \t错误:",err)
			continue
		}
		if string(line) == OUT {
			gpio.InOrOut = 1
		}else if string(line) == IN {
			gpio.InOrOut = 0
		}
		file.Close()

		file,err = os.Open("/sys/class/gpio/gpio"+strconv.Itoa(v)+"/value")
		if err != nil {
			log.Println("❌ \t错误:",err)
			continue
		}

		reader = bufio.NewReader(file)
		line,_,err = reader.ReadLine()
		if err != nil {
			log.Println("❌ \t错误:",err)
			continue
		}
		value := string(line)
		highOrLow,_ := strconv.Atoi(value)
		gpio.HighOrLow = highOrLow
		file.Close()
		activity[count] = *gpio
		count++
	}
	gpioList.List = *activity
	log.Println("➤ \tGPIO: ok \t\t✔")
}

// 对gpio限定口进行初始化
func (gpio *GPIO)Init(){

	for _, v := range CAN_CONTROL {
		// 通知系统内核调出该口
		file,err := os.OpenFile("/sys/class/gpio/export",os.O_WRONLY,0666)
		if err != nil {
			log.Println("❌ \t错误:",err)
			continue
		}
		file.WriteString(strconv.Itoa(v))
		file.Close()
		time.Sleep(50*time.Millisecond)
		// 对I/O口统一设置输出
		// direction <- out : 输出
		// direction <- in  : 输入
		file,err = os.OpenFile("/sys/class/gpio/gpio"+strconv.Itoa(v)+"/direction",os.O_WRONLY,0666)
		if err != nil {
			log.Println("❌ \t错误:",err)
			continue
		}
		if gpio.InOrOut == 1  {
			file.WriteString(OUT)
		}else if gpio.InOrOut == 0 {
			file.WriteString(IN)
		}
		file.Close()
		time.Sleep(20*time.Millisecond)
		// 设高低电平
		// 0: 低电平
		// 1: 高电平
		// 在这统一设低电平
		file,err = os.OpenFile("/sys/class/gpio/gpio"+strconv.Itoa(v)+"/value",os.O_WRONLY,0666)
		if err != nil {
			log.Println("❌ \t错误:",err)
			continue
		}
		if gpio.HighOrLow == 1 {
			file.WriteString(HIGH)
		}else if gpio.HighOrLow == 0 {
			file.WriteString(LOW)
		}
		file.Close()
	}
}
// 根据List数据设置I/O口状态
func (gpio *GPIO) doControl(){
	mutex.Lock()
	defer mutex.Unlock()
	// 判断文件是否存在
	b:=isExists("/sys/class/gpio/gpio"+strconv.Itoa(gpio.Pin))
	if !b {
		file,err := os.OpenFile("/sys/class/gpio/export",os.O_WRONLY,0666)
		if err != nil {
			log.Println("❌ \t错误:",err)
			return
		}
		file.WriteString(strconv.Itoa(gpio.Pin))
		file.Close()
	}

	file,err := os.OpenFile("/sys/class/gpio/gpio"+strconv.Itoa(gpio.Pin)+"/direction",os.O_WRONLY,0666)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return
	}
	if gpio.InOrOut == 1 {
		file.WriteString( OUT )
		//file.Close()
		//return
	}else if gpio.InOrOut == 0{
		file.WriteString( IN )
		//file.Close()
		//return
	}else {
		log.Println("gpio.InOrOut 参数错误："+strconv.Itoa(gpio.InOrOut))
	}
	file.Close()

	file,err = os.OpenFile("/sys/class/gpio/gpio"+strconv.Itoa(gpio.Pin)+"/value",os.O_WRONLY,0666)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return
	}
	if gpio.HighOrLow == 1 {
		file.WriteString(HIGH)
	}else if gpio.HighOrLow == 0 {
		file.WriteString(LOW)
	}else {
		log.Println("gpio.HighOrLow 参数错误："+strconv.Itoa(gpio.HighOrLow))
	}
	file.Close()
}
// 判断文件是否存在
func isExists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}