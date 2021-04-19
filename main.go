package main

import (
	"net/http"

	"faxer.heav.fr/server/api"
)

func main() {
	srv := api.NewServer()

	http.ListenAndServe("0.0.0.0:1234", srv)
}
