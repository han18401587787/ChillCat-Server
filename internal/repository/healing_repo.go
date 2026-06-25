package repository

import (
	"chillcat-server/internal/model"
	"time"

	"gorm.io/gorm"
)

// HealingRepo 稳情计划数据仓库
type HealingRepo struct{ db *gorm.DB }

// NewHealingRepo 创建稳情计划仓库实例
func NewHealingRepo(db *gorm.DB) *HealingRepo {
	return &HealingRepo{db: db}
}

// ─── 计划操作 ──────────────────────────────────────────────────

// CreatePlan 创建稳情计划
func (r *HealingRepo) CreatePlan(plan *model.HealingPlan) error {
	return r.db.Create(plan).Error
}

// FindActivePlan 查找用户当前生效中的计划
func (r *HealingRepo) FindActivePlan(userID int64) (*model.HealingPlan, error) {
	var plan model.HealingPlan
	err := r.db.Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at DESC").First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// FindPlanByID 根据 ID 查找计划（校验用户归属）
func (r *HealingRepo) FindPlanByID(planID, userID int64) (*model.HealingPlan, error) {
	var plan model.HealingPlan
	err := r.db.Where("id = ? AND user_id = ?", planID, userID).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// CompletePlan 完成计划
func (r *HealingRepo) CompletePlan(planID int64) error {
	return r.db.Model(&model.HealingPlan{}).Where("id = ?", planID).
		Update("status", "completed").Error
}

// ─── 任务操作 ──────────────────────────────────────────────────

// BatchCreateTasks 批量创建任务
func (r *HealingRepo) BatchCreateTasks(tasks []model.HealingPlanTask) error {
	return r.db.Create(&tasks).Error
}

// FindTasksByPlanID 根据计划 ID 获取所有任务
func (r *HealingRepo) FindTasksByPlanID(planID int64) ([]model.HealingPlanTask, error) {
	var tasks []model.HealingPlanTask
	err := r.db.Where("plan_id = ?", planID).Order("day_number ASC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []model.HealingPlanTask{}
	}
	return tasks, nil
}

// FindTaskByID 根据任务 ID 查找（校验计划归属）
func (r *HealingRepo) FindTaskByID(taskID, planID int64) (*model.HealingPlanTask, error) {
	var task model.HealingPlanTask
	err := r.db.Where("id = ? AND plan_id = ?", taskID, planID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// CompleteTask 完成任务
func (r *HealingRepo) CompleteTask(taskID int64) error {
	now := time.Now()
	return r.db.Model(&model.HealingPlanTask{}).Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": &now,
		}).Error
}

// CountIncompleteTasks 统计计划中未完成的任务数
func (r *HealingRepo) CountIncompleteTasks(planID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.HealingPlanTask{}).
		Where("plan_id = ? AND is_completed = ?", planID, false).Count(&count).Error
	return count, err
}
