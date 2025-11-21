package models

import (
	"time"

	"gorm.io/gorm"
)

// UserPermission represents a permission granted to a user
type UserPermission struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	UserID       uint      `gorm:"not null;index" json:"user_id"`
	PermissionID uint      `gorm:"not null;index" json:"permission_id"`
	GrantedBy    *uint     `json:"granted_by,omitempty"`              // Who granted this permission
	GrantedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"granted_at"`

	// Relationships
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	GrantedByUser *User   `gorm:"foreignKey:GrantedBy" json:"granted_by_user,omitempty"`
}

// TableName specifies the table name for UserPermission model
func (UserPermission) TableName() string {
	return "public.user_permissions"
}
