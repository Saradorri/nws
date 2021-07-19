package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"nws/config"
	"nws/pkg/rd"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var rooms = make(map[string]map[string]*websocket.Conn)
var rds rd.RedisClient

type ReceiveMessage struct {
	ID string `json:"id"`
}

type Message struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

func Index() {
	for _, value := range config.CNF.Queues {
		http.HandleFunc(fmt.Sprintf("/%s", value), handleConnections)
	}

	go Resend()
	log.Printf("http server started on :%v", config.CNF.Server.Port)
	err := http.ListenAndServe(":"+config.CNF.Server.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	queue := strings.Trim(r.URL.Path, "/")

	if len(rooms[queue]) == 0 {
		rooms[queue] = make(map[string]*websocket.Conn)
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	cid := uuid.NewV4().String()
	rooms[queue][cid] = ws

	for {
		var msg ReceiveMessage

		//Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)

		rds.Delete(msg.ID)

		if err != nil {
			fmt.Printf("error: %v", err)

			delete(rooms[queue], cid)
			break
		}
	}
}

func HandleMessages(q string, mb []byte) {

	if len(rooms[q]) == 0 {
		saveMessage(q, mb)
	} else {
		for id, client := range rooms[q] {
			msg := Message{
				ID:   uuid.NewV4().String(),
				Body: string(mb),
			}
			message := sendMessage(client, msg, q, id)

			info := rd.Info{
				Client:    client,
				Queue:     q,
				ClientID:  id,
				MessageID: message.ID,
				Message:   message.Body,
			}

			rds.Set(message.ID, info)
		}
	}
}

func sendMessage(client *websocket.Conn, msg Message, q, id string) *Message {
	defer func() {
		if r := recover(); r != nil {
			delete(rooms[q], id)
		}
	}()

	err := client.WriteJSON(msg)

	if err != nil {
		log.Printf("error: %v", err)
		_ = client.Close()
		_ = rooms[q][id].Close()
		rooms[q][id] = nil
		delete(rooms[q], id)
		return nil
	}
	return &msg
}

func saveMessage(q string, mb []byte) {
	msg := Message{
		ID:   uuid.NewV4().String(),
		Body: string(mb),
	}

	info := rd.Info{
		Client:    nil,
		Queue:     q,
		ClientID:  "",
		MessageID: msg.ID,
		Message:   msg.Body,
	}

	rds.Set(msg.ID, info)
}

func Resend() {

	var rdc rd.RedisClient

	data := rdc.GetAllKeys()
	for _, value := range data {
		res := rdc.Get(value.(string))
		m := Message{
			ID:   res.MessageID,
			Body: res.Message,
		}
		if res.Client == nil {
			if len(rooms[res.Queue]) != 0 {
				for id, client := range rooms[res.Queue] {
					msg := sendMessage(client, m, res.Queue, id)
					if msg != nil {
						rdc.Delete(res.MessageID)
					}
				}
			}
		} else {
			msg := sendMessage(res.Client, m, res.Queue, res.ClientID)
			if msg != nil {
				rdc.Delete(res.MessageID)
			}
		}
	}

	time.Sleep(time.Duration(config.CNF.Server.ResendInSeconds) * time.Second)
	Resend()
}
