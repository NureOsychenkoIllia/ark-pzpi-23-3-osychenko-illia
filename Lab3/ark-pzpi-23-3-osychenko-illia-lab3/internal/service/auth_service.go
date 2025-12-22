package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"busoptima/internal/model"
	"busoptima/internal/repository"
)

// AuthService інтерфейс для автентифікації
type AuthService interface {
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	DeviceAuth(ctx context.Context, serialNumber, token string) (*DeviceAuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	CreateUser(ctx context.Context, user *model.User, password string) error
	UpdateUser(ctx context.Context, user *model.User) error
	UpdateUserRole(ctx context.Context, userID, roleID int64) error
	GetUsers(ctx context.Context) ([]model.User, error)
}

// authService реалізація AuthService
type authService struct {
	userRepo   repository.UserRepository
	deviceRepo repository.DeviceRepository
	jwtSecret  string
}

// LoginResponse відповідь на успішну автентифікацію
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         *model.User `json:"user"`
}

// DeviceAuthResponse відповідь на автентифікацію пристрою
type DeviceAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	DeviceID    int64  `json:"device_id"`
}

// NewAuthService створює новий сервіс автентифікації
func NewAuthService(userRepo repository.UserRepository, deviceRepo repository.DeviceRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:   userRepo,
		deviceRepo: deviceRepo,
		jwtSecret:  jwtSecret,
	}
}

// Login автентифікує користувача
func (s *authService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	fmt.Printf("DEBUG: Attempting login for email: %s\n", email)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		fmt.Printf("DEBUG: GetByEmail error: %v\n", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	fmt.Printf("DEBUG: Found user: %s, hash: %s\n", user.Email, user.PasswordHash[:20]+"...")

	// Перевіряємо пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		fmt.Printf("DEBUG: Password comparison error: %v\n", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	fmt.Printf("DEBUG: Password verification successful\n")

	// Отримуємо дозволи користувача
	permissions, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Генеруємо JWT токени
	accessToken, err := s.generateAccessToken(user, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Приховуємо пароль у відповіді
	user.PasswordHash = ""

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 година
		User:         user,
	}, nil
}

// DeviceAuth автентифікує IoT-пристрій
func (s *authService) DeviceAuth(ctx context.Context, serialNumber, token string) (*DeviceAuthResponse, error) {
	// Отримуємо пристрій за серійним номером
	device, err := s.deviceRepo.GetBySerialNumber(ctx, serialNumber)
	if err != nil {
		return nil, fmt.Errorf("device not found")
	}

	// Перевіряємо токен автентифікації
	err = bcrypt.CompareHashAndPassword([]byte(device.AuthTokenHash), []byte(token))
	if err != nil {
		return nil, fmt.Errorf("invalid device credentials")
	}

	// Генеруємо токен для пристрою
	deviceToken, err := s.generateDeviceToken(device.ID, serialNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate device token: %w", err)
	}

	return &DeviceAuthResponse{
		AccessToken: deviceToken,
		ExpiresIn:   86400, // 24 години
		DeviceID:    device.ID,
	}, nil
}

// RefreshToken оновлює токен доступу
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Парсимо refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID := int64(claims["user_id"].(float64))

	// Отримуємо користувача
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Отримуємо дозволи
	permissions, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Генеруємо новий access token
	accessToken, err := s.generateAccessToken(user, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	user.PasswordHash = ""

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Повертаємо той самий refresh token
		ExpiresIn:    3600,
		User:         user,
	}, nil
}

// CreateUser створює нового користувача
func (s *authService) CreateUser(ctx context.Context, user *model.User, password string) error {
	// Хешуємо пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	user.IsActive = true

	return s.userRepo.Create(ctx, user)
}

// UpdateUser оновлює користувача
func (s *authService) UpdateUser(ctx context.Context, user *model.User) error {
	return s.userRepo.Update(ctx, user)
}

// UpdateUserRole оновлює роль користувача
func (s *authService) UpdateUserRole(ctx context.Context, userID, roleID int64) error {
	return s.userRepo.UpdateRole(ctx, userID, roleID)
}

// GetUsers повертає список користувачів
func (s *authService) GetUsers(ctx context.Context) ([]model.User, error) {
	return s.userRepo.GetAll(ctx)
}

// generateAccessToken генерує JWT токен доступу
func (s *authService) generateAccessToken(user *model.User, permissions []string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"email":       user.Email,
		"role":        user.Role.Name,
		"permissions": permissions,
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateRefreshToken генерує JWT refresh токен
func (s *authService) generateRefreshToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 днів
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateDeviceToken генерує JWT токен для пристрою
func (s *authService) generateDeviceToken(deviceID int64, serialNumber string) (string, error) {
	claims := jwt.MapClaims{
		"device_id":     deviceID,
		"serial_number": serialNumber,
		"type":          "device",
		"exp":           time.Now().Add(time.Hour * 24).Unix(), // 24 години
		"iat":           time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
