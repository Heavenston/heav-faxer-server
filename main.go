package main

import (
	"math/rand"
	"net/http"
	"os"
	"time"

	"faxer.heav.fr/server/api"
)

func main() {
	rand.Seed(time.Now().Unix())
	srv := api.NewServer()

	port, isPortSet := os.LookupEnv("PORT")
	if !isPortSet {
		port = "1234"
	}

	cert, isCertSet := os.LookupEnv("CERT")
	key, isKeySet := os.LookupEnv("KEY")
	if isCertSet && isKeySet {
		println("Starting the server with port", port, "on https")
		err := http.ListenAndServeTLS("0.0.0.0:"+port, cert, key, srv)
		if err != nil {
			println("Could not start the https server", err.Error())
		}
	} else {
		println("Starting the server with port", port, "on http")
		err := http.ListenAndServe("0.0.0.0:"+port, srv)
		if err != nil {
			println("Could not start the http server", err.Error())
		}
	}
}
