package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/lancer2672/BookingAppSubServer/db"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"

	"github.com/gin-gonic/gin"
)

func (server *Server) updateBankAccount(ctx *gin.Context) {
	// Parse bank account ID from path parameter
	bankID := ctx.Param("bankId")

	// Check if bank ID is valid
	id, err := strconv.Atoi(bankID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bank account ID"})
		return
	}

	// Start a transaction
	tx := server.store.Begin()

	// Fetch the bank account to update
	var bankAccount db.T_Banks
	if err := tx.Where("id = ?", id).First(&bankAccount).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find bank account"})
		return
	}

	// Parse form data fields
	bankName := ctx.PostForm("bankName")
	accountNumber := ctx.PostForm("accountNumber")
	accountName := ctx.PostForm("accountName")
	isDefaultStr := ctx.PostForm("isDefault")
	isDefault, err := strconv.ParseBool(isDefaultStr)
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for isDefault"})
		return
	}

	// Handle QR Code (Image) upload
	file, fileHeader, err := ctx.Request.FormFile("qrCode")
	if err == nil {
		defer file.Close()

		// Customize upload path and file name as needed
		filePath := fmt.Sprintf("uploads/%s", fileHeader.Filename)

		// Save file to server
		out, err := os.Create(filePath)
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		fullPath := fmt.Sprintf("%s/%s", utils.URL_API, filePath)
		qrCodeURL := &fullPath
		bankAccount.QR_Code = qrCodeURL

	}

	// Update bank account fields
	bankAccount.Bank_Name = bankName
	bankAccount.Account_Number = accountNumber
	bankAccount.Account_Name = accountName

	// Check if IsDefault is being updated
	if isDefault {
		// Set all other bank accounts of this agent to IsDefault = false
		if err := tx.Model(&db.T_Banks{}).Where("fk_argent_id = ?", bankAccount.Fk_Argent_Id).
			Where("id <> ?", id).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update other bank accounts"})
			return
		}
	}

	// Update Is_Default field of the current bank account
	bankAccount.Is_Default = isDefault

	// Save changes to the database
	if err := tx.Save(&bankAccount).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bank account"})
		return
	}

	// Commit the transaction
	tx.Commit()

	ctx.JSON(http.StatusOK, gin.H{"message": "Bank account updated successfully"})
}

func (server *Server) CreateBankAccount(ctx *gin.Context) {
	// Parse form data
	err := ctx.Request.ParseMultipartForm(10 << 20) // 10MB maximum form size
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse form data"})
		return
	}

	// Parse fields from form data
	bankName := ctx.Request.FormValue("bankName")
	accountNumber := ctx.Request.FormValue("accountNumber")
	accountName := ctx.Request.FormValue("accountName")
	isDefaultStr := ctx.Request.FormValue("isDefault")
	isDefault, err := strconv.ParseBool(isDefaultStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for isDefault"})
		return
	}

	// Retrieve agent ID from token or any other method
	agentID := uint(1) // Replace with actual agent ID retrieval logic

	// Handle QR Code (Image) upload
	file, fileHeader, err := ctx.Request.FormFile("qrCode")
	var qrCodeURL *string
	if err == nil {
		defer file.Close()

		// Customize upload path and file name as needed
		filePath := fmt.Sprintf("uploads/%s", fileHeader.Filename)

		// Save file to server
		out, err := os.Create(filePath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		fullPath := fmt.Sprintf("%s/%s", utils.URL_API, *qrCodeURL)
		qrCodeURL = &fullPath
	}

	// Create bank account record
	bankAccount := db.T_Banks{
		Bank_Name:      bankName,
		Account_Number: accountNumber,
		QR_Code:        qrCodeURL,
		Fk_Argent_Id:   agentID,
		Is_Default:     isDefault,
		Account_Name:   accountName,
		Create_At:      time.Now(),
	}

	// Save bank account to database
	if err := server.store.Create(&bankAccount).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bank account"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bank account created successfully", "bankAccount": bankAccount})
}

type BankResponse struct {
	ID            uint      `json:"id"`
	BankName      string    `json:"bankName"`
	AccountNumber string    `json:"accountNumber"`
	QRCode        *string   `json:"qrCode"`
	AgentID       uint      `json:"agentId"`
	IsDefault     bool      `json:"isDefault"`
	CreatedAt     time.Time `json:"createdAt"`
	AccountName   string    `json:"accountName"`
}

func (server *Server) GetListAccountByAgentId(ctx *gin.Context) {
	// Retrieve agent ID from token or any other method
	agentID := ctx.Param("agentId")

	// Query bank accounts for the given agent ID
	var bankAccounts = []db.T_Banks{}
	if err := server.store.Where("fk_argent_id = ?", agentID).Find(&bankAccounts).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching bank accounts"})
		return
	}

	// Convert to BankResponse format
	var bankResponses = []BankResponse{}
	for _, bank := range bankAccounts {
		bankResponse := BankResponse{
			ID:            bank.ID,
			BankName:      bank.Bank_Name,
			AccountNumber: bank.Account_Number,
			QRCode:        bank.QR_Code,
			AgentID:       bank.Fk_Argent_Id,
			IsDefault:     bank.Is_Default,
			CreatedAt:     bank.Create_At,
			AccountName:   bank.Account_Name,
		}
		bankResponses = append(bankResponses, bankResponse)
	}

	ctx.JSON(http.StatusOK, gin.H{"bankAccounts": bankResponses})
}
