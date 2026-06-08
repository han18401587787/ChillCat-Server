package service

import (
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/response"
)

type MessageService struct{ msgRepo *repository.MessageRepo }

func NewMessageService(msgRepo *repository.MessageRepo) *MessageService {
	return &MessageService{msgRepo: msgRepo}
}

type MessageVO struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	MsgType string `json:"msg_type"`
	IsRead  bool   `json:"is_read"`
	Created string `json:"created_at"`
}

func (s *MessageService) List(userID int64, page, pageSize int) (*response.Page, int, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 50 { pageSize = 10 }
	items, total, err := s.msgRepo.List(userID, page, pageSize)
	if err != nil { return nil, response.ErrInternal, err }
	var vos []MessageVO
	for _, m := range items {
		vos = append(vos, MessageVO{
			ID: m.ID, Title: m.Title, Content: m.Content,
			MsgType: m.MsgType, IsRead: m.IsRead,
			Created: m.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return &response.Page{List: vos, Total: total, Page: page, PageSize: pageSize}, response.CodeSuccess, nil
}

func (s *MessageService) UnreadCount(userID int64) (int64, int, error) {
	count, err := s.msgRepo.UnreadCount(userID)
	if err != nil { return 0, response.ErrInternal, err }
	return count, response.CodeSuccess, nil
}

func (s *MessageService) MarkRead(userID, id int64) (int, error) {
	if err := s.msgRepo.MarkRead(userID, id); err != nil {
		return response.ErrInternal, err
	}
	return response.CodeSuccess, nil
}
