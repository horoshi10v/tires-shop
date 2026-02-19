package models

type Warehouse struct {
	Base
	Name     string `gorm:"type:varchar(100);not null"`
	Location string `gorm:"type:varchar(255)"`
	IsActive bool   `gorm:"default:true"`

	// One to many relationship with Lots
	Lots []Lot `gorm:"foreignKey:WarehouseID"`
}
