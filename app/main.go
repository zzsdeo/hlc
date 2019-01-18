package main

import (
	"archive/zip"
	"bufio"
	"github.com/mailru/easyjson"
	"hlc/app/models"
	"hlc/app/rest"
	"log"
	"os"
	"strconv"
)

const (
	listenAddrEnvName = "SERVER_ADDR"
	defaultListenAddr = ":80"

	optionsFilePath = "/tmp/data/options.txt" //todo docker
	dataFilePath    = "/tmp/data/data.zip"

	//dataPath    = "/tmp/data/data/"

	//optionsFilePath = "/home/zzsdeo/tmp/data/options.txt" //todo hp
	//dataFilePath    = "/home/zzsdeo/tmp/data/data.zip"

	//optionsFilePath = "./tmp/data/options.txt" //todo home
	//dataFilePath    = "./tmp/data/data.zip"
)

type opts struct {
	listenAddr string
	now        int
}

func main() {
	opts := parseOpts()

	app := rest.App{}

	app.Initialize(opts.now)

	r, err := readZip()
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	for _, file := range r.File {
		data, err := parseData(file)
		if err != nil {
			log.Fatal("[ERROR] ", err)
		}
		app.LoadData(data.Accounts)
	}
	err = r.Close()
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	app.Run(opts.listenAddr)
}

func parseOpts() opts {
	opts := opts{}

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

func readZip() (*zip.ReadCloser, error) {
	r, err := zip.OpenReader(dataFilePath)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func parseData(f *zip.File) (models.Accounts, error) {
	accounts := models.Accounts{}
	file, err := f.Open()
	if err != nil {
		return accounts, err
	}

	err = easyjson.UnmarshalFromReader(file, &accounts)
	if err != nil {
		return accounts, err
	}

	err = file.Close()
	if err != nil {
		log.Println("[ERROR] ", err)
	}

	return accounts, nil
}
