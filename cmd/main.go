package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

const downstream = "https://cat-fact.herokuapp.com/facts"

func main() {
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		resp, err := http.Get(downstream)
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
	log.Println("binding to port 9090")
	if err := http.ListenAndServe(":9090", http.DefaultServeMux); err != nil {
		log.Printf("error while serving: %v", err)
	}
}
