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
			name: "color nil",
			args: args{
				ctx: context.Background(),
				tagID: 1,
			},
			want: nil,
			wantErr: true,
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