package usecase_test

import (
	"bernard/internal/domain/entity"
	"bernard/internal/usecase"
	mocks "bernard/internal/usecase/mocks"
	"errors"
	"io"

	"bernard/utils"
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_CreateProject(t *testing.T) {
	type args struct {
		ctx     context.Context
		project entity.Project
	}

	tests := []struct {
		name    string
		args    args
		mockSetup func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args)
		want *entity.Project
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				project: entity.Project{
					Name: "test",
				},
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("CreateProject", a.ctx, a.project).
					Return(&entity.Project{
						Name: "test",
					}, nil).Once()
				mr.On("InvalidateProjectCache", a.ctx, a.project.UserID, a.project.ID).
					Return(nil).
					Once()
			},
			want: &entity.Project{
				Name: "test",
			},
			wantErr: false,
		},
		{
			name: "bd error",
			args: args{
				ctx: context.Background(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("CreateProject", a.ctx, a.project).
					Return(nil, errors.New("db error")).
					Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "redis error",
			args: args{
				ctx: context.Background(),
				project: entity.Project{
					Name: "test",
				},
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("CreateProject", a.ctx, a.project).
					Return(&entity.Project{
						Name: "test",
					}, nil).
					Once()
				mr.On("InvalidateProjectCache", a.ctx, a.project.UserID, a.project.ID).
					Return(errors.New("redis error")).
					Once()
			},
			want: &entity.Project{
				Name: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			mockRedis := mocks.NewMockRedis(t)
			mockStorage := mocks.NewMockStorage(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage, mockRedis, tt.args)
			}

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
				Redis: mockRedis,
			}

			got, err := u.CreateProject(tt.args.ctx, tt.args.project)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUseCase_UpdateProject(t *testing.T) {
	type args struct {
		ctx       context.Context
		project   entity.UpdateProjectRequest
		projectID int64
		userID    uuid.UUID
	}

	tests := []struct {
		name    string
		args    args
		mockSetup func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args)
		want    *entity.Project
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				project: entity.UpdateProjectRequest{
					Name:        utils.Ptr("test"),
					Description: utils.Ptr("test"),
				},
				projectID: 1,
				userID:    uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("UpdateProject", a.ctx, a.project, a.projectID, a.userID).
					Return(&entity.Project{
						Name: "test",
						Description: utils.Ptr("test"),
					}, nil).
					Once()
				mr.On("InvalidateProjectCache", a.ctx, a.userID, a.projectID).
					Return(nil).
					Once()
			},
			want: &entity.Project{
				Name:        "test",
				Description: utils.Ptr("test"),
			},
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				project: entity.UpdateProjectRequest{},
				projectID: 0,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("UpdateProject", a.ctx, a.project, a.projectID, a.userID).
					Return(nil, errors.New("db error")).
					Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "redis error",
			args: args{
				ctx: context.Background(),
				project: entity.UpdateProjectRequest{
					Name:        utils.Ptr("test"),
					Description: utils.Ptr("test"),
				},
				projectID: 1,
				userID:    uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("UpdateProject", a.ctx, a.project, a.projectID, a.userID).
					Return(&entity.Project{
						Name: "test",
						Description: utils.Ptr("test"),
					}, nil).
					Once()
				mr.On("InvalidateProjectCache", a.ctx, a.userID, a.projectID).
					Return(errors.New("redis error")).
					Once()
			},
			want: &entity.Project{
				Name:        "test",
				Description: utils.Ptr("test"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			
			mockStorage := mocks.NewMockStorage(t)
			mockRedis := mocks.NewMockRedis(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage, mockRedis, tt.args)
			}
			
			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
				Redis: mockRedis,
			}

			got, err := u.UpdateProject(tt.args.ctx, tt.args.project, tt.args.projectID, tt.args.userID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUseCase_DeleteProject(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectID int64
		userID    uuid.UUID
	}

	tests := []struct {
		name    string
		args    args
		mockSetup func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:       context.Background(),
				projectID: 1,
				userID:    uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("DeleteProject", a.ctx, a.projectID, a.userID).Return(nil).Once()
				mr.On("InvalidateProjectCache", a.ctx, a.userID, a.projectID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "invalid id error",
			args: args{
				ctx:       context.Background(),
				projectID: -1,
				userID:    uuid.New(),
			},
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("DeleteProject", a.ctx, a.projectID, a.userID).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
		{
			name: "cache error",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				ms.On("DeleteProject", a.ctx, a.projectID, a.userID).Return(nil).Once()
				mr.On("InvalidateProjectCache", a.ctx, a.userID, a.projectID).Return(errors.New("cache error")).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockStorage := mocks.NewMockStorage(t)
			mockRedis := mocks.NewMockRedis(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage, mockRedis, tt.args)
			}

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
				Redis: mockRedis,
			}

			err := u.DeleteProject(tt.args.ctx, tt.args.projectID, tt.args.userID)
			
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCase_GetProject(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectID int64
		userID    uuid.UUID
		limit     uint64
		offset    uint64
	}

	tests := []struct {
		name    string
		args    args
		mockSetup func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args)
		want    *entity.Project
		wantErr bool
	}{
		{ 
			name: "success cache", 
			args: args{ 
				ctx:       context.Background(),
				projectID: 1,
				userID:    uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(&entity.Project{ID: 1}, nil).Once()
			},
			want: &entity.Project{
				ID: 1,
			},
			wantErr: false,
		},
		{
			name: "success no cache found",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(nil, errors.New("no cache")).Once()
				ms.On("GetProject", a.ctx, a.projectID, a.userID).Return(&entity.Project{ID: 1}, nil).Once()
				ms.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return([]*entity.Group{{ID: 1}}, nil).Once()
				ms.On("GetTasksByGroupIDs", a.ctx, []int64{1}).Return([]*entity.Task{{ID: 1}}, nil).Once()
				ms.On("GetTagsByTaskIDs", a.ctx, []int64{1}).Return([]*entity.Tag{{ID: 1}}, nil).Once()
				mr.On("SetTempProject", a.ctx, mock.AnythingOfType("*entity.Project")).Return(nil).Once()
			},
			want: &entity.Project{ID: 1, Groups: []*entity.Group{{ID: 1}}},
			wantErr: false,
		},
		{
			name: "error get project",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(nil, errors.New("no cache")).Once()
				ms.On("GetProject", a.ctx, a.projectID, a.userID).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "error get project groups",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(nil, errors.New("no cache")).Once()
				ms.On("GetProject", a.ctx, a.projectID, a.userID).Return(&entity.Project{ID: 1}, nil).Once()
				ms.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "error get group tasks",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(nil, errors.New("no cache")).Once()
				ms.On("GetProject", a.ctx, a.projectID, a.userID).Return(&entity.Project{ID: 1}, nil).Once()
				ms.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return([]*entity.Group{{ID: 1}}, nil).Once()
				ms.On("GetTasksByGroupIDs", a.ctx, []int64{1}).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "error get tasks tags",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(nil, errors.New("no cache")).Once()
				ms.On("GetProject", a.ctx, a.projectID, a.userID).Return(&entity.Project{ID: 1}, nil).Once()
				ms.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return([]*entity.Group{{ID: 1}}, nil).Once()
				ms.On("GetTasksByGroupIDs", a.ctx, []int64{1}).Return([]*entity.Task{{ID: 1}}, nil).Once()
				ms.On("GetTagsByTaskIDs", a.ctx, []int64{1}).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "success with error saving cache",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(nil, errors.New("no cache")).Once()
				ms.On("GetProject", a.ctx, a.projectID, a.userID).Return(&entity.Project{ID: 1}, nil).Once()
				ms.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return([]*entity.Group{{ID: 1}}, nil).Once()
				ms.On("GetTasksByGroupIDs", a.ctx, []int64{1}).Return([]*entity.Task{{ID: 1}}, nil).Once()
				ms.On("GetTagsByTaskIDs", a.ctx, []int64{1}).Return([]*entity.Tag{{ID: 1}}, nil).Once()
				mr.On("SetTempProject", a.ctx, mock.AnythingOfType("*entity.Project")).Return(errors.New("redis error")).Once()
			},
			want: &entity.Project{ID: 1, Groups: []*entity.Group{{ID: 1}}},
			wantErr: false,
		},
		{
			name: "success with task groups",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(ms *mocks.MockStorage, mr *mocks.MockRedis, a args) {
				mr.On("GetTempProject", a.ctx, a.userID, a.projectID).Return(nil, errors.New("no cache")).Once()
				ms.On("GetProject", a.ctx, a.projectID, a.userID).Return(&entity.Project{ID: 1}, nil).Once()
				ms.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return([]*entity.Group{{ID: 1}}, nil).Once()
				ms.On("GetTasksByGroupIDs", a.ctx, []int64{1}).Return([]*entity.Task{{ID: 1, GroupID: utils.Ptr(int64(1))}}, nil).Once()
				ms.On("GetTagsByTaskIDs", a.ctx, []int64{1}).Return([]*entity.Tag{{ID: 1}}, nil).Once()
				mr.On("SetTempProject", a.ctx, mock.AnythingOfType("*entity.Project")).Return(errors.New("redis error")).Once()
			},
			want: &entity.Project{ID: 1, Groups: []*entity.Group{{ID: 1, Tasks: []*entity.Task{{ID: 1, GroupID: utils.Ptr(int64(1))}}}}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			mockStorage := mocks.NewMockStorage(t)
			mockRedis := mocks.NewMockRedis(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage, mockRedis, tt.args)
			}

			u := &usecase.UseCase{
				Log:    log,
				DB:     mockStorage,
				Redis: mockRedis,
			}

			got, err := u.GetProject(tt.args.ctx, tt.args.projectID, tt.args.userID, tt.args.limit, tt.args.offset)
			
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUseCase_GetProjects(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		limit uint64
		offset uint64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want []*entity.Project
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				limit: 1,
				offset: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetProjects", a.ctx, a.userID, a.limit, a.offset).Return([]*entity.Project{{ID: 1}}, nil).Once()
			},
			want: []*entity.Project{{ID: 1}},
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				limit: 1,
				offset: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetProjects", a.ctx, a.userID, a.limit, a.offset).Return(nil, errors.New("db error")).Once()
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

		got, err := u.GetProjects(tt.args.ctx, tt.args.userID, tt.args.limit, tt.args.offset)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}