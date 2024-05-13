package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	portPtr := flag.Int("port", 3333, "укажите порт сервера")
	flag.Parse()

	userServer := newUserServer("users.db")

	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), userServer.mux)
}
