package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/shirou/gopsutil/v3/server"
)

var (
	listenPort = flag.Uint("port", 8080, "Listen on ports (1-65535)")
)

func main() {
	flag.Parse()
	if *listenPort > 65535 {
		flag.Usage()
		os.Exit(1)
	}
	http.Handle("/", server.AddRoutes(server.MakeRouter()))
	fmt.Printf("Starting up on port %d\n", *listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *listenPort), nil))
}
