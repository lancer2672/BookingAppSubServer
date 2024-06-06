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
	UserId     uint      `json:"userId" binding:"required"`
	RoomId     uint      `json:"roomId" binding:"required"`
	StartDate  time.Time `json:"startDate" binding:"required"`
	EndDate    time.Time `json:"endDate" binding:"required"`
	TotalPrice float64   `json:"totalPrice" binding:"required"`
}

type updateStatusRequest struct {
	BookingId uint   `json:"bookingId" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

func (server *Server) createBooking(ctx *gin.Context) {
	var req bookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Start a transaction
	tx := server.store.Begin()

	var room db.Room
	if err := tx.Where("id = ?", req.RoomId).First(&room).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	var hotel db.Property
	if err := tx.Where("id = ?", room.FkPropertyId).First(&hotel).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if hotel.Status != utils.HotelStatusAvaiable {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("hotel not available")))
		return

	}
	// Calculate expected total price
	duration := req.EndDate.Sub(req.StartDate).Hours() / 24
	expectedTotalPrice := float64(room.Price) * duration

	if req.TotalPrice != expectedTotalPrice {
		tx.Rollback()
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Total price does not match the expected value"})
		return
	}

	booking := db.Booking{
		FkUserId:   req.UserId,
		FkRoomId:   req.RoomId,
		Status:     utils.BookingStatus_NotCheckIn,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		TotalPrice: req.TotalPrice,
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
		utils.BookingStatus_NotCheckIn:  true,
		utils.BookingStatus_NotCheckOut: true,
		utils.BookingStatus_Canceled:    true,
		utils.BookingStatus_CheckOut:    true,
	}

	if !validStatuses[req.Status] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}
	// Start a transaction
	tx := server.store.Begin()

	var booking db.Booking
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
