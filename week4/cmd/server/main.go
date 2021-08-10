package main

import (
	"geek-time/week4/internal/global"
	"geek-time/week4/internal/setting"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func init() {
	err := setupSetting()
	if err != nil {
		log.Fatalf("init.setupSetting err: %v", err)
	}
}

func main() {
	gin.SetMode(global.ServerSetting.RunMode)
	//router := routers.NewRouter()
	router := InjectRouters()
	s := &http.Server{
		Addr:           ":" + global.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    global.ServerSetting.ReadTimeout,
		WriteTimeout:   global.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	_ = s.ListenAndServe()
}

func setupSetting() error {
	settings, err := setting.NewSetting()
	if err != nil {
		return err
	}

	err = settings.ReadSection("Server", &global.ServerSetting)
	if err != nil {
		return err
	}

	err = settings.ReadSection("App", &global.AppSetting)
	if err != nil {
		return err
	}

	err = settings.ReadSection("Database", &global.DatabaseSetting)
	if err != nil {
		return err
	}

	global.ServerSetting.ReadTimeout *= time.Second
	global.ServerSetting.WriteTimeout *= time.Second

	log.Printf("global.ServerSetting: %v", global.ServerSetting)
	log.Printf("global.AppSetting: %v", global.AppSetting)
	log.Printf("global.DatabaseSetting: %v", global.DatabaseSetting)

	return nil
}
