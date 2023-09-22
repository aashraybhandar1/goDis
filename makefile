compile:  
	protoc api/v1/*.proto \
	--go_out=internal \
	--go-grpc_out=internal \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	--proto_path=.
test:  
	go test -race ./...
