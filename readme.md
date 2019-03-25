
To start the server :- 

go run server\main.go

To run client 

go run client.go -h localhost:9006/status

Using curl to post data 

curl -d '{"ReportType":"value1", "ReportName":"value2", "RequestedBy" : "Jeremy"}' -H "Content-Type: application/json" -X POST localhost:9006/report
