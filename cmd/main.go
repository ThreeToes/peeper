package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/threetoes/peeper/internal/config"
	"github.com/threetoes/peeper/internal/service"
	"sort"
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

type endpointSorter []*config.Endpoint

func (e endpointSorter) Len() int {
	return len(e)
}

func (e endpointSorter) Less(i, j int) bool {
	return len(e[i].LocalPath) > len(e[j].LocalPath)
}

func (e endpointSorter) Swap(i, j int) {
	tmp := e[i]
	e[i] = e[j]
	e[j] = tmp
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

	sorter := endpointSorter{}

	for _, v := range conf.Endpoints {
		sorter = append(sorter, v)
	}

	sort.Sort(sorter)

	for _, e := range sorter {
		logrus.Infof("Mapping local endpoint %s to remote endpoint %s", e.LocalPath, e.RemotePath)
		svr.RegisterEndpoint(e)
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
