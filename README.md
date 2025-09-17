go get -u gorm.io/gorm

go get -u github.com/gin-gonic/gin

go get gorm.io/driver/mysql

go get github.com/spf13/viper@latest

go get google.golang.org/grpc

------------------Gen proto------------------
winget install protobuf // Install protobuf

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go-grpc_out=. proto/ai.proto

---------------------------------------------

docker run --name finance-admin-db -e MYSQL_ROOT_PASSWORD=mysecretpassword -d -p 3309:3306 mysql:latest

--------------------------------------

0. docker compose down -v

1. docker compose up --force-recreate --detach --build app

2. docker compose exec app ./chatbot_app outenv