package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lancer2672/BookingAppSubServer/db"

	"github.com/gin-gonic/gin"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"
)

type bookingRequest struct {
	// TODO: Retrieve from token
	UserId     uint      `form:"userId" binding:"required"`
	RoomIds    []uint    `form:"roomIds" binding:"required"`
	PropertyId uint      `form:"propertyId" binding:"required"`
	StartDate  time.Time `form:"startDate" binding:"required"`
	EndDate    time.Time `form:"endDate" binding:"required"`
	Deposit    float64   `form:"deposit"`
}

type updateStatusRequest struct {
	BookingId uint   `json:"bookingId"`
	Status    string `json:"status"`
}

func (server *Server) createBooking(ctx *gin.Context) {
	var req bookingRequest

	// Parse form data
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Start a transaction
	tx := server.store.Begin()

	var totalPrice float64 = 0
	// Iterate over each room ID to check availability and calculate the total price
	for _, roomId := range req.RoomIds {
		var room db.T_Rooms
		if err := tx.Where("id = ?", roomId).First(&room).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		// Check room availability within the requested time frame
		var overlappingBookings []db.T_Bookings
		if err := tx.Joins("JOIN t_booking_rooms ON t_booking_rooms.fk_booking_id = t_bookings.id").
			Where("t_booking_rooms.fk_room_id = ? AND ((t_bookings.start_date, t_bookings.end_date) OVERLAPS (?, ?))", room.Id, req.StartDate, req.EndDate).
			Find(&overlappingBookings).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		// Handle overlapping booking scenario
		if len(overlappingBookings) > 0 {
			tx.Rollback()
			ctx.JSON(http.StatusConflict, gin.H{"error": "Room already booked within this time frame"})
			return
		}

		if room.Status != utils.RoomStatusAvaiable {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("room %d not available", roomId)))
			return
		}

		duration := req.EndDate.Sub(req.StartDate).Hours() / 24
		totalPrice += float64(room.Price) * duration
	}

	var property db.T_Properties
	if err := tx.Where("id = ?", req.PropertyId).First(&property).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if property.Status != utils.HotelStatusAvaiable {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("hotel not available")))
		return
	}

	var status = utils.BookingStatus_Confirmed
	if req.Deposit != 0 {
		status = utils.BookingStatus_Pending
	}
	booking := db.T_Bookings{
		Fk_User_Id:     req.UserId,
		Status:         status,
		Start_Date:     req.StartDate,
		End_Date:       req.EndDate,
		Create_At:      time.Now(),
		Fk_Property_Id: property.Id,
		Total_Price:    totalPrice,
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	for _, roomId := range req.RoomIds {
		bookingRoom := &db.T_Booking_Rooms{
			Fk_Room_Id:    roomId,
			Fk_Booking_id: booking.Id,
		}
		if err := tx.Create(&bookingRoom).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	// Save uploaded images
	form, err := ctx.MultipartForm()
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	files := form.File["image"]
	if len(files) != 0 {

		for _, file := range files {
			filePath := fmt.Sprintf("uploads/%s", file.Filename) // Customize this path as needed

			deposit := db.T_Booking_Deposits{
				Fk_Booking_ID: booking.Id,
				Deposit:       req.Deposit,
			}

			// Save the file locally
			if err := ctx.SaveUploadedFile(file, filePath); err != nil {
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
			deposit.Image = &filePath
			if err := tx.Create(&deposit).Error; err != nil {
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}

			// Create property image record in the database
			propertyImage := db.T_Property_Images{
				Url:            filePath,
				Fk_Property_Id: property.Id,
			}
			if err := tx.Create(&propertyImage).Error; err != nil {
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
		}
	}

	// Create booking deposit record if deposit is provided

	// Commit the transaction
	tx.Commit()

	ctx.JSON(http.StatusOK, booking)
}

type RoomInfo struct {
	Id     uint   `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Price  uint   `json:"price"`
}

type BookingDepositInfo struct {
	ID      uint    `json:"id"`
	Image   string  `json:"image"`
	Deposit float64 `json:"deposit"`
}
type PropertyInfo struct {
	Id             uint    `json:"id"`
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	Fk_Ward_Id     uint    `json:"wardId"`
	Fk_District_Id uint    `json:"districtId"`
	Fk_Province_Id uint    `json:"provinceId"`
	Description    string  `json:"description,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Status         string  `json:"status"`
	Type           string  `json:"type"`
}
type BookingResponse struct {
	Id          uint                `json:"id"`
	Fk_User_Id  uint                `json:"userId"`
	Status      string              `json:"status"`
	Start_Date  time.Time           `json:"startDate"`
	End_Date    time.Time           `json:"endDate"`
	Create_At   time.Time           `json:"createAt"`
	Total_Price float64             `json:"totalPrice"`
	Rooms       []RoomInfo          `json:"rooms"`
	Deposit     *BookingDepositInfo `json:"deposit,omitempty"`
	Property    PropertyInfo        `json:"property"`
}

func (server *Server) getListBookingByUserId(ctx *gin.Context) {
	userId := ctx.Param("userId")

	var bookings []db.T_Bookings
	if err := server.store.Where("fk_user_id = ?", userId).Find(&bookings).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var bookingResponses []BookingResponse

	for _, booking := range bookings {
		var rooms []db.T_Rooms
		if err := server.store.Joins("JOIN t_booking_rooms ON t_booking_rooms.fk_room_id = t_rooms.id").
			Where("t_booking_rooms.fk_booking_id = ?", booking.Id).Find(&rooms).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		var roomInfos []RoomInfo
		for _, room := range rooms {
			roomInfos = append(roomInfos, RoomInfo{
				Id:     room.Id,
				Name:   room.Name,
				Status: room.Status,
				Price:  room.Price,
			})
		}

		var deposit db.T_Booking_Deposits
		var depositResponse *BookingDepositInfo = nil
		if err := server.store.Where("fk_booking_id = ?", booking.Id).First(&deposit).Error; err == nil {
			depositResponse = &BookingDepositInfo{
				ID:      deposit.ID,
				Image:   *deposit.Image,
				Deposit: deposit.Deposit,
			}
		}

		var property db.T_Properties
		if err := server.store.Where("id = ?", rooms[0].Fk_Property_Id).First(&property).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		propertyInfo := PropertyInfo{
			Id:             property.Id,
			Name:           property.Name,
			Address:        property.Address,
			Fk_Ward_Id:     property.Fk_Ward_Id,
			Fk_District_Id: property.Fk_District_Id,
			Fk_Province_Id: property.Fk_Province_Id,
			Description:    property.Description.String,
			Longitude:      property.Longitude.Float64,
			Latitude:       property.Latitude.Float64,
			Status:         property.Status,
			Type:           property.Type,
		}

		bookingResponse := BookingResponse{
			Id:          booking.Id,
			Fk_User_Id:  booking.Fk_User_Id,
			Status:      booking.Status,
			Start_Date:  booking.Start_Date,
			End_Date:    booking.End_Date,
			Create_At:   booking.Create_At,
			Total_Price: booking.Total_Price,
			Rooms:       roomInfos,
			Deposit:     depositResponse,
			Property:    propertyInfo,
		}

		bookingResponses = append(bookingResponses, bookingResponse)
	}

	ctx.JSON(http.StatusOK, bookingResponses)
}

func (server *Server) getListBookingByAgentId(ctx *gin.Context) {
	agentId := ctx.Param("agentId")

	// 1. Retrieve properties by AgentId
	var properties []db.T_Properties
	if err := server.store.Where("fk_argent_id = ?", agentId).Find(&properties).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if len(properties) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No properties found for the given agent"})
		return
	}

	var bookingResponses []BookingResponse

	for _, property := range properties {
		// 2. Get bookings associated with this property
		var bookings []db.T_Bookings
		if err := server.store.Where("fk_property_id = ?", property.Id).Find(&bookings).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		for _, booking := range bookings {
			// 3. Retrieve corresponding T_Booking_Deposits (if it exists)
			var deposit db.T_Booking_Deposits
			depositExists := server.store.Where("fk_booking_id = ?", booking.Id).First(&deposit).RowsAffected > 0

			// 4. Retrieve corresponding T_Booking_Rooms
			var bookingRooms []db.T_Booking_Rooms
			if err := server.store.Where("fk_booking_id = ?", booking.Id).Find(&bookingRooms).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}

			// 5. Loop through T_Booking_Rooms to get room details
			var roomInfos []RoomInfo
			for _, bookingRoom := range bookingRooms {
				var room db.T_Rooms
				if err := server.store.Where("id = ?", bookingRoom.Fk_Room_Id).First(&room).Error; err != nil {
					ctx.JSON(http.StatusInternalServerError, errorResponse(err))
					return
				}
				roomInfos = append(roomInfos, RoomInfo{
					Id:     room.Id,
					Name:   room.Name,
					Status: room.Status,
					Price:  room.Price,
				})
			}

			propertyInfo := PropertyInfo{
				Id:             property.Id,
				Name:           property.Name,
				Address:        property.Address,
				Fk_Ward_Id:     property.Fk_Ward_Id,
				Fk_District_Id: property.Fk_District_Id,
				Fk_Province_Id: property.Fk_Province_Id,
				Description:    property.Description.String,
				Longitude:      property.Longitude.Float64,
				Latitude:       property.Latitude.Float64,
				Status:         property.Status,
				Type:           property.Type,
			}

			bookingResponse := BookingResponse{
				Id:          booking.Id,
				Fk_User_Id:  booking.Fk_User_Id,
				Status:      booking.Status,
				Start_Date:  booking.Start_Date,
				End_Date:    booking.End_Date,
				Create_At:   booking.Create_At,
				Total_Price: booking.Total_Price,
				Rooms:       roomInfos,
				Property:    propertyInfo,
			}

			if depositExists {
				bookingResponse.Deposit = &BookingDepositInfo{
					ID:      deposit.ID,
					Image:   *deposit.Image,
					Deposit: deposit.Deposit,
				}
			}

			bookingResponses = append(bookingResponses, bookingResponse)
		}
	}

	ctx.JSON(http.StatusOK, bookingResponses)
}

func (server *Server) getById(ctx *gin.Context) {
	bookingId := ctx.Param("bookingId")

	var booking db.T_Bookings
	if err := server.store.Where("id = ?", bookingId).First(&booking).Error; err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, booking)
}
func (server *Server) updateBookingStatus(ctx *gin.Context) {
	var req updateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	validStatuses := map[string]bool{
		utils.BookingStatus_CheckOut:  true,
		utils.BookingStatus_Confirmed: true,
		utils.BookingStatus_Pending:   true,
		utils.BookingStatus_Canceled:  true,
		utils.BookingStatus_CheckIn:   true,
	}

	if !validStatuses[req.Status] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	// Start a transaction
	tx := server.store.Begin()

	var booking db.T_Bookings
	if err := tx.Where("id = ?", req.BookingId).First(&booking).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	// Check if the status is CheckIn and if it is within the allowed check-in time window
	if req.Status == utils.BookingStatus_CheckIn {
		now := time.Now()
		startTime := time.Date(booking.Start_Date.Year(), booking.Start_Date.Month(), booking.Start_Date.Day(), 12, 0, 0, 0, booking.Start_Date.Location())
		endTime := startTime.Add(15 * time.Hour) // From 12 PM to 3 AM next day

		if !(now.After(startTime) && now.Before(endTime)) {
			tx.Rollback()
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Check-in is only allowed between 12 PM and 3 AM the next day"})
			return
		}
	}

	booking.Status = req.Status

	if err := tx.Save(&booking).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Commit the transaction
	tx.Commit()

	ctx.JSON(http.StatusOK, booking)
}

type createHotelRequest struct {
	Name        string  `form:"name" binding:"required"`
	WardId      uint    `form:"wardId" binding:"required"`
	DistrictId  uint    `form:"districtId" binding:"required"`
	ProvinceId  uint    `form:"provinceId" binding:"required"`
	Description string  `form:"description"`
	Longitude   float64 `form:"longitude" binding:"required"`
	Latitude    float64 `form:"latitude" binding:"required"`
	Address     string  `form:"address" binding:"required"`
	AgentId     uint    `form:"agentId" binding:"required"`
	Type        string  `form:"type" binding:"required"`
	AmenityIds  []uint  `form:"amenityIds" binding:"required"`
}

type hotelResponse struct {
	Id          uint    `json:"id"`
	Name        string  `json:"name"`
	WardId      uint    `json:"wardId"`
	DistrictId  uint    `json:"districtId"`
	ProvinceId  uint    `json:"provinceId"`
	Description string  `json:"description"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Address     string  `json:"address"`
	AgentId     uint    `json:"agentId"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
}

func (server *Server) createHotel(ctx *gin.Context) {
	var req createHotelRequest

	// Parse the form data
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Handle image uploads
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	files := form.File["images"]

	hotel := db.T_Properties{
		Name:           req.Name,
		Fk_Ward_Id:     req.WardId,
		Fk_District_Id: req.DistrictId,
		Fk_Province_Id: req.ProvinceId,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		Longitude:      sql.NullFloat64{Float64: req.Longitude, Valid: true},
		Latitude:       sql.NullFloat64{Float64: req.Latitude, Valid: true},
		Address:        req.Address,
		Fk_Argent_Id:   req.AgentId,
		Status:         "AVAILABLE",
		Type:           req.Type,
	}

	// Create hotel record in the database
	if err := server.store.Create(&hotel).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Save uploaded images
	for _, file := range files {
		filePath := fmt.Sprintf("uploads/%s", file.Filename) // Customize this path as needed

		// Save the file locally
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		// Create property image record in the database
		propertyImage := db.T_Property_Images{
			Url:            filePath,
			Fk_Property_Id: hotel.Id,
		}
		if err := server.store.Create(&propertyImage).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}
	fmt.Println(">>>AmentiesIds", req.AmenityIds)
	for _, amenityId := range req.AmenityIds {
		propertyAmenity := db.T_Property_Amenities{
			Fk_Property_Id: hotel.Id,
			Fk_Amenity_Id:  amenityId,
		}
		if err := server.store.Create(&propertyAmenity).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}
	ctx.JSON(http.StatusOK, hotelResponse{
		Id:          hotel.Id,
		Name:        hotel.Name,
		WardId:      hotel.Fk_Ward_Id,
		DistrictId:  hotel.Fk_District_Id,
		ProvinceId:  hotel.Fk_Province_Id,
		Description: hotel.Description.String,
		Longitude:   hotel.Longitude.Float64,
		Latitude:    hotel.Latitude.Float64,
		Address:     hotel.Address,
		AgentId:     hotel.Fk_Argent_Id,
		Status:      hotel.Status,
		Type:        hotel.Type,
	})
}

func (server *Server) deleteRoom(ctx *gin.Context) {
	roomID := ctx.Param("roomId")

	// Check if room ID is valid
	id, err := strconv.Atoi(roomID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	// Start a transaction
	tx := server.store.Begin()

	// Update room status to DELETED
	if err := tx.Model(&db.T_Rooms{}).Where("id = ?", id).Update("status", "DELETED").Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room"})
		return
	}

	// Commit the transaction
	tx.Commit()

	ctx.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}
func (server *Server) deleteHotel(ctx *gin.Context) {
	hotelID := ctx.Param("hotelId")

	// Check if hotel ID is valid
	id, err := strconv.Atoi(hotelID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hotel ID"})
		return
	}

	// Start a transaction
	tx := server.store.Begin()

	// Update hotel status to DELETED
	if err := tx.Model(&db.T_Properties{}).Where("id = ?", id).Update("status", "DELETED").Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hotel"})
		return
	}

	// Commit the transaction
	tx.Commit()

	ctx.JSON(http.StatusOK, gin.H{"message": "Hotel deleted successfully"})
}

// HotelResponse struct for hotel (property) response
type HotelResponse struct {
	ID             uint              `json:"id"`
	Name           string            `json:"name"`
	WardID         uint              `json:"wardId"`
	DistrictID     uint              `json:"districtId"`
	ProvinceID     uint              `json:"provinceId"`
	Description    *string           `json:"description"`
	Longitude      *float64          `json:"longitude"`
	Latitude       *float64          `json:"latitude"`
	Address        string            `json:"address"`
	AgentID        uint              `json:"agentId"`
	Status         string            `json:"status"`
	Type           string            `json:"type"`
	HotelAmenities []AmenityResponse `json:"amenities"`
	HotelImages    []ImageResponse   `json:"images"`
	HotelRooms     []RoomResponse    `json:"rooms"`
}

// RoomResponse struct for room response
type RoomResponse struct {
	ID            uint              `json:"id"`
	PropertyID    uint              `json:"propertyId"`
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	Price         uint              `json:"price"`
	RoomAmenities []AmenityResponse `json:"amenities"`
	RoomImages    []ImageResponse   `json:"images"`
}

// AmenityResponse struct for amenity response
type AmenityResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ImageResponse struct for image response
type ImageResponse struct {
	ID  uint   `json:"id"`
	Url string `json:"url"`
}

func (server *Server) getHotelsByAgent(ctx *gin.Context) {
	agentID, err := strconv.Atoi(ctx.Param("agentId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var hotels []HotelResponse

	// Query properties for the given agentId
	var properties []db.T_Properties
	if err := server.store.Where("fk_argent_id = ? AND status <> ?", agentID, "DELETED").Find(&properties).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching hotels"})
		return
	}

	// Iterate through each property (hotel)
	for _, property := range properties {
		var rooms = []RoomResponse{}

		// Query rooms for the current hotel
		var dbRooms []db.T_Rooms
		if err := server.store.Where("fk_property_id = ? AND status != ?", property.Id, "DELETED").Find(&dbRooms).Error; err != nil {
			fmt.Println("ERR", err)
			continue // Skip this property if rooms cannot be fetched
		}

		// Map database rooms to RoomResponse structs
		for _, dbRoom := range dbRooms {
			room := RoomResponse{
				ID:         dbRoom.Id,
				PropertyID: dbRoom.Fk_Property_Id,
				Name:       dbRoom.Name,
				Status:     dbRoom.Status,
				Price:      dbRoom.Price,
			}

			// Query room amenities
			var roomAmenities []AmenityResponse
			if err := server.store.Table("t_room_amenities").
				Select("t_amenities.id, t_amenities.name , t_amenities.type").
				Joins("JOIN t_amenities ON t_amenities.id = t_room_amenities.fk_amenity_id").
				Where("t_amenities.is_deleted = ?", false).
				Where("t_room_amenities.fk_room_id = ?", dbRoom.Id).
				Find(&roomAmenities).Error; err != nil {
				continue // Skip this room if room amenities cannot be fetched
			}
			fmt.Printf("%+v \n", roomAmenities)
			room.RoomAmenities = roomAmenities

			// Query room images
			var roomImages []ImageResponse
			if err := server.store.Table("t_room_images").
				Select("t_room_images.id, t_room_images.url ").
				Where("t_room_images.fk_room_id = ?", dbRoom.Id).
				Find(&roomImages).Error; err != nil {
				continue // Skip this room if room amenities cannot be fetched
			}
			room.RoomImages = roomImages
			fmt.Printf("ROOM%+v \n", roomAmenities)

			// Append room to rooms list
			rooms = append(rooms, room)
		}
		var hotelImages []ImageResponse
		if err := server.store.Table("t_property_images").
			Select("t_property_images.id, t_property_images.url ").
			Where("t_property_images.fk_property_id = ?", property.Id).
			Find(&hotelImages).Error; err != nil {
			continue // Skip this room if room amenities cannot be fetched
		}
		var hotelAmenities []AmenityResponse
		if err := server.store.Table("t_property_amenities").
			Select("t_amenities.id, t_amenities.name , t_amenities.type").
			Joins("JOIN t_amenities ON t_amenities.id = t_property_amenities.fk_amenity_id").
			Where("t_amenities.is_deleted = ?", false).
			Where("t_property_amenities.fk_property_id = ?", property.Id).
			Find(&hotelAmenities).Error; err != nil {
			continue // Skip this room if room amenities cannot be fetched
		}
		// Prepare hotel response
		hotel := HotelResponse{
			ID:             property.Id,
			Name:           property.Name,
			WardID:         property.Fk_Argent_Id,
			DistrictID:     property.Fk_District_Id,
			ProvinceID:     property.Fk_Province_Id,
			Description:    &property.Description.String,
			Longitude:      &property.Longitude.Float64,
			Latitude:       &property.Latitude.Float64,
			Address:        property.Address,
			AgentID:        property.Fk_Argent_Id,
			Status:         property.Status,
			Type:           property.Type,
			HotelAmenities: hotelAmenities, // Populate if needed
			HotelImages:    hotelImages,    // Populate if needed
			HotelRooms:     rooms,
		}

		// Append hotel to hotels list
		hotels = append(hotels, hotel)
	}

	// Return JSON response with the list of hotels
	ctx.JSON(http.StatusOK, gin.H{"hotels": hotels})
}
