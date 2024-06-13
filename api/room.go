package api

import (
	"fmt"
	"net/http"

	"github.com/lancer2672/BookingAppSubServer/db"

	"github.com/gin-gonic/gin"
)

type createRoomRequest struct {
	PropertyId uint   `form:"propertyId" binding:"required"`
	Name       string `form:"name" binding:"required"`
	Price      uint   `form:"price" binding:"required"`
	AmenityIds []uint `form:"amenityIds" binding:"required"`
	Images     []uint `form:"amenityIds" binding:"required"`
}

func (server *Server) createRoom(ctx *gin.Context) {
	var req createRoomRequest

	// Parse the form data
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle image uploads
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	files := form.File["images"]

	room := db.T_Rooms{
		Fk_Property_Id: req.PropertyId,
		Name:           req.Name,
		Status:         "AVAILABLE",
		Price:          req.Price,
	}

	// Create room record in the database
	if err := server.store.Create(&room).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save uploaded images
	for _, file := range files {
		filePath := fmt.Sprintf("uploads/%s", file.Filename) // Customize this path as needed

		// Save the file locally
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create room image record in the database
		roomImage := db.T_Room_Images{
			Url:        filePath,
			Fk_Room_Id: room.Id,
		}
		if err := server.store.Create(&roomImage).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Save room amenities
	for _, amenityId := range req.AmenityIds {
		roomAmenity := db.T_Room_Amenities{
			Fk_Room_Id:    room.Id,
			Fk_Amenity_Id: amenityId,
		}
		if err := server.store.Create(&roomAmenity).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, RoomResponse{
		ID:         room.Id,
		PropertyID: room.Fk_Property_Id,
		Name:       room.Name,
		Status:     room.Status,
		Price:      room.Price,
	})
}

func (server *Server) getListRoomByHotelId(ctx *gin.Context) {
	hotelId := ctx.Param("propertyId")

	var rooms []db.T_Rooms

	// Query rooms by hotelId
	if err := server.store.Where("fk_property_id = ?", hotelId).Find(&rooms).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching rooms"})
		return
	}

	var roomResponses []RoomResponse

	// Iterate through each room and fetch amenities and images
	for _, room := range rooms {
		var amenities = []AmenityResponse{}
		if err := server.store.Table("t_room_amenities").
			Select("t_amenities.id, t_amenities.name , t_amenities.type").
			Joins("JOIN t_amenities ON t_amenities.id = t_room_amenities.fk_amenity_id").
			Where("t_amenities.is_deleted = ?", false).
			Where("t_room_amenities.fk_room_id = ?", room.Id).
			Find(&amenities).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching amenities"})
			return
		}

		var images = []ImageResponse{}
		if err := server.store.Table("t_room_images").
			Select("id, url").
			Where("fk_room_id = ?", room.Id).
			Find(&images).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching images"})
			return
		}
		roomResponse := RoomResponse{
			ID:            room.Id,
			PropertyID:    room.Fk_Property_Id,
			Name:          room.Name,
			Status:        room.Status,
			Price:         room.Price,
			RoomAmenities: amenities,
			RoomImages:    images,
		}

		roomResponses = append(roomResponses, roomResponse)

	}

	ctx.JSON(http.StatusOK, roomResponses)
}
