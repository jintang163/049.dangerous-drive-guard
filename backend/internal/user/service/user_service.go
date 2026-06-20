package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

type PageResult struct {
	List     []*model.User `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

func NewUserService(cfg *config.Config) *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

func (s *UserService) List(ctx context.Context, page, pageSize int, keyword, role string) (*PageResult, error) {
	var users []*model.User
	var total int64

	query := s.db.WithContext(ctx).Table("users").Where("status = 1")

	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count users error: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, fmt.Errorf("query users error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u model.User
		var lastLoginAt *time.Time
		rows.Scan(&u.ID, &u.Username, &u.Password, &u.RealName, &u.Phone,
			&u.Email, &u.Role, &u.OrgID, &u.AvatarURL, &u.IDCard,
			&u.LicenseNo, &u.LicenseType, &u.Status, &lastLoginAt,
			&u.CreatedAt, &u.UpdatedAt)
		u.LastLoginAt = lastLoginAt
		users = append(users, &u)
	}

	return &PageResult{
		List:     users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) Get(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ? AND status = 1", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user error: %w", err)
	}
	return &user, nil
}

func (s *UserService) Create(ctx context.Context, user *model.User, password string) (*model.User, error) {
	var count int64
	s.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password error: %w", err)
	}
	user.Password = string(hashedPwd)

	if user.Status == 0 {
		user.Status = 1
	}

	now := time.Now()
	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO users
		(username, password, real_name, phone, email, role, org_id,
		 avatar_url, id_card, license_no, license_type, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Password, user.RealName, user.Phone, user.Email,
		user.Role, user.OrgID, user.AvatarURL, user.IDCard, user.LicenseNo,
		user.LicenseType, user.Status, now, now,
	)
	if result.Error != nil {
		return nil, fmt.Errorf("create user error: %w", result.Error)
	}

	var id int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&id)
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id int64, data map[string]interface{}) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user error: %w", err)
	}

	data["updated_at"] = time.Now()

	if err := s.db.WithContext(ctx).Model(&user).Updates(data).Error; err != nil {
		return nil, fmt.Errorf("update user error: %w", err)
	}

	return s.Get(ctx, id)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("status", 0)
	if result.Error != nil {
		return fmt.Errorf("delete user error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, id int64) (string, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ? AND status = 1", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", fmt.Errorf("get user error: %w", err)
	}

	newPassword := generateRandomPassword(10)

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password error: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&user).Update("password", string(hashedPwd)).Error; err != nil {
		return "", fmt.Errorf("reset password error: %w", err)
	}

	return newPassword, nil
}

func generateRandomPassword(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "Abc123456!"
	}

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*"
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}

	password := string(b)
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		} else if c >= 'a' && c <= 'z' {
			hasLower = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasUpper {
		password = "A" + password[1:]
	}
	if !hasLower {
		password = password[:1] + "a" + password[2:]
	}
	if !hasDigit {
		password = password[:2] + "1" + password[3:]
	}

	return password
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
