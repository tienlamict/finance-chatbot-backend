# Get lib
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
# 1. Open your terminal on project

- Remove volume (optional):

docker compose down -v

- Build app:

docker compose up --force-recreate --detach --build app


# 2. API
- Register:
```shell
curl --location 'http://localhost:3001/v1/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "lamnt@chatbot.com",
    "password": "12345678",
    "last_name": "Microservices",
    "first_name": "Lam "
}'
```

- Login:
```shell
curl --location 'http://localhost:3001/v1/authenticate' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "lamnt@chatbot.com",
    "password": "12345678"
}'
```


- Send message:
```shell
curl --location 'http://localhost:3001/v1/chat/send-message' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlNTMycW9zOGpqTTIiLCJleHAiOjE3NTg3MjY2MDIsIm5iZiI6MTc1ODEyMTgwMiwiaWF0IjoxNzU4MTIxODAyLCJqdGkiOiI3MGY5NGQ5OC04ZGI0LTRmNWYtYjU4NC01NWRkODRlNjdjNTQifQ.fy2AT5W4FL89e1B8iGeyKkJ3Di8j-fWZBh1H6JCwORg' \
--data '{
    "user_id": "u1",
    "content": "Hello, how is my account balance?"
  }'
  ```