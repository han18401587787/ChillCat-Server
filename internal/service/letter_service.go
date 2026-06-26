package service

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/response"
	"errors"
	"time"

	"gorm.io/gorm"
)

// LetterService 感谢信服务
type LetterService struct{ repo *repository.LetterRepo }

// NewLetterService 创建感谢信服务
func NewLetterService(repo *repository.LetterRepo) *LetterService {
	return &LetterService{repo: repo}
}

// CreateLetterRequest 创建感谢信请求
type CreateLetterRequest struct {
	ReceiverID  int64  `json:"receiver_id" binding:"required"`
	Content     string `json:"content" binding:"required,max=1000"`
	IsAnonymous bool   `json:"is_anonymous"`
	IsPublic    bool   `json:"is_public"`
}

// LetterVO 感谢信视图对象
type LetterVO struct {
	ID           int64  `json:"id"`
	SenderID     int64  `json:"sender_id"`
	SenderName   string `json:"sender_name"`
	ReceiverID   int64  `json:"receiver_id"`
	ReceiverName string `json:"receiver_name,omitempty"`
	Content      string `json:"content"`
	IsAnonymous  bool   `json:"is_anonymous"`
	IsPublic     bool   `json:"is_public"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// Create 创建感谢信
func (s *LetterService) Create(senderID int64, req *CreateLetterRequest) (*LetterVO, int, error) {
	// 不能给自己发信
	if senderID == req.ReceiverID {
		return nil, response.ErrValidation, errors.New("不能给自己发感谢信")
	}

	// 检查每日发送上限（3封）
	todayCount, err := s.repo.CountTodayBySender(senderID)
	if err != nil {
		return nil, response.ErrInternal, err
	}
	if todayCount >= 3 {
		return nil, response.ErrRateLimit, errors.New("今日发送已达上限（3封），明天再来吧")
	}

	// is_public 默认 true
	isPublic := true           //nolint:staticcheck
	if !req.IsPublic {
		isPublic = false
	}

	letter := &model.ThankYouLetter{
		SenderID:    senderID,
		ReceiverID:  req.ReceiverID,
		Content:     req.Content,
		IsAnonymous: req.IsAnonymous,
		IsPublic:    isPublic,
		Status:      "sent",
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(letter); err != nil {
		return nil, response.ErrInternal, err
	}

	return s.toVO(letter), response.CodeSuccess, nil
}

// Sent 查询我发出的信件
func (s *LetterService) Sent(senderID int64, page, pageSize int) (*response.Page, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	items, total, err := s.repo.ListBySender(senderID, page, pageSize)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	vos := make([]LetterVO, 0, len(items))
	for _, l := range items {
		vos = append(vos, *s.toVO(&l))
	}

	return &response.Page{List: vos, Total: total, Page: page, PageSize: pageSize}, response.CodeSuccess, nil
}

// Received 查询我收到的信件
func (s *LetterService) Received(receiverID int64, page, pageSize int) (*response.Page, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	items, total, err := s.repo.ListByReceiver(receiverID, page, pageSize)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	vos := make([]LetterVO, 0, len(items))
	for _, l := range items {
		v := s.toVO(&l)
		// 匿名信隐藏发送者信息
		if l.IsAnonymous {
			v.SenderID = 0
			v.SenderName = "匿名用户"
		}
		vos = append(vos, *v)
	}

	return &response.Page{List: vos, Total: total, Page: page, PageSize: pageSize}, response.CodeSuccess, nil
}

// Get 获取信件详情
func (s *LetterService) Get(letterID int64, userID int64) (*LetterVO, int, error) {
	letter, err := s.repo.GetByID(letterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrNotFound, err
		}
		return nil, response.ErrInternal, err
	}

	// 只有发件人或收件人可以查看
	if letter.SenderID != userID && letter.ReceiverID != userID {
		return nil, response.ErrForbidden, errors.New("无权查看该信件")
	}

	v := s.toVO(letter)
	if letter.IsAnonymous && letter.SenderID != userID {
		v.SenderID = 0
		v.SenderName = "匿名用户"
	}

	return v, response.CodeSuccess, nil
}

// toVO 转换为视图对象
func (s *LetterService) toVO(letter *model.ThankYouLetter) *LetterVO {
	return &LetterVO{
		ID:          letter.ID,
		SenderID:    letter.SenderID,
		ReceiverID:  letter.ReceiverID,
		Content:     letter.Content,
		IsAnonymous: letter.IsAnonymous,
		IsPublic:    letter.IsPublic,
		Status:      letter.Status,
		CreatedAt:   letter.CreatedAt.Format("2006-01-02 15:04"),
	}
}
