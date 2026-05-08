package cache

import (
	"bernard/internal/domain/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (c *Cache) SetTempProject(ctx context.Context, project *entity.Project) error {
	const op = "cache.SetTempProject"

	key := fmt.Sprintf("TempUserProject:%v:%v", project.ID, project.UserID)

	data, err := json.Marshal(project)
	if err != nil {
		return fmt.Errorf("%s: error marshal project", op)
	}

	return c.DB.Set(ctx, key, data, 60*time.Minute).Err()
}

func (c *Cache) GetTempProject(ctx context.Context, userID uuid.UUID, projectID int64) (*entity.Project, error) {
	const op = "cache.GetTempProject"
	
	key := fmt.Sprintf("TempUserProject:%v:%v", projectID, userID)

	val, err := c.DB.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%s: cache miss", op) 
		}

		return nil, fmt.Errorf("%s: cannot get data from redis - %w", op, err)
	}

	var project entity.Project

	err = json.Unmarshal(val, &project)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot unmarshal project tree - %w", op, err)
	}

	project.UserID = userID

	return &project, nil
}

func (c *Cache) InvalidateProjectCache(ctx context.Context, userID uuid.UUID, projectID int64) error {
    key := fmt.Sprintf("TempUserProject:%v:%v", projectID, userID)
    
    return c.DB.Del(ctx, key).Err()
}