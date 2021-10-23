package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

var user string
var password string

func callAPI(method, url string, body io.Reader) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 3 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	color.Blue("Calling %s", url)
	req, err := http.NewRequest(method, url, body)
	req.SetBasicAuth(user, password)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}


	return response, err
}


func main()  {
	var apiAddress string
	var reqHeaders cli.StringSlice

	app := cli.NewApp()
	app.Name = "purge CLI"
	app.Version = "1.0.0"
	app.Description = "CLI for purge cache and remove media"
	app.UsageText= "CLI for purge cache and remove media"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "api",
			Value: "http://purge-api.k8s.m39",
			Usage: "API address",
			Destination: &apiAddress,
		},
		cli.StringSliceFlag{
			Name: "header",
			Usage: "Additional requests headers header_name:header_value",
			Value: &reqHeaders,
		},
		cli.StringFlag{
			Name: "user",
			Value: "api",
			Usage: "API address",
			Destination: &user,
		},
		cli.StringFlag{
			Name: "password",
			Value: "123-qwe",
			Usage: "API address",
			Destination: &password,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "purge",
			Aliases: []string{"p"},
			Usage:   "purge list of urls",
			Action:  func(c *cli.Context) error {
				data := make(map[string][]string)
				data["urls"] = c.Args()
				body, err := json.Marshal(data)
				if err != nil {
					return err
				}

				headers := make(map[string]string)
				for _, h := range reqHeaders {
					parts := strings.Split(h, ":")
					headers[parts[0]] = parts[1]
				}

				res, err := callAPI("POST", apiAddress + "/purge", bytes.NewReader(body))
				if err != nil {
					return err
				}

				if res.StatusCode == 202 {
					color.Green("Successfully purge %s", data["urls"])
				} else {
					color.Red("There was an error. SC %d", res.StatusCode)
					fmt.Println()
					for h, v := range res.Header {
						fmt.Println("       ", h, ":", v)
					}

					b, _ := ioutil.ReadAll(res.Body)
					fmt.Println(string(b))

				}
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
