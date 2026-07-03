package service

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/jwt"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/response"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userRepo   *repository.UserRepo
	memberRepo *repository.MemberRepo
	jwtSecret  string
	jwtExpire  int
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepo, memberRepo *repository.MemberRepo, jwtSecret string, jwtExpire int) *UserService {
	return &UserService{
		userRepo:   userRepo,
		memberRepo: memberRepo,
		jwtSecret:  jwtSecret,
		jwtExpire:  jwtExpire,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required" validate:"username"`
	Email    string `json:"email" binding:"required" validate:"email"`
	Password string `json:"password" binding:"required" validate:"password"`
	Nickname string `json:"nickname"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// Register 用户注册
func (s *UserService) Register(req *RegisterRequest) (*RegisterResponse, int, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, response.ErrUserExists, errors.New("用户名已存在")
	}


	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("密码加密失败: %v", err)
		return nil, response.ErrInternal, err
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(user); err != nil {
		logger.Errorf("创建用户失败: %v", err)
		return nil, response.ErrInternal, err
	}

	// 生成 Token
	token, err := jwt.GenerateToken(user.ID, user.Username, s.jwtSecret, s.jwtExpire)
	if err != nil {
		logger.Errorf("生成 Token 失败: %v", err)
		return nil, response.ErrInternal, err
	}
	refreshToken, _ := jwt.GenerateRefreshToken(user.ID, user.Username, s.jwtSecret, s.jwtExpire*3)

	return &RegisterResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtExpire * 3600,
	}, response.CodeSuccess, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*LoginResponse, int, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrUserNotFound, errors.New("用户不存在")
		}
		logger.Errorf("查询用户失败: %v", err)
		return nil, response.ErrInternal, err
	}

	if user.Status == 0 {
		return nil, response.ErrUserDisabled, errors.New("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, response.ErrPasswordWrong, errors.New("密码错误")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, s.jwtSecret, s.jwtExpire)
	if err != nil {
		logger.Errorf("生成 Token 失败: %v", err)
		return nil, response.ErrInternal, err
	}

	// 同时生成 refresh token
	refreshToken, _ := jwt.GenerateRefreshToken(user.ID, user.Username, s.jwtSecret, s.jwtExpire*3)

	return &LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtExpire * 3600,
	}, response.CodeSuccess, nil
}

// GetProfile 获取用户信息
func (s *UserService) GetProfile(userID int64) (*model.User, int, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrUserNotFound, err
		}
		return nil, response.ErrInternal, err
	}
	return user, response.CodeSuccess, nil
}

// AnonymousRegister creates a guest user and returns a token directly
func (s *UserService) AnonymousRegister() (*LoginResponse, int, error) {
	username := fmt.Sprintf("guest_%d", time.Now().UnixMilli())
	nickname := "匿名用户"
	guestPassword := fmt.Sprintf("xa_%d", time.Now().UnixNano())
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(guestPassword), bcrypt.DefaultCost)
	user := &model.User{Username: username, Email: username + "@anonymous.xa", Password: string(hashedPassword), Nickname: nickname, Status: 1}
	if err := s.userRepo.Create(user); err != nil {
		logger.Errorf("匿名用户创建失败: %v", err)
		return nil, response.ErrInternal, err
	}
	token, err := jwt.GenerateToken(user.ID, user.Username, s.jwtSecret, s.jwtExpire)
	if err != nil { return nil, response.ErrInternal, err }
	refreshToken, _ := jwt.GenerateRefreshToken(user.ID, user.Username, s.jwtSecret, s.jwtExpire*3)
	return &LoginResponse{UserID: user.ID, Username: user.Username, Nickname: user.Nickname, Token: token, RefreshToken: refreshToken, ExpiresIn: s.jwtExpire * 3600}, response.CodeSuccess, nil
}
