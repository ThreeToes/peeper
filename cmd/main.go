package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"peeper/internal/config"
	"peeper/internal/service"
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

	svr := service.New(fmt.Sprintf("%s:%d", conf.Network.BindInterface, conf.Network.BindPort))

	for _, v := range conf.Endpoints {
		logrus.Infof("Mapping local endpoint %s to remote endpoint %s", v.LocalPath, v.RemotePath)
		svr.RegisterEndpoint(v)
	}

	logrus.Infof("binding to %s:%d", conf.Network.BindInterface, conf.Network.BindPort)
	if err := svr.Start(); err != nil {
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
