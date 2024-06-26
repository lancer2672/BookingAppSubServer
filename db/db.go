package db

import (
	"fmt"

	"github.com/lancer2672/BookingAppSubServer/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDatabase(config utils.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.DBSource), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	fmt.Println("Connected to db")
	// Assign the database instance to the store variable

	// Migrate the schema
	// db.AutoMigrate(
	// 	&User{},
	// 	&Property{},
	// 	&Room{},
	// 	&Amenity{},
	// 	&RoomImage{},
	// 	&PropertyAmenity{},
	// 	&PropertyImage{},
	// 	&Booking{},
	// 	&Province{},
	// 	&District{},
	// 	&Ward{},
	// )
	return db
}
