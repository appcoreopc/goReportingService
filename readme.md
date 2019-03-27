
To start the server :- 

go run server\main.go

To run client 

go run client.go -h localhost:9006/status

Using curl to post data 

curl -d '{"ReportType":"value1", "ReportName":"value2", "RequestedBy" : "Jeremy"}' -H "Content-Type: application/json" -X POST localhost:9006/report


Starting docker instance 

docker run -d --name zookeeper -p 2181:2181 confluent/zookeeper

docker run -d --name kafka -p 9092:9092 --link zookeeper:zookeeper confluent/kafka


