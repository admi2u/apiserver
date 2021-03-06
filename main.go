package main

import (
	"apiserver/config"
	"apiserver/model"
	v "apiserver/pkg/version"
	"apiserver/router"
	"apiserver/router/middleware"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
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
	cfg     string
	version bool
)

func init() {
	flag.StringVar(&cfg, "c", "", "apiserver config file path.")
	flag.BoolVar(&version, "v", false, "show version info")
}

// @title Apiserver Example API
// @version 1.0
// @description apiserver demo
// @contact.name admi2u
// @contact.url http://www.swagger.io/support
// @contact.email admi2u@qq.com
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name bearer
// @host localhost:8080
// @BathPath /
func main() {
	flag.Parse()
	if version {
		versionInfo := v.Get()
		data, err := json.MarshalIndent(&versionInfo, "", "")
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	// init config
	if err := config.Init(cfg); err != nil {
		panic(err)
	}

	// init db
	model.DB.Init()
	defer model.DB.Close()

	gin.SetMode(viper.GetString("runmode"))
	// Create the Gin engine.
	g := gin.New()

	// routes
	router.Load(
		g,

		// Middlwares.
		middleware.Logging(),
		middleware.RequestId(),
	)

	// Ping the server to make sure the router is working.
	go func() {
		if err := pingServer(); err != nil {
			log.Fatal("The router has no response, or it might took too long to start up.", err)
		}
		log.Info("The router has been deployed successfully.")
	}()

	// ???????????????https??????????????????https??????
	cert := viper.GetString("tls.cert")
	key := viper.GetString("tls.key")
	if cert != "" && key != "" {
		go func() {
			log.Infof("Start to listening the incoming requests on https address: %s", viper.GetString("tls.addr"))
			log.Info(http.ListenAndServeTLS(viper.GetString("tls.addr"), cert, key, g).Error())
		}()
	}

	log.Infof("Start to listening the incoming requests on http address: %s", viper.GetString("addr"))
	log.Info(http.ListenAndServe(viper.GetString("addr"), g).Error())
}
