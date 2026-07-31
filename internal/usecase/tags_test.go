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

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUseCase_CreateTag(t *testing.T) {
	type args struct {
		ctx context.Context
		tag entity.Tag
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want *entity.Tag
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				tag: entity.Tag{
					ID: 1,
					UserID: uuid.New(),
					ProjectID: int64(1),
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.tag.UserID, a.tag.ProjectID).Return(true, nil).Once()
				m.On("CreateTag", a.ctx, a.tag).Return(&entity.Tag{ID: 1}, nil).Once()
			},
			want: &entity.Tag{ID: 1},
			wantErr: false,
		},
		{
			name: "permission error",
			args: args{
				ctx: context.Background(),
				tag: entity.Tag{
					ID: 1,
					UserID: uuid.New(),
					ProjectID: 1,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.tag.UserID, a.tag.ProjectID).Return(false, errors.New("permission denied")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				tag: entity.Tag{
					ID: 1,
					UserID: uuid.New(),
					ProjectID: 1,
				},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckProjectOwnership", a.ctx, a.tag.UserID, a.tag.ProjectID).Return(true, nil).Once()
				m.On("CreateTag", a.ctx, a.tag).Return(nil, errors.New("db error")).Once()
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

		u := &usecase.UseCase{
			Log: log,
			DB: mockStorage,
		}

		got, err := u.CreateTag(tt.args.ctx, tt.args.tag)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestUseCase_UpdateTag(t *testing.T) {
	type args struct {
		ctx context.Context
		tagID int64
		tagName *string
		tagColor *string
		userID uuid.UUID
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want *entity.Tag
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				tagID: int64(1),
				tagColor: utils.Ptr("#111111"),
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("UpdateTag", a.ctx, a.tagID, a.tagName, a.tagColor, a.userID).Return(&entity.Tag{ID: 1}, nil).Once()
			},
			want: &entity.Tag{ID: 1},
			wantErr: false,
		},
		{
			name: "invalid color",
			args: args{
				ctx: context.Background(),
				tagID: 1,
				tagColor: utils.Ptr("asdfasdf"),
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				tagID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("UpdateTag", a.ctx, a.tagID, a.tagName, a.tagColor, a.userID).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		mockStorage := mocks.NewMockStorage(t)

		if tt.mockSetup != nil {
			tt.mockSetup(mockStorage, tt.args)
		}

		u := &usecase.UseCase{
			DB: mockStorage,
		}

		got, err := u.UpdateTag(tt.args.ctx, tt.args.tagID, tt.args.tagName, tt.args.tagColor, tt.args.userID)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestUseCase_DeleteTag(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		tagsID int64
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
				tagsID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("DeleteTag", a.ctx, a.tagsID, a.userID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				tagsID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("DeleteTag", a.ctx, a.tagsID, a.userID).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		mockStorage := mocks.NewMockStorage(t)

		if tt.mockSetup != nil {
			tt.mockSetup(mockStorage, tt.args)
		}

		u := usecase.UseCase{
			DB: mockStorage,
		}

		err := u.DeleteTag(tt.args.ctx, tt.args.tagsID, tt.args.userID)

		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestUseCase_GetTag(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		tagID int64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want *entity.Tag
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				tagID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTag", a.ctx, a.tagID, a.userID).Return(&entity.Tag{ID: 1}, nil).Once()
			},
			want: &entity.Tag{ID: 1},
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				tagID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTag", a.ctx, a.tagID, a.userID).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		mockStorage := mocks.NewMockStorage(t)

		if tt.mockSetup != nil {
			tt.mockSetup(mockStorage, tt.args)
		}

		u := usecase.UseCase{
			DB: mockStorage,
		}

		got, err := u.GetTag(tt.args.ctx, tt.args.tagID, tt.args.userID)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestUseCase_GetTagList(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		projectID int64
		limit uint64
		offset uint64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want []*entity.Tag
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				projectID: 1,
				limit: 1,
				offset: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTagList", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return([]*entity.Tag{{ID: 1}}, nil).Once()
			},
			want: []*entity.Tag{{ID: 1}},
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				projectID: 1,
				limit: 1,
				offset: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("GetTagList", a.ctx, a.userID, a.projectID, a.limit, a.offset).Return(nil, errors.New("db error")).Once()
			},
			want: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		mockStorage := mocks.NewMockStorage(t)

		if tt.mockSetup != nil {
			tt.mockSetup(mockStorage, tt.args)
		}

		u := usecase.UseCase{
			DB: mockStorage,
		}

		got, err := u.GetTagList(tt.args.ctx, tt.args.userID, tt.args.projectID, tt.args.limit, tt.args.offset)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestUseCase_GetTaskTags(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		taskID int64
	}

	tests := []struct {
		name string
		args args
		mockSetup func(m *mocks.MockStorage, a args)
		want []*entity.Tag
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
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("GetTaskTags", a.ctx, a.taskID).Return([]*entity.Tag{{ID: 1}}, nil).Once()
			},
			want: []*entity.Tag{{ID: 1}},
			wantErr: false,
		},
		{
			name: "permission error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, errors.New("permission error")).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "permission denied",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(false, nil).Once()
			},
			want: nil,
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("GetTaskTags", a.ctx, a.taskID).Return(nil, errors.New("db error")).Once()
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

		got, err := u.GetTaskTags(tt.args.ctx, tt.args.userID, tt.args.taskID)

		if tt.wantErr {
			require.Error(t, err)
			require.Empty(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestUseCase_AddTagsTotask(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		tagsIDs []int64
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
				tagsIDs: []int64{1},
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("AddTagsToTask", a.ctx, a.tagsIDs, a.taskID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "permission error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(false, errors.New("permission denied")).Once()
			},
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				tagsIDs: []int64{1},
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("AddTagsToTask", a.ctx, a.tagsIDs, a.taskID).Return(errors.New("db error")).Once()
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

		err := u.AddTagsToTask(tt.args.ctx, tt.args.userID, tt.args.tagsIDs, tt.args.taskID)

		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestUseCase_RemoveTagsFromTask(t *testing.T) {
	type args struct {
		ctx context.Context
		userID uuid.UUID
		tagsIDs []int64
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
				tagsIDs: []int64{1},
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("RemoveTagsFromTask", a.ctx, a.tagsIDs, a.taskID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "permission error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(false, errors.New("permission denied")).Once()
			},
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
				tagsIDs: []int64{1},
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("RemoveTagsFromTask", a.ctx, a.tagsIDs, a.taskID).Return(errors.New("db error")).Once()
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

		err := u.RemoveTagsFromTask(tt.args.ctx, tt.args.userID, tt.args.tagsIDs, tt.args.taskID)

		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestUseCase_RemoveAllTagsFromtask(t *testing.T) {
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
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("RemoveAllTagsFromTask", a.ctx, a.taskID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "permission error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(false, errors.New("permission denied")).Once()
			},
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx: context.Background(),
				userID: uuid.New(),
				taskID: 1,
			},
			mockSetup: func(m *mocks.MockStorage, a args) {
				m.On("CheckTaskOwnership", a.ctx, a.userID, a.taskID).Return(true, nil).Once()
				m.On("RemoveAllTagsFromTask", a.ctx, a.taskID).Return(errors.New("db error")).Once()
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

		err := u.RemoveAllTagsFromTask(tt.args.ctx, tt.args.userID, tt.args.taskID)

		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}