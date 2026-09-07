package main

import (
	"log"

	"simplechat/server/config"
	"simplechat/server/dao"
	"simplechat/server/router"
)

func main() {
	cfg := config.Load()

	dao.InitDB(&cfg.MySQL)
	dao.InitRedis(&cfg.Redis)
	dao.InitMinio(&cfg.Minio)

	r := router.Setup(cfg)
	log.Printf("简聊服务端启动于 :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}