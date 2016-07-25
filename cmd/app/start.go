package main

import (
	"flag"

	"github.com/tophers42/go-naivebayes/naivebayes"
)

// Create the application, register endpoints and start it.
func main() {
	configFile := flag.String("config_file", nil, "Config file path. Other options override config")
	portPointer := flag.String("port", ":8080", "Port to take requests on format = ':<port_number>'")
	flag.Parse()

	naivebayes.NewApp(*portPointer).StartServer()
}
