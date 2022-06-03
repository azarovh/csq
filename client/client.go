package client

import (
	"bufio"
	"csq/common"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type Config struct {
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Host         string `yaml:"host"`
	Port         uint16 `yaml:"port"`
	SendInterval uint   `yaml:"send_interval"`
	Input        string `yaml:"input_file"`
}

type Client struct {
	config Config
}

func CreateClient(config Config) *Client {
	return &Client{config: config}
}

func (c *Client) Send() {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", c.config.User, c.config.Password,
		c.config.Host, c.config.Port))
	if err != nil {
		log.Fatalf("Failed to dial queue: %s", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to create a channel to queue: %s", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("csq", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
		return
	}

	publish := func(body []byte) {
		ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(body),
			})
	}

	id := uuid.New()
	f, err := os.Open(c.config.Input)
	if err != nil {
		log.Fatalf("Error opening input file: %v\n", err)
		return
	}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		if len(line) > 0 {
			var serializedMsg []byte
			switch line[0] {
			case "add":
				if len(line) < 3 {
					log.Fatalln("Wrong number of argument to AddItem")
					continue
				}
				item, _ := json.Marshal(common.AddItem{TheItem: common.Item{Key: line[1], Value: line[2]}})
				msg := &common.Message{Sender: id.String(), Type: common.AddItemMessage, Data: item}
				serializedMsg, _ = json.Marshal(msg)
			case "rm":
				if len(line) < 2 {
					log.Fatalln("Wrong number of argument to AddItem")
					continue
				}
				item, _ := json.Marshal(common.RemoveItem{Key: line[1]})
				msg := &common.Message{Sender: id.String(), Type: common.RemoveItemMessage, Data: item}
				serializedMsg, _ = json.Marshal(msg)
			case "get":
				if len(line) < 2 {
					log.Fatalln("Wrong number of argument to AddItem")
					continue
				}
				item, _ := json.Marshal(common.GetItem{Key: line[1]})
				msg := &common.Message{Sender: id.String(), Type: common.GetItemMessage, Data: item}
				serializedMsg, _ = json.Marshal(msg)
			case "getall":
				item, _ := json.Marshal(common.GetAllItems{})
				msg := &common.Message{Sender: id.String(), Type: common.GetAllItemsMessage, Data: item}
				serializedMsg, _ = json.Marshal(msg)
			default:
				log.Println("Unknown message type")
			}
			publish(serializedMsg)
			time.Sleep(time.Duration(c.config.SendInterval) * time.Millisecond)
		}
	}
}
