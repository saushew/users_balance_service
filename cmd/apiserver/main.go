package main

import (
	"flag"
	"log"

	"github.com/saushew/users-balance-service/app/apiserver"
	"github.com/saushew/users-balance-service/lib/j"
)

//	@title		User's Balance Service
//	@version	1.0

//	@host		localhost:8000
//	@usePath	/

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/apiserver.json", "path to config file")
}

func main() {
	flag.Parse()

	config := apiserver.NewConfig()
	if err := j.ParseFile(configPath, config); err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}

	if err := apiserver.Start(config); err != nil {
		log.Fatalf("error occured while running http server: %s", err.Error())
	}
}
