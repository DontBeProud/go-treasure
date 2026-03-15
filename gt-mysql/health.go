package gtmysql

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"gorm.io/gorm"
)

// DBWithHealthCheck 包含健康检查功能的数据库连接
type DBWithHealthCheck interface {
	// Check 执行健康检查并更新状态
	Check(timeout time.Duration) (bool, error)
	// GetDBHealthStatus 获取数据库健康状态信息
	GetDBHealthStatus() (*DBHealthStatus, error)
	// GetHealthyThreshold 获取健康阈值
	GetHealthyThreshold() int
	// SetHealthyThreshold 设置健康阈值
	SetHealthyThreshold(threshold int)
	// StartAutoCheck 启动自动健康检查
	StartAutoCheck(pool *ants.Pool, interval, timeout time.Duration) error
	// StopAutoCheck 停止自动健康检查
	StopAutoCheck()
	// IsAutoCheckRunning 检查是否正在自动检查
	IsAutoCheckRunning() bool
	// Close 关闭数据库连接，自动停止健康检查
	Close() error
	// GetDB 获取一个新的数据库会话（Fork模式）
	GetDB() *gorm.DB
	// GetDBWithContext 获取一个带上下文的新数据库会话（Fork模式）
	GetDBWithContext(ctx context.Context) *gorm.DB
	// GetRawDB 获取原始数据库对象
	GetRawDB() *gorm.DB
}

// DBHealthStatus 数据库健康状态信息
type DBHealthStatus struct {
	IsHealthy           bool          `json:"is_healthy"`
	LastCheckTime       time.Time     `json:"last_check_time"`
	LastError           string        `json:"last_error"`
	LastErrorTime       time.Time     `json:"last_error_time"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
	TotalChecks         int           `json:"total_checks"`
	TotalSuccesses      int           `json:"total_successes"`
	TotalFailures       int           `json:"total_failures"`
	MaxOpenConnections  int           `json:"max_open_connections"`
	OpenConnections     int           `json:"open_connections"`
	InUse               int           `json:"in_use"`
	Idle                int           `json:"idle"`
	WaitCount           int64         `json:"wait_count"`
	WaitDuration        time.Duration `json:"wait_duration"`
	MaxIdleClosed       int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed   int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed   int64         `json:"max_lifetime_closed"`
}

// NewDBWithHealthCheck 创建一个带健康检查的数据库连接
// healthyThreshold: 连续失败多少次后标记为不健康，默认为3
func NewDBWithHealthCheck(db *gorm.DB, healthyThreshold int) (DBWithHealthCheck, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if healthyThreshold <= 0 {
		healthyThreshold = 3
	}
	return &dbWithHealthCheck{
		DB:               db,
		isHealthy:        true,
		lastCheckTime:    time.Now(),
		healthyThreshold: healthyThreshold,
	}, nil
}

type dbWithHealthCheck struct {
	*gorm.DB
	healthyThreshold    int
	mu                  sync.RWMutex
	isHealthy           bool
	lastCheckTime       time.Time
	lastError           error
	lastErrorTime       time.Time
	consecutiveFailures int
	totalChecks         int
	totalSuccesses      int
	totalFailures       int

	autoCheckMu      sync.Mutex
	autoCheckRunning bool
	autoCheckStop    chan struct{}
	autoCheckPool    *ants.Pool
	autoCheckWg      sync.WaitGroup
}

// Check 执行健康检查并更新状态
func (d *dbWithHealthCheck) Check(timeout time.Duration) (bool, error) {
	if d.DB == nil {
		err := fmt.Errorf("db is nil")
		d.updateFailStatus(err)
		return false, err
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		d.updateFailStatus(err)
		return false, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		d.updateFailStatus(err)
		return false, fmt.Errorf("ping failed: %w", err)
	}

	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		err = fmt.Errorf("no open connections")
		d.updateFailStatus(err)
		return false, err
	}

	if err = d.DB.Exec("SELECT 1").Error; err != nil {
		d.updateFailStatus(err)
		return false, fmt.Errorf("test query failed: %w", err)
	}

	d.updateSuccessStatus()
	return true, nil
}

func (d *dbWithHealthCheck) updateFailStatus(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.totalChecks++
	d.consecutiveFailures++
	d.totalFailures++
	d.lastError = err
	d.lastErrorTime = time.Now()
	d.lastCheckTime = time.Now()
	if d.consecutiveFailures >= d.healthyThreshold {
		d.isHealthy = false
	}
}

func (d *dbWithHealthCheck) updateSuccessStatus() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.totalChecks++
	d.totalSuccesses++
	d.consecutiveFailures = 0
	d.isHealthy = true
	d.lastCheckTime = time.Now()
}

// GetDBHealthStatus 获取数据库健康状态信息
func (d *dbWithHealthCheck) GetDBHealthStatus() (*DBHealthStatus, error) {
	d.mu.RLock()
	status := &DBHealthStatus{
		IsHealthy:           d.isHealthy,
		LastCheckTime:       d.lastCheckTime,
		LastErrorTime:       d.lastErrorTime,
		ConsecutiveFailures: d.consecutiveFailures,
		TotalChecks:         d.totalChecks,
		TotalSuccesses:      d.totalSuccesses,
		TotalFailures:       d.totalFailures,
	}
	if d.lastError != nil {
		status.LastError = d.lastError.Error()
	}
	d.mu.RUnlock()

	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			return status, fmt.Errorf("failed to get sql.DB: %w", err)
		}
		stats := sqlDB.Stats()
		status.MaxOpenConnections = stats.MaxOpenConnections
		status.OpenConnections = stats.OpenConnections
		status.InUse = stats.InUse
		status.Idle = stats.Idle
		status.WaitCount = stats.WaitCount
		status.WaitDuration = stats.WaitDuration
		status.MaxIdleClosed = stats.MaxIdleClosed
		status.MaxIdleTimeClosed = stats.MaxIdleTimeClosed
		status.MaxLifetimeClosed = stats.MaxLifetimeClosed
	}
	return status, nil
}

// GetHealthyThreshold 获取健康阈值
func (d *dbWithHealthCheck) GetHealthyThreshold() int { return d.healthyThreshold }

// SetHealthyThreshold 设置健康阈值
func (d *dbWithHealthCheck) SetHealthyThreshold(threshold int) { d.healthyThreshold = threshold }

// StartAutoCheck 启动自动健康检查
func (d *dbWithHealthCheck) StartAutoCheck(pool *ants.Pool, interval, timeout time.Duration) error {
	d.autoCheckMu.Lock()
	defer d.autoCheckMu.Unlock()

	if d.autoCheckRunning {
		return fmt.Errorf("auto check is already running")
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	d.autoCheckStop = make(chan struct{})
	d.autoCheckPool = pool
	d.autoCheckRunning = true

	checkTask := func() {
		d.autoCheckWg.Add(1)
		defer d.autoCheckWg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-d.autoCheckStop:
				return
			case <-ticker.C:
				_, _ = d.Check(timeout)
			}
		}
	}

	if pool != nil {
		if err := pool.Submit(checkTask); err != nil {
			d.autoCheckRunning = false
			d.autoCheckStop = nil
			d.autoCheckPool = nil
			return fmt.Errorf("failed to submit task to pool: %w", err)
		}
	} else {
		go checkTask()
	}
	return nil
}

// StopAutoCheck 停止自动健康检查
func (d *dbWithHealthCheck) StopAutoCheck() {
	d.autoCheckMu.Lock()
	if !d.autoCheckRunning {
		d.autoCheckMu.Unlock()
		return
	}
	d.autoCheckRunning = false
	close(d.autoCheckStop)
	d.autoCheckMu.Unlock()

	d.autoCheckWg.Wait()

	d.autoCheckMu.Lock()
	d.autoCheckStop = nil
	d.autoCheckPool = nil
	d.autoCheckMu.Unlock()
}

// IsAutoCheckRunning 检查是否正在自动检查
func (d *dbWithHealthCheck) IsAutoCheckRunning() bool {
	d.autoCheckMu.Lock()
	defer d.autoCheckMu.Unlock()
	return d.autoCheckRunning
}

// Close 关闭数据库连接，自动停止健康检查
func (d *dbWithHealthCheck) Close() error {
	d.StopAutoCheck()
	if d.DB == nil {
		return nil
	}
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	err = sqlDB.Close()
	d.DB = nil
	return err
}

// GetDB 获取一个新的数据库会话（Fork模式）
func (d *dbWithHealthCheck) GetDB() *gorm.DB { return ForkDB(d.DB) }

// GetDBWithContext 获取一个带上下文的新数据库会话（Fork模式）
func (d *dbWithHealthCheck) GetDBWithContext(ctx context.Context) *gorm.DB {
	return ForkDBWithContext(ctx, d.DB)
}

// GetRawDB 获取原始数据库对象
func (d *dbWithHealthCheck) GetRawDB() *gorm.DB { return d.DB }
