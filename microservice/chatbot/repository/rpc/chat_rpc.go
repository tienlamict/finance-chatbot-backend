package rpc

import (
	"context"
	"time"

	"finance-chatbot/microservice/chatbot/entity"
	ai "finance-chatbot/proto/pb" // gRPC stub đã generate từ proto/ai/ai.proto

	"google.golang.org/grpc"
)

type aiClient struct {
	conn *grpc.ClientConn
	cli  ai.AIServiceClient
}

func NewAIClient(addr string) (entity.AIClient, error) {
	// tuỳ môi trường bạn có thể dùng credentials thay vì Insecure
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
	if err != nil {
		return nil, err
	}
	return &aiClient{conn: conn, cli: ai.NewAIServiceClient(conn)}, nil
}

func (c *aiClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *aiClient) GenerateReply(userID string, history []entity.ChatMessage, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Map history (oldest-first tuỳ AI service yêu cầu)
	items := make([]*ai.ChatItem, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		h := history[i]
		items = append(items, &ai.ChatItem{
			Role:      string(h.Role),
			Content:   h.Content,
			Timestamp: h.CreatedAt.Unix(),
		})
	}

	res, err := c.cli.GenerateReply(ctx, &ai.GenerateReplyRequest{
		UserId:  userID,
		Prompt:  prompt,
		History: items,
	})
	if err != nil {
		return "", err
	}
	return res.GetReply(), nil
}
