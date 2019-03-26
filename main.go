package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"syscall"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/websocket"
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
var epoller *epoll

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

	fmt.Println("using ws upgrader and epoller!")

	// c, err := upgrader.Upgrade(w, r, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// clients[c] = true
	// fmt.Println(clients)

	conn, _, _, err := ws.UpgradeHTTP(r, w)

	if err != nil {
		fmt.Println("error upgrading")
		return
	}

	if err := epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection %v", err)
		conn.Close()
	}

}

func ListenToIncomingStatus() {

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

func StartPolling() {

	for {

		connections, err := epoller.Wait()

		status := <-statusChannel

		fmt.Println("we are in start pooling section")
		fmt.Println("getting conection info", connections)

		fmt.Println(status.Status)
		fmt.Println(status.ReportName)

		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}

		for _, conn := range connections {

			if conn == nil {
				break
			}

			// if _, _, err := wsutil.ReadClientData(conn); err != nil {
			// 	if err := epoller.Remove(conn); err != nil {
			// 		log.Printf("Failed to remove %v", err)
			// 	}
			// } else {

			log.Println("sending out value to clients")
			wsutil.WriteClientText(conn, []byte(status.ReportName))
			// This is commented out since in demo usage, stdout is showing messages sent from > 1M connections at very high rate
			log.Printf("msg: %s", "writting ")
			//}
		}

	}
}

func main() {

	var err error

	// Increase resources limitations

	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}

	go StartPolling()

	//r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	//r.HandleFunc("/report", createReportRequest).Methods("POST")
	//r.HandleFunc("/status", handleReportStatus)

	http.HandleFunc("/report", createReportRequest)
	http.HandleFunc("/status", handleReportStatus)

	//go ListenToIncomingStatus()

	log.Println("Serving services on port 9006")
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":9006", nil))

}
