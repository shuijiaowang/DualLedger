package main

import (
	"fmt"
	"log"

	"SService/config"
	"SService/db"
	"SService/routes"
)

func main() {
	config.InitConfig()

	if err := db.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	r := routes.SetupRouter()

	port := config.AppConfig.Server.Port
	if port <= 0 {
		port = 7789
	}
	addr := fmt.Sprintf(":%d", port)
	log.Printf("HTTP 监听 %s（来自 config server.port）\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
