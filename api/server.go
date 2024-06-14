package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"
	"gorm.io/gorm"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config utils.Config
	store  *gorm.DB

	router *gin.Engine
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config utils.Config, store *gorm.DB) (*Server, error) {
	// tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot create token maker: %w", err)
	// }

	server := &Server{
		config: config,
		store:  store,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	// config := cors.DefaultConfig()
	// config.AllowAllOrigins = true
	// config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	// config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}

	// router.Use(cors.New(config))

	router.StaticFS("/uploads", gin.Dir("./uploads", true))
	router.POST("/api/booking/v2", server.createBookingV2)
	// router.POST("/api/bookings", server.createBooking)
	router.GET("/healthcheck", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "OK")
	})
	router.PATCH("/api/bookings", server.updateBookingStatus)
	router.GET("/api/bookings/user/:userId", server.getListBookingByUserId)
	router.GET("/api/bookings/agent/:agentId", server.getListBookingByAgentId)
	router.GET("/api/bookings/:bookingId", server.getById)

	router.POST("api/hotels", server.createHotel)
	// router.POST("api/hotels/v2", server.createHotel)
	router.DELETE("api/hotels/:hotelId", server.deleteHotel)
	router.GET("api/hotels/:agentId", server.getHotelsByAgent)

	router.GET("api/rooms/:propertyId", server.getListRoomByHotelId)
	router.POST("api/rooms/", server.createRoom)
	router.DELETE("api/rooms/:roomId", server.deleteRoom)

	router.POST("api/banks/", server.CreateBankAccount)
	router.PUT("api/banks/:bankId", server.updateBankAccount)
	router.GET("api/banks/:agentId", server.GetListAccountByAgentId)

	router.GET("api/staffs/:agentId", server.GetStaffByAgentId)
	router.POST("api/staffs", server.CreateStaff)

	// router.POST("/tokens/renew_access", server.renewAccessToken)

	// authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
