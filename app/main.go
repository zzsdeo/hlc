package main

import (
	"archive/zip"
	"bufio"
	"hlc/app/models"
	"hlc/app/rest"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/bcicen/jstream"
)

const (
	listenAddrEnvName = "SERVER_ADDR"
	defaultListenAddr = ":80"

	//optionsFilePath = "/tmp/data/options.txt" //todo docker
	//dataFilePath    = "/tmp/data/data.zip"

	//optionsFilePath = "/home/zzsdeo/tmp/data/options.txt" //todo hp
	//dataFilePath    = "/home/zzsdeo/tmp/data/data.zip"

	optionsFilePath = "./tmp/data/options.txt" //todo home
	dataFilePath    = "./tmp/data/data.zip"
)

type opts struct {
	listenAddr string
	now        int
}

func main() {
	opts := parseOpts()

	app := rest.App{}

	app.Initialize(opts.now)

	parseData(app)

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

func parseData(app rest.App) {
	r, err := zip.OpenReader(dataFilePath)
	if err != nil {
		log.Fatal(err)
	}

	for _, zf := range r.File {
		file, err := zf.Open()
		if err != nil {
			log.Fatal(err)
		}

		decoder := jstream.NewDecoder(file, 2)
		var fname, sname, phone, country, city string
		var interests []string
		var premium *models.Premium
		var likes []models.Like
		for mv := range decoder.Stream() {
			accMap := mv.Value.(map[string]interface{})
			fname, sname, phone, country, city = "", "", "", "", ""
			if v, ok := accMap["fname"]; ok {
				fname = v.(string)
			}
			if v, ok := accMap["sname"]; ok {
				sname = v.(string)
			}
			if v, ok := accMap["phone"]; ok {
				phone = v.(string)
			}
			if v, ok := accMap["country"]; ok {
				country = v.(string)
			}
			if v, ok := accMap["city"]; ok {
				city = v.(string)
			}
			interests = []string{}
			if _, ok := accMap["interests"]; ok {
				for _, interest := range accMap["interests"].([]interface{}) {
					interests = append(interests, interest.(string))
				}
			}
			premium = nil
			if premMap, ok := accMap["premium"]; ok {
				premium = &models.Premium{
					Start:  int(premMap.(map[string]interface{})["start"].(float64)),
					Finish: int(premMap.(map[string]interface{})["finish"].(float64)),
				}
			}
			likes = []models.Like{}
			if likeMap, ok := accMap["likes"]; ok {
				for _, l := range likeMap.([]interface{}) {
					likes = append(likes, models.Like{
						ID: int(l.(map[string]interface{})["id"].(float64)),
						TS: int(l.(map[string]interface{})["ts"].(float64)),
					})
				}
			}
			app.AddAccount(models.Account{
				ID:        int(accMap["id"].(float64)),
				Email:     accMap["email"].(string),
				FName:     fname,
				SName:     sname,
				Phone:     phone,
				Sex:       accMap["sex"].(string),
				Birth:     int(accMap["birth"].(float64)),
				Country:   country,
				City:      city,
				Joined:    int(accMap["joined"].(float64)),
				Status:    accMap["status"].(string),
				Interests: interests,
				Premium:   premium,
				Likes:     likes,
			})
		}

		runtime.GC()

		err = file.Close()
		if err != nil {
			log.Println("[ERROR] ", err)
		}
	}

	err = r.Close()
	if err != nil {
		log.Fatal(err)
	}
}
