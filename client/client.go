package client

import (
	"csq/common"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Config struct {
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Host         string `yaml:"host"`
	Port         uint16 `yaml:"port"`
	SendInterval uint16 `yaml:"send_interval"`
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
		fmt.Println(string(body))
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

	additem, _ := json.Marshal(common.AddItem{TheItem: common.Item{Key: "a", Value: "v"}})
	msg1, _ := json.Marshal(common.Message{Type: common.AddItemMessage, Data: additem})
	publish(msg1)

	additem1, _ := json.Marshal(common.AddItem{TheItem: common.Item{Key: "b", Value: "v"}})
	msg2, _ := json.Marshal(common.Message{Type: common.AddItemMessage, Data: additem1})
	publish(msg2)

	additem2, _ := json.Marshal(common.AddItem{TheItem: common.Item{Key: "c", Value: "v"}})
	msg3, _ := json.Marshal(common.Message{Type: common.AddItemMessage, Data: additem2})
	publish(msg3)

	additem3, _ := json.Marshal(common.AddItem{TheItem: common.Item{Key: "d", Value: "v"}})
	msg4, _ := json.Marshal(common.Message{Type: common.AddItemMessage, Data: additem3})
	publish(msg4)

	getallitems, _ := json.Marshal(common.GetAllItems{})
	msg5, _ := json.Marshal(common.Message{Type: common.GetAllItemsMessage, Data: getallitems})
	publish(msg5)
}
