package shutdown

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ShutdownManager 智能断电管理器
type ShutdownManager struct {
	sequences map[int]*ShutdownSequence
	mutex     sync.RWMutex
	logger    *log.Logger
}

// ShutdownSequence 断电序列
type ShutdownSequence struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	BreakerID   int                `json:"breaker_id"`
	Steps       []*ShutdownStep    `json:"steps"`
	Status      string             `json:"status"`
	Progress    int                `json:"progress"`
	StartTime   time.Time          `json:"start_time"`
	EndTime     *time.Time         `json:"end_time,omitempty"`
	ErrorMsg    string             `json:"error_msg,omitempty"`
	Context     context.Context    `json:"-"`
	CancelFunc  context.CancelFunc `json:"-"`
}

// ShutdownStep 断电步骤
type ShutdownStep struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`        // server_shutdown, delay, breaker_off
	ServerID    int           `json:"server_id,omitempty"`
	DelayTime   time.Duration `json:"delay_time,omitempty"`
	Status      string        `json:"status"`      // pending, running, completed, failed
	StartTime   *time.Time    `json:"start_time,omitempty"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	ErrorMsg    string        `json:"error_msg,omitempty"`
	RetryCount  int           `json:"retry_count"`
	MaxRetries  int           `json:"max_retries"`
}

// NewShutdownManager 创建断电管理器
func NewShutdownManager() *ShutdownManager {
	return &ShutdownManager{
		sequences: make(map[int]*ShutdownSequence),
		logger:    log.New(log.Writer(), "[ShutdownManager] ", log.LstdFlags),
	}
}

// CreateShutdownSequence 创建断电序列
func (sm *ShutdownManager) CreateShutdownSequence(name string, breakerID int, serverIDs []int) *ShutdownSequence {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sequenceID := len(sm.sequences) + 1
	ctx, cancel := context.WithCancel(context.Background())
	
	sequence := &ShutdownSequence{
		ID:         sequenceID,
		Name:       name,
		BreakerID:  breakerID,
		Status:     "created",
		Progress:   0,
		Context:    ctx,
		CancelFunc: cancel,
	}
	
	// 创建断电步骤
	stepID := 1
	
	// 1. 服务器关机步骤
	for _, serverID := range serverIDs {
		step := &ShutdownStep{
			ID:         stepID,
			Name:       fmt.Sprintf("关闭服务器 %d", serverID),
			Type:       "server_shutdown",
			ServerID:   serverID,
			Status:     "pending",
			MaxRetries: 3,
		}
		sequence.Steps = append(sequence.Steps, step)
		stepID++
	}
	
	// 2. 延时步骤
	delayStep := &ShutdownStep{
		ID:        stepID,
		Name:      "断电延时等待",
		Type:      "delay",
		DelayTime: 30 * time.Second,
		Status:    "pending",
	}
	sequence.Steps = append(sequence.Steps, delayStep)
	stepID++
	
	// 3. 断路器断电步骤
	breakerStep := &ShutdownStep{
		ID:     stepID,
		Name:   fmt.Sprintf("断路器 %d 断电", breakerID),
		Type:   "breaker_off",
		Status: "pending",
	}
	sequence.Steps = append(sequence.Steps, breakerStep)
	
	sm.sequences[sequenceID] = sequence
	sm.logger.Printf("创建断电序列 %d: %s", sequenceID, name)
	
	return sequence
}

// ExecuteShutdownSequence 执行断电序列
func (sm *ShutdownManager) ExecuteShutdownSequence(sequenceID int) error {
	sm.mutex.RLock()
	sequence, exists := sm.sequences[sequenceID]
	sm.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("断电序列 %d 不存在", sequenceID)
	}
	
	if sequence.Status == "running" {
		return fmt.Errorf("断电序列 %d 正在执行中", sequenceID)
	}
	
	// 启动执行
	go sm.executeSequence(sequence)
	
	return nil
}

// executeSequence 执行序列
func (sm *ShutdownManager) executeSequence(sequence *ShutdownSequence) {
	sequence.Status = "running"
	sequence.StartTime = time.Now()
	sequence.Progress = 0
	
	sm.logger.Printf("开始执行断电序列 %d: %s", sequence.ID, sequence.Name)
	
	totalSteps := len(sequence.Steps)
	
	for i, step := range sequence.Steps {
		// 检查是否被取消
		select {
		case <-sequence.Context.Done():
			sequence.Status = "cancelled"
			sm.logger.Printf("断电序列 %d 被取消", sequence.ID)
			return
		default:
		}
		
		// 执行步骤
		err := sm.executeStep(sequence, step)
		if err != nil {
			step.Status = "failed"
			step.ErrorMsg = err.Error()
			sequence.Status = "failed"
			sequence.ErrorMsg = fmt.Sprintf("步骤 %d 执行失败: %s", step.ID, err.Error())
			sm.logger.Printf("断电序列 %d 执行失败: %s", sequence.ID, err.Error())
			return
		}
		
		step.Status = "completed"
		sequence.Progress = int(float64(i+1) / float64(totalSteps) * 100)
		
		sm.logger.Printf("断电序列 %d 步骤 %d 完成，进度: %d%%", 
			sequence.ID, step.ID, sequence.Progress)
	}
	
	// 序列执行完成
	sequence.Status = "completed"
	sequence.Progress = 100
	endTime := time.Now()
	sequence.EndTime = &endTime
	
	sm.logger.Printf("断电序列 %d 执行完成，耗时: %v", 
		sequence.ID, endTime.Sub(sequence.StartTime))
}

// executeStep 执行步骤
func (sm *ShutdownManager) executeStep(sequence *ShutdownSequence, step *ShutdownStep) error {
	step.Status = "running"
	startTime := time.Now()
	step.StartTime = &startTime
	
	var err error
	
	switch step.Type {
	case "server_shutdown":
		err = sm.executeServerShutdown(step)
	case "delay":
		err = sm.executeDelay(sequence.Context, step)
	case "breaker_off":
		err = sm.executeBreakerOff(step)
	default:
		err = fmt.Errorf("未知的步骤类型: %s", step.Type)
	}
	
	endTime := time.Now()
	step.EndTime = &endTime
	
	if err != nil && step.RetryCount < step.MaxRetries {
		step.RetryCount++
		sm.logger.Printf("步骤 %d 执行失败，重试 %d/%d: %s", 
			step.ID, step.RetryCount, step.MaxRetries, err.Error())
		
		// 重试延时
		time.Sleep(5 * time.Second)
		return sm.executeStep(sequence, step)
	}
	
	return err
}

// executeServerShutdown 执行服务器关机
func (sm *ShutdownManager) executeServerShutdown(step *ShutdownStep) error {
	sm.logger.Printf("正在关闭服务器 %d...", step.ServerID)
	
	// 模拟服务器关机过程
	// 实际实现中这里会调用SSH连接执行关机命令
	time.Sleep(10 * time.Second)
	
	// 模拟关机成功
	sm.logger.Printf("服务器 %d 关机完成", step.ServerID)
	return nil
}

// executeDelay 执行延时
func (sm *ShutdownManager) executeDelay(ctx context.Context, step *ShutdownStep) error {
	sm.logger.Printf("开始延时等待 %v...", step.DelayTime)
	
	select {
	case <-time.After(step.DelayTime):
		sm.logger.Printf("延时等待完成")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("延时被取消")
	}
}

// executeBreakerOff 执行断路器断电
func (sm *ShutdownManager) executeBreakerOff(step *ShutdownStep) error {
	sm.logger.Printf("正在执行断路器断电...")
	
	// 模拟断路器断电过程
	// 实际实现中这里会调用断路器控制接口
	time.Sleep(2 * time.Second)
	
	sm.logger.Printf("断路器断电完成")
	return nil
}

// GetShutdownSequence 获取断电序列
func (sm *ShutdownManager) GetShutdownSequence(sequenceID int) (*ShutdownSequence, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	sequence, exists := sm.sequences[sequenceID]
	return sequence, exists
}

// GetAllShutdownSequences 获取所有断电序列
func (sm *ShutdownManager) GetAllShutdownSequences() []*ShutdownSequence {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	sequences := make([]*ShutdownSequence, 0, len(sm.sequences))
	for _, sequence := range sm.sequences {
		sequences = append(sequences, sequence)
	}
	
	return sequences
}

// CancelShutdownSequence 取消断电序列
func (sm *ShutdownManager) CancelShutdownSequence(sequenceID int) error {
	sm.mutex.RLock()
	sequence, exists := sm.sequences[sequenceID]
	sm.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("断电序列 %d 不存在", sequenceID)
	}
	
	if sequence.Status != "running" {
		return fmt.Errorf("断电序列 %d 未在运行中", sequenceID)
	}
	
	sequence.CancelFunc()
	sm.logger.Printf("取消断电序列 %d", sequenceID)
	
	return nil
}
