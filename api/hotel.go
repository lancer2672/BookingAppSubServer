package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lancer2672/BookingAppSubServer/db"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"
)

type bookingRequest struct {
	UserId     uint      `json:"userId"`
	RoomId     uint      `json:"roomId"`
	Status     string    `json:"status"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	TotalPrice float64   `json:"totalPrice"`
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
	// Check room availability within the requested time frame
	var overlappingBookings []db.T_Bookings
	if err := tx.Where("fk_room_id = ? AND ((start_date, end_date) OVERLAPS (?, ?))", req.RoomId, req.StartDate, req.EndDate).Find(&overlappingBookings).Error; err != nil {
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
	var room db.T_Rooms
	if err := tx.Where("id = ?", req.RoomId).First(&room).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if room.Status != utils.RoomStatusAvaiable {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("hotel not available")))
		return
	}
	var property db.T_Properties
	if err := tx.Where("id = ?", room.Fk_Property_Id).First(&property).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if property.Status != utils.HotelStatusAvaiable {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("hotel not available")))
		return
	}

	// Calculate expected total price
	duration := req.EndDate.Sub(req.StartDate).Hours() / 24
	expectedTotalPrice := float64(room.Price) * (duration)
	log.Printf("ExpectedPrice %v %v %v", req.TotalPrice, duration, expectedTotalPrice)
	if req.TotalPrice != expectedTotalPrice {
		tx.Rollback()
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Total price does not match the expected value"})
		return
	}
	var status = utils.BookingStatus_Confirmed
	// if(property.Deposit == true){
	// 	status = utils.BookingStatus_Pending
	// }
	booking := db.T_Bookings{
		Fk_User_Id: req.UserId,
		Fk_Room_Id: req.RoomId,
		Status:     status,
		// Start_Date: req.StartDate,
		// End_Date:   req.EndDate,
		Start_Date: db.Timetz{
			Time: req.StartDate,
		},
		End_Date: db.Timetz{
			Time: req.EndDate,
		},
		Created_At:  time.Now(),
		Total_Price: req.TotalPrice,
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Commit the transaction
	tx.Commit()

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
