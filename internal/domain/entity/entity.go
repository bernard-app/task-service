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

const (
	None = iota
	TagFilter
	PriorityFilter
	DateFilter
)

type Task struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Tags        []*Tag     `json:"tags"`
	Priority    *int       `json:"priority"`
	Status      *string    `json:"status"`
	StartTime   *time.Time `json:"start_time"`
	Deadline    *time.Time `json:"deadline"`
	GroupID     *int64     `json:"group_id"`
	ProjectID   *int64     `json:"project_id"`
	UserID      uuid.UUID  `json:"user_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	IsArchived  bool       `json:"is_archived"`
}

type Group struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Tasks     []*Task   `json:"tasks"`
	TaskCount int       `json:"task_count"`
	ProjectID int64     `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
}

type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Groups      []*Group  `json:"groups"`
	TaskCount   int       `json:"task_count"`
	UserID      uuid.UUID `json:"user_id"`
}

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	UserID    uuid.UUID `json:"user_id"`
	ProjectID int64     `json:"project_id"`
}

type UserTasksTab struct {
	Task        Task   `json:"task"`
	GroupName   string `json:"group_name"`
	ProjectName string `json:"project_name"`
	ProjectID   int64  `json:"project_id"`
}

type TasksFilter struct {
	UserID     uuid.UUID
	FilterType int
	Priority   *int
	Tag        *int64
	From       *time.Time
	To         *time.Time
	Limit      uint64
	Offset     uint64
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type UpdateTaskRequest struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	TagsIDs     *[]int64   `json:"tags"`
	Priority    *int       `json:"priority"`
	Status      *string    `json:"status"`
	GroupID     *int64     `json:"group_id"`
	ProjectID   *int64     `json:"project_id"`
	StartTime   *time.Time `json:"start_time"`
	Deadline    *time.Time `json:"deadline"`
}

type UpdateGroupRequest struct {
	Name      *string `json:"name"`
	ProjectID *int64  `json:"project_id"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
