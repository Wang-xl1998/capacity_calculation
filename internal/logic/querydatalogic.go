package logic

import (
	"context"
	"time"

	"power/internal/svc"
	"power/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDataLogic {
	return &QueryDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDataLogic) QueryData(req *types.QueryRequest) (*types.QueryResponse, error) {
	// 记录收到的请求信息
	l.Logger.Infof("Received QueryData request: startTime=%s, endTime=%s, company=%s", req.StartTime, req.EndTime, req.Company)

	// 加载上海时区，确保时区与上传时保持一致
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		l.Logger.Error("Failed to load location: ", err)
		return nil, err
	}

	// 使用上海时区解析时间
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartTime, location)
	if err != nil {
		l.Logger.Error("Failed to parse start time: ", err)
		return nil, err
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", req.EndTime, location)
	if err != nil {
		l.Logger.Error("Failed to parse end time: ", err)
		return nil, err
	}

	// 记录解析后的时间信息
	l.Logger.Infof("Parsed times: startTime=%v, endTime=%v", startTime, endTime)

	// 从数据库中查询数据
	data, err := l.svcCtx.Model.QueryData(l.ctx, startTime, endTime, req.Company)
	if err != nil {
		l.Logger.Error("Database query failed: ", err)
		return nil, err
	}

	// 记录查询到的数据条数
	l.Logger.Infof("Retrieved %d data points from database", len(data))

	// 构造返回数据
	var result []types.PowerData
	for _, d := range data {
		result = append(result, types.PowerData{
			Time:  d.DataTime.Format("2006-01-02 15:04:05"),
			Power: d.Power,
		})
	}

	// 记录返回的数据条数
	l.Logger.Infof("Returning %d data points in response", len(result))

	// 返回查询的 company 和数据
	return &types.QueryResponse{
		Data: result,
	}, nil
}
