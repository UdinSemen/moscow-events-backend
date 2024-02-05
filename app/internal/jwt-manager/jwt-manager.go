package jwt_manager

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
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

type tokenClaims struct {
	jwt.StandardClaims
	Uuid string `json:"uuid"`
	Role string `json:"role"`
}

func NewManager(signingKey string, tokenTTL *time.Duration) (*Manager, error) {
	const op = "jwt-manager.NewManager"
	if signingKey == "" {
		return nil, fmt.Errorf("%s:%w", op, errors.New("empty signing key"))
	}
	if tokenTTL == nil {
		return nil, fmt.Errorf("%s:%w", op, errors.New("empty tokenTTL key"))
	}

	zap.S().Infow("tokenTTL",
		"access", tokenTTL)
	return &Manager{
		signingKey: signingKey,
		tokenTTL:   *tokenTTL}, nil
}

func (m *Manager) GenerateToken(userID, role string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(m.tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Uuid: userID,
		Role: role,
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
		Uuid: claims["uuid"].(string),
		Role: claims["role"].(string),
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
