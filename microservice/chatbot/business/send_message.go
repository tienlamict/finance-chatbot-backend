// finance-chatbot-backend/microservice/chatbot/business/send_message.go
package business

import (
	"context"
	"fmt"
	"time"

	"finance-chatbot/addon/common"
	minioc "finance-chatbot/addon/component/minioc"
	"finance-chatbot/microservice/chatbot/entity"
	aiclient "finance-chatbot/microservice/chatbot/repository/rpc"
)

func (uc *chatUsecase) ensureConversation(ctx context.Context, req entity.SendMessageRequest, userID string) (string, error) {
	if req.ConversationID != "" {
		// Validate that the conversation exists and belongs to the authenticated user
		hasAccess, err := uc.sql.CheckConversationAccess(ctx, req.ConversationID, userID)
		if err != nil {
			return "", fmt.Errorf("failed to check conversation access: %w", err)
		}
		if !hasAccess {
			return "", fmt.Errorf("conversation not found or access denied")
		}
		return req.ConversationID, nil
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

	// 2) Call AI service with full context (history + attachments)
	start := time.Now()
	aiRes, err := uc.callAIWithContext(ctx, userID, convID, req.Content, req.DeepResearch)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		// Save assistant message with error code
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

// callAIWithContext calls the AI service with conversation history and attachments
func (uc *chatUsecase) callAIWithContext(ctx context.Context, userID, conversationID, prompt string, deepResearch bool) (*aiclient.AIResult, error) {
	// Check if AI client supports enhanced context
	enhancedClient, isEnhanced := uc.ai.(aiclient.EnhancedAIClient)
	if !isEnhanced {
		// Fallback to basic Generate method
		return uc.ai.Generate(ctx, userID, conversationID, prompt)
	}

	// Fetch conversation history for AI context
	history, err := uc.buildConversationHistory(ctx, conversationID, enhancedClient.GetHistoryMaxTurns())
	if err != nil {
		// Log warning but continue without history
		fmt.Printf("[WARN] Failed to fetch conversation history: %v\n", err)
		history = []aiclient.HistoryItem{}
	}

	// Build attachment references with presigned URLs
	attachFiles, err := uc.buildAttachmentReferences(ctx, conversationID, enhancedClient.GetPresignedExpirySec())
	if err != nil {
		// Log warning but continue without attachments
		fmt.Printf("[WARN] Failed to build attachment references: %v\n", err)
		attachFiles = []aiclient.AttachFile{}
	}

	// Build RAG context if enabled
	ragDocs, err := uc.buildRAGContext(ctx, prompt)
	if err != nil {
		// Log warning but continue without RAG
		fmt.Printf("[WARN] Failed to build RAG context: %v\n", err)
		ragDocs = []aiclient.RAGContextSnippet{}
	}

	// Call AI with full context (history + attachments + RAG)
	return enhancedClient.GenerateWithContext(ctx, userID, conversationID, prompt, history, deepResearch, attachFiles, ragDocs)
}

// buildConversationHistory fetches recent messages and converts them to HistoryItem format
func (uc *chatUsecase) buildConversationHistory(ctx context.Context, conversationID string, maxTurns int) ([]aiclient.HistoryItem, error) {
	if maxTurns <= 0 {
		return []aiclient.HistoryItem{}, nil
	}

	// Fetch recent messages from DB
	messages, err := uc.sql.GetRecentMessagesForAI(ctx, conversationID, maxTurns)
	if err != nil {
		return nil, err
	}

	// Convert to HistoryItem format
	history := make([]aiclient.HistoryItem, 0, len(messages))
	for _, msg := range messages {
		// Skip messages without content
		if msg.Content == nil || *msg.Content == "" {
			continue
		}

		// Map role
		role := string(msg.Role)
		if role != "user" && role != "assistant" {
			continue // Skip other roles (e.g., "tool")
		}

		history = append(history, aiclient.HistoryItem{
			Role:      role,
			Message:   *msg.Content,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		})
	}

	return history, nil
}

// buildAttachmentReferences generates presigned URLs for recent attachments in the conversation
func (uc *chatUsecase) buildAttachmentReferences(ctx context.Context, conversationID string, expirySec int) ([]aiclient.AttachFile, error) {
	if uc.storage == nil {
		return []aiclient.AttachFile{}, nil
	}

	// Fetch recent messages to find their attachments
	// We'll look at the last few messages (e.g., last 10 messages)
	messages, err := uc.sql.ListByConversation(ctx, conversationID, 10, 0)
	if err != nil {
		return nil, err
	}

	var attachFiles []aiclient.AttachFile

	// Process attachments from recent messages
	for _, msg := range messages {
		if msg.Role != entity.RoleUser {
			continue // Only look at user messages for attachments
		}

		// Get attachments for this message
		attachments, err := uc.sql.GetAttachmentsByMessageID(ctx, msg.ID)
		if err != nil || len(attachments) == 0 {
			continue
		}

		// Generate presigned URLs for each attachment
		for _, att := range attachments {
			// Extract object name from storage key (format: bucket/path)
			objectName := extractObjectName(att.StorageKey)

			// Generate presigned URL
			expiry := time.Duration(expirySec) * time.Second
			presignedURL, err := uc.storage.GetPresignedURL(ctx, objectName, expiry)
			if err != nil {
				fmt.Printf("[WARN] Failed to generate presigned URL for %s: %v\n", att.Filename, err)
				continue
			}

			// Build AttachFile
			attachFile := aiclient.AttachFile{
				URL:      presignedURL,
				Filename: att.Filename,
			}
			if att.MimeType != nil {
				attachFile.MimeType = *att.MimeType
			}
			if att.SHA256 != nil {
				attachFile.SHA256 = *att.SHA256
			}
			if att.Pages != nil {
				attachFile.Pages = att.Pages
			}

			attachFiles = append(attachFiles, attachFile)
		}
	}

	return attachFiles, nil
}

// extractObjectName extracts the object name from a storage key (format: bucket/path)
func extractObjectName(storageKey string) string {
	// Storage key format: "bucket-name/path/to/file.ext"
	// We need to extract "path/to/file.ext"
	parts := splitStorageKey(storageKey)
	if len(parts) > 1 {
		return joinPath(parts[1:])
	}
	return storageKey
}

// splitStorageKey splits a storage key by '/'
func splitStorageKey(key string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '/' {
			if i > start {
				result = append(result, key[start:i])
			}
			start = i + 1
		}
	}
	if start < len(key) {
		result = append(result, key[start:])
	}
	return result
}

// joinPath joins path segments with '/'
func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "/" + parts[i]
	}
	return result
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
