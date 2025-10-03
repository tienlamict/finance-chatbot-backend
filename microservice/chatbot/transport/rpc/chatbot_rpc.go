package rpc

import (
	"context"
	"time"

	"finance-chatbot/microservice/chatbot/business"
	"finance-chatbot/microservice/chatbot/entity"
	"finance-chatbot/proto/pb"
)

type Server struct {
	pb.UnimplementedChatbotServiceServer
	uc business.ChatUsecase
}

func NewRPCServer(uc business.ChatUsecase) *Server {
	return &Server{uc: uc}
}

func (s *Server) SendMessage(ctx context.Context, r *pb.SendMessageRequest) (*pb.SendMessageReply, error) {
	req := entity.SendMessageRequest{
		ConversationID: r.GetConversationId(),
		Content:        r.GetContent(),
		UserID:         r.GetUserId(),
	}
	// optional
	if org := r.GetOrgId(); org != "" {
		req.OrgID = &[]string{org}[0]
	}
	if title := r.GetTitle(); title != "" {
		req.Title = &[]string{title}[0]
	}

	res, err := s.uc.SendMessage(ctx, req, req.UserID)
	if err != nil {
		return nil, err
	}

	return &pb.SendMessageReply{
		ConversationId:         res.ConversationID,
		UserMessageId:          res.UserMessageID,
		AssistantMessageId:     res.AssistantMessageID,
		AssistantContent:       res.AssistantContent,
		ModelName:              res.ModelName,
		TokensInput:            int32(res.TokensInput),
		TokensOutput:           int32(res.TokensOutput),
		LatencyMs:              int32(res.LatencyMs),
		UserMessageCreatedAtMs: toMs(res.UserMessageCreatedAt),
		AiMessageCreatedAtMs:   toMs(res.AIMessageCreatedAt),
	}, nil
}

func (s *Server) ListMessages(ctx context.Context, r *pb.ListMessagesRequest) (*pb.ListMessagesReply, error) {
	req := entity.ListMessagesRequest{
		ConversationID: r.GetConversationId(),
		Limit:          int(r.GetLimit()),
		Offset:         int(r.GetOffset()),
	}
	res, err := s.uc.ListMessages(ctx, req)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.MessageItem, 0, len(res.Items))
	for i := range res.Items {
		m := res.Items[i]
		var content, parentID, model, errCode string
		if m.Content != nil {
			content = *m.Content
		}
		if m.ParentID != nil {
			parentID = *m.ParentID
		}
		if m.ModelName != nil {
			model = *m.ModelName
		}
		if m.ErrorCode != nil {
			errCode = *m.ErrorCode
		}
		items = append(items, &pb.MessageItem{
			Id:             m.ID,
			ConversationId: m.ConversationID,
			ParentId:       parentID,
			Role:           string(m.Role),
			Content:        content,
			ContentType:    m.ContentType,
			TokensInput:    int32(m.TokensInput),
			TokensOutput:   int32(m.TokensOutput),
			LatencyMs:      int32(m.LatencyMs),
			ModelName:      model,
			ErrorCode:      errCode,
			CreatedAtMs:    toMs(m.CreatedAt),
		})
	}

	return &pb.ListMessagesReply{
		ConversationId: res.ConversationID,
		Total:          res.Total,
		Items:          items,
	}, nil
}

func toMs(t time.Time) int64 { return t.UnixMilli() }
