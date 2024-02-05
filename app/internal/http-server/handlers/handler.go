package handlers

import (
	"net/http"
	"time"

	logmiddlewares "github.com/UdinSemen/moscow-events-backend/internal/http-server/log-middlewares"
	jwtmanager "github.com/UdinSemen/moscow-events-backend/internal/jwt-manager"
	"github.com/UdinSemen/moscow-events-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    *services.Service
	jwtManager jwtmanager.TokenManager
}

func NewHandler(service *services.Service, jwtManager jwtmanager.TokenManager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/ping_category", func(c *gin.Context) {
		c.JSON(http.StatusOK, inputGetEvent{
			Category: "test_category",
			Date:     []time.Time{time.Now(), time.Now().Add(time.Hour * 24)},
		})
	})

	auth := router.Group("/auth", logmiddlewares.RequestLogger)
	{
		auth.POST("/sign-in", h.signIn)
		auth.GET("/sign-in-ws", h.signInWebSocket)
		auth.POST("/sign-up", h.signUp)
		auth.POST("/refresh", h.refresh)
	}

	api := router.Group("/api", logmiddlewares.RequestLogger, h.userIdentity)
	{
		event := api.Group("/event")
		{
			event.GET("/", h.getEvent)
		}

		user := api.Group("/user")
		{
			user.GET("/", h.moderateGetUser)
			user.PUT("/", h.moderateAddUser)
		}
	}

	moderate := router.Group("/moderate")
	{
		modEvent := moderate.Group("/event")
		{
			modEvent.GET("/", h.moderateGetEvent)
			modEvent.POST("/:id", h.moderateAddEvent)
			modEvent.PUT("/:id", h.moderateUpdateEvent)
			modEvent.DELETE("/:id", h.moderateDeleteEvent)
		}

	}

	return router
}
