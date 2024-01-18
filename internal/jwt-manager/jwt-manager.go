package jwt_manager

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/dgrijalva/jwt-go"
)

type TokenManager interface {
	GenerateToken(userID, role string) (string, error)
	ParseToken(accessToken string) (models.UserDTO, error)
	NewRefreshToken() (string, error)
}

type Manager struct {
	signingKey string
	tokenTTL   time.Duration
}

func NewManager(signingKey string, tokenTTL *time.Duration) (*Manager, error) {
	const op = "jwt-manager.NewManager"
	if signingKey == "" {
		return nil, fmt.Errorf("%s:%w", op, errors.New("empty signing key"))
	}
	if tokenTTL == nil {
		return nil, fmt.Errorf("%s:%w", op, errors.New("empty tokenTTL key"))
	}

	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) GenerateToken(userID, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(m.tokenTTL).Unix(),
		Id:        userID,
		Subject:   role,
	})

	return token.SignedString([]byte(m.signingKey))
}

func (m *Manager) ParseToken(accessToken string) (models.UserDTO, error) {
	const op = "jwt-manager.ParseToken"
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s:%w", op,
				fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		return models.UserDTO{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return models.UserDTO{}, fmt.Errorf("error get user claims from token")
	}

	return models.UserDTO{
		UserID: claims["jti"].(string),
		Role:   claims["sub"].(string),
	}, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	const op = "jwt-manager.NewRefreshToken"
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	return fmt.Sprintf("%x", b), nil
}
