package main

import (
	"hlc/app/rest"
	"io/ioutil"
	"log"
	"os"
)

const (
	mongoAddrEnvName = "MONGO_ADDR"
	defaultMongoAddr = "mongodb://localhost:27017"

	listenAddrEnvName = "SERVER_ADDR"
	defaultListenAddr = "localhost:8000"

	optionsFilePath = "/tmp/data/options.txt"
)

type opts struct {
	mongoAddr  string
	listenAddr string
	now        int
}

func main() {
	opts := parseOpts()

	app := rest.App{}

	app.Initialize(opts.mongoAddr)

	app.SetNow(opts.now)

	app.DropCollection()

	app.Run(opts.listenAddr)
}

func parseOpts() opts {
	opts := opts{}

	opts.mongoAddr = os.Getenv(mongoAddrEnvName)
	if opts.mongoAddr == "" {
		opts.mongoAddr = defaultMongoAddr
	}

	opts.listenAddr = os.Getenv(listenAddrEnvName)
	if opts.listenAddr == "" {
		opts.listenAddr = defaultListenAddr
	}

	b, err := ioutil.ReadFile(optionsFilePath)
	if err != nil {
		log.Println("[ERROR] ", err)
	}

	return opts
}
