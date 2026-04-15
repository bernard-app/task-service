package entity

type Task struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Priority    int    `json:"priority"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	GroupID     uint64 `json:"group_id"`
	UserID      uint64 `json:"user_id"`
}

type Group struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Tasks     []Task `json:"tasks"`
	ProjectID uint64 `json:"project_id"`
}

type Project struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Groups      []Group `json:"groups"`
}

type User struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}
