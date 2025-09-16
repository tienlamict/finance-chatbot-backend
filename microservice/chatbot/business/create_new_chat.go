package business

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/chatbot/entity"
)

func (biz *business) CreateNewChat(ctx context.Context, data *entity.ChatDataCreation) error {
	requester := core.GetRequester(ctx)

	uid, _ := core.FromBase58(requester.GetSubject())
	requesterId := int(uid.GetLocalID()) // chat owner id, id of who creates this new chat

	data.Prepare(requesterId, "user")

	if err := biz.chatRepo.AddNewMessage(ctx, data); err != nil {
		return core.ErrInternalServerError.WithError(entity.ErrCannotCreateMessage.Error())
	}

	return nil
}
