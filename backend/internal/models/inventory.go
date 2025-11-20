package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents inventory items (dental materials, medicines, etc)
type Product struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Product info
	Name        string  `gorm:"not null" json:"name"`
	Code        string  `gorm:"index" json:"code"`
	Description string  `gorm:"type:text" json:"description"`
	Category    string  `json:"category"` // material, medicine, equipment, consumable

	// Supplier
	SupplierID  *uint    `json:"supplier_id"`
	Supplier    *Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`

	// Stock
	Quantity       int     `gorm:"default:0" json:"quantity"`
	MinimumStock   int     `gorm:"default:0" json:"minimum_stock"`
	Unit           string  `json:"unit"` // un, kg, ml, box, etc

	// Pricing
	CostPrice      float64 `json:"cost_price"`
	SalePrice      float64 `json:"sale_price"`

	// Validity
	ExpirationDate *time.Time `json:"expiration_date"`

	// Status
	Active         bool    `gorm:"default:true" json:"active"`

	// Barcode
	Barcode        string  `json:"barcode"`

	// Relationships
	Movements      []StockMovement `gorm:"foreignKey:ProductID" json:"movements,omitempty"`
}

// Supplier represents product suppliers
type Supplier struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name      string `gorm:"not null" json:"name"`
	CNPJ      string `gorm:"index" json:"cnpj"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`

	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`

	Active    bool   `gorm:"default:true" json:"active"`

	Notes     string `gorm:"type:text" json:"notes"`
}

// StockMovement represents stock entry/exit
type StockMovement struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProductID uint     `gorm:"not null;index" json:"product_id"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	Type      string `json:"type"` // entry, exit, adjustment
	Quantity  int    `json:"quantity"`
	Reason    string `json:"reason"` // purchase, sale, loss, adjustment, usage

	UserID    uint   `gorm:"not null;index" json:"user_id"`
	User      *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Notes     string `gorm:"type:text" json:"notes"`
}
