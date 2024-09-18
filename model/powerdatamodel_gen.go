// Code generated by goctl. DO NOT EDIT.

package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	powerDataFieldNames          = builder.RawFieldNames(&PowerData{})
	powerDataRows                = strings.Join(powerDataFieldNames, ",")
	powerDataRowsExpectAutoSet   = strings.Join(stringx.Remove(powerDataFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	powerDataRowsWithPlaceHolder = strings.Join(stringx.Remove(powerDataFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"
)

type (
	powerDataModel interface {
		Insert(ctx context.Context, data *PowerData) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*PowerData, error)
		Update(ctx context.Context, data *PowerData) error
		Delete(ctx context.Context, id int64) error
	}

	defaultPowerDataModel struct {
		conn  sqlx.SqlConn
		table string
	}

	PowerData struct {
		Id       int64     `db:"id"`
		DataTime time.Time `db:"data_time"`
		Power    float64   `db:"power"`
		Company  string    `db:"company"`
	}
)

func newPowerDataModel(conn sqlx.SqlConn) *defaultPowerDataModel {
	return &defaultPowerDataModel{
		conn:  conn,
		table: "`power_data`",
	}
}

func (m *defaultPowerDataModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *defaultPowerDataModel) FindOne(ctx context.Context, id int64) (*PowerData, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", powerDataRows, m.table)
	var resp PowerData
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultPowerDataModel) Insert(ctx context.Context, data *PowerData) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?)", m.table, powerDataRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.DataTime, data.Power, data.Company)
	return ret, err
}

func (m *defaultPowerDataModel) Update(ctx context.Context, data *PowerData) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, powerDataRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.DataTime, data.Power, data.Company, data.Id)
	return err
}

func (m *defaultPowerDataModel) tableName() string {
	return m.table
}