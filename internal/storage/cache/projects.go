package cache

import (
	"bernard/internal/domain/entity"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (c *Cache) SetTempProject(ctx context.Context, userProject *entity.Project) error {
	const op = "cache.SetTempProject"

	projectGroups, err := json.Marshal(userProject.Groups)
	if err != nil {
		return fmt.Errorf("%s: cannot marshal data - %s", op, err)
	}

	err = c.DB.HSet(ctx, fmt.Sprintf("TempUserProject:%v:%v", userProject.ID, userProject.UserID), map[string]any{
		"id":          userProject.ID,
		"name":        userProject.Name,
		"description": userProject.Description,
		"groups":      projectGroups,
		"user_id":     userProject.UserID,
	}).Err()

	if err != nil {
		return fmt.Errorf("%s: cannot set data - %s", op, err)
	}

	err = c.DB.Expire(ctx, fmt.Sprintf("TempUserProject:%v:%v", userProject.ID, userProject.UserID), 2*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("%s: cannot set ttl - %s", op, err)
	}

	return nil
}

func (c *Cache) GetTempProject(ctx context.Context, userID uuid.UUID, projectID int64) (*entity.Project, error) {
	const op = "cache.GetTempProject"

	val, err := c.DB.HGetAll(ctx, fmt.Sprintf("TempUserProject:%v:%v", projectID, userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot get data - %s", op, err)
	}

	var project entity.Project

	err = json.Unmarshal([]byte(val["id"]), &project.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot unmarshal data - %s", op, err)
	}

	err = json.Unmarshal([]byte(val["name"]), &project.Name)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot unmarshal data - %s", op, err)
	}

	err = json.Unmarshal([]byte(val["description"]), &project.Description)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot unmarshal data - %s", op, err)
	}

	err = json.Unmarshal([]byte(val["groups"]), &project.Groups)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot unmarshal data - %s", op, err)
	}

	project.UserID = userID

	return &project, nil
}