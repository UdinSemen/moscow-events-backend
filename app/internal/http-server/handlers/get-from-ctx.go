package handlers

import (
	"errors"
	"net/http"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/gin-gonic/gin"
)

func getUserDTOFromCtx(c *gin.Context) (models.UserDTO, error) {
	userDTO, ok := c.Get(UserCtx)
	if !ok {
		newErrorResponse(c, http.StatusInternalServerError, "userDTO not found")
		return models.UserDTO{}, errors.New("userDTO not found")
	}

	user, ok := userDTO.(models.UserDTO)
	if !ok {
		newErrorResponse(c, http.StatusInternalServerError, "userDTO have invalid type")
		return models.UserDTO{}, errors.New("userDTO have invalid type")
	}

	return user, nil
}
