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
					Return(nil).
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

// func TestUseCase_DeleteProject(t *testing.T) {
// 	type fields struct {
// 		log *slog.Logger
// 		db  *mocks.MockStorage
// 	}

// 	type args struct {
// 		ctx       context.Context
// 		projectID int64
// 		userID    uuid.UUID
// 	}

// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *entity.Project
// 		wantErr bool
// 	}{
// 		{
// 			name: "success",
// 			args: args{
// 				ctx:       context.Background(),
// 				projectID: 1,
// 				userID:    uuid.New(),
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "error",
// 			args: args{
// 				ctx:       context.Background(),
// 				projectID: -1,
// 				userID:    uuid.New(),
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
// 			mockStorage := mocks.NewMockStorage(t)

// 			mockStorage.On("DeleteProject", tt.args.ctx, tt.args.projectID).Maybe().Return(tt.want, nil)

// 			u := &usecase.UseCase{
// 				Log:    log,
// 				DB:     mockStorage,
// 			}

// 			err := u.DeleteProject(tt.args.ctx, tt.args.projectID, tt.args.userID)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DeleteProject() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			if tt.wantErr {
// 				require.Error(t, err)
// 			}
// 		})
// 	}
// }

// func TestUseCase_GetProjectTree(t *testing.T) {
// 	type fields struct {
// 		cfg *config.Config
// 		log *slog.Logger
// 		db  *mocks.MockStorage
// 	}

// 	type args struct {
// 		ctx       context.Context
// 		projectID int64
// 		userID    uuid.UUID
// 		limit     uint64
// 		offset    uint64
// 	}

// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *entity.Project
// 		wantErr bool
// 	}{
// 		{
// 			name: "success",
// 			args: args{
// 				ctx:       context.Background(),
// 				projectID: 1,
// 				userID:    uuid.New(),
// 			},
// 			want: &entity.Project{
// 				ID: 1,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "error",
// 			args: args{
// 				ctx:       context.Background(),
// 				projectID: -1,
// 				userID:    uuid.New(),
// 			},
// 			want:    nil,
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
// 			mockStorage := mocks.NewMockStorage(t)

// 			mockStorage.On("GetProject", tt.args.ctx, tt.args.projectID, mock.Anything).Maybe().Return(tt.want, nil)

// 			u := &usecase.UseCase{
// 				Log:    log,
// 				DB:     mockStorage,
// 			}

// 			got, err := u.GetProject(tt.args.ctx, tt.args.projectID, tt.args.userID, tt.args.limit, tt.args.offset)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetProjectTree() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			if tt.want != nil {
// 				require.Equal(t, tt.want, got)
// 			}

// 			if tt.wantErr {
// 				require.Error(t, err)
// 			}
// 		})
// 	}
// }

// func TestUseCase_GetProjects(t *testing.T) {
// 	type fields struct {
// 		cfg *config.Config
// 		log *slog.Logger
// 		db  *mocks.MockStorage
// 	}

// 	type args struct {
// 		ctx    context.Context
// 		userID uuid.UUID
// 		limit  uint64
// 		offset uint64
// 	}

// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    []*entity.Project
// 		wantErr bool
// 	}{
// 		{
// 			name: "success",
// 			args: args{
// 				ctx:    context.Background(),
// 				userID: uuid.New(),
// 			},
// 			want:    []*entity.Project{},
// 			wantErr: false,
// 		},
// 		{
// 			name: "error",
// 			args: args{
// 				ctx:    context.Background(),
// 				userID: uuid.Nil,
// 			},
// 			want:    nil,
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
// 			mockStorage := mocks.NewMockStorage(t)

// 			mockStorage.On("GetProjects", tt.args.ctx, tt.args.userID, tt.args.limit, tt.args.offset).Maybe().Return(tt.want, nil)

// 			u := &usecase.UseCase{
// 				Log:    log,
// 				DB:     mockStorage,
// 			}

// 			got, err := u.GetProjects(tt.args.ctx, tt.args.userID, tt.args.limit, tt.args.offset)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetProjects() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			if tt.want != nil {
// 				require.Equal(t, tt.want, got)
// 			}

// 			if tt.wantErr {
// 				require.Error(t, err)
// 			}
// 		})
// 	}
// }
