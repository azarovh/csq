package server

import (
	"csq/common"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/streadway/amqp"
)

type Config struct {
	Workers uint8 `yaml:"server_workers"`

	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`

	Output string `yaml:"output_file"`
}

type Server struct {
	config Config

	store     *OrderedMap
	storeLock sync.RWMutex

	//workersChan map[string]chan []byte
	//workersLock sync.Mutex

	resultLog *log.Logger
}

type worker struct {
}

func CreateServer(config Config) *Server {
	return &Server{config: config, store: CreateOrderedMap(), resultLog: log.New(os.Stdout, "", log.LstdFlags)}
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

	workersChan := make(chan []byte, s.config.Workers)
	var wg sync.WaitGroup
	for t := uint8(0); t < s.config.Workers; t++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range workersChan {
				s.demuxMessage(msg)
			}
		}()
	}

	for msg := range msgs {
		workersChan <- msg.Body
	}

	close(workersChan)
	wg.Wait()
}

func (s *Server) demuxMessage(data []byte) {
}

func (s *Server) processMessage(data []byte) {
	msg := &common.Message{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		log.Printf("Unable to unmarshal a message: %s", err)
		return
	}

	switch msg.Type {
	case common.AddItemMessage:
		addMsg := common.AddItem{}
		if err := json.Unmarshal(msg.Data, &addMsg); err != nil {
			log.Printf("Unable to unmarshal AddItem message: %s", err)
			return
		}

		s.storeLock.Lock()
		defer s.storeLock.Unlock()

		s.store.Add(addMsg.TheItem.Key, addMsg.TheItem.Value)
		s.resultLog.Printf("Item added {%v}", addMsg.TheItem)

	case common.RemoveItemMessage:
		rmMsg := common.RemoveItem{}
		if err := json.Unmarshal(msg.Data, &rmMsg); err != nil {
			log.Printf("Unable to unmarshal RemoveItem message: %s", err)
			return
		}

		s.storeLock.Lock()
		defer s.storeLock.Unlock()

		s.store.Remove(rmMsg.Key)
		s.resultLog.Printf("Item removed {%s}", rmMsg.Key)
	case common.GetItemMessage:
		getMsg := common.GetItem{}
		if err := json.Unmarshal(msg.Data, &getMsg); err != nil {
			log.Printf("Unable to unmarshal GetItem message: %s", err)
			return
		}

		s.storeLock.RLock()
		defer s.storeLock.RUnlock()

		if v := s.store.Get(getMsg.Key); v != nil {
			s.resultLog.Printf("Item requested {%s}:{%s}", getMsg.Key, *v)
		} else {
			s.resultLog.Printf("Item requested {%s} but not found", getMsg.Key)
		}
	case common.GetAllItemsMessage:
		getMsg := common.GetAllItems{}
		if err := json.Unmarshal(msg.Data, &getMsg); err != nil {
			log.Printf("Unable to unmarshal GetAllItems message: %s", err)
			return
		}

		s.storeLock.RLock()
		defer s.storeLock.RUnlock()

		l := s.store.GetAll()
		s.resultLog.Println("All items requested:")
		for el := l.Front(); el != nil; el = el.Next() {
			fmt.Printf("%s, ", el.Value.(*Item).Key)
		}
	default:
		log.Printf("Unknown message type: %d", msg.Type)
	}
}
