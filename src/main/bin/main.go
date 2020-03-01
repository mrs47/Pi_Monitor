package main

import (
	"../module/controller"
	"../module/encrypt"
	"../module/info"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type DeviceData struct {
	Code       int     `json:"code"` //66:传递设备信息，50:上传GPIO信息 51：GPIO 控制信息 52: 初始化端口,98:心跳包，70：文件更新包括软件更新查询
	ProductKey string  `json:"productKey"`// 产品key
	DeviceKey  string  `json:"deviceKey"`// 设备key
	Random     string  `json:"random"`// 使用AES算法Token为密钥 Random为初始化向量
	Device     info.Info `json:"data"`
}

type FileData struct {
	Code       int     `json:"code"` //66:传递设备信息，50:上传GPIO信息 51：GPIO 控制信息 52: 初始化端口,98:心跳包，70：文件更新包括软件更新查询
	ProductKey string  `json:"productKey"`// 产品key
	DeviceKey  string  `json:"deviceKey"`// 设备key
	Random     string  `json:"random"`// 使用AES算法Token为密钥 Random为初始化向量
	File       controller.FileControlInfo `json:"data"`
}
type GPIOData struct {
	Code       int     `json:"code"` //66:传递设备信息，50:上传GPIO信息 51：GPIO 控制信息 52: 初始化端口,98:心跳包，70：文件更新包括软件更新查询
	ProductKey string  `json:"productKey"`// 产品key
	DeviceKey  string  `json:"deviceKey"`// 设备key
	Random     string  `json:"random"`// 使用AES算法Token为密钥 Random为初始化向量
	GpioList   controller.GPIOList `json:"data"`
}
type config struct {
	ProductKey  string
	DeviceKey string
	Token string
	Random string
	HttpSeverDeviceInfoAPI string
	HttpSeverFileAPI string
	HttpSeverGPIOAPI string
	getInfoTime string
	version string
	OS string
}

var conf *config

// 主配置文件
const config_path string = "/conf/config.json"

func main() {
	printInit()
	log.Println("➤\t⚡ 主程序启动 ⚡")
	log.Println("➤\t程序初始化 \t\t🔥")
	conf = new(config)
	err := conf.init()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("➤\t程序初始化完成   \t✔")
	service()
}

func service(){
	log.Println("➤\t开启业务   \t\t✔")
	wg := new(sync.WaitGroup)
	for {
		wg.Add(1)
		go doDeviceService(wg)
		wg.Add(1)
		go doGPIOService(wg)
		wg.Add(1)
		go doFileService(wg)
		wg.Wait()

	}
}

// 上传设备信息到服务器
func doDeviceService(wg *sync.WaitGroup){
	log.Println("➤ \tdoDeviceService \t↩")
	defer wg.Done()
	var data = new(DeviceData)
	var deviceInfo = new(info.Info)
	deviceInfo.Update()

	data.Device = *deviceInfo
	data.ProductKey = conf.ProductKey
	data.DeviceKey = conf.DeviceKey
	data.Code = 66
	defer catchError()
	err := data.uploadByDeviceInfo()
	if err != nil {
		log.Println("❌ 错误: 发送设备信息错误",err)
	}
	log.Println("➤ \tdoDeviceService \t↪")
}

// 读取配置文件
func (conf *config)init() error{
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("❌ \t错误:",err)
		return err
	}
	file,err := os.Open(dir+"/.."+config_path)
	if err != nil {
		log.Fatal("❌ \t错误:",err)
		return err
	}
	defer file.Close()
	buf,err := ioutil.ReadAll(bufio.NewReader(file))
	if err != nil {
		log.Fatal("❌ \t错误:",err)
		return err
	}
	err = json.Unmarshal(buf,&conf)
	if err != nil {
		log.Fatal("❌ \t错误:",err)
		return err
	}
	return nil
}

// data.Random 作为使用随机数产生的初始化向量才能达到语义安全（散列函数与消息验证码也有相同要求），
// 并让攻击者难以对同一把密钥的密文进行破解 初始化向量的值依密码算法而不同。最基本的要求是“唯一性”，
// 也就是说同一把密钥不重复使用同一个初始化向量。这个特性无论在区块加密或流加密中都非常重要。
// 初始为”0000000000000000“ 16位0
// 每次http返回下一次的初始化向量
// 错误处理待优化
func (data DeviceData)uploadByDeviceInfo() error{
	log.Println("📡\t",conf.HttpSeverDeviceInfoAPI)
	temp:=[]byte(data.infoToJson())
	infoAES,err:=encrypt.AesEncryptSimple(temp,conf.Token,conf.Random)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	info := make(url.Values)
	info["info"] = []string{infoAES}
	info["deviceKey"] = []string{conf.DeviceKey}
	info["productKey"] = []string{conf.ProductKey}
	resp,err := http.PostForm(conf.HttpSeverDeviceInfoAPI,info)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	defer resp.Body.Close()
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	responseInfo,err :=encrypt.AesDecryptSimple(string(body),conf.Token,conf.Random)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	log.Println("📃 \tresponseDeviceInfo: ",string(responseInfo))
	err = json.Unmarshal(responseInfo,&data)
	// 更新初始化向量的配置
	conf.Random = data.Random
	return nil
}


func doGPIOService(wg *sync.WaitGroup){
	log.Println("➤ \tdoGPIOService \t\t↩")
	defer wg.Done()
	for  {
		data := new(GPIOData)
		data.Code = 50
		data.ProductKey = conf.ProductKey
		data.DeviceKey = conf.DeviceKey
		gpio :=new(controller.GPIOList)
		gpio.SelectAll()
		data.GpioList=*gpio
		defer catchError()
		data.postByGPIO()
		if data.Code == 51{
			gpio = &data.GpioList
			for _, value := range gpio.List {
				value.GPIOControl()

			}
           continue
		}else if data.Code == 52 {
			gpio = &data.GpioList
			for _, v := range gpio.List {
				v.Init()
				break
			}
			continue
		}else {
			break
		}
	}
	log.Println("➤ \tdoGPIOService \t\t↪")
}

// GPIO
// 发送 I/O 口状态信息
// 服务器端只保留最新一份
// 如需 控制返回 控制信息
func (data *GPIOData)postByGPIO()error{
	log.Println("📡\t",conf.HttpSeverGPIOAPI)
	infoAES,err:=encrypt.AesEncryptSimple([]byte(data.infoToJson()),conf.Token,conf.Random)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	info := make(url.Values)
	info["info"] = []string{infoAES}
	info["deviceKey"] = []string{conf.DeviceKey}
	info["productKey"] = []string{conf.ProductKey}
	resp,err := http.PostForm(conf.HttpSeverGPIOAPI,info)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	defer resp.Body.Close()
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	responseInfo,err:=encrypt.AesDecryptSimple(string(body),conf.Token,conf.Random)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	log.Println("📃 \tresponseGPIOInfo: ",string(responseInfo))
	err = json.Unmarshal(responseInfo,&data)
	return nil
}
// ！！！！！！！！！！！！！！！！！！！！！！！！！
// 此业务逻辑有巨大安全漏洞，巨大，非常巨大！！！！
// 说明：
// 文件传输使用FTP协议，且全系统只分配一个账号
// 服务器端传来用于FTP文件服务器认证的账号密码，
// 如果被恶意利用将造成，全服务器文件泄漏。
// 备选方案：
// 1.改为Http协议，服务器端做权限判断，文件映射转发。
// 2.为每个用户分配独立的只拥有读权限的FTP账号
// ！！！！！！！！！！！！！！！！！！！！！！！！！
func doFileService(wg *sync.WaitGroup){
	log.Println("➤ \tdoFileService \t\t↩")
	defer wg.Done()
	var data = new(FileData)
	data.Code = 70
	data.DeviceKey = conf.DeviceKey
	data.ProductKey = conf.ProductKey
	defer catchError()
	data.postByFile()
	if data.Code == 71 {
		data.File.FileControl()
	}
	log.Println("➤ \tdoFileService \t\t↪")
}

func (data *FileData)postByFile()error{
	log.Println("📡\t",conf.HttpSeverFileAPI)
	info := make(url.Values)
	info["deviceKey"] = []string{conf.DeviceKey}
	info["productKey"] = []string{conf.ProductKey}
	resp,err := http.PostForm(conf.HttpSeverFileAPI,info)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	defer resp.Body.Close()
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	responseInfo,err:=encrypt.AesDecryptSimple(string(body),conf.Token,conf.Random)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	log.Println("📃 \tresponseFileInfo: ",string(responseInfo))
	err = json.Unmarshal(responseInfo,&data)
	return nil
}
// infoToJson() 转换成Json字符串
func (data DeviceData)infoToJson() string{
	json,err := json.Marshal(data)
	if err != nil{
		log.Println("❌ \t错误:",err)
		return ""
	}
	return string(json)
}
// infoToJson() 转换成Json字符串
func (data GPIOData)infoToJson() string{
	json,err := json.Marshal(data)
	if err != nil{
		log.Println("❌ \t错误:",err)
		return ""
	}
	return string(json)
}
// infoToJson() 转换成Json字符串
func (data FileData)infoToJson() string{
	json,err := json.Marshal(data)
	if err != nil{
		log.Println("❌ \t错误:",err)
		return ""
	}
	return string(json)
}

func catchError() {
	if err := recover(); err != nil {
		log.Println("❌ \t错误:", err)
	}
}

func printInit(){
	fmt.Println(" ......................。。。。......................")
	fmt.Println("   Author: Mrs47                                    ")
	fmt.Println("   OS: Linux on Raspberry(树莓派)                    ")
	fmt.Println("   Version: 1.0                                     ")
	fmt.Println("   主配置文件: ./conf/config.json                    ")
	fmt.Println("                                                    ")
	fmt.Println("   Copyright © 2019 Mrs47. All rights reserved.     ")
	fmt.Println(" ......................阿弥陀佛......................")
	fmt.Println("                      _oo0oo_                       ")
	fmt.Println("                     o8888888o                      ")
	fmt.Println("                     88\" . \"88                    ")
	fmt.Println("                     (| -_- |)                      ")
	fmt.Println("                     0\\  =  /0                     ")
	fmt.Println("                   ___/‘---’\\___                   ")
	fmt.Println("                  .' \\|       |/ '.                ")
	fmt.Println("                 / \\\\|||  :  |||// \\             ")
	fmt.Println("                / _||||| -卍-|||||_ \\              ")
	fmt.Println("               |   | \\\\\\  -  /// |   |           ")
	fmt.Println("               | \\_|  ''\\---/''  |_/ |            ")
	fmt.Println("               \\  .-\\__  '-'  ___/-. /            ")
	fmt.Println("             ___'. .'  /--.--\\  '. .'___           ")
	fmt.Println("         .\"\" ‘<  ‘.___\\_<|>_/___.’>’ \"\".       ")
	fmt.Println("       | | :  ‘- \\‘.;‘\\ _ /’;.’/ - ’ : | |        ")
	fmt.Println("         \\  \\ ‘_.   \\_ __\\ /__ _/   .-’ /  /    ")
	fmt.Println("    =====‘-.____‘.___ \\_____/___.-’___.-’=====     ")
	fmt.Println("                       ‘=---=’                      ")
	fmt.Println("                                                    ")
	fmt.Println("..................佛祖保佑 ,永无BUG..................")
}