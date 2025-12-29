package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"yinyue/config"
	"yinyue/controllers"
	"yinyue/routes"
	"yinyue/storage"
)

// DataDir 全局数据目录
var DataDir string

func initLogger() *os.File {
	// 获取数据目录
	DataDir = "./data"

	// 确保目录存在
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		log.Printf("创建数据目录失败: %v", err)
		return nil
	}

	// 日志文件路径
	logPath := filepath.Join(DataDir, "app.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("打开日志文件失败: %v", err)
		return nil
	}

	// 同时输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("日志文件: %s", logPath)
	return logFile
}

func main() {
	// 初始化日志
	logFile := initLogger()
	if logFile != nil {
		defer logFile.Close()
	}

	log.Println("========== TuneHub Music 启动 ==========")
	log.Printf("数据目录: %s", DataDir)

	// 加载配置
	log.Println("正在加载配置...")
	config.Init()
	log.Println("配置加载完成")

	// 初始化存储（传入数据目录）
	log.Println("正在初始化存储...")
	if err := storage.Init(DataDir); err != nil {
		log.Printf("初始化存储失败: %v", err)
	} else {
		log.Println("存储初始化完成")
	}

	// 初始化音乐库（传入数据目录）
	log.Println("正在初始化音乐库...")
	controllers.InitLibrary(DataDir)
	log.Println("音乐库初始化完成")

	// 初始化路由（传入嵌入的静态文件）
	log.Println("正在初始化路由...")
	r := routes.SetupRouter(StaticFS, TemplatesFS)
	log.Println("路由初始化完成")

	// 启动服务
	log.Println("正在启动 HTTP 服务，端口: 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
