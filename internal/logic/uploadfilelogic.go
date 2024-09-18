package logic

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"power/internal/svc"
	"power/internal/types"
	"power/model"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type UploadFileLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	httpReq *http.Request
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext, httpReq *http.Request) *UploadFileLogic {
	return &UploadFileLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		httpReq: httpReq,
	}
}

func (l *UploadFileLogic) UploadFile(req *types.UploadRequest) (*types.UploadResponse, error) {
	// 直接解析文件字段
	file, header, err := l.httpReq.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get form file 'file': %v", err)
	}
	defer file.Close()

	// 获取文件名并判断扩展名
	filename := header.Filename
	fileExt := filepath.Ext(filename)

	// 保存文件到临时路径
	tempFile, err := os.CreateTemp("", "upload-*.xlsx") // 无论上传什么文件，都将其保存为 .xlsx 文件
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// 如果是 CSV 文件，先转换为 Excel 格式再保存
	if fileExt == ".csv" {
		err = l.convertCSVToExcel(file, tempFile.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to convert CSV to Excel: %v", err)
		}
	} else if fileExt == ".xlsx" {
		// 如果是 Excel 文件，直接复制内容
		_, err = io.Copy(tempFile, file)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file content: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported file type: %s", fileExt)
	}

	// 使用背景上下文启动异步任务
	go l.processFileAsync(context.Background(), tempFile.Name(), req.Company)

	// 返回上传成功的响应
	return &types.UploadResponse{
		Message: "文件上传成功，数据正在处理",
	}, nil
}

func (l *UploadFileLogic) convertCSVToExcel(csvFile multipart.File, excelFilePath string) error {
	// 打开 CSV 文件
	reader := csv.NewReader(csvFile)

	// 读取所有行
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %v", err)
	}

	// 创建 Excel 文件
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	// 将 CSV 内容写入 Excel
	for i, row := range rows {
		for j, cell := range row {
			cellRef, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue(sheet, cellRef, cell)
		}
	}

	// 保存 Excel 文件
	err = f.SaveAs(excelFilePath)
	if err != nil {
		return fmt.Errorf("failed to save as excel file: %v", err)
	}

	return nil
}

func (l *UploadFileLogic) processFileAsync(ctx context.Context, filePath, company string) {
	//打印 company字段的值
	logx.Infof("Processing file for company: %s", company)

	defer os.Remove(filePath) // 异步任务完成后删除临时文件

	// 解析Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		logx.Errorf("Failed to open excel file: %v", err)
		return
	}

	// 判断 Excel 文件的格式
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		logx.Errorf("Failed to get rows from excel file: %v", err)
		return
	}

	if len(rows) > 0 && len(rows[0]) > 0 && rows[0][0] == "数据日期" {
		// 如果 Excel 文件第1行第1列为 "数据日期"，使用第一种逻辑
		l.processFirstFormat(ctx, rows, company)
	} else {
		// 使用第二种逻辑处理
		l.processSecondFormat(ctx, rows, company)
	}
}

func (l *UploadFileLogic) processFirstFormat(ctx context.Context, rows [][]string, company string) {
	validData := make(map[string][]model.PowerData) // 使用 map 来存储每个日期对应的有效数据

	// 加载时区
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logx.Errorf("Failed to load location: %v", err)
		return
	}

	for i, row := range rows {
		if i < 2 { // 跳过标题行
			continue
		}
		dateStr := row[0]
		if row[1] != "有功功率" { // 确保只处理有功功率行
			continue
		}

		valid := true
		tempReadings := []model.PowerData{}
		for j, cell := range row[3:] {
			if cell == "" || cell == "0" {
				valid = false
				break
			}
			power, err := strconv.ParseFloat(cell, 64)
			if err != nil {
				logx.Errorf("Invalid power value on row %d, col %d: %v", i+1, j+4, err)
				valid = false
				break
			}
			hour := j / 4
			minute := (j % 4) * 15
			timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %02d:%02d:00", dateStr, hour, minute), location)
			if err != nil {
				logx.Errorf("Error parsing date on row %d, col %d: %v", i+1, j+4, err)
				valid = false
				break
			}
			tempReadings = append(tempReadings, model.PowerData{
				DataTime: timestamp,
				Power:    power,
				Company:  company,
			})
		}

		// 如果数据有效，且该日期的数据点数量为96个，才存储
		if valid && len(tempReadings) == 96 {
			validData[dateStr] = tempReadings
		} else if len(tempReadings) < 96 {
			logx.Errorf("日期 %s 的数据不足 96 条，已删除", dateStr)
		}
	}

	// Flatten the map to a slice
	var allReadings []model.PowerData
	for _, readings := range validData {
		allReadings = append(allReadings, readings...)
	}

	// Correct sorting to order by ascending timestamp
	sort.Slice(allReadings, func(i, j int) bool {
		return allReadings[i].DataTime.Before(allReadings[j].DataTime)
	})

	// 将清洗后的数据存入MySQL数据库
	for _, data := range allReadings {
		_, err := l.svcCtx.Model.Insert(ctx, &data)
		if err != nil {
			logx.Errorf("Failed to store data into database: %v", err)
			continue
		}
		logx.Infof("Inserted data: Time=%s, Power=%f, Company=%s", data.DataTime, data.Power, data.Company)
	}

	logx.Info("文件上传并处理成功")
}

func (l *UploadFileLogic) processSecondFormat(ctx context.Context, rows [][]string, company string) {
	// 查找日期和功率的列
	dateCol, powerCol := -1, -1
	for index, cell := range rows[0] {
		switch cell {
		case "日期", "数据时间":
			dateCol = index
		case "瞬时有功", "功率有功", "E", "总", "总有功功率":
			powerCol = index
		}
	}

	if dateCol == -1 || powerCol == -1 {
		logx.Errorf("Excel文件中未找到所需的列")
		return
	}

	// 加载时区
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logx.Errorf("Failed to load location: %v", err)
		return
	}

	var cleanedData []model.PowerData

	// 假设文件中的年份是 2024
	year := "2024"

	// 定义符合完整格式的正则表达式，用于验证日期是否已包含年份
	fullDatePattern := `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`
	fullDateRegex := regexp.MustCompile(fullDatePattern)

	// 临时保存每个日期的数据
	dailyData := make(map[string][]model.PowerData)

	// 清洗数据
	for _, row := range rows[1:] {
		if len(row) <= powerCol || len(row) <= dateCol {
			continue
		}

		dateStr := strings.TrimSpace(row[dateCol])
		powerStr := strings.TrimSpace(row[powerCol])

		if powerStr == "" || powerStr == "0" {
			continue
		}

		power, err := strconv.ParseFloat(powerStr, 64)
		if err != nil || power == 0 {
			continue
		}

		var dateTime time.Time

		// 判断日期是否符合 `yyyy-MM-dd HH:mm:ss` 格式
		if fullDateRegex.MatchString(dateStr) {
			// 如果日期已经是完整格式，直接解析
			dateTime, err = time.ParseInLocation("2006-01-02 15:04:05", dateStr, location)
			if err != nil {
				logx.Errorf("Failed to parse full date: %v", err)
				continue
			}
		} else {
			// 否则假设是 `MM-DD HH:mm` 格式，补充年份并解析
			fullDateStr := fmt.Sprintf("%s-%s:00", year, dateStr) // 添加年份并补全秒数
			dateTime, err = time.ParseInLocation("2006-01-02 15:04:05", fullDateStr, location)
			if err != nil {
				logx.Errorf("Failed to parse partial date: %v", err)
				continue
			}
		}

		// 将数据按日期（不包括时间部分）分组
		dateKey := dateTime.Format("2006-01-02")
		dailyData[dateKey] = append(dailyData[dateKey], model.PowerData{
			DataTime: dateTime,
			Power:    power,
			Company:  company,
		})
	}

	// 检查每一天是否有96个数据点，如果不足则删除该天的数据3
	for date, data := range dailyData {
		if len(data) == 96 {
			cleanedData = append(cleanedData, data...)
		} else {
			logx.Errorf("日期 %s 的数据不足 96 条，已删除", date)
		}
	}

	// Correct sorting to order by ascending timestamp
	sort.Slice(cleanedData, func(i, j int) bool {
		return cleanedData[i].DataTime.Before(cleanedData[j].DataTime)
	})

	// 将清洗后的数据存入MySQL数据库
	for _, data := range cleanedData {
		_, err := l.svcCtx.Model.Insert(ctx, &data)
		if err != nil {
			logx.Errorf("Failed to store data into database: %v", err)
			continue
		}
		logx.Infof("Inserted data: Time=%s, Power=%f, Company=%s", data.DataTime, data.Power, data.Company)
	}

	logx.Info("文件上传并处理成功")
}
