package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"peeper/internal/config"

	"github.com/BurntSushi/toml"
)

const downstream = "https://cat-fact.herokuapp.com/facts"

type opts struct {
	ConfigFile *string
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
		log.Fatalf("error parsing command line options: %v", err)
	}

	if err := opts.verify(); err != nil {
		log.Fatalf("error parsing command line options: %v", err)
	}

	var conf config.AppOptions

	_, err = toml.DecodeFile(*opts.ConfigFile, &conf)

	if err != nil {
		log.Fatalf("could not decode config file: %v", err)
	}

	mux := http.NewServeMux()

	for k, v := range conf.Endpoints {
		log.Printf("Registering endpoint %s", k)
		mux.HandleFunc(v.LocalPath, func(rw http.ResponseWriter, req *http.Request) {
			forwardedReq, err := http.NewRequest(v.RemoteMethod, v.RemotePath, req.Body)

			client := http.DefaultClient

			resp, err := client.Do(forwardedReq)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}

			rw.WriteHeader(resp.StatusCode)
			rw.Write(body)
		})
	}
	log.Println("binding to port 9090")
	if err := http.ListenAndServe(":9090", mux); err != nil {
		log.Printf("error while serving: %v", err)
	}
}

func parseOpts() (*opts, error) {
	var options opts
	options.ConfigFile = flag.String("config", "", "Path to TOML config file")

	flag.Parse()

	return &options, nil
}
