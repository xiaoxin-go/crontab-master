package main

import (
	"crontab/worker"
	"flag"
	"fmt"
	"runtime"
)

var (
	confFile string
)

// 解析命令行参数
func initArgs(){
	flag.StringVar(&confFile, "config", "./worker.json", "指定worker.json")
	flag.Parse()
}

// 初始化线程数量
func initEnv(){
	runtime.GOMAXPROCS(runtime.NumCPU())
}


func main(){
	var (
		err error
	)
	// 1. 初始化命令行参数
	initArgs()

	// 2. 初始化线程数量
	initEnv()

	// 3. 初始化配置文件
	if err = worker.InitConfig(confFile); err != nil{
		fmt.Println("加载配置文件异常：", err.Error())
		return
	}

	// 4. 注册服务
	if err = worker.InitRegister(); err != nil{
		fmt.Println("节点注册异常：", err.Error())
		return
	}

	// 启动日志协程
	if err = worker.InitLogSink(); err != nil{
		fmt.Println("日志协程启动异常： ", err.Error())
		return
	}

	// 初始化执行器
	worker.InitExecutor()

	// 启动调度协程
	worker.InitScheduler()
}
