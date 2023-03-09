package main

import (
	"WowjoyProject/WowjoyCallRadiant/global"
	"WowjoyProject/WowjoyCallRadiant/pkg/object"
	"WowjoyProject/WowjoyCallRadiant/pkg/workpattern"
	"net"
	"os"
	"strings"
)

// @title 调用Radiant程序
// @version 1.0.0.1
// @description 调用Radiant看图软件
// @termsOfService https://github.com/jianghuxiaoloulou/CallRadiant.git
func main() {
	global.RadiantParam = make([]string, 0)
	global.Logger.Info("*******开始运行调用Radiant程序********")

	global.ObjectDataChan = make(chan global.RadiantData)

	// 注册工作池，传入任务
	// 参数1 初始化worker(工人)设置最大线程数
	wokerPool := workpattern.NewWorkerPool(global.GeneralSetting.MaxThreads)
	// 有任务就去做，没有就阻塞，任务做不过来也阻塞
	wokerPool.Run()
	// 处理任务
	go func() {
		for {
			select {
			case data := <-global.ObjectDataChan:
				sc := &Dosomething{key: data}
				wokerPool.JobQueue <- sc
			}
		}
	}()

	// 1、net.ListenUDP() 监听UDP服务
	// 2、net.UDPConn.ReadFromUDP()  循环读取数据
	// 3、net.UDPConn.WriteToUDP() 写数据
	address := "127.0.0.1" + ":" + global.ServerSetting.UDPPort

	// ResolveUDPAddr函数构造服务端的地址信息
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		global.Logger.Error(err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		global.Logger.Error(err)
		os.Exit(1)
	}

	defer conn.Close()

	for {
		// Here must use make and give the lenth of buffer
		data := make([]byte, 1024)
		_, _, err := conn.ReadFromUDP(data)
		if err != nil {
			global.Logger.Error(err)
			continue
		}
		strData := string(data)
		strData = strings.Trim(strData, "\x00")
		global.Logger.Info("Received Data :", strData)
		object.ParseUDPData(strData)
	}
}

type Dosomething struct {
	key global.RadiantData
}

func (d *Dosomething) Do() {
	global.Logger.Info("正在处理的数据是：", d.key)
	//处理封装对象
	obj := object.NewObject(d.key)
	global.Logger.Debug(obj)
	switch d.key.ParamType {
	case global.IMAGE_REVIEW:
		// 影像查看
		obj.Image_Review()
	case global.APPEND_IMAGE_VIEW:
		// 追加影像查看
		obj.Append_Image_View()
	case global.FILM_PRINT:
		// 胶片直接打印
		obj.Film_Print()
	case global.APPEND_FILM_PRINT:
		// 胶片追加打印
		obj.Append_Film_Print()
	}
}
