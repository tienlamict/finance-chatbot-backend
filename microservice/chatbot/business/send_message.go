// finance-chatbot-backend/microservice/chatbot/business/send_message.go
package business

import (
	"context"
	"fmt"
	"time"

	"finance-chatbot/addon/common"
	minioc "finance-chatbot/addon/component/minioc"
	"finance-chatbot/microservice/chatbot/entity"
)

func (uc *chatUsecase) ensureConversation(ctx context.Context, req entity.SendMessageRequest, userID string) (string, error) {
	if req.ConversationID != "" {
		// Đảm bảo conv tồn tại
		if _, err := uc.sql.FindByID(ctx, req.ConversationID); err == nil {
			return req.ConversationID, nil
		}
		// Nếu không tìm thấy thì fallback tạo mới
	}

	// Validate userID is not empty when creating new conversation
	if userID == "" {
		return "", fmt.Errorf("user ID is required to create a new conversation")
	}

	conv := &entity.Conversation{
		ID:        common.NewV7(),
		UserID:    userID,
		OrgID:     req.OrgID,
		Title:     req.Title,
		Status:    entity.ConversationActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.sql.Create(ctx, conv); err != nil {
		return "", err
	}
	return conv.ID, nil
}

func (uc *chatUsecase) SendMessage(ctx context.Context, req entity.SendMessageRequest, userID string) (*entity.SendMessageResponse, error) {
	convID, err := uc.ensureConversation(ctx, req, userID)
	if err != nil {
		return nil, err
	}

	// 1) Lưu message của user
	now := time.Now()
	content := req.Content
	userMsg := &entity.Message{
		ID:             common.NewV7(),
		ConversationID: convID,
		Role:           entity.RoleUser,
		Content:        &content,
		ContentType:    "text",
		CreatedAt:      now,
	}
	if err := uc.sql.CreateMessage(ctx, userMsg); err != nil {
		return nil, err
	}

	// 1.5) Process and store attachments if any
	var attachments []*entity.MessageAttachment
	if len(req.Files) > 0 {
		attachments, err = uc.processAttachments(ctx, userMsg.ID, userID, req.Files)
		if err != nil {
			return nil, fmt.Errorf("failed to process attachments: %w", err)
		}
	}

	// 2) Gọi AI service
	start := time.Now()
	aiRes, err := uc.ai.Generate(ctx, userID, convID, req.Content)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		// Lưu 1 message assistant báo lỗi (optional)
		errCode := "ai_generate_error"
		msg := &entity.Message{
			ID:             common.NewV7(),
			ConversationID: convID,
			Role:           entity.RoleAssistant,
			Content:        nil,
			ContentType:    "text",
			ErrorCode:      &errCode,
			LatencyMs:      latency,
			CreatedAt:      time.Now(),
		}
		_ = uc.sql.CreateMessage(ctx, msg)
		return nil, err
	}

	// 3) Lưu message của assistant
	assistantContent := aiRes.Content
	model := aiRes.Model
	assistantMsg := &entity.Message{
		ID:             common.NewV7(),
		ConversationID: convID,
		Role:           entity.RoleAssistant,
		Content:        &assistantContent,
		ContentType:    "text",
		TokensInput:    aiRes.TokensInput,
		TokensOutput:   aiRes.TokensOutput,
		LatencyMs:      latency,
		ModelName:      &model,
		CreatedAt:      time.Now(),
	}
	if err := uc.sql.CreateMessage(ctx, assistantMsg); err != nil {
		return nil, err
	}

	// Convert attachments to DTOs
	attachmentDTOs := make([]entity.AttachmentDTO, len(attachments))
	for i, att := range attachments {
		attachmentDTOs[i] = entity.AttachmentDTO{
			ID:         att.ID,
			Filename:   att.Filename,
			MimeType:   *att.MimeType,
			SHA256:     *att.SHA256,
			Pages:      att.Pages,
			CreatedAt:  att.CreatedAt,
			StorageKey: att.StorageKey,
		}
	}

	return &entity.SendMessageResponse{
		ConversationID:       convID,
		UserMessageID:        userMsg.ID,
		AssistantMessageID:   assistantMsg.ID,
		AssistantContent:     assistantContent,
		ModelName:            model,
		TokensInput:          aiRes.TokensInput,
		TokensOutput:         aiRes.TokensOutput,
		LatencyMs:            latency,
		UserMessageCreatedAt: userMsg.CreatedAt,
		AIMessageCreatedAt:   assistantMsg.CreatedAt,
		Attachments:          attachmentDTOs,
	}, nil
}

// processAttachments handles file uploads and creates attachment records
func (uc *chatUsecase) processAttachments(ctx context.Context, messageID, userID string, files []entity.FileUploadInfo) ([]*entity.MessageAttachment, error) {
	if uc.storage == nil {
		return nil, fmt.Errorf("storage provider not configured")
	}

	attachments := make([]*entity.MessageAttachment, 0, len(files))

	for _, fileInfo := range files {
		// Generate storage key
		storageKey := minioc.GenerateStorageKey(userID, messageID, fileInfo.FileHeader.Filename)

		// Upload file to object storage
		finalStorageKey, err := uc.storage.UploadFile(ctx, storageKey, fileInfo.Content, fileInfo.MimeType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %s: %w", fileInfo.FileHeader.Filename, err)
		}

		// Create attachment record
		attachment := &entity.MessageAttachment{
			ID:         common.NewV7(),
			MessageID:  messageID,
			Filename:   common.SanitizeFilename(fileInfo.FileHeader.Filename),
			StorageKey: finalStorageKey,
			MimeType:   &fileInfo.MimeType,
			SHA256:     &fileInfo.SHA256Hash,
			Pages:      fileInfo.Pages,
			CreatedAt:  time.Now(),
		}

		if err := uc.sql.CreateAttachment(ctx, attachment); err != nil {
			// Attempt to delete uploaded file on DB error
			_ = uc.storage.DeleteFile(ctx, storageKey)
			return nil, fmt.Errorf("failed to save attachment record: %w", err)
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}
