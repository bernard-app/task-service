package usecase_test

import (
	"bernard/internal/domain/entity"
	"bernard/internal/usecase"
	mocks "bernard/internal/usecase/mocks"
	"bernard/utils"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_CreateTask(t *testing.T) {
	type fields struct {
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx     context.Context
		task    entity.Task
		tagsIDs []int64
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.Task
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				task: entity.Task{
					Name:        "test",
					Description: utils.Ptr("test"),
				},
				tagsIDs: []int64{},
			},
			want: &entity.Task{
				Name:        "test",
				Description: utils.Ptr("test"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("CreateTask", tt.args.ctx, tt.args.task).Return(tt.want, nil)

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
			}

			got, err := u.CreateTask(tt.args.ctx, tt.args.task, tt.args.tagsIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.want != nil {
				require.Equal(t, tt.want, got)
			}

			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}

func TestUseCase_UpdateTask(t *testing.T) {
	type fields struct {
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx    context.Context
		task   entity.UpdateTaskRequest
		userID uuid.UUID
		taskID int64
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.Task
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					Name:        utils.Ptr("test"),
					Description: utils.Ptr("test"),
				},
				userID: uuid.New(),
				taskID: 1,
			},
			want: &entity.Task{
				Name:        "test",
				Description: utils.Ptr("test"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("UpdateTask", tt.args.ctx, tt.args.task).Return(tt.want, nil)

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
			}

			got, err := u.UpdateTask(tt.args.ctx, tt.args.task, tt.args.userID, tt.args.taskID)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.want != nil {
				require.Equal(t, tt.want, got)
			}

			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}

func TestUseCase_DeleteTask(t *testing.T) {
	type fields struct {
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx    context.Context
		userID uuid.UUID
		taskID int64
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.Task
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			want: &entity.Task{
				ID: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid userID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.Nil,
				taskID: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid taskID",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
				taskID: -1,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("DeleteTask", tt.args.ctx, tt.args.userID, tt.args.taskID).Maybe().Return(tt.want, nil)

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
			}

			err := u.DeleteTask(tt.args.ctx, tt.args.userID, tt.args.taskID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}

func TestUseCase_GetTask(t *testing.T) {
	type fields struct {
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx    context.Context
		taskID int64
		userID uuid.UUID
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.Task
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				taskID: 1,
				userID: uuid.New(),
			},
			want: &entity.Task{
				ID:          1,
				Description: utils.Ptr("test"),
			},
			wantErr: false,
		},
		{
			name: "invalid userID",
			args: args{
				ctx:    context.Background(),
				taskID: -1,
				userID: uuid.New(),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("GetTask", tt.args.ctx, tt.args.taskID).Maybe().Return(tt.want, nil)

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
			}

			got, err := u.GetTask(tt.args.ctx, tt.args.taskID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTask() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.want != nil {
				require.Equal(t, tt.want, got)
			}

			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}

func TestUseCase_GetListTask(t *testing.T) {
	type fields struct {
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx      context.Context
		taskFilter entity.TasksFilter
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*entity.UserTasksTab
		wantErr bool
	}{
		{
			name: "success non-filter",
			args: args{
				ctx:    context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Limit: uint64(5),
					Offset: uint64(0),
				},
			},
			want:    []*entity.UserTasksTab{},
			wantErr: false,
		},
		{
			name: "error with userID",
			args: args{
				ctx:    context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Limit: uint64(5),
					Offset: uint64(0),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error with priority",
			args: args{
				ctx:      context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Priority: utils.Ptr(1),
					Limit: uint64(5),
					Offset: uint64(0),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error with date to",
			args: args{
				ctx:    context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					From: utils.Ptr(time.Now()),
					To:  utils.Ptr(time.Now()),
					Limit: uint64(5),
					Offset: uint64(0),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("ListTasks", tt.args.ctx, mock.Anything).Maybe().Return(tt.want, nil)

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
			}

			got, err := u.GetListTask(tt.args.ctx, tt.args.taskFilter)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetListTask() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.want != nil {
				require.Equal(t, tt.want, got)
			}

			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}
