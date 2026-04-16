package cache

import (
	"bernard/internal/domain/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TempUserTasks struct {
	UserID uuid.UUID      `json:"user_id"`
	Tasks  []*entity.Task `json:"tasks"`
}

func (s *Cache) SetTempTasks(ctx context.Context, userTasks *TempUserTasks) error {
	const op = "cache.SetTempTasks"

	tasks, err := json.Marshal(userTasks.Tasks)
	if err != nil {
		return errors.New(fmt.Sprintf("%s: cannot marshal data - %s", op, err))
	}

	err = s.DB.HSet(ctx, fmt.Sprintf("TempUserTasks:%v", userTasks.UserID), map[string]any{
		"user_id": userTasks.UserID,
		"tasks":   tasks,
	}).Err()

	if err != nil {
		return errors.New(fmt.Sprintf("%s: cannot set data - %s", op, err))
	}

	err = s.DB.Expire(ctx, fmt.Sprintf("TempUserTasks:%v", userTasks.UserID), time.Hour).Err()
	if err != nil {
		return errors.New(fmt.Sprintf("%s: cannot set ttl - %s", op, err))
	}

	return nil
}

func (s *Cache) GetTempTasks(ctx context.Context, userID uuid.UUID) (*TempUserTasks, error) {
	const op = "cache.GetTempTasks"

	val, err := s.DB.HGetAll(ctx, fmt.Sprintf("TempUserTasks:%v", userID)).Result()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot get data - %s", op, err))
	}

	var tasks []*entity.Task

	err = json.Unmarshal([]byte(val["tasks"]), &tasks)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot unmarshal data - %s", op, err))
	}

	tempUserTasks := &TempUserTasks{
		UserID: userID,
		Tasks:  tasks,
	}

	return tempUserTasks, nil
}
