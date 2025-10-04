# Get lib
go get -u gorm.io/gorm

go get -u github.com/gin-gonic/gin

go get gorm.io/driver/mysql

go get github.com/spf13/viper@latest

go get google.golang.org/grpc

go get github.com/gofrs/uuid

go get gorm.io/datatypes

------------------Gen proto------------------
winget install protobuf // Install protobuf

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

cd proto
protoc --go_out=. --go-grpc_out=. chatbot.proto 

---------------------------------------------

docker run --name finance-admin-db -e MYSQL_ROOT_PASSWORD=mysecretpassword -d -p 3309:3306 mysql:latest

--------------------------------------
# 1. Open your terminal on project

- Remove volume (optional):
```shell
docker compose down -v
```
- Build app:
```shell
docker compose up --force-recreate --detach --build app
```

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

- Promt:
```shell
curl --location 'http://localhost:3001/v1/chatbot/promt' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlNTMycW9zOGpqTTIiLCJleHAiOjE3NjAxNTI3NjIsIm5iZiI6MTc1OTU0Nzk2MiwiaWF0IjoxNzU5NTQ3OTYyLCJqdGkiOiI4NDZlMmI1OC1iNmY2LTQyZTUtOTlmYS02N2E5ZjMzM2NkY2YifQ.ltRCc-5ynvgzMlWpMT2Dchp6J-LMMtsmIlRqCcN5w-o' \
--data '{
    "user_id": "lamnt",
    "content": "Báo cáo tỷ giá ngoại tệ 2 ngày gần đây nhất"
  }'
  ```