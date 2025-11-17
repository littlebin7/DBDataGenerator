package poolmonitor

import (
	"fmt"

	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
)

// PoolStatus 连接池状态
type PoolStatus struct {
	ConnectionID      string `json:"connection_id"`
	Type              string `json:"type"`               // 数据库类型
	MaxConnections    int    `json:"max_connections"`    // 最大连接数
	ActiveConnections int    `json:"active_connections"` // 活动连接数
	IdleConnections   int    `json:"idle_connections"`   // 空闲连接数
	WaitingRequests   int    `json:"waiting_requests"`   // 等待请求数
}

// PoolMonitor 连接池监控器
type PoolMonitor struct {
	connMgr database.ConnectionManagerInterface
	logger  *zap.Logger
}

// NewPoolMonitor 创建连接池监控器
func NewPoolMonitor(connMgr database.ConnectionManagerInterface, logger *zap.Logger) *PoolMonitor {
	return &PoolMonitor{
		connMgr: connMgr,
		logger:  logger,
	}
}

// GetPoolStatus 获取连接池状态
func (pm *PoolMonitor) GetPoolStatus(connectionID string) (*PoolStatus, error) {
	conn, err := pm.connMgr.GetConnection(connectionID)
	if err != nil {
		return nil, fmt.Errorf("连接不存在: %w", err)
	}

	status := &PoolStatus{
		ConnectionID: connectionID,
		Type:         conn.Config.Type,
	}

	// 尝试获取连接池信息
	// 注意：不同数据库的连接池实现不同，这里简化处理
	if db, ok := conn.Database.(interface {
		GetPoolStats() (active, idle, max int)
	}); ok {
		status.ActiveConnections, status.IdleConnections, status.MaxConnections = db.GetPoolStats()
	} else {
		// 默认值
		status.MaxConnections = 10
		status.ActiveConnections = 0
		status.IdleConnections = 0
	}

	return status, nil
}

// GetAllPoolStatus 获取所有连接池状态
func (pm *PoolMonitor) GetAllPoolStatus() ([]*PoolStatus, error) {
	connections := pm.connMgr.GetAllConnections()
	statuses := make([]*PoolStatus, 0, len(connections))

	for _, conn := range connections {
		status, err := pm.GetPoolStatus(conn.ID)
		if err != nil {
			continue
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// GetPoolRecommendations 获取连接池配置建议
func (pm *PoolMonitor) GetPoolRecommendations(connectionID string) []string {
	var recommendations []string

	status, err := pm.GetPoolStatus(connectionID)
	if err != nil {
		return recommendations
	}

	// 检查活动连接数
	if status.MaxConnections > 0 {
		activeRatio := float64(status.ActiveConnections) / float64(status.MaxConnections)

		if status.ActiveConnections >= status.MaxConnections {
			recommendations = append(recommendations,
				"⚠️ 活动连接数已达到最大值，建议增加最大连接数以避免连接等待")
		} else if activeRatio >= 0.8 {
			recommendations = append(recommendations,
				"⚠️ 活动连接数接近最大值（80%+），建议考虑增加最大连接数")
		}
	}

	// 检查空闲连接数
	if status.MaxConnections > 0 {
		idleRatio := float64(status.IdleConnections) / float64(status.MaxConnections)

		if idleRatio > 0.7 {
			recommendations = append(recommendations,
				"💡 空闲连接数较多（70%+），可以考虑减少最大连接数以节省资源")
		} else if idleRatio < 0.1 && status.ActiveConnections > 0 {
			recommendations = append(recommendations,
				"💡 空闲连接数较少，当前配置较为合理")
		}
	}

	// 检查等待请求数
	if status.WaitingRequests > 0 {
		recommendations = append(recommendations,
			"⚠️ 有等待的请求，建议增加最大连接数或优化查询性能")
	}

	// 连接池利用率建议
	if status.MaxConnections > 0 {
		totalConnections := status.ActiveConnections + status.IdleConnections
		utilization := float64(totalConnections) / float64(status.MaxConnections) * 100

		if utilization < 30 {
			recommendations = append(recommendations,
				"💡 连接池利用率较低（<30%），可以考虑减少最大连接数")
		} else if utilization > 90 {
			recommendations = append(recommendations,
				"⚠️ 连接池利用率很高（>90%），建议增加最大连接数")
		}
	}

	// 如果没有问题，给出正面反馈
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"✅ 连接池配置合理，当前状态良好")
	}

	return recommendations
}
