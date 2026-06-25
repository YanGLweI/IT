package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ExportDepartmentConfirmation 导出部门用户确认表Excel
func ExportDepartmentConfirmation(c *gin.Context) {
	deptIDStr := c.Query("department_id")
	if deptIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请指定部门"})
		return
	}

	deptID, err := strconv.ParseUint(deptIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	// 查询部门
	var dept models.Department
	if err := database.GetDB().First(&dept, deptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "部门不存在"})
		return
	}

	// 查询该部门下所有用户
	var users []models.UserPermission
	if err := database.GetDB().Where("department_id = ?", deptID).Order("created_at asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户失败"})
		return
	}

	// 当前年月
	now := time.Now()
	yearMonth := fmt.Sprintf("%d年%d月份", now.Year(), now.Month())

	// 创建Excel
	f := excelize.NewFile()
	sheetName := dept.Name
	f.SetSheetName("Sheet1", sheetName)

	// ---- 样式定义 ----
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

	headerInfoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
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
			Vertical:   "center",
			WrapText:   true,
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	nameCellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			WrapText:   true,
			Horizontal: "left",
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

	// ---- Row 1: 标题 ----
	f.SetCellValue(sheetName, "A1", "用户确认表")
	f.MergeCell(sheetName, "A1", "F1")
	f.SetCellStyle(sheetName, "A1", "F1", titleStyle)
	f.SetRowHeight(sheetName, 1, 36)

	// ---- Row 2: 部门信息 ----
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("部门：%s", dept.Name))
	f.MergeCell(sheetName, "A2", "B2")
	f.SetCellStyle(sheetName, "A2", "B2", headerInfoStyle)

	f.SetCellValue(sheetName, "C2", "确认人（部门领导）：                         确认日期：")
	f.MergeCell(sheetName, "C2", "F2")
	f.SetCellStyle(sheetName, "C2", "F2", headerInfoStyle)
	f.SetRowHeight(sheetName, 2, 30)

	// ---- Row 3: 备注 ----
	f.SetCellValue(sheetName, "A3", "注：A:保留  B:不保留  C:权限有误")
	f.MergeCell(sheetName, "A3", "F3")
	f.SetCellStyle(sheetName, "A3", "F3", headerInfoStyle)
	f.SetRowHeight(sheetName, 3, 22)

	// ---- Row 4: 表头 ----
	headers := []string{"序号", "姓名", "岗位", "系统", "角色", "确认结果"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"4", h)
		f.SetCellStyle(sheetName, col+"4", col+"4", tableHeaderStyle)
	}
	f.SetRowHeight(sheetName, 4, 28)

	// ---- 数据行 ----
	rowNum := 5
	seq := 1
	for _, user := range users {
		// 解析系统角色
		var systemRoles []struct {
			System string   `json:"system"`
			Roles  []string `json:"roles"`
		}
		if user.SystemRolesJSON != "" && user.SystemRolesJSON != "[]" {
			if err := json.Unmarshal([]byte(user.SystemRolesJSON), &systemRoles); err != nil {
				systemRoles = nil
			}
		}

		// 过滤掉没有角色的系统
		var validRoles []struct {
			System string
			Roles  string
		}
		for _, sr := range systemRoles {
			if len(sr.Roles) > 0 {
				validRoles = append(validRoles, struct {
					System string
					Roles  string
				}{
					System: sr.System,
					Roles:  joinRoleList(sr.Roles),
				})
			}
		}

		if len(validRoles) == 0 {
			// 没有系统角色，仍然输出一行
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), seq)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), user.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), user.PositionName)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), "-")
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), "-")
			for _, col := range []string{"A", "B", "C", "D", "E", "F"} {
				f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), dataCellStyle)
			}
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), nameCellStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("C%d", rowNum), nameCellStyle)
			f.SetRowHeight(sheetName, rowNum, 24)
			rowNum++
			seq++
		} else {
			// 有系统角色，每个系统一行，第一行显示序号/姓名/岗位
			startRow := rowNum
			for i, vr := range validRoles {
				if i == 0 {
					f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), seq)
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), user.Name)
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), user.PositionName)
				}
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), vr.System)
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), vr.Roles)

				for _, col := range []string{"A", "B", "C", "D", "E", "F"} {
					f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), dataCellStyle)
				}
				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), nameCellStyle)
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("C%d", rowNum), nameCellStyle)
				f.SetRowHeight(sheetName, rowNum, 24)
				rowNum++
			}
			// 合并序号、姓名、岗位列（如果多行）
			if len(validRoles) > 1 {
				endRow := rowNum - 1
				f.MergeCell(sheetName, fmt.Sprintf("A%d", startRow), fmt.Sprintf("A%d", endRow))
				f.MergeCell(sheetName, fmt.Sprintf("B%d", startRow), fmt.Sprintf("B%d", endRow))
				f.MergeCell(sheetName, fmt.Sprintf("C%d", startRow), fmt.Sprintf("C%d", endRow))
			}
			seq++
		}
	}

	// 如果没有用户，显示空行提示
	if len(users) == 0 {
		f.SetCellValue(sheetName, "A5", "（该部门暂无用户）")
		f.MergeCell(sheetName, "A5", "F5")
		f.SetCellStyle(sheetName, "A5", "F5", dataCellStyle)
		rowNum = 6
	}

	// ---- 为F列添加下拉选择验证（A:保留 / B:不保留 / C:权限有误）----
	lastDataRow := rowNum - 1
	if lastDataRow >= 5 {
		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("F5:F%d", lastDataRow)
		dv.SetDropList([]string{"A:保留", "B:不保留", "C:权限有误"})
		dv.SetError(excelize.DataValidationErrorStyleStop, "输入错误", "请从下拉列表中选择 A/B/C")
		dv.SetInput("确认结果", "请选择：A-保留 / B-不保留 / C-权限有误")
		f.AddDataValidation(sheetName, dv)
	}

	// ---- 空行 ----
	rowNum++

	// ---- 复核行 ----
	footerText := "复核部门：IT          复核人：                         复核日期："
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), footerText)
	f.MergeCell(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("F%d", rowNum))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("F%d", rowNum), footerStyle)
	f.SetRowHeight(sheetName, rowNum, 30)

	// ---- 设置列宽 ----
	f.SetColWidth(sheetName, "A", "A", 8)   // 序号
	f.SetColWidth(sheetName, "B", "B", 14)  // 姓名
	f.SetColWidth(sheetName, "C", "C", 20)  // 岗位
	f.SetColWidth(sheetName, "D", "D", 22)  // 系统
	f.SetColWidth(sheetName, "E", "E", 35)  // 角色
	f.SetColWidth(sheetName, "F", "F", 14)  // 确认结果

	// 输出Excel文件
	fileName := fmt.Sprintf("IT07-2.0 用户确认表(%s)-%s.xlsx", yearMonth, dept.Name)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")

	if err := f.Write(c.Writer); err != nil {
		fmt.Printf("导出Excel失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出失败"})
		return
	}
}

// joinRoleList 将角色列表用逗号连接
func joinRoleList(roles []string) string {
	result := ""
	for i, r := range roles {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}
