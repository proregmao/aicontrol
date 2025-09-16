package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronScheduler Cron调度器
type CronScheduler struct {
	cron    *cron.Cron
	tasks   map[int]*ScheduledTask
	mutex   sync.RWMutex
	logger  *log.Logger
	running bool
}

// ScheduledTask 定时任务
type ScheduledTask struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CronExpr    string                 `json:"cron_expr"`
	TaskType    string                 `json:"task_type"`
	TaskConfig  map[string]interface{} `json:"task_config"`
	Enabled     bool                   `json:"enabled"`
	Handler     TaskHandler            `json:"-"`
	EntryID     cron.EntryID           `json:"-"`
	LastRun     time.Time              `json:"last_run"`
	NextRun     time.Time              `json:"next_run"`
	RunCount    int64                  `json:"run_count"`
	SuccessCount int64                 `json:"success_count"`
	FailureCount int64                 `json:"failure_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TaskHandler 任务处理器接口
type TaskHandler interface {
	Execute(ctx context.Context, config map[string]interface{}) error
	GetType() string
}

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID        int64     `json:"id"`
	TaskID    int       `json:"task_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int64     `json:"duration"`
	Output    string    `json:"output"`
	Error     string    `json:"error"`
}

// NewCronScheduler 创建新的Cron调度器
func NewCronScheduler() *CronScheduler {
	return &CronScheduler{
		cron:   cron.New(cron.WithSeconds()),
		tasks:  make(map[int]*ScheduledTask),
		logger: log.New(log.Writer(), "[SCHEDULER] ", log.LstdFlags),
	}
}

// Start 启动调度器
func (s *CronScheduler) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("调度器已经在运行")
	}

	s.cron.Start()
	s.running = true
	s.logger.Println("Cron调度器已启动")
	return nil
}

// Stop 停止调度器
func (s *CronScheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("调度器未运行")
	}

	ctx := s.cron.Stop()
	<-ctx.Done()
	s.running = false
	s.logger.Println("Cron调度器已停止")
	return nil
}

// AddTask 添加定时任务
func (s *CronScheduler) AddTask(task *ScheduledTask) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if task.Handler == nil {
		return fmt.Errorf("任务处理器不能为空")
	}

	// 验证Cron表达式
	_, err := cron.ParseStandard(task.CronExpr)
	if err != nil {
		return fmt.Errorf("无效的Cron表达式: %v", err)
	}

	// 如果任务已存在，先移除
	if existingTask, exists := s.tasks[task.ID]; exists {
		if existingTask.EntryID != 0 {
			s.cron.Remove(existingTask.EntryID)
		}
	}

	// 添加到Cron调度器
	if task.Enabled {
		entryID, err := s.cron.AddFunc(task.CronExpr, func() {
			s.executeTask(task)
		})
		if err != nil {
			return fmt.Errorf("添加Cron任务失败: %v", err)
		}
		task.EntryID = entryID
		
		// 计算下次运行时间
		entries := s.cron.Entries()
		for _, entry := range entries {
			if entry.ID == entryID {
				task.NextRun = entry.Next
				break
			}
		}
	}

	task.UpdatedAt = time.Now()
	s.tasks[task.ID] = task
	s.logger.Printf("任务已添加: %s (ID: %d)", task.Name, task.ID)
	return nil
}

// RemoveTask 移除定时任务
func (s *CronScheduler) RemoveTask(taskID int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %d", taskID)
	}

	if task.EntryID != 0 {
		s.cron.Remove(task.EntryID)
	}

	delete(s.tasks, taskID)
	s.logger.Printf("任务已移除: %s (ID: %d)", task.Name, taskID)
	return nil
}

// EnableTask 启用任务
func (s *CronScheduler) EnableTask(taskID int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %d", taskID)
	}

	if task.Enabled {
		return nil // 已经启用
	}

	// 添加到Cron调度器
	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executeTask(task)
	})
	if err != nil {
		return fmt.Errorf("启用任务失败: %v", err)
	}

	task.EntryID = entryID
	task.Enabled = true
	task.UpdatedAt = time.Now()

	// 计算下次运行时间
	entries := s.cron.Entries()
	for _, entry := range entries {
		if entry.ID == entryID {
			task.NextRun = entry.Next
			break
		}
	}

	s.logger.Printf("任务已启用: %s (ID: %d)", task.Name, taskID)
	return nil
}

// DisableTask 禁用任务
func (s *CronScheduler) DisableTask(taskID int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %d", taskID)
	}

	if !task.Enabled {
		return nil // 已经禁用
	}

	if task.EntryID != 0 {
		s.cron.Remove(task.EntryID)
		task.EntryID = 0
	}

	task.Enabled = false
	task.NextRun = time.Time{}
	task.UpdatedAt = time.Now()

	s.logger.Printf("任务已禁用: %s (ID: %d)", task.Name, taskID)
	return nil
}

// ExecuteTaskNow 立即执行任务
func (s *CronScheduler) ExecuteTaskNow(taskID int) error {
	s.mutex.RLock()
	task, exists := s.tasks[taskID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("任务不存在: %d", taskID)
	}

	go s.executeTask(task)
	return nil
}

// GetTask 获取任务
func (s *CronScheduler) GetTask(taskID int) (*ScheduledTask, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %d", taskID)
	}

	return task, nil
}

// GetAllTasks 获取所有任务
func (s *CronScheduler) GetAllTasks() []*ScheduledTask {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// executeTask 执行任务
func (s *CronScheduler) executeTask(task *ScheduledTask) {
	execution := &TaskExecution{
		ID:        time.Now().UnixNano(),
		TaskID:    task.ID,
		Status:    "running",
		StartTime: time.Now(),
	}

	s.logger.Printf("开始执行任务: %s (ID: %d)", task.Name, task.ID)

	// 更新任务统计
	s.mutex.Lock()
	task.LastRun = execution.StartTime
	task.RunCount++
	s.mutex.Unlock()

	// 创建执行上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 执行任务
	err := task.Handler.Execute(ctx, task.TaskConfig)
	execution.EndTime = time.Now()
	execution.Duration = execution.EndTime.Sub(execution.StartTime).Milliseconds()

	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		
		s.mutex.Lock()
		task.FailureCount++
		s.mutex.Unlock()
		
		s.logger.Printf("任务执行失败: %s (ID: %d), 错误: %v", task.Name, task.ID, err)
	} else {
		execution.Status = "success"
		execution.Output = "任务执行成功"
		
		s.mutex.Lock()
		task.SuccessCount++
		s.mutex.Unlock()
		
		s.logger.Printf("任务执行成功: %s (ID: %d), 耗时: %dms", task.Name, task.ID, execution.Duration)
	}

	// 这里可以将执行记录保存到数据库
	s.saveExecution(execution)
}

// saveExecution 保存执行记录
func (s *CronScheduler) saveExecution(execution *TaskExecution) {
	// 这里应该保存到数据库，目前只是日志记录
	executionJSON, _ := json.Marshal(execution)
	s.logger.Printf("任务执行记录: %s", string(executionJSON))
}

// GetStatus 获取调度器状态
func (s *CronScheduler) GetStatus() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	totalTasks := len(s.tasks)
	enabledTasks := 0
	runningTasks := 0

	for _, task := range s.tasks {
		if task.Enabled {
			enabledTasks++
		}
		// 这里可以添加更复杂的运行状态检查
	}

	return map[string]interface{}{
		"running":       s.running,
		"total_tasks":   totalTasks,
		"enabled_tasks": enabledTasks,
		"running_tasks": runningTasks,
		"cron_entries":  len(s.cron.Entries()),
	}
}
