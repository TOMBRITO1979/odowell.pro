package models

import (
	"time"

	"gorm.io/gorm"
)

// Permission represents an action that can be performed on a module
type Permission struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	ModuleID    uint   `gorm:"not null;index" json:"module_id"`
	Action      string `gorm:"not null;size:20" json:"action"`        // 'view', 'create', 'edit', 'delete'
	Description string `gorm:"size:200" json:"description,omitempty"`

	// Relationships
	Module          Module           `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	UserPermissions []UserPermission `gorm:"foreignKey:PermissionID" json:"-"`
}

// TableName specifies the table name for Permission model
func (Permission) TableName() string {
	return "public.permissions"
}

// Permission actions constants
const (
	ActionView   = "view"
	ActionCreate = "create"
	ActionEdit   = "edit"
	ActionDelete = "delete"
)
