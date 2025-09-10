go get -u gorm.io/gorm

go get -u github.com/gin-gonic/gin

go get gorm.io/driver/mysql

go get github.com/spf13/viper@latest

go get google.golang.org/grpc

docker run --name finance-admin-db -e MYSQL_ROOT_PASSWORD=mysecretpassword -d -p 3309:3306 mysql:latest

--------------------------------------