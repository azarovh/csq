package main

import (
	"csq/client"
	"csq/server"
	"flag"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func main() {
	modePtr := flag.String("mode", "", "[client|server]")

	flag.Parse()

	if modePtr == nil {
		log.Fatalf("Mode was not provided")
	} else if *modePtr == "server" {
		configFile, err := ioutil.ReadFile("config.yaml")
		if err != nil {
			log.Printf("Failed to read config.yaml: %v", err)
			return
		}

		config := &server.Config{}
		err = yaml.Unmarshal(configFile, config)
		if err != nil {
			log.Fatalf("Failed to unmarshal config: %v", err)
		}
		s := server.CreateServer(*config)
		s.Run()
	} else if *modePtr == "client" {
		configFile, err := ioutil.ReadFile("config.yaml")
		if err != nil {
			log.Printf("Failed to read config.yaml: %v", err)
			return
		}

		config := &client.Config{}
		err = yaml.Unmarshal(configFile, config)
		if err != nil {
			log.Fatalf("Failed to unmarshal config: %v", err)
		}
		c := client.CreateClient(*config)
		c.Send()
	} else {
		log.Fatalf("Unknown mode")
	}
}
