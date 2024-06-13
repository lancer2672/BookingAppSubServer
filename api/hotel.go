package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lancer2672/BookingAppSubServer/db"

	"github.com/gin-gonic/gin"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"
)

type bookingRequest struct {
	//TODO: retrive from token
	UserId       uint      `json:"userId"`
	RoomIds      []uint    `json:"roomIds"`
	PropertyId   uint      `json:"propertyId"`
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	Deposit      float64   `json:"deposit"`
	DepositImage *string   `json:"depositImage"`
}

type updateStatusRequest struct {
	BookingId uint   `json:"bookingId"`
	Status    string `json:"status"`
}

func (server *Server) createBooking(ctx *gin.Context) {
	var req bookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
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
	// Create booking deposit record if deposit is provided
	if req.Deposit != 0 {
		deposit := db.T_Booking_Deposit{
			Fk_Booking_ID: booking.Id,
			Deposit:       req.Deposit,
			Image:         req.DepositImage,
		}

		if err := tx.Create(&deposit).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

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
	Id          uint               `json:"id"`
	Fk_User_Id  uint               `json:"userId"`
	Status      string             `json:"status"`
	Start_Date  time.Time          `json:"startDate"`
	End_Date    time.Time          `json:"endDate"`
	Create_At   time.Time          `json:"createAt"`
	Total_Price float64            `json:"totalPrice"`
	Rooms       []RoomInfo         `json:"rooms"`
	Deposit     BookingDepositInfo `json:"deposit,omitempty"`
	Property    PropertyInfo       `json:"property"`
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

		var deposit db.T_Booking_Deposit
		if err := server.store.Where("fk_booking_id = ?", booking.Id).First(&deposit).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
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
			Deposit: BookingDepositInfo{
				ID:      deposit.ID,
				Image:   *deposit.Image,
				Deposit: deposit.Deposit,
			},
			Property: propertyInfo,
		}

		bookingResponses = append(bookingResponses, bookingResponse)
	}

	ctx.JSON(http.StatusOK, bookingResponses)
}

func (server *Server) getListBookingByAgentId(ctx *gin.Context) {
	agentId := ctx.Param("agentId")

	// 1. Retrieve properties by AgentId
	var properties []db.T_Properties
	if err := server.store.Where("fk_agent_id = ?", agentId).Find(&properties).Error; err != nil {
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
			// 3. Retrieve corresponding T_Booking_Deposit (if it exists)
			var deposit db.T_Booking_Deposit
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
				bookingResponse.Deposit = BookingDepositInfo{
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
