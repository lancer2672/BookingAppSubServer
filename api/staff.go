package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/lancer2672/BookingAppSubServer/db"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"

	"github.com/gin-gonic/gin"
)

func (server *Server) CreateStaff(ctx *gin.Context) {
	// Parse form data
	err := ctx.Request.ParseMultipartForm(10 << 20) // 10MB maximum form size
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse form data"})
		return
	}

	// Retrieve fields from form data
	agentIDStr := ctx.Request.FormValue("agentId")
	agentID, err := strconv.ParseUint(agentIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	firstName := ctx.Request.FormValue("firstName")
	lastName := ctx.Request.FormValue("lastName")
	email := ctx.Request.FormValue("email")
	phoneNumber := ctx.Request.FormValue("phoneNumber")
	role := "STAFF"
	password := "$2a$10$sW1Loq.Jo8LAwuaXzCRcj.KeXSegN15xCZDLFfV3woiu0MaI8sc5."
	avatar, avatarHeader, err := ctx.Request.FormFile("avatar")
	var avatarURL string
	if err == nil {
		defer avatar.Close()

		// Customize upload path and file name as needed
		avatarPath := fmt.Sprintf("uploads/%s", avatarHeader.Filename)

		// Save avatar file to server
		out, err := os.Create(avatarPath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar file"})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, avatar)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar file"})
			return
		}
		fullPath := fmt.Sprintf("%s/%s", utils.URL_API, avatarPath)

		avatarURL = fullPath
	}

	// Create a new user (staff)
	newUser := db.T_Users{
		First_Name:   firstName,
		Last_Name:    lastName,
		Email:        &email,
		Phone_Number: phoneNumber,
		Role:         role,
		Avatar:       avatarURL,
		Status:       "ACTIVE", // Assuming staff is active upon creation
		Password:     password,
	}

	// Save user to database
	if err := server.store.Create(&newUser).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff"})
		return
	}

	// Create agent-staff relationship
	agentStaff := db.T_Agent_Staffs{
		Agent_Id: uint(agentID),
		Staff_Id: newUser.Id,
	}

	// Save agent-staff relationship to database
	if err := server.store.Create(&agentStaff).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent-staff relationship"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Staff created successfully", "staff": newUser})
}

type StaffResponse struct {
	ID          uint   `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	Avatar      string `json:"avatar"`
	Status      string `json:"status"`
}

func (server *Server) GetStaffByAgentId(ctx *gin.Context) {
	// Extract agent ID from path parameter
	agentID := ctx.Param("agentId")
	agentIDUint, err := strconv.ParseUint(agentID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	// Query staffs by agent ID
	var staffs []db.T_Users
	if err := server.store.Table("t_users").
		Joins("JOIN t_agent_staffs ON t_users.id = t_agent_staffs.staff_id").
		Where("t_agent_staffs.agent_id = ?", agentIDUint).
		Find(&staffs).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch staffs"})
		return
	}

	// Map database results to response format
	var staffResponses []StaffResponse
	for _, staff := range staffs {
		staffResponses = append(staffResponses, StaffResponse{
			ID:          staff.Id,
			FirstName:   staff.First_Name,
			LastName:    staff.Last_Name,
			Email:       *staff.Email,
			PhoneNumber: staff.Phone_Number,
			Role:        staff.Role,
			Avatar:      staff.Avatar,
			Status:      staff.Status,
		})
	}

	ctx.JSON(http.StatusOK, staffResponses)
}
