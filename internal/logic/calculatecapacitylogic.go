package logic

import (
	"context"
	"fmt"
	"math"
	"time"

	"power/internal/svc"
	"power/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CalculateCapacityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCalculateCapacityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CalculateCapacityLogic {
	return &CalculateCapacityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CalculateCapacityLogic) CalculateCapacity(req *types.CapacityConfigRequest) (resp *types.CapacityConfigResponse, err error) {
	l.Logger.Info("Starting capacity calculation for company: ", req.Company)

	// 调用查询逻辑来获取功率数据
	queryLogic := NewQueryDataLogic(l.ctx, l.svcCtx)

	// 第一次充电时段
	firstChargeAvgPower, err := l.getAveragePower(queryLogic, req.FirstChargePeriod[0], req.FirstChargePeriod[1], req.Company)
	if err != nil {
		return nil, fmt.Errorf("Failed to query power data for first charge period: %v", err)
	}

	// 第一次放电时段
	firstDischargeAvgPower, err := l.getAveragePower(queryLogic, req.FirstDischargePeriod[0], req.FirstDischargePeriod[1], req.Company)
	if err != nil {
		return nil, fmt.Errorf("Failed to query power data for first discharge period: %v", err)
	}

	// 第二次充电时段
	secondChargeAvgPower, err := l.getAveragePower(queryLogic, req.SecondChargePeriod[0], req.SecondChargePeriod[1], req.Company)
	if err != nil {
		return nil, fmt.Errorf("Failed to query power data for second charge period: %v", err)
	}

	// 第二次放电时段
	secondDischargeAvgPower, err := l.getAveragePower(queryLogic, req.SecondDischargePeriod[0], req.SecondDischargePeriod[1], req.Company)
	if err != nil {
		return nil, fmt.Errorf("Failed to query power data for second discharge period: %v", err)
	}

	// 使用查询到的功率数据进行容量计算
	meterMultiplier := req.MeterMultiplier
	powerFactor := req.PowerFactor
	transformerCapacity := req.TransformerCapacity

	// 计算每个时段的充电量和放电量
	firstChargeAmount := (transformerCapacity*powerFactor - firstChargeAvgPower*meterMultiplier) * l.getHours(req.FirstChargePeriod[0], req.FirstChargePeriod[1])
	firstDischargeAmount := firstDischargeAvgPower * meterMultiplier * l.getHours(req.FirstDischargePeriod[0], req.FirstDischargePeriod[1]) * powerFactor

	secondChargeAmount := (transformerCapacity*powerFactor - secondChargeAvgPower*meterMultiplier) * l.getHours(req.SecondChargePeriod[0], req.SecondChargePeriod[1])
	secondDischargeAmount := secondDischargeAvgPower * meterMultiplier * l.getHours(req.SecondDischargePeriod[0], req.SecondDischargePeriod[1]) * powerFactor

	// 计算储能柜台数
	firstChargeCabinets := firstChargeAmount / req.ChargeCapacity
	firstDischargeCabinets := firstDischargeAmount / req.DischargeCapacity
	secondChargeCabinets := secondChargeAmount / req.ChargeCapacity
	secondDischargeCabinets := secondDischargeAmount / req.DischargeCapacity

	// 取最小储能柜台数
	minCabinetCount := int(math.Min(math.Min(firstChargeCabinets, firstDischargeCabinets), math.Min(secondChargeCabinets, secondDischargeCabinets)))

	// 返回结果
	resp = &types.CapacityConfigResponse{
		MinCabinetCount:       minCabinetCount,
		FirstChargeAmount:     firstChargeAmount,
		FirstDischargeAmount:  firstDischargeAmount,
		SecondChargeAmount:    secondChargeAmount,
		SecondDischargeAmount: secondDischargeAmount,
	}
	return resp, nil
}

// 获取平均功率的辅助方法
func (l *CalculateCapacityLogic) getAveragePower(queryLogic *QueryDataLogic, startTime, endTime, company string) (float64, error) {
	queryReq := types.QueryRequest{
		StartTime: startTime,
		EndTime:   endTime,
		Company:   company,
	}

	queryResp, err := queryLogic.QueryData(&queryReq)
	if err != nil {
		return 0, err
	}

	// 从查询结果中计算平均功率
	return l.calculateAveragePower(queryResp), nil
}

// 计算平均功率
func (l *CalculateCapacityLogic) calculateAveragePower(queryResp *types.QueryResponse) float64 {
	var totalPower float64
	for _, data := range queryResp.Data {
		totalPower += data.Power
	}
	if len(queryResp.Data) == 0 {
		l.Logger.Infof("No data points available for power calculation")
		return 0
	}
	avgPower := totalPower / float64(len(queryResp.Data))
	l.Logger.Infof("Calculated average power: %f", avgPower)
	return avgPower
}

// 计算两个时间点之间的时长（小时）
func (l *CalculateCapacityLogic) getHours(startTimeStr, endTimeStr string) float64 {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
	if err != nil {
		l.Logger.Errorf("Invalid start time: %s", startTimeStr)
		return 0
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, loc)
	if err != nil {
		l.Logger.Errorf("Invalid end time: %s", endTimeStr)
		return 0
	}
	// 计算数据点数量
	duration := endTime.Sub(startTime)
	dataPoints := int(duration.Minutes()/15) + 1 // 包括起始点
	// 每 4 个数据点算 1 小时
	return float64(dataPoints) / 4
}
