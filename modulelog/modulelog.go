package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type PostMsg struct {
	Timestamp    time.Time `json:"timestamp"`
	Address      int       `json:"address"`
	Description  string    `json:"description"`
	Label        string    `json:"label"`
	Lqi          int       `json:"lqi"`
	Rssi         int       `json:"rssi"`
	Uptime       int       `json:"uptime"`
	Tempcpu      int       `json:"tempcpu"`
	Vrefcpu      int       `json:"vrefcpu"`
	Ntc0         int       `json:"ntc0"`
	Ntc1         int       `json:"ntc1"`
	Photores     int       `json:"photores"`
	Pressure     int       `json:"pressure"`
	Temppressure int       `json:"temppressure"`
}

var (
	debug   = flag.Bool("debug", false, "Send fake post to server")
	url     = flag.String("url", "http://radiolog.asterix.cloud/import", "Server post url")
	postRow = flag.String("postRow", "$13;50;180;733978;2600;1300;906;904;0;1016342;701;", "Data row to post")
)

func main() {
	flag.Parse()
	var port io.ReadWriteCloser
	var err error
	if !*debug {
		// Set up options.
		options := serial.OpenOptions{
			PortName:        "/dev/ttyS2",
			BaudRate:        115200,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 1,
		}

		// Open the port.
		port, err = serial.Open(options)
		if err != nil {
			fmt.Println("serial.Open: %v", err)
		}
		// Make sure to close it later.
		defer port.Close()
	}

	for {

		line := ""
		if !*debug {
			for {
				buf := make([]byte, 1)
				n, err := port.Read(buf)
				if err != nil {
					fmt.Println("port.Read: %v", err)
				}
				if string(buf[:n]) == "\n" {
					break
				}
				line = line + string(buf[:n])
			}
		} else {
			line = *postRow
		}

		var loc *time.Location
		//set timezone,
		loc, err = time.LoadLocation("Europe/Rome")
		if err != nil {
			fmt.Println("Unable go get time clock..")
			panic(err)
		}

		now := time.Now().In(loc)
		t := now.Format("Mon Jan _2 15:04:05 2006 ")
		if len(line) > 1 {
			fmt.Println(t, line)
		}
		if strings.HasPrefix(line, "$") {
			line = line[1:]
			fields := strings.Split(line, ";")

			var postData PostMsg
			for item := 0; item < len(fields); item++ {
				switch item {
				case 0:
					postData.Address, _ = strconv.Atoi(fields[item])
					label := "Slave"
					if fields[item] == "0" {
						label = "Master"
					}
					postData.Label = label
					postData.Description = "-"
					postData.Timestamp = time.Now().In(loc)
				case 1:
					postData.Lqi, _ = strconv.Atoi(fields[item])
				case 2:
					postData.Rssi, _ = strconv.Atoi(fields[item])
				case 3:
					postData.Uptime, _ = strconv.Atoi(fields[item])
				case 4:
					postData.Tempcpu, _ = strconv.Atoi(fields[item])
				case 5:
					postData.Vrefcpu, _ = strconv.Atoi(fields[item])
				case 6:
					postData.Ntc0, _ = strconv.Atoi(fields[item])
				case 7:
					postData.Ntc1, _ = strconv.Atoi(fields[item])
				case 8:
					postData.Photores, _ = strconv.Atoi(fields[item])
				case 9:
					postData.Pressure, _ = strconv.Atoi(fields[item])
				case 10:
					postData.Temppressure, _ = strconv.Atoi(fields[item])
				}
			}

			jsonData, _ := json.Marshal(postData)
			fmt.Println(*url)
			req, err := http.NewRequest("POST", *url, bytes.NewBuffer(jsonData))

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Print(err)
				continue
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(t, resp.Status, string(body))

		}
	}
}
