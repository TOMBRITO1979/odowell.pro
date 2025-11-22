package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	DueDate     *time.Time     `json:"due_date"`
	Priority    string         `gorm:"size:20;default:'medium'" json:"priority"` // low, medium, high, urgent
	Status      string         `gorm:"size:20;default:'pending'" json:"status"`  // pending, in_progress, completed, cancelled
	CreatedBy   uint           `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Creator      User             `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Assignments  []TaskAssignment `gorm:"foreignKey:TaskID" json:"assignments,omitempty"`
	Responsibles []TaskUser       `gorm:"foreignKey:TaskID" json:"responsibles,omitempty"`
}

// TaskAssignment represents the polymorphic relationship between tasks and other entities
type TaskAssignment struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	TaskID         uint           `gorm:"not null;index" json:"task_id"`
	AssignableType string         `gorm:"size:50;not null;index" json:"assignable_type"` // patient, supplier, user, etc.
	AssignableID   uint           `gorm:"not null;index" json:"assignable_id"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Task Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// TaskUser represents multiple users responsible for a task
type TaskUser struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	TaskID    uint           `gorm:"not null;index" json:"task_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Task Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for Task
func (Task) TableName() string {
	return "tasks"
}

// TableName specifies the table name for TaskAssignment
func (TaskAssignment) TableName() string {
	return "task_assignments"
}

// TableName specifies the table name for TaskUser
func (TaskUser) TableName() string {
	return "task_users"
}
