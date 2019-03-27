package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/appcoreopc/reportingService/services"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/optiopay/kafka"
	"github.com/optiopay/kafka/proto"
)

type ReportRequest struct {
	ReportType  string
	ReportName  string
	RequestedBy string
}

type ReportStatus struct {
	ReportName string
	Status     string
}

var statusChannel = make(chan *ReportStatus, 2)
var clients = make(map[*websocket.Conn]bool)

func createReportRequest(w http.ResponseWriter, r *http.Request) {

	var reportReq ReportRequest

	if err := json.NewDecoder(r.Body).Decode(&reportReq); err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}

	fmt.Println("getting report status from rest request")

	defer r.Body.Close()

	statusReport := ReportStatus{ReportName: reportReq.ReportName, Status: "Running"}

	UpdateReportStatus(&statusReport)

}

func UpdateReportStatus(status *ReportStatus) {

	statusChannel <- status
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleReportStatus(w http.ResponseWriter, r *http.Request) {

	fmt.Println("status hit!")

	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
	}

	clients[c] = true

	fmt.Println(clients)
}

const (
	topic     = "test"
	partition = 0
)

var kafkaAddrs = []string{"localhost:9092", "localhost:9093"}

func main() {

	msgSvc := services.MessagingService{}

	msgSvc.Init(kafkaAddrs, "test-client", "test")

	go msgSvc.Subscribe()

	time.Sleep(2000)

	go msgSvc.Send("testintetest")

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/report", createReportRequest).Methods("POST")
	r.HandleFunc("/status", handleReportStatus)

	go ListenToIncomingStatus()

	log.Println("Serving services on port 9006")
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":9006", r))

}

func ListenToIncomingStatus() {

	fmt.Println("listening to incoming update from app")
	for {

		status := <-statusChannel
		fmt.Println(status.Status)
		fmt.Println(status.ReportName)

		for c := range clients {

			err := c.WriteMessage(websocket.TextMessage, []byte(status.ReportName))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func printConsumed(broker kafka.Client) {
	conf := kafka.NewConsumerConf(topic, partition)
	conf.StartOffset = kafka.StartOffsetNewest
	consumer, err := broker.Consumer(conf)
	if err != nil {
		log.Fatalf("cannot create kafka consumer for %s:%d: %s", topic, partition, err)
	}

	for {
		msg, err := consumer.Consume()
		if err != nil {
			if err != kafka.ErrNoData {
				log.Printf("cannot consume %q topic message: %s", topic, err)
			}
			break
		}
		log.Printf("message %d: %s", msg.Offset, msg.Value)
	}
	log.Print("consumer quit")
}

func sendMessage(broker kafka.Client) {

	fmt.Println("will be sending out messagess...." + time.Now().String())
	time.Sleep(2000)
	producer := broker.Producer(kafka.NewProducerConf())

	msg := &proto.Message{Value: []byte("hello world")}
	if _, err := producer.Produce(topic, partition, msg); err != nil {
		log.Fatalf("cannot produce message to %s:%d: %s", topic, partition, err)
	}

}
