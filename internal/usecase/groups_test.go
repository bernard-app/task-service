package usecase_test

import (
	"bernard/internal/config"
	"bernard/internal/domain/entity"
	"bernard/internal/usecase"
	"bernard/utils"
	"context"
	"log/slog"
	"os"
	"testing"

	mocks "bernard/internal/usecase/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUseCase_CreateGroup(t *testing.T) {
	type fields struct {
		config *config.Config
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx   context.Context
		group entity.Group
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.Group
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				group: entity.Group{
					Name:      "test_group",
					Tasks:     []*entity.Task{},
					TaskCount: 0,
					ProjectID: 1,
				},
			},
			want: &entity.Group{
				Name:      "test_group",
				Tasks:     []*entity.Task{},
				TaskCount: 0,
				ProjectID: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("CreateGroup", tt.args.ctx, tt.args.group).Return(tt.want, nil)

			u := &usecase.UseCase{
				Config: cfg,
				Log:    log,
				DB:     mockStorage,
			}

			got, err := u.CreateGroup(tt.args.ctx, tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGroup() error = %v, wantErr %v", err, tt.wantErr)
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

func TestUseCase_UpdateGroup(t *testing.T) {
	type fields struct {
		config *config.Config
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx     context.Context
		group   entity.UpdateGroupRequest
		userID  uuid.UUID
		groupID int64
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
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
			want: &entity.Group{
				Name:      "test",
				ProjectID: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("UpdateGroup", tt.args.ctx, tt.args.group).Return(tt.want, nil)

			u := &usecase.UseCase{
				Config: cfg,
				Log:    log,
				DB:     mockStorage,
			}

			got, err := u.UpdateGroup(tt.args.ctx, tt.args.group, tt.args.userID, tt.args.groupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateGroup() error = %v, wantErr %v", err, tt.wantErr)
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

func TestUseCase_DeleteGroup(t *testing.T) {
	type fields struct {
		config *config.Config
		log    *slog.Logger
		db     *mocks.MockStorage
	}

	type args struct {
		ctx     context.Context
		groupID int64
		userID  uuid.UUID
	}

	tests := []struct {
		name    string
		fields  fields
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockStorage := mocks.NewMockStorage(t)

			mockStorage.On("DeleteGroup", tt.args.ctx, tt.args.groupID).Maybe().Return(nil)

			u := &usecase.UseCase{
				Config: cfg,
				Log:    log,
				DB:     mockStorage,
			}

			err := u.DeleteGroup(tt.args.ctx, tt.args.groupID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteGroup() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}
