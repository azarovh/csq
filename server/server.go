package server

import (
	"csq/common"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type Config struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`

	Output string `yaml:"output_file"`
}

type Server struct {
	config Config

	workersChan map[string]chan *common.Message

	resultLog *log.Logger
}

func CreateServer(config Config) *Server {
	return &Server{config: config, workersChan: make(map[string]chan *common.Message),
		resultLog: log.New(os.Stdout, "", log.LstdFlags)}
}

func (s *Server) Run() {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", s.config.User, s.config.Password,
		s.config.Host, s.config.Port))
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

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to crate consumer: %s", err)
		return
	}

	f, err := os.OpenFile(s.config.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		s.resultLog.SetOutput(f)
	} else {
		log.Fatalf("Failed to open output file: %s. Writing result to stdout", err)
	}
	defer f.Close()

	log.Println("--- Listening for messages ---")

	// TODO: exit consumer loop
	for msg := range msgs {
		s.demuxMessage(msg.Body)
	}

	for _, c := range s.workersChan {
		close(c)
	}
}

func (s *Server) demuxMessage(data []byte) {
	msg := &common.Message{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		log.Printf("Unable to unmarshal a message: %s", err)
		return
	}

	if _, exist := s.workersChan[msg.Sender]; !exist {
		workerChannel := make(chan *common.Message, 100)
		s.workersChan[msg.Sender] = workerChannel
		go func() {
			store := CreateOrderedMap()
			for msg := range workerChannel {
				processMessage(store, msg, s.resultLog)
			}
		}()
	}
	s.workersChan[msg.Sender] <- msg
}

func processMessage(store *OrderedMap, msg *common.Message, resultLog *log.Logger) {
	switch msg.Type {
	case common.AddItemMessage:
		addMsg := common.AddItem{}
		if err := json.Unmarshal(msg.Data, &addMsg); err != nil {
			log.Printf("Unable to unmarshal AddItem message: %s", err)
			return
		}

		store.Add(addMsg.TheItem.Key, addMsg.TheItem.Value)
		resultLog.Printf("Item added {%v}", addMsg.TheItem)

	case common.RemoveItemMessage:
		rmMsg := common.RemoveItem{}
		if err := json.Unmarshal(msg.Data, &rmMsg); err != nil {
			log.Printf("Unable to unmarshal RemoveItem message: %s", err)
			return
		}

		store.Remove(rmMsg.Key)
		resultLog.Printf("Item removed {%s}", rmMsg.Key)
	case common.GetItemMessage:
		getMsg := common.GetItem{}
		if err := json.Unmarshal(msg.Data, &getMsg); err != nil {
			log.Printf("Unable to unmarshal GetItem message: %s", err)
			return
		}

		if v := store.Get(getMsg.Key); v != nil {
			resultLog.Printf("Item requested {%s}:{%s}", getMsg.Key, *v)
		} else {
			resultLog.Printf("Item requested {%s} but not found", getMsg.Key)
		}
	case common.GetAllItemsMessage:
		getMsg := common.GetAllItems{}
		if err := json.Unmarshal(msg.Data, &getMsg); err != nil {
			log.Printf("Unable to unmarshal GetAllItems message: %s", err)
			return
		}

		l := store.GetAll()
		var keys string
		for el := l.Front(); el != nil; el = el.Next() {
			keys += fmt.Sprintf("%s, ", el.Value.(*Item).Key)
		}
		resultLog.Printf("All items requested: %s", keys)
	default:
		log.Printf("Unknown message type: %d", msg.Type)
	}
}
