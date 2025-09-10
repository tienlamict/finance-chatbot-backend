go get -u gorm.io/gorm

go get -u github.com/gin-gonic/gin

go get gorm.io/driver/mysql

go get github.com/spf13/viper@latest

go get google.golang.org/grpc

docker run --name finance-admin-db -e MYSQL_ROOT_PASSWORD=mysecretpassword -d -p 3309:3306 mysql:latest

--------------------------------------

finance-chatbot-backend/
├── cmd/
│   └── api/
│       └── main.go            # entrypoint của server API
│
├── configs/
│   └── config.yaml            # file config (DB, Redis, S3, JWT…)
│
├── internal/
│   ├── app/
│   │   └── app.go             # khởi tạo toàn bộ app: DB, router, worker
│   │
│   ├── http/
│   │   ├── router.go          # khai báo routes
│   │   └── middleware/
│   │       ├── auth.go        # JWT middleware
│   │       ├── rbac.go        # Role-based access control
│   │       ├── logger.go      # log request/response
│   │       └── rate_limit.go  # giới hạn request
│   │
│   ├── auth/
│   │   ├── service.go         # login/register, phát hành JWT
│   │   ├── jwt.go
│   │   └── password.go
│   │
│   ├── users/
│   │   ├── model.go
│   │   ├── repo.go
│   │   └── handler.go
│   │
│   ├── data/
│   │   ├── model.go           # Issuer, Report, Document, Source
│   │   ├── repo.go
│   │   ├── handler.go         # upload/list reports
│   │   └── storage.go         # kết nối S3/MinIO
│   │
│   ├── crawl/
│   │   ├── job.go             # enqueue job crawl
│   │   ├── worker.go          # xử lý Asynq job
│   │   ├── collectors/        # crawler từng nguồn (SEC, HOSE…)
│   │   └── parsers/           # parse PDF/XLS/XBRL
│   │
│   ├── chat/
│   │   ├── model.go           # Conversation, Message, ToolRun
│   │   ├── repo.go
│   │   ├── handler.go         # API chat
│   │   ├── llm.go             # orchestrator gọi LLM
│   │   └── retrieval.go       # search/report retrieval
│   │
│   ├── db/
│   │   └── db.go              # mở DB connection, chạy migrations
│   │
│   ├── search/
│   │   └── fts.go             # full-text search hoặc OpenSearch client
│   │
│   └── vector/
│       └── pgvector.go        # hỗ trợ embedding search (nếu dùng)
│
├── migrations/                # file SQL migrations (golang-migrate)
│
├── scripts/                   # tiện ích, init db, seed data
│
├── go.mod
├── go.sum
└── docker-compose.yml         # services: API, Postgres, Redis, MinIO
