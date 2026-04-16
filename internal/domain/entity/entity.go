package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Task struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        string    `json:"tags"`
	Priority    int       `json:"priority"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	Deadline    time.Time `json:"deadline"`
	GroupID     int64     `json:"group_id"`
	UserID      uuid.UUID `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Group struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Tasks     []Task `json:"tasks"`
	TaskCount int    `json:"task_count"`
	ProjectID int64  `json:"project_id"`
}

type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Groups      []Group   `json:"groups"`
	TaskCount   int       `json:"task_count"`
	UserID      uuid.UUID `json:"user_id"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
}
