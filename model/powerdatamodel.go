package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// PowerDataModel 接口，包含插入和查询方法
type PowerDataModel interface {
	Insert(ctx context.Context, data *PowerData) (sql.Result, error)
	QueryData(ctx context.Context, startTime, endTime time.Time, company string) ([]PowerData, error)
}

// NewPowerDataModel 创建一个新的 PowerDataModel 实例
func NewPowerDataModel(conn sqlx.SqlConn) PowerDataModel {
	return &defaultPowerDataModel{
		conn:  conn,
		table: "power_data",
	}
}

// QueryData 方法用于根据时间范围和公司名称查询数据
func (m *defaultPowerDataModel) QueryData(ctx context.Context, startTime, endTime time.Time, company string) ([]PowerData, error) {
	query := `SELECT id, data_time, power, company FROM ` + m.table + ` WHERE data_time BETWEEN ? AND ? AND company = ? ORDER BY data_time ASC`
	var data []PowerData
	err := m.conn.QueryRowsCtx(ctx, &data, query, startTime, endTime, company)
	if err != nil {
		return nil, err
	}
	return data, nil
}
