package business

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/task/entity"
)

type TaskRepository interface {
	AddNewTask(ctx context.Context, data *entity.TaskDataCreation) error
	UpdateTask(ctx context.Context, id int, data *entity.TaskDataUpdate) error
	DeleteTask(ctx context.Context, id int) error
	GetTaskById(ctx context.Context, id int) (*entity.Task, error)
	ListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error)
}

type UserRepository interface {
	GetUsersByIds(ctx context.Context, ids []int) ([]core.SimpleUser, error)
	GetUserById(ctx context.Context, id int) (*core.SimpleUser, error)
}

type business struct {
	taskRepo TaskRepository
	userRepo UserRepository
}

func NewBusiness(taskRepo TaskRepository, userRepo UserRepository) *business {
	return &business{
		taskRepo: taskRepo,
		userRepo: userRepo,
	}
}
