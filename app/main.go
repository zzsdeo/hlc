package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"hlc/app/models"
	"hlc/app/rest"
	"log"
	"os"
	"strconv"
)

const (
	mongoAddrEnvName = "MONGO_ADDR"
	defaultMongoAddr = "mongodb://:27017"

	listenAddrEnvName = "SERVER_ADDR"
	defaultListenAddr = ":80"

	optionsFilePath = "./tmp/data/options.txt" //todo for docker without dot
	dataFilePath    = "./tmp/data/data.zip"    //todo for docker without dot
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

	app.LoadData(parseData().Accounts)

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

	file, err := os.Open(optionsFilePath)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Println("[ERROR] ", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	opts.now, _ = strconv.Atoi(scanner.Text())

	log.Println("[DEBUG] ", opts)
	return opts
}

func parseData() models.Accounts {
	r, err := zip.OpenReader(dataFilePath)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	defer func() {
		err = r.Close()
		if err != nil {
			log.Println("[ERROR] ", err)
		}
	}()

	file, err := r.File[0].Open()
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Println("[ERROR] ", err)
		}
	}()

	accounts := models.Accounts{}

	err = json.NewDecoder(file).Decode(&accounts)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	return accounts
}
