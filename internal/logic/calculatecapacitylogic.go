package logic

import (
	"context"
	"fmt"
	"math"
	"sort"
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
	firstChargePower, err := l.getPower(queryLogic, req.FirstChargePeriod[0], req.FirstChargePeriod[1], req.Company, req.CalculationMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to query power data for first charge period: %v", err)
	}
	firstChargeHours := l.getHours(req.FirstChargePeriod[0], req.FirstChargePeriod[1], firstChargePower)

	// 第一次放电时段
	firstDischargePower, err := l.getPower(queryLogic, req.FirstDischargePeriod[0], req.FirstDischargePeriod[1], req.Company, req.CalculationMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to query power data for first discharge period: %v", err)
	}
	firstDischargeHours := l.getHours(req.FirstDischargePeriod[0], req.FirstDischargePeriod[1], firstDischargePower)

	// 第二次充电时段
	secondChargePower, err := l.getPower(queryLogic, req.SecondChargePeriod[0], req.SecondChargePeriod[1], req.Company, req.CalculationMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to query power data for second charge period: %v", err)
	}
	secondChargeHours := l.getHours(req.SecondChargePeriod[0], req.SecondChargePeriod[1], secondChargePower)

	// 第二次放电时段
	secondDischargePower, err := l.getPower(queryLogic, req.SecondDischargePeriod[0], req.SecondDischargePeriod[1], req.Company, req.CalculationMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to query power data for second discharge period: %v", err)
	}
	secondDischargeHours := l.getHours(req.SecondDischargePeriod[0], req.SecondDischargePeriod[1], secondDischargePower)

	// 如果任何一个时段的时长为0，说明数据无效，返回0的结果
	if firstChargeHours == 0 || firstDischargeHours == 0 || secondChargeHours == 0 || secondDischargeHours == 0 {
		l.Logger.Infof("One or more periods have no valid data, returning 0 for all values")
		return &types.CapacityConfigResponse{
			MinCabinetCount:       0,
			FirstChargeAmount:     0,
			FirstDischargeAmount:  0,
			SecondChargeAmount:    0,
			SecondDischargeAmount: 0,
		}, nil
	}

	// 使用查询到的功率数据进行容量计算
	meterMultiplier := req.MeterMultiplier
	powerFactor := req.PowerFactor
	transformerCapacity := req.TransformerCapacity

	// 计算每个时段的充电量和放电量
	firstChargeAmount := math.Round(transformerCapacity*powerFactor-firstChargePower*meterMultiplier) * firstChargeHours
	firstDischargeAmount := math.Round(firstDischargePower * meterMultiplier * firstDischargeHours * powerFactor)

	secondChargeAmount := math.Round(transformerCapacity*powerFactor-secondChargePower*meterMultiplier) * secondChargeHours
	secondDischargeAmount := math.Round(secondDischargePower * meterMultiplier * secondDischargeHours * powerFactor)

	// 打印每个时段的充放电量
	l.Logger.Infof("First Charge Amount: %f kWh", firstChargeAmount)
	l.Logger.Infof("First Discharge Amount: %f kWh", firstDischargeAmount)
	l.Logger.Infof("Second Charge Amount: %f kWh", secondChargeAmount)
	l.Logger.Infof("Second Discharge Amount: %f kWh", secondDischargeAmount)

	// 计算储能柜台数
	firstChargeCabinets := firstChargeAmount / req.ChargeCapacity
	firstDischargeCabinets := firstDischargeAmount / req.DischargeCapacity
	secondChargeCabinets := secondChargeAmount / req.ChargeCapacity
	secondDischargeCabinets := secondDischargeAmount / req.DischargeCapacity

	// 打印每个时段的储能柜数量
	l.Logger.Infof("First Charge Cabinets: %f", firstChargeCabinets)
	l.Logger.Infof("First Discharge Cabinets: %f", firstDischargeCabinets)
	l.Logger.Infof("Second Charge Cabinets: %f", secondChargeCabinets)
	l.Logger.Infof("Second Discharge Cabinets: %f", secondDischargeCabinets)

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

// 获取功率的辅助方法，根据请求的计算方法计算功率
func (l *CalculateCapacityLogic) getPower(queryLogic *QueryDataLogic, startTime, endTime, company, method string) (float64, error) {
	queryReq := types.QueryRequest{
		StartTime: startTime,
		EndTime:   endTime,
		Company:   company,
	}

	queryResp, err := queryLogic.QueryData(&queryReq)
	if err != nil {
		return 0, err
	}

	if len(queryResp.Data) == 0 {
		l.Logger.Infof("No data points available for power calculation")
		return 0, nil
	}

	// 根据请求的方法选择计算功率的方式
	switch method {
	case "average":
		return l.calculateAveragePower(queryResp), nil
	case "median":
		return l.calculateMedian(queryResp), nil
	case "mode":
		return l.calculateMode(queryResp), nil
	case "percentile90":
		return l.calculatePercentile(queryResp, 90), nil
	case "quartile":
		return l.calculateQuartile(queryResp), nil
	case "stddev_mean":
		return l.calculateStdDevMean(queryResp), nil
	default:
		return 0, fmt.Errorf("unsupported calculation method: %s", method)
		//return l.calculateAveragePower(queryResp), nil
	}
}

// 计算平均数
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
	l.Logger.Infof("Calculated average power: %.2f", avgPower) // 打印平均功率
	return avgPower
}

// 计算中位数
func (l *CalculateCapacityLogic) calculateMedian(queryResp *types.QueryResponse) float64 {
	var powers []float64
	for _, data := range queryResp.Data {
		powers = append(powers, data.Power)
	}
	sort.Float64s(powers)
	length := len(powers)
	if length == 0 {
		l.Logger.Infof("No data points available for median calculation")
		return 0
	}
	var median float64
	if length%2 == 0 {
		median = (powers[length/2-1] + powers[length/2]) / 2
	} else {
		median = powers[length/2]
	}
	l.Logger.Infof("Calculated median power: %f", median) // 打印中位数
	return median
}

// 计算众数
func (l *CalculateCapacityLogic) calculateMode(queryResp *types.QueryResponse) float64 {
	frequencyMap := make(map[float64]int)
	var mode float64
	maxFrequency := 0
	for _, data := range queryResp.Data {
		frequencyMap[data.Power]++
		if frequencyMap[data.Power] > maxFrequency {
			mode = data.Power
			maxFrequency = frequencyMap[data.Power]
		}
	}
	if maxFrequency == 0 {
		l.Logger.Infof("No data points available for mode calculation")
		return 0
	}
	l.Logger.Infof("Calculated mode power: %f", mode) // 打印众数
	return mode
}

// 计算百分位数
func (l *CalculateCapacityLogic) calculatePercentile(queryResp *types.QueryResponse, percentile float64) float64 {
	var powers []float64
	for _, data := range queryResp.Data {
		powers = append(powers, data.Power)
	}
	sort.Float64s(powers)
	length := len(powers)
	if length == 0 {
		l.Logger.Infof("No data points available for percentile calculation")
		return 0
	}
	index := percentile / 100 * float64(length-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	var percentileValue float64
	if lower == upper {
		percentileValue = powers[lower]
	} else {
		percentileValue = powers[lower] + (powers[upper]-powers[lower])*(index-float64(lower))
	}
	l.Logger.Infof("Calculated %f percentile power: %f", percentile, percentileValue) // 打印百分位数
	return percentileValue
}

// 计算四分位数（Q1、Q2、Q3的平均值）
func (l *CalculateCapacityLogic) calculateQuartile(queryResp *types.QueryResponse) float64 {
	q1 := l.calculatePercentile(queryResp, 25)
	q2 := l.calculatePercentile(queryResp, 50) // 中位数
	q3 := l.calculatePercentile(queryResp, 75)
	quartileAverage := (q1 + q2 + q3) / 3
	l.Logger.Infof("Calculated Quartile Average (Q1, Q2, Q3): %f", quartileAverage) // 打印四分位数平均值
	return quartileAverage
}

// 计算标准差和均值的组合
func (l *CalculateCapacityLogic) calculateStdDevMean(queryResp *types.QueryResponse) float64 {
	var powers []float64
	var sum, mean, variance float64
	for _, data := range queryResp.Data {
		powers = append(powers, data.Power)
		sum += data.Power
	}
	length := len(powers)
	if length == 0 {
		l.Logger.Infof("No data points available for standard deviation calculation")
		return 0
	}
	mean = sum / float64(length)
	for _, power := range powers {
		variance += math.Pow(power-mean, 2)
	}
	stddev := math.Sqrt(variance / float64(length))
	// 返回均值加标准差
	stddevMean := mean + stddev
	l.Logger.Infof("Calculated Mean + Standard Deviation: %f", stddevMean) // 打印均值+标准差
	return stddevMean
}

// 计算两个时间点之间的时长（小时）
// 计算两个时间点之间的时长（小时），如果时间无效或数据不存在返回 0
func (l *CalculateCapacityLogic) getHours(startTimeStr, endTimeStr string, power float64) float64 {
	// 如果功率为 0，表示该时间段没有数据，返回 0
	if power == 0 {
		l.Logger.Infof("Power data is zero, indicating no data exists for the time period %s - %s", startTimeStr, endTimeStr)
		return 0
	}

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
	// 如果结束时间早于开始时间，返回 0
	if endTime.Before(startTime) {
		l.Logger.Errorf("End time is before start time: %s < %s", endTimeStr, startTimeStr)
		return 0
	}
	// 计算数据点数量
	duration := endTime.Sub(startTime)
	dataPoints := int(duration.Minutes()/15) + 1 // 包括起始点
	// 每 4 个数据点算 1 小时
	hours := float64(dataPoints) / 4
	l.Logger.Infof("Calculated hours between %s and %s: %f hours", startTimeStr, endTimeStr, hours) // 打印时长
	return hours
}
