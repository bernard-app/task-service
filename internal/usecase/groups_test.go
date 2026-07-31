package usecase_test

import (
	"bernard/internal/domain/entity"
	"bernard/internal/usecase"

	"bernard/utils"
	"context"
	"errors"
	"io"
	"log/slog"

	"testing"

	mocks "bernard/internal/usecase/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUseCase_CreateGroup(t *testing.T) {
	type args struct {
		ctx   context.Context
		group entity.Group
	}

	tests := []struct {
		name    string
		args    args
		mockSetup func(m *mocks.MockStorage, a args)
		want    *entity.Group
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				group: entity.Group{
					Name:      "test_group",
					ProjectID: 1,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership",  a.ctx, a.group.UserID, a.group.ProjectID).
					Return(true, nil).
					Once()
				m.On("CreateGroup", a.ctx, a.group).
					Return(&entity.Group{
						Name: "test_group",
						TaskCount: 0,
						ProjectID: 1,
					}, nil).
					Once()
			},
			want: &entity.Group{
				Name:      "test_group",
				TaskCount: 0,
				ProjectID: 1,
			},
			wantErr: false,
		},
		{
			name: "ownership error",
			args: args{
				ctx: context.Background(),
				group: entity.Group{
					Name: "test_group",
					TaskCount: 0,
					ProjectID: 11,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.group.UserID, a.group.ProjectID).
					Return(false, errors.New("not your project")).
					Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				group: entity.Group{
					Name: "error_group",
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.group.UserID, a.group.ProjectID).
					Return(true, nil).
					Once()
				m.On("CreateGroup", a.ctx, a.group).
					Return(nil, errors.New("database error")).
					Once()
			},
			want: nil,
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
				Log:    log,
				DB:     mockStorage,
			}
			
			got, err := u.CreateGroup(tt.args.ctx, tt.args.group)

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

func TestUseCase_UpdateGroup(t *testing.T) {
	type args struct {
		ctx     context.Context
		group   entity.UpdateGroupRequest
		userID  uuid.UUID
		groupID int64
	}

	tests := []struct {
		name    string
		args    args
		mockSetup func(m *mocks.MockStorage, a args)
		want    *entity.Group
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				group: entity.UpdateGroupRequest{
					Name:      utils.Ptr("test"),
					ProjectID: utils.Ptr(int64(1)),
				},
				userID:  uuid.New(),
				groupID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.userID, *a.group.ProjectID).
					Return(true, nil).
					Once()
				m.On("UpdateGroup", a.ctx, a.group, a.userID, a.groupID).
					Return(&entity.Group{
						Name: "test",
						ProjectID: 1,
					}, nil).
					Once()
			},
			want: &entity.Group{
				Name:      "test",
				ProjectID: 1,
			},
			wantErr: false,
		},
		{
			name: "ownership error",
			args: args{
				ctx: context.Background(),
				group: entity.UpdateGroupRequest{
					Name: utils.Ptr("test"),
					ProjectID: utils.Ptr(int64(1)),
				},
				userID: uuid.New(),
				groupID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.userID, *a.group.ProjectID).
					Return(false, errors.New("not your project")).
					Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				group: entity.UpdateGroupRequest{
					Name: utils.Ptr("error_group"),
				},
				userID: uuid.Nil,
				groupID: 0,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("UpdateGroup", a.ctx, a.group, a.userID, a.groupID).
					Return(nil, errors.New("db error")).
					Once()
			},
			want: nil,
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
				Log:    log,
				DB:     mockStorage,
			}

			got, err := u.UpdateGroup(tt.args.ctx, tt.args.group, tt.args.userID, tt.args.groupID)

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

func TestUseCase_DeleteGroup(t *testing.T) {
	type args struct {
		ctx     context.Context
		groupID int64
		userID  uuid.UUID
	}

	tests := []struct {
		name    string
		mockSetup func(m *mocks.MockStorage, a args)
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				groupID: 1,
				userID:  uuid.New(),
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("DeleteGroup", a.ctx, a.groupID, a.userID).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "invalid group id",
			args: args{
				ctx:     context.Background(),
				groupID: 0,
				userID:  uuid.New(),
			},
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				groupID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("DeleteGroup", a.ctx, a.groupID, a.userID).
					Return(errors.New("db error")).
					Once()
			},
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
				Log:    log,
				DB:     mockStorage,
			}

			err := u.DeleteGroup(tt.args.ctx, tt.args.groupID, tt.args.userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCase_GetGroup(t *testing.T) {
	type args struct {
		ctx context.Context
		groupID int64 
		userID uuid.UUID
	}

	testUserID := uuid.New()
	groupID := int64(1)
	taskID := int64(1)
	
	tests := []struct {
		name string 
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want *entity.Group
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				groupID: groupID,
				userID: testUserID,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetGroup", a.ctx, a.userID, a.groupID).Return(&entity.Group{ID: 1}, nil).Once()
				m.On("GetTasksByGroupIDs", a.ctx, []int64{a.groupID}).Return([]*entity.Task{{ID: 1, GroupID: &a.groupID}}, nil).Once()
				m.On("GetTagsByTaskIDs", a.ctx, []int64{1}).Return([]*entity.Tag{{ID: 1, TaskID: 1}}, nil).Once()
			},
			want: &entity.Group{
				ID: 1,
				Tasks: []*entity.Task{{ID: 1, GroupID: &groupID, Tags: []*entity.Tag{{ID: 1, TaskID: taskID}}}},
			},
			wantErr: false,
		},
		{
			name: "error get group",
			args: args{
				ctx: context.Background(),
				groupID: groupID,
				userID: testUserID,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetGroup", a.ctx, a.userID, a.groupID).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "get tasks error",
			args: args{
				ctx: context.Background(),
				groupID: groupID,
				userID: testUserID,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetGroup", a.ctx, a.userID, a.groupID).Return(&entity.Group{ID: 1}, nil).Once()
				m.On("GetTasksByGroupIDs", a.ctx, []int64{a.groupID}).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "get tags error",
			args: args{
				ctx: context.Background(),
				groupID: groupID,
				userID: testUserID,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetGroup", a.ctx, a.userID, a.groupID).Return(&entity.Group{ID: 1}, nil).Once()
				m.On("GetTasksByGroupIDs", a.ctx, []int64{a.groupID}).Return([]*entity.Task{{ID: 1, GroupID: &a.groupID}}, nil).Once()
				m.On("GetTagsByTaskIDs", a.ctx, []int64{1}).Return(nil, errors.New("db error")).Once()
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

		got, err := u.GetGroup(tt.args.ctx, tt.args.groupID, tt.args.userID)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestUseCase_GetProjectGroups(t *testing.T) {
	type args struct {
		ctx context.Context
		projectID int64 
		userID uuid.UUID
		limit uint64
		offset uint64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want []*entity.Group
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
				limit: 1,
				offset: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.userID, a.projectID).Return(true, nil).Once()
				m.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return([]*entity.Group{{ID: 1}}, nil).Once()
			},
			want: []*entity.Group{{ID: 1}},
			wantErr: false,
		},
		{
			name: "permission error",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.userID, a.projectID).Return(true, errors.New("permission error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "permission denied",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.userID, a.projectID).Return(false, nil).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				projectID: 1,
				userID: uuid.New(),
				limit: 1,
				offset: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.userID, a.projectID).Return(true, nil).Once()
				m.On("GetProjectGroups", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return(nil, errors.New("db error")).Once()
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

		got, err := u.GetProjectGroups(tt.args.ctx, tt.args.projectID, tt.args.userID, tt.args.limit, tt.args.offset)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}