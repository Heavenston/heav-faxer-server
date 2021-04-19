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

	http.ListenAndServe("0.0.0.0:"+port, srv)
}
