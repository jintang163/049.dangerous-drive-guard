package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db     *gorm.DB
	jwtCfg *config.JWTConfig
}

type LoginResult struct {
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	ExpiresIn   int            `json:"expires_in"`
	User        *model.User    `json:"user"`
	Permissions []string       `json:"permissions"`
}

type RefreshResult struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
}

type CustomClaims struct {
	UserID   int64           `json:"user_id"`
	Username string          `json:"username"`
	Role     model.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		db:     database.GetDB(),
		jwtCfg: &cfg.JWT,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("username = ? AND status = 1", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("query user error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, expiresAt, err := s.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("generate token error: %w", err)
	}

	now := time.Now()
	s.db.WithContext(ctx).Model(&user).Update("last_login_at", &now)

	permissions := s.getPermissionsByRole(user.Role)

	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(expiresAt).Seconds()),
		User:        &user,
		Permissions: permissions,
	}, nil
}

func (s *AuthService) GenerateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.jwtCfg.ExpireHours) * time.Hour)

	claims := CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "dangerous-drive-guard",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	s.saveToken(user.ID, tokenStr, expiresAt)

	return tokenStr, expiresAt, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, oldToken string) (*RefreshResult, error) {
	claims, err := s.parseToken(oldToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return nil, errors.New("token is still valid, no need to refresh")
	}

	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ? AND status = 1", claims.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	token, expiresAt, err := s.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("generate token error: %w", err)
	}

	s.deleteToken(oldToken)

	return &RefreshResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(expiresAt).Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.deleteToken(token)
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ? AND status = 1", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("get user error: %w", err)
	}
	return &user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPwd, newPwd string) error {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPwd)); err != nil {
		return errors.New("old password is incorrect")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password error: %w", err)
	}

	return s.db.WithContext(ctx).Model(&user).Update("password", string(hashedPwd)).Error
}

func (s *AuthService) ValidateToken(tokenStr string) (*CustomClaims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}

	if !s.isTokenValid(tokenStr) {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

func (s *AuthService) parseToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) saveToken(userID int64, token string, expiresAt time.Time) {
	userToken := &model.UserToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	s.db.Create(userToken)
}

func (s *AuthService) deleteToken(token string) error {
	return s.db.Where("token = ?", token).Delete(&model.UserToken{}).Error
}

func (s *AuthService) isTokenValid(token string) bool {
	var count int64
	s.db.Model(&model.UserToken{}).Where("token = ? AND expires_at > ?", token, time.Now()).Count(&count)
	return count > 0
}

func (s *AuthService) getPermissionsByRole(role model.UserRole) []string {
	permissions := make(map[model.UserRole][]string)
	permissions[model.RoleAdmin] = []string{
		"user:create", "user:update", "user:delete", "user:list",
		"vehicle:create", "vehicle:update", "vehicle:delete", "vehicle:list",
		"route:plan", "route:view",
		"fatigue:view", "fatigue:export",
		"monitor:view", "monitor:control",
		"system:config", "system:log",
	}
	permissions[model.RoleDispatcher] = []string{
		"user:list", "user:view",
		"vehicle:list", "vehicle:view",
		"route:plan", "route:view",
		"fatigue:view",
		"monitor:view",
	}
	permissions[model.RoleDriver] = []string{
		"user:view",
		"vehicle:view",
		"route:view",
		"fatigue:view",
	}
	permissions[model.RoleEscort] = []string{
		"vehicle:view",
		"route:view",
		"monitor:view",
	}
	permissions[model.RoleViewer] = []string{
		"vehicle:view",
		"route:view",
		"fatigue:view",
		"monitor:view",
	}

	if perms, ok := permissions[role]; ok {
		return perms
	}
	return []string{}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
