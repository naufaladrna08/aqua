GRPC_PORT = 9090
HTTP_PORT = 8080
DB_HOST = localhost
DB_PORT = 3306
DB_USER = root
DB_PASSWORD = 1234
DB_SCHEMA = aqua
DB_ARGS = -db-host=$(DB_HOST) -db-port=$(DB_PORT) -db-user=$(DB_USER) -db-password=$(DB_PASSWORD) -db-schema=$(DB_SCHEMA)

init:
	@echo "Creating directories"
	mkdir -pv vendor api/proto api/swagger cmd/aqua pkg/api/aqua

	@echo "Installing dependencies"
	go mod download
	go mod vendor

build:
	@echo "Building the application"
	go build -o cmd/aqua/aqua cmd/aqua/main.go

proto:
	@echo "Generating go package for proto file"
	protoc --proto_path=api/proto --proto_path=third_party --go_out=pkg/api --go-grpc_out=pkg/api model.proto
	protoc --proto_path=api/proto --proto_path=third_party --grpc-gateway_out=logtostderr=true:pkg/api model.proto
	protoc --proto_path=api/proto --proto_path=third_party --swagger_out=logtostderr=true:api/swagger model.proto

run-server: build
	@echo "Running the server"
	./cmd/aqua/aqua -grpc-port=$(GRPC_PORT) -http-port=$(HTTP_PORT) $(DB_ARGS)
all: proto build
run: run-server

clean:
	@echo "Cleaning the application"
	rm -rf cmd/aqua/aqua
	rm -rf pkg/api/model.pb.go
	rm -rf pkg/api/model_grpc.pb.go
	rm -rf pkg/api/model.pb.gw.go
	rm -rf api/swagger/model.swagger.json