package models

import (
	"time"

	"gorm.io/gorm"
)

// Module represents a system module that can have permissions
type Module struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Code        string `gorm:"unique;not null;size:50" json:"code"`        // Ex: 'patients', 'appointments'
	Name        string `gorm:"not null;size:100" json:"name"`              // Ex: 'Pacientes', 'Agenda'
	Description string `json:"description,omitempty"`
	Icon        string `gorm:"size:50" json:"icon,omitempty"`              // Ex: 'UserOutlined'
	Active      bool   `gorm:"default:true" json:"active"`

	// Relationships
	Permissions []Permission `gorm:"foreignKey:ModuleID" json:"permissions,omitempty"`
}

// TableName specifies the table name for Module model
func (Module) TableName() string {
	return "public.modules"
}
