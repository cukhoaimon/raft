generate:
	rm -rf pb/* && protoc --proto_path=proto proto/*.proto \
		   --go_out=pb --go_opt=paths=source_relative \
           --go-grpc_out=pb --go-grpc_opt=paths=source_relative \


bootstrap:
	rm -rf peers.json &&\
 	go build . &&\
 	./main --bootstrap=true
