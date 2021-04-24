package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"faxer.heav.fr/server/api"
)

type config struct {
	Port             string
	SslCrt           string
	SslKey           string
	GoogleAccessId   string
	GooglePrivateKey string
	GoogleBucket     string
}

var Config config

func main() {
	rand.Seed(time.Now().Unix())
	rawConfig, err := ioutil.ReadFile("settings.json")
	json.Unmarshal(rawConfig, &Config)

	if err != nil {
		println("Could not read config:", err.Error())
	}

	srv, err := api.NewServer(Config.GoogleAccessId, Config.GooglePrivateKey, Config.GoogleBucket)
	if err != nil {
		println("Could not create server:", err.Error())
	}

	if len(Config.Port) == 0 {
		Config.Port = "1234"
	}

	if len(Config.SslCrt) > 0 && len(Config.SslKey) > 0 {
		println("Starting the server with port", Config.Port, "on https")
		err := http.ListenAndServeTLS("0.0.0.0:"+Config.Port, Config.SslCrt, Config.SslKey, srv)
		if err != nil {
			println("Could not start the https server", err.Error())
		}
	} else {
		println("Starting the server with port", Config.Port, "on http")
		err := http.ListenAndServe("0.0.0.0:"+Config.Port, srv)
		if err != nil {
			println("Could not start the http server", err.Error())
		}
	}
}
