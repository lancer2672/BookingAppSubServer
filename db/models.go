package db

import (
	"database/sql"
	"time"
)

// BaseModel struct definition with common fields
type BaseModel struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// User struct definition with embedded BaseModel
type User struct {
	BaseModel
	FirstName   string  `gorm:"type:varchar(100)"`
	LastName    string  `gorm:"type:varchar(100)"`
	Email       *string `gorm:"type:varchar(100);uniqueIndex"`
	PhoneNumber string  `gorm:"type:varchar(15)"`
	Role        string  `gorm:"type:varchar(50)"`
	Avatar      string  `gorm:"type:varchar(255)"`
	Status      string  `gorm:"type:varchar(50)"`
	Password    string  `gorm:"type:varchar(255)"`
}

// Property struct definition with embedded BaseModel
type Property struct {
	BaseModel
	Name         string         `gorm:"type:varchar(100)"`
	FkWardId     uint           `gorm:"not null"`
	FkDistrictId uint           `gorm:"not null"`
	FkProvinceId uint           `gorm:"not null"`
	Description  sql.NullString `gorm:"type:text"`
	Longitude    sql.NullFloat64
	Latitude     sql.NullFloat64
	Address      string `gorm:"type:varchar(255)"`
	FkAgentId    uint   `gorm:"not null"`
	Status       string `gorm:"type:varchar(50)"`
	Type         string `gorm:"type:varchar(50)"`
}

// Room struct definition with embedded BaseModel
type Room struct {
	BaseModel
	FkPropertyId uint   `gorm:"not null"`
	Name         string `gorm:"type:varchar(100)"`
	Status       string `gorm:"type:varchar(50)"`
	Price        uint   `gorm:"not null"`
}

// Amenity struct definition with embedded BaseModel
type Amenity struct {
	BaseModel
	Name      string `gorm:"type:varchar(100)"`
	Type      string `gorm:"type:varchar(50)"`
	IsDeleted bool   `gorm:"default:false"`
}

// RoomImage struct definition with embedded BaseModel
type RoomImage struct {
	BaseModel
	Url      string `gorm:"type:varchar(255)"`
	FkRoomId uint   `gorm:"not null"`
}

// PropertyAmenity struct definition with embedded BaseModel
type PropertyAmenity struct {
	BaseModel
	FkPropertyId uint `gorm:"not null"`
	FkAmenityId  uint `gorm:"not null"`
}

// PropertyImage struct definition with embedded BaseModel
type PropertyImage struct {
	BaseModel
	Url          string `gorm:"type:varchar(255)"`
	FkPropertyId uint   `gorm:"not null"`
}

// Booking struct definition with embedded BaseModel
type Booking struct {
	BaseModel
	FkUserId   uint      `gorm:"not null"`
	FkRoomId   uint      `gorm:"not null"`
	Status     string    `gorm:"type:varchar(50)"`
	StartDate  time.Time `gorm:"not null"`
	EndDate    time.Time `gorm:"not null"`
	TotalPrice float64   `gorm:"not null"`
}

// Province struct definition with embedded BaseModel
type Province struct {
	BaseModel
	ProvinceName string `gorm:"type:varchar(100)"`
	ProvinceType string `gorm:"type:varchar(50)"`
}

// District struct definition with embedded BaseModel
type District struct {
	BaseModel
	DistrictName string `gorm:"type:varchar(100)"`
	DistrictType string `gorm:"type:varchar(50)"`
	Latitude     float64
	Longitude    float64
	ProvinceId   uint `gorm:"not null"`
}

// Ward struct definition with embedded BaseModel
type Ward struct {
	BaseModel
	WardName     string `gorm:"type:varchar(100)"`
	WardType     string `gorm:"type:varchar(50)"`
	FkDistrictId uint   `gorm:"not null"`
}
