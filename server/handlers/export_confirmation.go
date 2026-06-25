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

	specialConfirmStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			WrapText:   true,
			Horizontal: "right", // 右对齐，留出左侧签字空间
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
	f.MergeCell(sheetName, "A1", "G1") // 固定7列
	f.SetCellStyle(sheetName, "A1", "G1", titleStyle)
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
	f.MergeCell(sheetName, "A3", "G3")
	f.SetCellStyle(sheetName, "A3", "G3", headerInfoStyle)
	f.SetRowHeight(sheetName, 3, 22)

	// ---- Row 4: 表头 ----
	headers := []string{"序号", "姓名", "岗位", "系统", "角色", "确认结果", "特殊确认人"}
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
			// F列：显示三个复选框 □A □B □C
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), "□A  □B  □C")
			// G列：特殊确认人（留空）
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), "/")
			// 设置所有列的样式
			for i := 0; i < 7; i++ {
				col := string(rune('A' + i))
				f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), dataCellStyle)
			}
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), nameCellStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("C%d", rowNum), nameCellStyle)
			// G列使用右对齐样式
			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), specialConfirmStyle)
			f.SetRowHeight(sheetName, rowNum, 24)
			rowNum++
			seq++
		} else {
			// 有系统角色，每个系统一行，第一行显示序号/姓名/岗位
			startRow := rowNum
			
			// 检查是否为密钥团队的密钥经理（需要合并G列）
			isKeyManagerInKeyTeam := dept.Name == "密钥团队" && containsKeyword(user.PositionName, []string{"密钥经理", "Key Manager"})
			
			for i, vr := range validRoles {
				if i == 0 {
					f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), seq)
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), user.Name)
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), user.PositionName)
				}
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), vr.System)
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), vr.Roles)
				// F列：显示三个复选框 □A □B □C
				f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), "□A  □B  □C")
				// G列：特殊确认人 - 根据系统和角色判断，使用右对齐样式
				specialConfirm := getSpecialConfirm(vr.System, vr.Roles)
				f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), specialConfirm)

				for j := 0; j < 7; j++ {
					col := string(rune('A' + j))
					f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), dataCellStyle)
				}
				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), nameCellStyle)
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("C%d", rowNum), nameCellStyle)
				// G列使用右对齐样式，留出左侧签字空间
				f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), specialConfirmStyle)
				f.SetRowHeight(sheetName, rowNum, 24)
				rowNum++
			}
			// 合并序号、姓名、岗位列（如果多行）
			if len(validRoles) > 1 {
				endRow := rowNum - 1
				f.MergeCell(sheetName, fmt.Sprintf("A%d", startRow), fmt.Sprintf("A%d", endRow))
				f.MergeCell(sheetName, fmt.Sprintf("B%d", startRow), fmt.Sprintf("B%d", endRow))
				f.MergeCell(sheetName, fmt.Sprintf("C%d", startRow), fmt.Sprintf("C%d", endRow))
				
				// 如果是密钥团队的密钥经理，合并G列并显示(CISO)
				if isKeyManagerInKeyTeam {
					f.MergeCell(sheetName, fmt.Sprintf("G%d", startRow), fmt.Sprintf("G%d", endRow))
					f.SetCellValue(sheetName, fmt.Sprintf("G%d", startRow), "(CISO)")
				}
			}
			seq++
		}
	}

	// 如果没有用户，显示空行提示
	if len(users) == 0 {
		f.SetCellValue(sheetName, "A5", "（该部门暂无用户）")
		f.MergeCell(sheetName, "A5", "G5")
		f.SetCellStyle(sheetName, "A5", "G5", dataCellStyle)
		rowNum = 6
	}

	// ---- 空行 ----
	rowNum++

	// ---- 复核行 ----
	footerText := "复核部门：IT          复核人：                         复核日期："
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), footerText)
	f.MergeCell(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("G%d", rowNum))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("G%d", rowNum), footerStyle)
	f.SetRowHeight(sheetName, rowNum, 30)

	// ---- 设置列宽 ----
	f.SetColWidth(sheetName, "A", "A", 8)   // 序号
	f.SetColWidth(sheetName, "B", "B", 14)  // 姓名
	f.SetColWidth(sheetName, "C", "C", 20)  // 岗位
	f.SetColWidth(sheetName, "D", "D", 22)  // 系统
	f.SetColWidth(sheetName, "E", "E", 35)  // 角色
	f.SetColWidth(sheetName, "F", "F", 14)  // 确认结果
	f.SetColWidth(sheetName, "G", "G", 20)  // 特殊确认人（增加宽度，留出签字空间）

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

// containsKeyword 检查字符串是否包含任一关键词
func containsKeyword(str string, keywords []string) bool {
	for _, kw := range keywords {
		if str != "" && (str == kw || len(str) >= len(kw) && (str[:len(kw)] == kw || str[len(str)-len(kw):] == kw || containsSubstring(str, kw))) {
			return true
		}
	}
	return false
}

// containsSubstring 检查字符串是否包含子串（简单实现）
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// getSpecialConfirm 根据系统和角色返回特殊确认人标识
func getSpecialConfirm(system, roles string) string {
	// 产业部：数据库系统 + 服务用户 -> DBA
	if containsKeyword(system, []string{"数据库", "Database", "DB"}) && 
	   containsKeyword(roles, []string{"服务用户", "Service User"}) {
		return "(DBA)"
	}
	
	// 产业部：数据库系统 + DBA角色 -> CISO
	if containsKeyword(system, []string{"数据库", "Database", "DB"}) && 
	   containsKeyword(roles, []string{"DBA", "数据库管理员"}) {
		return "(CISO)"
	}
	
	// 密钥团队：任何系统 + 密钥经理角色 -> CISO
	if containsKeyword(roles, []string{"密钥经理", "Key Manager", "密钥管理"}) {
		return "(CISO)"
	}
	
	// 默认返回斜杠
	return "/"
}
