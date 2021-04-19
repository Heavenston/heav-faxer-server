package main

import (
	"math/rand"
	"net/http"
	"time"

	"faxer.heav.fr/server/api"
)

func main() {
	rand.Seed(time.Now().Unix())
	srv := api.NewServer()

	http.ListenAndServe("0.0.0.0:1234", srv)
}
