package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	r := gin.Default()

	r.StaticFile("/", viper.GetString("serve.indexFile"))
	r.Static("/dev", viper.GetString("serve.staticFolder"))

	_ = r.Run(viper.GetString("serve.serverAddress"))
}
