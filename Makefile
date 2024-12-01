generate:
	rm -rf pb/* && protoc --proto_path=proto proto/*.proto \
		   --go_out=pb --go_opt=paths=source_relative \
           --go-grpc_out=pb --go-grpc_opt=paths=source_relative \


bootstrap:
	rm -rf peers.json &&\
 	go build . &&\
 	./main --bootstrap=true

node1:
	./main --raft-port=8081

node2:
	./main --raft-port=8082

node3:
	./main --raft-port=8083

node4:
	./main --raft-port=8084


.PHONY: run_node1 run_node2 run_node3 run_node4
