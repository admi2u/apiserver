package main

import (
	"apiserver/config"
	"apiserver/router"
	"errors"
	"flag"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lexkong/log"
	"github.com/spf13/viper"
)

// pingServer pings the http server to make sure the router is working.
func pingServer() error {
	for i := 0; i < viper.GetInt("max_ping_count"); i++ {
		// Ping the server by sending a GET request to `/health`.
		resp, err := http.Get(viper.GetString("url") + "/sd/health")
		if err == nil && resp.StatusCode == 200 {
			return nil
		}

		// Sleep for a second to continue the next ping.
		log.Info("Waiting for the router, retry in 1 second.")
		time.Sleep(time.Second)
	}
	return errors.New("cannot connect to the router")
}

var (
	cfg string
)

func init() {
	flag.StringVar(&cfg, "c", "", "apiserver config file path.")
}

func main() {
	flag.Parse()
	// init config
	if err := config.Init(cfg); err != nil {
		panic(err)
	}

	gin.SetMode(viper.GetString("runmode"))
	// Create the Gin engine.
	g := gin.New()

	// gin middlewares
	middlewares := []gin.HandlerFunc{}

	// routes
	router.Load(g, middlewares...)

	// Ping the server to make sure the router is working.
	go func() {
		if err := pingServer(); err != nil {
			log.Fatal("The router has no response, or it might took too long to start up.", err)
		}
		log.Info("The router has been deployed successfully.")
	}()

	log.Infof("Start to listening the incoming requests on http address: %s", viper.GetString("addr"))
	log.Info(http.ListenAndServe(viper.GetString("addr"), g).Error())
}