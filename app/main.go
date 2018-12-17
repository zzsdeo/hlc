package main

import (
	"os"
)

const (
	mongoUrlEnvName = "MONGO_URL"
	defaultMongoUrl = "mongodb://localhost:27017"

	urlEnvName = "PROJECT_DB_URL"
	defaultUrl = "localhost:8000"
)

type opts struct {
	mongoUrl string
	url      string
}

func main() {
	opts := parseOpts()
	app := rest.App{}
	app.Initialize(opts.mongoUrl)
	app.Run(opts.url)
}

func parseOpts() opts {
	opts := opts{}

	opts.mongoUrl = os.Getenv(mongoUrlEnvName)
	if opts.mongoUrl == "" {
		opts.mongoUrl = defaultMongoUrl
	}

	opts.url = os.Getenv(urlEnvName)
	if opts.url == "" {
		opts.url = defaultUrl
	}

	return opts
}
