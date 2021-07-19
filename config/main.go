package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func init() {
	initializeConfigs()
}

type config struct {
	RabbitMQ rabbitmq `json:"rabbitmq"`
	Queues   []string `json:"queues"`
	Server   server   `json:"server"`
	Redis    redis    `json:"redis"`
}

type rabbitmq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

type server struct {
	Port            string `json:"port"`
	ResendInSeconds int    `json:"resendInSeconds"`
}

type redis struct {
	Name     int    `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
}

var CNF config
var Path string

func initializeConfigs() {
	var configFile *os.File
	var err error
	if !strings.Contains(os.Args[0], "___Test") && !strings.Contains(os.Args[0], ".test") {
		c := flag.String("config", "/etc/nws/nws.json", "path to the config file")
		flag.Parse()
		Path = *c
	} else {
		Path = "/etc/nws/nws.json"
	}
	configFile, err = os.Open(Path)
	if err != nil {
		fmt.Print("config file does not exists")
		return
	}
	defer configFile.Close()
	byteValue, _ := ioutil.ReadAll(configFile)
	errUnmarshal := json.Unmarshal(byteValue, &CNF)
	if errUnmarshal != nil {
		fmt.Print("config file is malformed")
		return
	}
}
