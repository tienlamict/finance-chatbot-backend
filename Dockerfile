FROM golang:1.24.2-alpine AS builder

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chatbot_app .

FROM alpine
WORKDIR /app/
COPY --from=builder /app/chatbot_app .
ENTRYPOINT ["./chatbot_app"]