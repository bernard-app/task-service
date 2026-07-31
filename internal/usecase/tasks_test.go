package usecase_test

import (
	"bernard/internal/domain/entity"
	"bernard/internal/usecase"
	mocks "bernard/internal/usecase/mocks"
	"bernard/utils"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_CreateTask(t *testing.T) {
	type args struct {
		ctx     context.Context
		task    entity.Task
		tagsIDs []int64
	}

	testUserID := uuid.New()
	projectID := int64(10)
	groupID := int64(20)
	taskID := int64(100)

	tests := []struct {
		name      string
		args      args
		mockSetup func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args)
		want      *entity.Task
		wantErr   bool
	}{
		{
			name: "success with group and project",
			args: args{
				ctx: context.Background(),
				task: entity.Task{
					Name:        "test",
					Description: utils.Ptr("test"),
					GroupID:     &groupID,
					ProjectID:   &projectID,
					UserID:      testUserID,
				},
				tagsIDs: []int64{},
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckProjectOwnership", a.ctx, a.task.UserID, *a.task.ProjectID).Return(true, nil).Once()
				ms.On("CheckGroupOwnership", a.ctx, a.task.UserID, *a.task.GroupID).Return(true, nil).Once()
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				created := &entity.Task{ID: taskID, UserID: a.task.UserID}
				ms.On("CreateTask", mock.Anything, mock.Anything).Return(created, nil).Once()
				ms.On("AddTagsToTask", mock.Anything, a.tagsIDs, created.ID).Return(nil).Once()
			},
			want: &entity.Task{
				ID:     taskID,
				UserID: testUserID,
			},
			wantErr: false,
		},
		{
			name: "project permission error",
			args: args{
				ctx: context.Background(),
				task: entity.Task{
					Name:      "test",
					ProjectID: &projectID,
					UserID:    testUserID,
				},
				tagsIDs: []int64{},
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckProjectOwnership", a.ctx, a.task.UserID, *a.task.ProjectID).Return(false, errors.New("project permission denied")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "group permission error",
			args: args{
				ctx: context.Background(),
				task: entity.Task{
					Name:    "test",
					GroupID: &groupID,
					UserID:  testUserID,
				},
				tagsIDs: []int64{},
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckGroupOwnership", a.ctx, a.task.UserID, *a.task.GroupID).Return(false, errors.New("group permission denied")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "db error on creating task",
			args: args{
				ctx: context.Background(),
				task: entity.Task{
					Name:   "test",
					UserID: testUserID,
				},
				tagsIDs: []int64{},
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				ms.On("CreateTask", mock.Anything, a.task).Return(nil, errors.New("error with creating task")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "db error on adding tags to task",
			args: args{
				ctx: context.Background(),
				task: entity.Task{
					Name:   "test",
					UserID: testUserID,
				},
				tagsIDs: []int64{},
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				created := entity.Task{ID: taskID, UserID: testUserID}
				ms.On("CreateTask", mock.Anything, a.task).Return(&created, nil).Once()
				ms.On("AddTagsToTask", mock.Anything, a.tagsIDs, created.ID).Return(errors.New("error adding tags")).Once()
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockStorage := mocks.NewMockStorage(t)
			mockTx := mocks.NewMockTxManager(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage, mockTx, tt.args)
			}

			u := &usecase.UseCase{
				Log: log,
				DB:  mockStorage,
				Tx:  mockTx,
			}

			got, err := u.CreateTask(tt.args.ctx, tt.args.task, tt.args.tagsIDs)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUseCase_UpdateTask(t *testing.T) {
	type args struct {
		ctx    context.Context
		task   entity.UpdateTaskRequest
		userID uuid.UUID
		taskID int64
	}

	testUserID := uuid.New()
	projectID := int64(10)
	groupID := int64(20)
	tagsForUpdate := []int64{1}
	taskID := int64(100)
	updateName := "test"

	tests := []struct {
		name      string
		args      args
		mockSetup func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args)
		want      *entity.Task
		wantErr   bool
	}{
		{
			name: "success update with new tags",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					Name:      &updateName,
					ProjectID: &projectID,
					GroupID:   &groupID,
					TagsIDs:   &tagsForUpdate,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckProjectOwnership", a.ctx, a.userID, *a.task.ProjectID).Return(true, nil).Once()
				ms.On("CheckGroupOwnership", a.ctx, a.userID, *a.task.GroupID).Return(true, nil).Once()
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				updatedTask := entity.Task{ID: taskID, Name: updateName, UserID: testUserID}
				ms.On("UpdateTask", mock.Anything, a.task, a.userID, a.taskID).Return(&updatedTask, nil).Once()
				ms.On("RemoveAllTagsFromTask", mock.Anything, a.taskID).Return(nil).Once()
				ms.On("AddTagsToTask", mock.Anything, *a.task.TagsIDs, a.taskID).Return(nil).Once()
				ms.On("GetTaskTags", a.ctx, a.taskID).Return([]*entity.Tag{{ID: 1}}, nil).Once()
			},
			want: &entity.Task{
				ID:     taskID,
				Name:   updateName,
				UserID: testUserID,
				Tags:   []*entity.Tag{{ID: 1}},
			},
			wantErr: false,
		},
		{
			name: "success - no updated with new tags",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					TagsIDs: &tagsForUpdate,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				updatedTask := entity.Task{ID: taskID, UserID: testUserID}
				ms.On("GetTask", mock.Anything, a.taskID, a.userID).Return(&updatedTask, nil).Once()
				ms.On("RemoveAllTagsFromTask", mock.Anything, a.taskID).Return(nil).Once()
				ms.On("AddTagsToTask", mock.Anything, *a.task.TagsIDs, a.taskID).Return(nil).Once()
				ms.On("GetTaskTags", a.ctx, a.taskID).Return([]*entity.Tag{{ID: 1}}, nil).Once()
			},
			want: &entity.Task{
				ID:     taskID,
				UserID: testUserID,
				Tags:   []*entity.Tag{{ID: 1}},
			},
			wantErr: false,
		},
		{
			name: "success - update with no tags",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					Name: &updateName,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				updatedTask := entity.Task{ID: taskID, Name: updateName, UserID: testUserID}
				ms.On("UpdateTask", mock.Anything, a.task, a.userID, a.taskID).Return(&updatedTask, nil).Once()
				ms.On("GetTaskTags", a.ctx, a.taskID).Return([]*entity.Tag{{ID: 1}}, nil).Once()
			},
			want: &entity.Task{
				ID:     taskID,
				Name:   updateName,
				UserID: testUserID,
				Tags:   []*entity.Tag{{ID: 1}},
			},
			wantErr: false,
		},
		{
			name: "project permission denied",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					ProjectID: &projectID,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckProjectOwnership", a.ctx, a.userID, *a.task.ProjectID).Return(false, nil).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "group permission denied",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					GroupID: &groupID,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckGroupOwnership", a.ctx, a.userID, *a.task.GroupID).Return(false, nil).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "project permission error",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					ProjectID: &projectID,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckProjectOwnership", a.ctx, a.userID, *a.task.ProjectID).Return(false, errors.New("permission denied")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "group permission error",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					GroupID: &groupID,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				ms.On("CheckGroupOwnership", a.ctx, a.userID, *a.task.GroupID).Return(false, errors.New("permission denied")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "update task error",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					Name: &updateName,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				ms.On("UpdateTask", mock.Anything, a.task, a.userID, a.taskID).Return(nil, errors.New("db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get task error",
			args: args{
				ctx:    context.Background(),
				task:   entity.UpdateTaskRequest{},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				ms.On("GetTask", mock.Anything, a.taskID, a.userID).Return(nil, errors.New("db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "remove tags error",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					TagsIDs: &tagsForUpdate,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				task := entity.Task{ID: 1}
				ms.On("GetTask", mock.Anything, a.taskID, a.userID).Return(&task, nil).Once()
				ms.On("RemoveAllTagsFromTask", mock.Anything, a.taskID).Return(errors.New("db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "add tasks error",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					TagsIDs: &tagsForUpdate,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				task := entity.Task{ID: 1}
				ms.On("GetTask", mock.Anything, a.taskID, a.userID).Return(&task, nil).Once()
				ms.On("RemoveAllTagsFromTask", mock.Anything, a.taskID).Return(nil).Once()
				ms.On("AddTagsToTask", mock.Anything, *a.task.TagsIDs, a.taskID).Return(errors.New("db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get task tags error",
			args: args{
				ctx: context.Background(),
				task: entity.UpdateTaskRequest{
					TagsIDs: &tagsForUpdate,
				},
				userID: testUserID,
				taskID: taskID,
			},
			mockSetup: func(ms *mocks.MockStorage, mtx *mocks.MockTxManager, a args) {
				mtx.On("ReadWrite", a.ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					}).Once()
				task := entity.Task{ID: 1}
				ms.On("GetTask", mock.Anything, a.taskID, a.userID).Return(&task, nil).Once()
				ms.On("RemoveAllTagsFromTask", mock.Anything, a.taskID).Return(nil).Once()
				ms.On("AddTagsToTask", mock.Anything, *a.task.TagsIDs, a.taskID).Return(nil).Once()
				ms.On("GetTaskTags", a.ctx, a.taskID).Return(nil, errors.New("db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockStorage := mocks.NewMockStorage(t)
			mockTx := mocks.NewMockTxManager(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage, mockTx, tt.args)
			}

			u := &usecase.UseCase{
				Log: log,
				DB:  mockStorage,
				Tx:  mockTx,
			}

			got, err := u.UpdateTask(tt.args.ctx, tt.args.task, tt.args.userID, tt.args.taskID)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUseCase_DeleteTask(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		taskID int64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("DeleteTask", a.ctx, a.userID, a.taskID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("DeleteTask", a.ctx, a.userID, a.taskID).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		mockStorage := mocks.NewMockStorage(t)

		if tt.mockSetup != nil {
			tt.mockSetup(mockStorage, tt.args)
		}

		u := usecase.UseCase{
			Log: log,
			DB: mockStorage,
		}

		err := u.DeleteTask(tt.args.ctx, tt.args.userID, tt.args.taskID)

		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestUseCase_GetTask(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		taskID int64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want *entity.Task
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTask", a.ctx, a.taskID, a.userID).Return(&entity.Task{ID: 1}, nil).Once()
				m.On("GetTaskTags", a.ctx, a.taskID).Return([]*entity.Tag{{ID: 1}}, nil).Once()
			},
			want: &entity.Task{ID: 1, Tags: []*entity.Tag{{ID: 1}}},
			wantErr: false,
		},
		{
			name: "error getting task tags",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTask", a.ctx, a.taskID, a.userID).Return(&entity.Task{ID: 1}, nil).Once()
				m.On("GetTaskTags", a.ctx, a.taskID).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "error getting task",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTask", a.ctx, a.taskID, a.userID).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		mockStorage := mocks.NewMockStorage(t)

		if tt.mockSetup != nil {
			tt.mockSetup(mockStorage, tt.args)
		}

		u := usecase.UseCase{
			Log: log,
			DB: mockStorage,
		}

		got, err := u.GetTask(tt.args.ctx, tt.args.taskID, tt.args.userID)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestUseCase_GetListTask(t *testing.T) {
	type args struct {
		ctx        context.Context
		taskFilter entity.TasksFilter
	}

	tests := []struct {
		name      string
		args      args
		mockSetup func(m *mocks.MockStorage, a args)
		want      []*entity.Task
		wantErr   bool
	}{
		{
			name: "success non-filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("ListTasks", a.ctx, a.taskFilter.UserID, a.taskFilter.Limit, a.taskFilter.Offset).Return([]*entity.Task{}, nil).Once()
			},
			want:    []*entity.Task{},
			wantErr: false,
		},
		{
			name: "success tag-filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Tag: utils.Ptr(int64(5)),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTasksByTag", a.ctx, a.taskFilter.UserID, *a.taskFilter.Tag, a.taskFilter.Limit, a.taskFilter.Offset).Return([]*entity.Task{}, nil).Once()
			},
			want:    []*entity.Task{},
			wantErr: false,
		},
		{
			name: "success priority-filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Priority: utils.Ptr(5),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTasksByPriority", a.ctx, a.taskFilter.UserID, *a.taskFilter.Priority, a.taskFilter.Limit, a.taskFilter.Offset).Return([]*entity.Task{}, nil).Once()
			},
			want:    []*entity.Task{},
			wantErr: false,
		},
		{
			name: "success date-filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					From: utils.Ptr(time.Now()),
					To: utils.Ptr(time.Now()),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTasksByDate", a.ctx, a.taskFilter.UserID, *a.taskFilter.From, *a.taskFilter.To, a.taskFilter.Limit, a.taskFilter.Offset).Return([]*entity.Task{}, nil).Once()
			},
			want:    []*entity.Task{},
			wantErr: false,
		},
		{
			name: "error with userID",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.Nil,
					Limit:  uint64(5),
					Offset: uint64(0),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error with date to",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					From:   utils.Ptr(time.Now()),
					To:     nil,
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error tag filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Tag: utils.Ptr(int64(5)),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTasksByTag", a.ctx, a.taskFilter.UserID, *a.taskFilter.Tag, a.taskFilter.Limit, a.taskFilter.Offset).Return(nil, errors.New("tag filter db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error priority filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Priority: utils.Ptr(5),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTasksByPriority", a.ctx, a.taskFilter.UserID, *a.taskFilter.Priority, a.taskFilter.Limit, a.taskFilter.Offset).Return(nil, errors.New("priority filter db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error date filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					From: utils.Ptr(time.Now()),
					To: utils.Ptr(time.Now()),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTasksByDate", a.ctx, a.taskFilter.UserID, *a.taskFilter.From, *a.taskFilter.To, a.taskFilter.Limit, a.taskFilter.Offset).Return(nil, errors.New("date filter db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error no filter",
			args: args{
				ctx: context.Background(),
				taskFilter: entity.TasksFilter{
					UserID: uuid.New(),
					Limit:  uint64(5),
					Offset: uint64(0),
					FilterType: 0,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("ListTasks", a.ctx, a.taskFilter.UserID, a.taskFilter.Limit, a.taskFilter.Offset).Return(nil, errors.New("no filter db error")).Once()
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockStorage := mocks.NewMockStorage(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage, tt.args)
			}

			u := &usecase.UseCase{
				Log: log,
				DB:  mockStorage,
			}

			got, err := u.GetListTask(tt.args.ctx, tt.args.taskFilter)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUseCase_ArchiveTask(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		id int64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want *entity.Task
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				id: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("ArchiveTask", a.ctx, a.userID, a.id).Return(&entity.Task{ID: 1}, nil).Once()
			},
			want: &entity.Task{ID: 1},
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				id: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("ArchiveTask", a.ctx, a.userID, a.id).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		log := slog.New(slog.NewJSONHandler(io.Discard, nil))
		mockStorage := mocks.NewMockStorage(t)

		if tt.mockSetup != nil {
			tt.mockSetup(mockStorage, tt.args)
		}

		u := usecase.UseCase{
			Log: log,
			DB: mockStorage,
		}

		got, err := u.ArchiveTask(tt.args.ctx, tt.args.userID, tt.args.id)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}