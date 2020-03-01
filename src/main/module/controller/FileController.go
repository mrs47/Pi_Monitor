// Author： mrs47
// Since:   2019/12/04
// Version: 1.0
// Describe:用于处理文件下载
package controller

import (
	"encoding/json"
	"github.com/jlaffaye/ftp"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)
type FileControlInfo struct {
	Tag      int `json:"tag"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Uri      string `json:"uri"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}
type SoftInfo struct {
	Name     string
	Version  string
	Time     string
}
var prefix = ""
func (fileControlInfo *FileControlInfo)FileControl(){
	log.Println("➤ \tFileControl: ",fileControlInfo.Tag)
	if fileControlInfo.Tag == 1 {
		err:=fileControlInfo.updateSoft()
		if err != nil {
			log.Println("❌ \t软件更新错误:",err)
			return
		}
		fileControlInfo.updateSoftInfo()
		log.Println("➤ \t软件更新成功：",fileControlInfo.Name)
	}else if fileControlInfo.Tag == 0 {
		err:=fileControlInfo.downloadFile()
		if err != nil {
			log.Println("❌ \t文件下载错误:",err)
			return
		}
		log.Println("➤ \t文件下载成功：",fileControlInfo.Name)
	}else {
		log.Println("➤ \t文件下载参数有误Tag：",fileControlInfo.Tag)
	}
}

func (fileControlInfo *FileControlInfo)downloadFile() error {
	prefix = "../files/"
	err:=fileControlInfo.download()
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	return nil
}
func (fileControlInfo FileControlInfo)updateSoft() error{
	prefix = "../program/"

	err:=fileControlInfo.download()
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	return nil
}
func (fileControlInfo *FileControlInfo)download() error {

	cil,err := fileControlInfo.login()
	if err != nil {
		log.Println("❌ \t登陆失败:",err)
		return err
	}
	err = fileControlInfo.getFile(cil)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	fileControlInfo.quit(cil)
	return nil
}
func (fileControlInfo *FileControlInfo)login() (*ftp.ServerConn,error){
	cli, err := ftp.Dial(fileControlInfo.Host+":"+fileControlInfo.Port, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Println("❌ \t错误:",err)
		return cli,err
	}
	err = cli.Login(fileControlInfo.UserName, fileControlInfo.Password)
	if err != nil {
		log.Println("❌ \t密码账号认证错误:",err)
		return cli,err
	}
	return cli,err
}
// 从FTP服务器下载文件
func (fileControlInfo FileControlInfo)getFile(cli *ftp.ServerConn)error{
	// 分离Uri，把路径和文件名分开
	uri := strings.Split(fileControlInfo.Uri,"/")
	// ftp服务器上的文件名
	var filename string
	for _, v := range uri[len(uri)-1:] {
		filename+=v
	}
	// 清空并重新拼接路径
	fileControlInfo.Uri = ""
	for _, v := range uri[:len(uri)-1] {
		fileControlInfo.Uri += v + "/"
	}
	// 选择目录
	err := cli.ChangeDir(fileControlInfo.Uri)
	if err != nil {
		log.Println("❌ \t目录选择错误:",err)
		return err
	}
	// 读取文件内容，默认开启二进制模式
	reader, err := cli.Retr(filename)
	if err != nil {
		log.Println("❌ \t文件内容读取错误:",err)
		return err
	}

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	// 写入本地文件夹
	str :=strings.Split(filename,".")
	suffix :=str[len(str)-1:]
	file,err := os.Create(prefix+fileControlInfo.Name+"."+suffix[0])
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	defer file.Close()

	_,err = file.Write(buf)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return err
	}
	return nil
}
func (fileControlInfo *FileControlInfo)quit(cli *ftp.ServerConn) {
	if err := cli.Quit(); err != nil {
		log.Println("❌ \t错误:",err)
	}
}
// 当更新软件完成时更新配置文件
func (fileControlInfo *FileControlInfo) updateSoftInfo() {
	file,err:=os.OpenFile("./conf/userProgram.conf",os.O_WRONLY|os.O_CREATE|os.O_TRUNC,0644)
	if err != nil {
		log.Println("❌ \t错误:",err)
		return
	}
	defer file.Close()
	softInfo:=new(SoftInfo)
	softInfo.Name=fileControlInfo.Name
	softInfo.Version=fileControlInfo.Version
	softInfo.Time=getTimeNow()
	file.WriteString(softInfo.infoToJson()+"\n")
}
// infoToJson() 转换成Json字符串
func (softInfo *SoftInfo)infoToJson() string{
	json,err:=json.Marshal(softInfo)
	if err != nil{
		log.Println("❌ \t错误:",err)
		return "error: 1"
	}
	return string(json)
}
// 获取当前时间
func getTimeNow() string {
	return strings.Split(time.Now().String(),".")[0]
}