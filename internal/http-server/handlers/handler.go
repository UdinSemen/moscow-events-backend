package handlers

import (
	jwt_manager "github.com/UdinSemen/moscow-events-backend/internal/jwt-manager"
	"github.com/UdinSemen/moscow-events-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    *services.Service
	jwtManager jwt_manager.TokenManager
}

func NewHandler(service *services.Service, jwtManager jwt_manager.TokenManager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	auth := router.Group("/auth")
	{
		auth.POST("/sign-in", h.signIn)
		auth.POST("/sign-up", h.signUp)
		auth.POST("/refresh", h.refresh)
	}

	api := router.Group("/api")
	{
		event := api.Group("/event")
		{
			event.GET("/:event_type", h.getEvent)
		}

	}

	moderate := router.Group("/moderate")
	{
		modEvent := moderate.Group("/event")
		{
			modEvent.GET("/:id/:name/:date", h.moderateGetEvent)
			modEvent.POST("/:id", h.moderateAddEvent)
			modEvent.PUT("/:id", h.moderateUpdateEvent)
			modEvent.DELETE("/:id", h.moderateDeleteEvent)
		}
		user := moderate.Group("/user")
		{
			user.GET("/:id", h.moderateGetUser)
			user.DELETE("/:id", h.moderateDeleteUser)
			user.PUT("/:id", h.moderateUpdateUser)
			user.POST("/add", h.moderateAddUser)
		}
	}

	return router
}
