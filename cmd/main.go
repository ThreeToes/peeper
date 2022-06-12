package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"peeper/internal/config"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type opts struct {
	ConfigFile *string
	LogFormat  *string
}

func (o *opts) verify() error {
	if o.ConfigFile == nil {
		return fmt.Errorf("config file must be set")
	}
	return nil
}

func main() {
	opts, err := parseOpts()
	if err != nil {
		logrus.Fatalf("error parsing command line options: %v", err)
	}

	if err := opts.verify(); err != nil {
		logrus.Fatalf("error parsing command line options: %v", err)
	}

	if *opts.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	var conf config.AppOptions

	_, err = toml.DecodeFile(*opts.ConfigFile, &conf)

	if err != nil {
		logrus.Fatalf("could not decode config file: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)

	svr := gin.New()
	svr.Use(gin.Recovery())

	for k, v := range conf.Endpoints {
		logrus.Infof("Registering endpoint %s", k)
		svr.Handle(v.LocalMethod, v.LocalPath, func(ctx *gin.Context) {
			forwardedReq, err := http.NewRequest(v.RemoteMethod, v.RemotePath, ctx.Request.Body)

			client := http.DefaultClient

			resp, err := client.Do(forwardedReq)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "")
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "")
				return
			}
			ctx.String(resp.StatusCode, string(body))
		})
	}

	logrus.Infof("binding to %s:%d", conf.Network.BindInterface, conf.Network.BindPort)
	if err := svr.Run(fmt.Sprintf("%s:%d", conf.Network.BindInterface, conf.Network.BindPort)); err != nil {
		logrus.Infof("error while serving: %v", err)
	}
}

func parseOpts() (*opts, error) {
	var options opts
	options.ConfigFile = flag.String("config", "", "Path to TOML config file")
	options.LogFormat = flag.String("logformat", "text", "The log format to use. Supported formats: json, text")

	flag.Parse()

	return &options, nil
}
