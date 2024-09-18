package svc

import (
	"power/internal/config"
	"power/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	Model  model.PowerDataModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource) // 修改为使用 c.Mysql.DataSource
	return &ServiceContext{
		Config: c,
		Model:  model.NewPowerDataModel(conn),
	}
}
