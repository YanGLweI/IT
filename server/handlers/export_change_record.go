package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
			Vertical: "center",
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
			Vertical: "center",
		},
	})

	// ==================== Sheet1: 用户变更记录表 ====================
	sheet1 := "用户变更记录表"
	f.SetSheetName("Sheet1", sheet1)

	// Row 1: 标题
	f.SetCellValue(sheet1, "A1", "用户变更记录表")
	f.MergeCell(sheet1, "A1", "D1")
	f.SetCellStyle(sheet1, "A1", "D1", titleStyle)
	f.SetRowHeight(sheet1, 1, 36)

	// Row 2: 申请信息
	f.SetCellValue(sheet1, "A2", "申请部门：                                   申请人：                                    申请日期：")
	f.MergeCell(sheet1, "A2", "D2")
	f.SetCellStyle(sheet1, "A2", "D2", infoStyle)
	f.SetRowHeight(sheet1, 2, 30)

	// Row 3: 岗位(下拉) + 审批人
	f.SetCellValue(sheet1, "A3", "岗位：")
	f.SetCellStyle(sheet1, "A3", "A3", infoStyle)
	// C3 将由数据验证设置为下拉选择
	f.SetCellValue(sheet1, "B3", "")
	f.SetCellStyle(sheet1, "B3", "B3", infoStyle)
	f.SetCellValue(sheet1, "C3", "部门审批人：                           审批日期：")
	f.MergeCell(sheet1, "C3", "D3")
	f.SetCellStyle(sheet1, "C3", "D3", infoStyle)
	f.SetRowHeight(sheet1, 3, 30)

	// Row 4: 注释
	f.SetCellValue(sheet1, "A4", "注：不需要的权限请划竖线丨，删除的权限请打×，需要的权限请打√")
	f.MergeCell(sheet1, "A4", "D4")
	f.SetCellStyle(sheet1, "A4", "D4", noteStyle)
	f.SetRowHeight(sheet1, 4, 22)

	// Row 5: 表头
	headers := []string{"系统", "角色", "账号名", "备注"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheet1, col+"5", h)
		f.SetCellStyle(sheet1, col+"5", col+"5", tableHeaderStyle)
	}
	f.SetRowHeight(sheet1, 5, 28)

	// Row 6~25: 数据行（公式），预留20行
	// 查找表起始行（Sheet2 的 Row 29 + 1 表头 = 数据从 Row 30 开始）
	lookupStartRow := 30
	lookupEndRow := lookupStartRow + len(lookupRows) - 1
	if lookupEndRow < lookupStartRow {
		lookupEndRow = lookupStartRow
	}

	// Sheet2 名称（后面创建）
	sheet2Name := "权限规则参考"

	for row := 6; row <= 25; row++ {
		n := row - 5 // 第几个匹配项 (1, 2, 3, ...)

		// A列公式: 系统名 - 使用 INDEX + SMALL + IF 数组公式
		formulaA := fmt.Sprintf(
			`IFERROR(INDEX('%s'!B:B,SMALL(IF('%s'!$A$%d:$A$%d=$B$3,ROW('%s'!$A$%d:$A$%d)),%d)),"")`,
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
		f.SetCellStyle(sheet1, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), dataCellStyle)
		f.SetRowHeight(sheet1, row, 30)
	}

	// Row 26: 签字栏
	signRow := 26
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", signRow),
		"开通人：                          开通日期：                          复核人：                          复核日期：                          申请人确认：                          确认日期：")
	f.MergeCell(sheet1, fmt.Sprintf("A%d", signRow), fmt.Sprintf("D%d", signRow))
	f.SetCellStyle(sheet1, fmt.Sprintf("A%d", signRow), fmt.Sprintf("D%d", signRow), footerStyle)
	f.SetRowHeight(sheet1, signRow, 30)

	// 设置列宽
	f.SetColWidth(sheet1, "A", "A", 20)  // 系统
	f.SetColWidth(sheet1, "B", "B", 35)  // 角色
	f.SetColWidth(sheet1, "C", "C", 18)  // 账号名
	f.SetColWidth(sheet1, "D", "D", 18)  // 备注

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

	// 为 B3 设置数据验证（岗位下拉框）
	// 数据源在 Sheet2 的 H1:H{posCount}
	posCount := len(rules)
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "B3"
	dv.SetSqrefDropList(fmt.Sprintf("'%s'!$H$1:$H$%d", sheet2Name, posCount))
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

	// 输出Excel文件
	fileName := "用户变更记录表.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
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
