package db

import (
	"database/sql"
	"time"
)

// User struct definition with embedded
type T_Users struct {
	Id           uint    `json:"id"`
	First_Name   string  `gorm:"type:varchar(100)" json:"first_name"`
	Last_Name    string  `gorm:"type:varchar(100)" json:"last_name"`
	Email        *string `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Phone_Number string  `gorm:"type:varchar(15)" json:"phone_number"`
	Role         string  `gorm:"type:varchar(50)" json:"role"`
	Avatar       string  `gorm:"type:varchar(255)" json:"avatar"`
	Status       string  `gorm:"type:varchar(50)" json:"status"`
	Password     string  `gorm:"type:varchar(255)" json:"password"`
}

type T_Argents struct {
	Id                  uint   `json:"id"`
	Fk_User_Id          uint   `json:"fk_user_id"`
	Identity_Number     string `gorm:"type:varchar(15)" json:"identity_number"`
	Front_Identity_Card string `gorm:"type:varchar(15)" json:"front_identity_card"`
	Back_Identity_Card  string `gorm:"type:varchar(15)" json:"back_identity_card"`
	Selfie_Img          string `gorm:"type:varchar(15)" json:"selfie_img"`
}

// Property struct definition with embedded
type T_Properties struct {
	Id              uint            `json:"id"`
	Name            string          `gorm:"type:varchar(100)" json:"name"`
	Deposit_Percent float64         ` json:"deposit_percent"`
	Fk_Ward_Id      uint            `gorm:"not null" json:"fk_ward_id"`
	Fk_District_Id  uint            `gorm:"not null" json:"fk_district_id"`
	Fk_Province_Id  uint            `gorm:"not null" json:"fk_province_id"`
	Description     sql.NullString  `gorm:"type:text" json:"description"`
	Longitude       sql.NullFloat64 `json:"longitude"`
	Latitude        sql.NullFloat64 `json:"latitude"`
	Address         string          `gorm:"type:varchar(255)" json:"address"`
	Fk_Argent_Id    uint            `gorm:"not null" json:"fk_argent_id"`
	Status          string          `gorm:"type:varchar(50)" json:"status"`
	Type            string          `gorm:"type:varchar(50)" json:"type"`
}

// Room struct definition with embedded
type T_Rooms struct {
	Id             uint   `json:"id"`
	Fk_Property_Id uint   `gorm:"not null" json:"fk_property_id"`
	Name           string `gorm:"type:varchar(100)" json:"name"`
	Status         string `gorm:"type:varchar(50)" json:"status"`
	Price          uint   `gorm:"not null" json:"price"`
}
type T_Agent_Staffs struct {
	Id       uint `json:"id"`
	Agent_Id uint `json:"agent_id"`
	Staff_Id uint `json:"staff_id"`
}
type T_Booking_Deposits struct {
	ID            uint    `json:"id"`
	Fk_Booking_ID uint    `gorm:"not null" json:"fk_booking_id"`
	Image         *string `gorm:"type:varchar(255)" json:"image"`
	Deposit       float64 `gorm:"not null" json:"deposit"`
}

type T_Banks struct {
	ID             uint      `json:"id"`
	Bank_Name      string    `" json:"bank_name"`
	Account_Number string    ` json:"image"`
	QR_Code        *string   ` json:"deposit"`
	Fk_Argent_Id   uint      `json:"fk_argent_id"`
	Is_Default     bool      `json:"is_default"`
	Create_At      time.Time `json:"create_at"`
	Account_Name   string    ` json:"account_name"`
}

// Amenity struct definition with embedded
type T_Amenities struct {
	Id         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string `gorm:"type:varchar(100)" json:"name"`
	Type       string `gorm:"type:varchar(50)" json:"type"`
	Is_Deleted bool   `gorm:"default:false" json:"is_deleted"`
}

// RoomImage struct definition with embedded
type T_Room_Images struct {
	Id         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Url        string `gorm:"type:varchar(255)" json:"url"`
	Fk_Room_Id uint   `gorm:"not null" json:"fk_room_id"`
}

// PropertyAmenity struct definition with embedded
type T_Room_Amenities struct {
	Id            uint `gorm:"primaryKey;autoIncrement" json:"id"`
	Fk_Room_Id    uint `gorm:"not null" json:"fk_room_id"`
	Fk_Amenity_Id uint `gorm:"not null" json:"fk_amenity_id"`
}

// PropertyAmenity struct definition with embedded
type T_Property_Amenities struct {
	Id             uint `gorm:"primaryKey;autoIncrement" json:"id"`
	Fk_Property_Id uint `gorm:"not null" json:"fk_property_id"`
	Fk_Amenity_Id  uint `gorm:"not null" json:"fk_amenity_id"`
}

// PropertyImage struct definition with embedded
type T_Property_Images struct {
	Id             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Url            string `gorm:"type:varchar(255)" json:"url"`
	Fk_Property_Id uint   `gorm:"not null" json:"fk_property_id"`
}

// Booking struct definition with embedded
type T_Bookings struct {
	Id             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Fk_User_Id     uint      `gorm:"not null" json:"fk_user_id"`
	Status         string    `gorm:"type:varchar(50)" json:"status"`
	Start_Date     time.Time ` json:"start_date"`
	End_Date       time.Time ` json:"end_date"`
	Create_At      time.Time ` json:"create_at"`
	Total_Price    float64   `gorm:"not null" json:"total_price"`
	Fk_Property_Id uint      `gorm:"not null" json:"fk_property_id"`
}
type T_Booking_Rooms struct {
	Id            uint ` json:"id"`
	Fk_Room_Id    uint `gorm:"not null" json:"fk_room_id"`
	Fk_Booking_id uint `gorm:"not null" json:"fk_booking_id"`
}

// Province struct definition with embedded
type T_Provinces struct {
	Id            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Province_Name string `gorm:"type:varchar(100)" json:"province_name"`
	Province_Type string `gorm:"type:varchar(50)" json:"province_type"`
}

// District struct definition with embedded
type T_Districts struct {
	Id            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	District_Name string  `gorm:"type:varchar(100)" json:"district_name"`
	District_Type string  `gorm:"type:varchar(50)" json:"district_type"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Province_Id   uint    `gorm:"not null" json:"province_id"`
}

// Ward struct definition with embedded
type T_Wards struct {
	Id             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Ward_Name      string `gorm:"type:varchar(100)" json:"ward_name"`
	Ward_Type      string `gorm:"type:varchar(50)" json:"ward_type"`
	Fk_District_Id uint   `gorm:"not null" json:"fk_district_id"`
}

type T_User_Web_Session struct {
	Id             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Ward_Name      string `gorm:"type:varchar(100)" json:"ward_name"`
	Ward_Type      string `gorm:"type:varchar(50)" json:"ward_type"`
	Fk_District_Id uint   `gorm:"not null" json:"fk_district_id"`
}
