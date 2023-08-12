package main

import (
	"log"

	"github.com/aashraybhandar1/goDis/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":6969")
	log.Fatal(srv.ListenAndServe())
}
