// Code generated by goctl. DO NOT EDIT.
package types

type CapacityConfigRequest struct {
	Company               string   `json:"company"`
	PowerFactor           float64  `json:"powerFactor"`           // 功率因数
	TransformerCapacity   float64  `json:"transformerCapacity"`   // 变压器容量 (kW)
	MeterMultiplier       float64  `json:"meterMultiplier"`       // 电表倍率
	DischargeCapacity     float64  `json:"dischargeCapacity"`     // 储能柜实际放电容量 (kWh)
	ChargeCapacity        float64  `json:"chargeCapacity"`        // 储能柜实际充电容量 (kWh)
	FirstChargePeriod     []string `json:"firstChargePeriod"`     // 第一次充电时段，例如 [9, 11] 表示9-11点
	FirstDischargePeriod  []string `json:"firstDischargePeriod"`  // 第一次放电时段
	SecondChargePeriod    []string `json:"secondChargePeriod"`    // 第二次充电时段
	SecondDischargePeriod []string `json:"secondDischargePeriod"` // 第二次放电时段
	CalculationMethod     string   `json:"calculationMethod"`     //计算方法：平均数、中位数、众数、百分位数等
}

type CapacityConfigResponse struct {
	MinCabinetCount       int     // 储能柜最小台数
	FirstChargeAmount     float64 // 第一次充电量 (kWh)
	FirstDischargeAmount  float64 // 第一次放电量 (kWh)
	SecondChargeAmount    float64 // 第二次充电量 (kWh)
	SecondDischargeAmount float64 // 第二次放电量 (kWh)
}

type PowerData struct {
	Time  string  // 数据时间
	Power float64 // 功率
}

type QueryRequest struct {
	StartTime string `form:"startTime"` // 查询开始时间，格式：YYYY-MM-DD HH:MM:SS
	EndTime   string `form:"endTime"`   // 查询结束时间，格式：YYYY-MM-DD HH:MM:SS
	Company   string `form:"company"`   // 公司名称
}

type QueryResponse struct {
	Data []PowerData
}

type UploadRequest struct {
	File    string `form:"file"`    // 文件内容作为Base64字符串上传
	Company string `form:"company"` // 公司名称
}

type UploadResponse struct {
	Message string
}
