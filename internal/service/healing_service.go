package service

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/response"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ──────────────────────────────────────────────────────────────
// 稳情计划服务 — 一周治愈计划生成与管理
// 根据当前情绪状态为用户定制 7 天治愈任务清单
// ──────────────────────────────────────────────────────────────

// ─── 请求/响应结构 ─────────────────────────────────────────────

// PlanVO 稳情计划视图对象
type PlanVO struct {
	PlanID    int64        `json:"plan_id"`
	StartDate string       `json:"start_date"`
	EndDate   string       `json:"end_date"`
	Status    string       `json:"status"`
	Tasks     []PlanTaskVO `json:"tasks"`
}

// PlanTaskVO 任务视图对象
type PlanTaskVO struct {
	ID          int64  `json:"id"`
	DayNumber   int    `json:"day_number"`
	TaskType    string `json:"task_type"`
	TaskTitle   string `json:"task_title"`
	TaskDesc    string `json:"task_desc"`
	IsCompleted bool   `json:"is_completed"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// ─── 稳情计划服务 ──────────────────────────────────────────────

// HealingService 稳情计划服务
type HealingService struct {
	repo      *repository.HealingRepo
	aiService *AIService
}

// NewHealingService 创建稳情计划服务实例
func NewHealingService(repo *repository.HealingRepo, aiService *AIService) *HealingService {
	return &HealingService{repo: repo, aiService: aiService}
}

// ─── 预设的 7 天治愈任务模板 ────────────────────────────────────

// healingTemplates 每日任务模板（day_number → 当日任务列表）
var healingTemplates = [][]struct {
	TaskType  string
	TaskTitle string
	TaskDesc  string
}{
	// 第 1 天：觉察与呼吸
	{
		{"breathing", "4-7-8呼吸法", "找一个安静的地方，吸气4秒，屏息7秒，缓慢呼气8秒，重复5轮"},
		{"journal", "情绪天气日记", "用「晴/多云/小雨/暴雨」描述今天的心情，然后写下为什么是这个天气"},
	},
	// 第 2 天：感恩与连接
	{
		{"journal", "感恩三件事", "写下今天让你感到温暖或感激的三件小事，不论多么微小"},
		{"active", "微小连接", "给一位很久没联系的朋友发一句问候，哪怕只是一个表情包"},
	},
	// 第 3 天：身体关怀
	{
		{"active", "身体扫描", "闭上眼睛，从头顶到脚尖慢慢感受身体的每个部位，注意紧张的地方"},
		{"breathing", "能量呼吸", "吸气时想象吸收大地能量，呼气时释放所有疲惫，重复8轮"},
	},
	// 第 4 天：自我对话
	{
		{"journal", "给自己的一封信", "像对待最好的朋友一样给自己写一封信，表达理解、支持和鼓励"},
		{"meditation", "正念行走", "慢走10分钟，注意每一步脚底与地面的接触，感受此时此刻"},
	},
	// 第 5 天：创造力释放
	{
		{"music", "治愈歌单", "创建一份属于你的治愈歌单，听歌时闭上眼睛感受旋律的流动"},
		{"journal", "自由书写", "设定10分钟计时器，不停笔地写下脑海中所有想法，不加评判"},
	},
	// 第 6 天：放下与接纳
	{
		{"meditation", "放下冥想", "想象把烦恼写在叶子上，放入溪流，看着它慢慢漂远，不再回头"},
		{"breathing", "阳光呼吸", "想象吸入温暖的金色阳光，呼出灰色的负担，重复5轮"},
	},
	// 第 7 天：庆祝与展望
	{
		{"journal", "本周收获", "回顾这一周的治愈旅程，写下你的收获、发现和想继续保持的习惯"},
		{"active", "庆祝仪式", "为自己做一件期待已久的小事——一杯好咖啡、一束花、一次散步"},
	},
}

// ─── 公开方法 ──────────────────────────────────────────────────

// GetCurrentPlan 获取当前生效中的稳情计划
func (s *HealingService) GetCurrentPlan(userID int64) (*PlanVO, int, error) {
	plan, err := s.repo.FindActivePlan(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有生效中的计划，返回空
			return nil, response.CodeSuccess, nil
		}
		return nil, response.ErrInternal, err
	}

	tasks, err := s.repo.FindTasksByPlanID(plan.ID)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	return s.buildPlanVO(plan, tasks), response.CodeSuccess, nil
}

// GeneratePlan 生成一周稳情计划
func (s *HealingService) GeneratePlan(userID int64) (*PlanVO, int, error) {
	// 检查是否已有生效中的计划
	existing, _ := s.repo.FindActivePlan(userID)
	if existing != nil {
		// 已有生效计划，直接返回
		tasks, err := s.repo.FindTasksByPlanID(existing.ID)
		if err != nil {
			return nil, response.ErrInternal, err
		}
		return s.buildPlanVO(existing, tasks), response.CodeSuccess, nil
	}

	// 生成新计划
	today := time.Now()
	startDate := today.Format("2006-01-02")
	endDate := today.AddDate(0, 0, 6).Format("2006-01-02")

	plan := &model.HealingPlan{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    "active",
	}

	if err := s.repo.CreatePlan(plan); err != nil {
		return nil, response.ErrInternal, err
	}

	// 生成 7 天任务
	tasks := make([]model.HealingPlanTask, 0, 14)
	for dayIdx, dayTemplates := range healingTemplates {
		for _, tmpl := range dayTemplates {
			tasks = append(tasks, model.HealingPlanTask{
				PlanID:    plan.ID,
				DayNumber: dayIdx + 1,
				TaskType:  tmpl.TaskType,
				TaskTitle: tmpl.TaskTitle,
				TaskDesc:  tmpl.TaskDesc,
			})
		}
	}

	if err := s.repo.BatchCreateTasks(tasks); err != nil {
		return nil, response.ErrInternal, err
	}

	createdTasks, err := s.repo.FindTasksByPlanID(plan.ID)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	return s.buildPlanVO(plan, createdTasks), response.CodeSuccess, nil
}

// CompleteTask 完成稳情计划中的某个任务
func (s *HealingService) CompleteTask(userID int64, taskID int64) (*PlanVO, int, error) {
	// 校验用户是否有生效计划
	plan, err := s.repo.FindActivePlan(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrNotFound, fmt.Errorf("没有生效中的稳情计划")
		}
		return nil, response.ErrInternal, err
	}

	// 校验任务归属于该计划
	task, err := s.repo.FindTaskByID(taskID, plan.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrNotFound, fmt.Errorf("任务不存在")
		}
		return nil, response.ErrInternal, err
	}

	// 如果已完成则跳过
	if task.IsCompleted {
		// 返回当前计划状态
		tasks, _ := s.repo.FindTasksByPlanID(plan.ID)
		return s.buildPlanVO(plan, tasks), response.CodeSuccess, nil
	}

	// 完成任务
	if err := s.repo.CompleteTask(taskID); err != nil {
		return nil, response.ErrInternal, err
	}

	// 检查是否所有任务都完成，若是则自动完成计划
	remaining, _ := s.repo.CountIncompleteTasks(plan.ID)
	if remaining == 0 {
		_ = s.repo.CompletePlan(plan.ID)
	}

	// 刷新计划数据
	plan, _ = s.repo.FindActivePlan(userID)
	if plan == nil {
		// 计划已完成，重新查询
		plan, _ = s.repo.FindPlanByID(task.PlanID, userID)
	}
	tasks, err := s.repo.FindTasksByPlanID(task.PlanID)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	return s.buildPlanVO(plan, tasks), response.CodeSuccess, nil
}

// ─── 内部辅助方法 ──────────────────────────────────────────────

// buildPlanVO 构建计划视图对象
func (s *HealingService) buildPlanVO(plan *model.HealingPlan, tasks []model.HealingPlanTask) *PlanVO {
	taskVOs := make([]PlanTaskVO, 0, len(tasks))
	for _, t := range tasks {
		vo := PlanTaskVO{
			ID:          t.ID,
			DayNumber:   t.DayNumber,
			TaskType:    t.TaskType,
			TaskTitle:   t.TaskTitle,
			TaskDesc:    t.TaskDesc,
			IsCompleted: t.IsCompleted,
		}
		if t.CompletedAt != nil {
			vo.CompletedAt = t.CompletedAt.Format("2006-01-02 15:04")
		}
		taskVOs = append(taskVOs, vo)
	}

	return &PlanVO{
		PlanID:    plan.ID,
		StartDate: plan.StartDate,
		EndDate:   plan.EndDate,
		Status:    plan.Status,
		Tasks:     taskVOs,
	}
}
