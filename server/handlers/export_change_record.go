package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "image/png"

	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// lookupRow 查找表的一行数据
type lookupRow struct {
	Position string
	System   string
	Roles    string // "□角色1 □角色2 □角色3"
}

// ExportChangeRecord 导出用户变更记录表Excel
func ExportChangeRecord(c *gin.Context) {
	// 查询所有岗位权限规则
	var rules []models.PermissionRule
	if err := database.GetDB().Order("sort_order asc").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询岗位权限数据失败"})
		return
	}

	if len(rules) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "暂无岗位权限数据"})
		return
	}

	// 解析每个岗位的权限规则，构建查找表数据
	type roleItem struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	type systemRule struct {
		System string     `json:"system"`
		Roles  []roleItem `json:"roles"`
	}

	var lookupRows []lookupRow
	// 收集所有系统名称（用于Sheet2全量矩阵）
	allSystems := []string{}
	systemSet := map[string]bool{}

	for _, rule := range rules {
		var sysRules []systemRule
		if err := json.Unmarshal([]byte(rule.RulesJSON), &sysRules); err != nil {
			continue
		}
		for _, sr := range sysRules {
			// 收集系统名称
			if !systemSet[sr.System] {
				systemSet[sr.System] = true
				allSystems = append(allSystems, sr.System)
			}
			// 过滤 enabled 角色
			var enabledRoles []string
			for _, r := range sr.Roles {
				if r.Enabled {
					enabledRoles = append(enabledRoles, r.Name)
				}
			}
			if len(enabledRoles) > 0 {
				rolesStr := ""
				for i, r := range enabledRoles {
					if i > 0 {
						rolesStr += " "
					}
					rolesStr += "□" + r
				}
				lookupRows = append(lookupRows, lookupRow{
					Position: rule.PositionName,
					System:   sr.System,
					Roles:    rolesStr,
				})
			}
		}
	}

	// 创建Excel
	f := excelize.NewFile()

	// ==================== 样式定义 ====================
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	infoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "bottom",
			WrapText: true,
		},
	})

	noteStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
			Color:  "#FF0000",
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#D9E1F2"},
		},
	})

	dataCellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	centerDataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	footerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "bottom",
		},
	})

	// ==================== Sheet1: 用户变更记录表 ====================
	sheet1 := "用户变更记录表"
	f.SetSheetName("Sheet1", sheet1)

	// Row 1: 标题
	f.SetCellValue(sheet1, "A1", "用户变更记录表")
	f.MergeCell(sheet1, "A1", "C1")
	f.SetCellStyle(sheet1, "A1", "C1", titleStyle)
	f.SetRowHeight(sheet1, 1, 36)

	// Row 2: 申请信息（带下划线）
	f.SetCellValue(sheet1, "A2", "申请部门：_________  申请人：_________  申请日期：_________")
	f.MergeCell(sheet1, "A2", "C2")
	f.SetCellStyle(sheet1, "A2", "C2", infoStyle)
	f.SetRowHeight(sheet1, 2, 28)

	// Row 3: 岗位(下拉)
	f.SetCellValue(sheet1, "A3", "岗位：")
	f.SetCellStyle(sheet1, "A3", "A3", infoStyle)
	// B3: 岗位下拉选择（设置占位提示文字，选择后自动替换）
	f.SetCellValue(sheet1, "B3", "▼ 请下拉选择岗位")
	f.SetCellStyle(sheet1, "B3", "B3", infoStyle)
	f.SetRowHeight(sheet1, 3, 30)

	// Row 4: 审批人 + 审批日期（带下划线）
	f.SetCellValue(sheet1, "A4", "部门审批人：_________              审批日期：_________")
	f.MergeCell(sheet1, "A4", "C4")
	f.SetCellStyle(sheet1, "A4", "C4", infoStyle)
	f.SetRowHeight(sheet1, 4, 28)

	// Row 5: 注释
	f.SetCellValue(sheet1, "A5", "注：不需要的权限请划竖线丨，删除的权限请打×，需要的权限请打√")
	f.SetCellStyle(sheet1, "A5", "A5", noteStyle)
	f.SetRowHeight(sheet1, 5, 22)

	// Row 6: 表头
	headers := []string{"系统", "角色", "账号名"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheet1, col+"6", h)
		f.SetCellStyle(sheet1, col+"6", col+"6", tableHeaderStyle)
	}
	f.SetRowHeight(sheet1, 6, 26)

	// Row 7~23: 数据行（公式），预留17行
	// 查找表起始行（Sheet2 的 Row 29 + 1 表头 = 数据从 Row 30 开始）
	lookupStartRow := 30
	lookupEndRow := lookupStartRow + len(lookupRows) - 1
	if lookupEndRow < lookupStartRow {
		lookupEndRow = lookupStartRow
	}

	// Sheet2 名称（后面创建）
	sheet2Name := "权限规则参考"

	for row := 7; row <= 23; row++ {
		n := row - 6 // 第几个匹配项 (1, 2, 3, ...)

		// A列公式: 系统名 - 排除占位提示文字
		formulaA := fmt.Sprintf(
			`IF(OR($B$3="",$B$3="▼ 请下拉选择岗位"),"",IFERROR(INDEX('%s'!B:B,SMALL(IF('%s'!$A$%d:$A$%d=$B$3,ROW('%s'!$A$%d:$A$%d)),%d)),""))`,
			sheet2Name, sheet2Name, lookupStartRow, lookupEndRow, sheet2Name, lookupStartRow, lookupEndRow, n,
		)
		cellA := fmt.Sprintf("A%d", row)
		formulaType := excelize.STCellFormulaTypeArray
		cellRef := fmt.Sprintf("%s:%s", cellA, cellA)
		f.SetCellFormula(sheet1, cellA, formulaA, excelize.FormulaOpts{Ref: &cellRef, Type: &formulaType})

		// B列公式: 角色 - 如果A列为空则空白
		formulaB := fmt.Sprintf(
			`IF(A%d="","",IFERROR(INDEX('%s'!C:C,SMALL(IF('%s'!$A$%d:$A$%d=$B$3,ROW('%s'!$A$%d:$A$%d)),%d)),""))`,
			row, sheet2Name, sheet2Name, lookupStartRow, lookupEndRow, sheet2Name, lookupStartRow, lookupEndRow, n,
		)
		cellB := fmt.Sprintf("B%d", row)
		cellRefB := fmt.Sprintf("%s:%s", cellB, cellB)
		f.SetCellFormula(sheet1, cellB, formulaB, excelize.FormulaOpts{Ref: &cellRefB, Type: &formulaType})

		// 设置样式
		f.SetCellStyle(sheet1, cellA, cellA, dataCellStyle)
		f.SetCellStyle(sheet1, cellB, cellB, dataCellStyle)
		f.SetCellStyle(sheet1, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), centerDataStyle)
		f.SetRowHeight(sheet1, row, 28)
	}

	// Row 24: 备注行（带全边框）
	noteRow := 24
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", noteRow), "备注：")
	f.MergeCell(sheet1, fmt.Sprintf("A%d", noteRow), fmt.Sprintf("C%d", noteRow))
	// 创建带全边框的备注样式
	noteBorderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet1, fmt.Sprintf("A%d", noteRow), fmt.Sprintf("C%d", noteRow), noteBorderStyle)
	f.SetRowHeight(sheet1, noteRow, 28)

	// Row 25: 签字栏第一行（开通人、复核人）- 合并显示，带下划线
	signRow1 := 25
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", signRow1),
		"开通人：_________  开通日期：_________  复核人：_________  复核日期：_________")
	f.MergeCell(sheet1, fmt.Sprintf("A%d", signRow1), fmt.Sprintf("C%d", signRow1))
	f.SetCellStyle(sheet1, fmt.Sprintf("A%d", signRow1), fmt.Sprintf("C%d", signRow1), footerStyle)
	f.SetRowHeight(sheet1, signRow1, 28)

	// Row 26: 签字栏第二行（申请人确认）- 合并显示，带下划线
	signRow2 := 26
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", signRow2),
		"申请人确认：_________              确认日期：_________")
	f.MergeCell(sheet1, fmt.Sprintf("A%d", signRow2), fmt.Sprintf("C%d", signRow2))
	f.SetCellStyle(sheet1, fmt.Sprintf("A%d", signRow2), fmt.Sprintf("C%d", signRow2), footerStyle)
	f.SetRowHeight(sheet1, signRow2, 28)

	// 设置列宽（仅A-C三列，适配A4打印）
	f.SetColWidth(sheet1, "A", "A", 18)  // 系统
	f.SetColWidth(sheet1, "B", "B", 35)  // 角色
	f.SetColWidth(sheet1, "C", "C", 20)  // 账号名

	// 设置Sheet1页面: A4纵向
	f.SetPageLayout(sheet1, &excelize.PageLayoutOptions{
		Size:            intPtr(9), // A4
		Orientation:     stringPtr("portrait"),
		AdjustTo:        uintPtr(100),
		FitToHeight:     intPtr(1),
		FitToWidth:      intPtr(1),
		BlackAndWhite:   boolPtr(false),
	})
	f.SetPageMargins(sheet1, &excelize.PageLayoutMarginsOptions{
		Bottom: float64Ptr(0.5),
		Footer: float64Ptr(0.3),
		Header: float64Ptr(0.3),
		Left:   float64Ptr(0.5),
		Right:  float64Ptr(0.5),
		Top:    float64Ptr(0.5),
	})

	// 设置右侧页脚：版本号 + 信息等级（淡灰色）
	footerText := fmt.Sprintf("&R&K999999%s\n信息等级：内部公开 Info Class: Internal Disclosure", config.Cfg.Document.PermissionDocumentVersion)
	f.SetHeaderFooter(sheet1, &excelize.HeaderFooterOptions{
		OddFooter: footerText,
	})

	// 为 B3 设置数据验证（岗位下拉框）
	// 数据源在 Sheet2 的 H1:H{posCount}
	posCount := len(rules)
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "B3"
	dv.SetSqrefDropList(fmt.Sprintf("'%s'!$H$1:$H$%d", sheet2Name, posCount))
	dv.SetError(excelize.DataValidationErrorStyleWarning, "提示", "请从下拉列表中选择岗位")
	f.AddDataValidation(sheet1, dv)

	// ==================== Sheet2: 权限规则参考 ====================
	idx, _ := f.NewSheet(sheet2Name)

	// 区域1: H列 - 岗位名称列表（供Sheet1下拉引用）
	for i, rule := range rules {
		cell := fmt.Sprintf("H%d", i+1)
		f.SetCellValue(sheet2Name, cell, rule.PositionName)
	}

	// 区域2: 查找表（从 Row 29 开始）
	f.SetCellValue(sheet2Name, "A29", "岗位")
	f.SetCellValue(sheet2Name, "B29", "系统")
	f.SetCellValue(sheet2Name, "C29", "角色")

	lookupHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#E2EFDA"},
		},
	})
	f.SetCellStyle(sheet2Name, "A29", "C29", lookupHeaderStyle)

	for i, lr := range lookupRows {
		row := 30 + i
		f.SetCellValue(sheet2Name, fmt.Sprintf("A%d", row), lr.Position)
		f.SetCellValue(sheet2Name, fmt.Sprintf("B%d", row), lr.System)
		f.SetCellValue(sheet2Name, fmt.Sprintf("C%d", row), lr.Roles)
	}

	// 区域3: 全量权限矩阵（在查找表下方，间隔3行）
	matrixStartRow := 30 + len(lookupRows) + 3

	f.SetCellValue(sheet2Name, fmt.Sprintf("A%d", matrixStartRow), "全量权限规则参考矩阵")
	f.MergeCell(sheet2Name, fmt.Sprintf("A%d", matrixStartRow),
		fmt.Sprintf("%s%d", columnLetter(len(allSystems)), matrixStartRow))
	matrixTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheet2Name, fmt.Sprintf("A%d", matrixStartRow),
		fmt.Sprintf("%s%d", columnLetter(len(allSystems)), matrixStartRow), matrixTitleStyle)
	f.SetRowHeight(sheet2Name, matrixStartRow, 30)

	// 矩阵表头行
	matrixHeaderRow := matrixStartRow + 1
	f.SetCellValue(sheet2Name, fmt.Sprintf("A%d", matrixHeaderRow), "岗位")
	f.SetCellStyle(sheet2Name, fmt.Sprintf("A%d", matrixHeaderRow), fmt.Sprintf("A%d", matrixHeaderRow), tableHeaderStyle)
	for j, sys := range allSystems {
		col := columnLetter(j + 1)
		f.SetCellValue(sheet2Name, fmt.Sprintf("%s%d", col, matrixHeaderRow), sys)
		f.SetCellStyle(sheet2Name, fmt.Sprintf("%s%d", col, matrixHeaderRow), fmt.Sprintf("%s%d", col, matrixHeaderRow), tableHeaderStyle)
	}

	// 矩阵数据行
	matrixDataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   9,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	matrixPosStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   9,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	for i, rule := range rules {
		dataRow := matrixHeaderRow + 1 + i
		f.SetCellValue(sheet2Name, fmt.Sprintf("A%d", dataRow), rule.PositionName)
		f.SetCellStyle(sheet2Name, fmt.Sprintf("A%d", dataRow), fmt.Sprintf("A%d", dataRow), matrixPosStyle)

		// 解析该岗位的权限规则
		var sysRules []systemRule
		json.Unmarshal([]byte(rule.RulesJSON), &sysRules)

		// 构建 system -> enabled roles 的映射
		sysRoleMap := map[string][]string{}
		for _, sr := range sysRules {
			var enabled []string
			for _, r := range sr.Roles {
				if r.Enabled {
					enabled = append(enabled, r.Name)
				}
			}
			if len(enabled) > 0 {
				sysRoleMap[sr.System] = enabled
			}
		}

		// 填充每个系统列
		for j, sys := range allSystems {
			col := columnLetter(j + 1)
			roles, ok := sysRoleMap[sys]
			if ok && len(roles) > 0 {
				f.SetCellValue(sheet2Name, fmt.Sprintf("%s%d", col, dataRow), strings.Join(roles, ", "))
			} else {
				f.SetCellValue(sheet2Name, fmt.Sprintf("%s%d", col, dataRow), "-")
			}
			f.SetCellStyle(sheet2Name, fmt.Sprintf("%s%d", col, dataRow), fmt.Sprintf("%s%d", col, dataRow), matrixDataStyle)
		}
	}

	// Sheet2 列宽
	f.SetColWidth(sheet2Name, "A", "A", 18)
	f.SetColWidth(sheet2Name, "B", "B", 20)
	f.SetColWidth(sheet2Name, "C", "C", 40)
	f.SetColWidth(sheet2Name, "H", "H", 20)

	// 设置Sheet2为不打印（设置一个极小的打印区域或隐藏）
	// excelize 不支持直接设置"不打印"，但我们可以在Sheet2设置页面布局
	f.SetPageLayout(sheet2Name, &excelize.PageLayoutOptions{
		Size:        intPtr(9),
		Orientation: stringPtr("landscape"),
	})

	// 设置活动Sheet为Sheet1
	f.SetActiveSheet(idx)
	// 确保Sheet1为第一个sheet
	sheet1Idx, _ := f.GetSheetIndex(sheet1)
	f.SetActiveSheet(sheet1Idx)

	// 设置左侧页头：插入Logo图片到左上角（缩小尺寸，避免遮挡文字）
	logoPath := config.Cfg.Document.LogoPath
	if logoPath != "" {
		logoData, err := os.ReadFile(logoPath)
		if err != nil {
			fmt.Printf("读取Logo文件失败: %v\n", err)
		} else {
			if err := f.AddPictureFromBytes(sheet1, "A1", &excelize.Picture{
				Extension: ".png",
				File:      logoData,
				Format: &excelize.GraphicOptions{
					ScaleX:       0.08,
					ScaleY:       0.08,
				},
			}); err != nil {
				fmt.Printf("插入Logo失败: %v\n", err)
			}
		}
	}

	// 输出Excel文件
	// 优先使用保管区中定义的文件名（通过 context 传递）
	fileName := "用户变更记录表.xlsx"
	if preferredName, exists := c.Get("preferred_filename"); exists {
		if nameStr, ok := preferredName.(string); ok && nameStr != "" {
			fileName = nameStr
		}
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// 同时设置 filename（ASCII 回退）和 filename*（UTF-8）以兼容不同浏览器
	asciiName := toASCIIFallback(fileName)
	c.Header("Content-Disposition", fmt.Sprintf(
		"attachment; filename=\"%s\"; filename*=UTF-8''%s",
		asciiName,
		url.PathEscape(fileName),
	))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")

	if err := f.Write(c.Writer); err != nil {
		fmt.Printf("导出用户变更记录表失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出失败"})
		return
	}
}

// columnLetter 将数字索引转换为Excel列字母（1=A, 2=B, ..., 26=Z, 27=AA）
func columnLetter(idx int) string {
	result := ""
	for idx > 0 {
		idx--
		result = string(rune('A'+idx%26)) + result
		idx /= 26
	}
	return result
}

// 辅助指针函数
func intPtr(v int) *int          { return &v }
func uintPtr(v uint) *uint       { return &v }
func stringPtr(v string) *string { return &v }
func boolPtr(v bool) *bool       { return &v }
func float64Ptr(v float64) *float64 { return &v }
