package utils

import (
	"fmt"
	"strings"
	"unicode"

	"it-platform-server/config"
)

// ValidatePasswordComplexity 校验密码复杂度
func ValidatePasswordComplexity(password, username string) error {
	policy := &config.Cfg.PasswordPolicy

	// 1. 最小长度检查
	if len(password) < policy.MinLength {
		return fmt.Errorf("密码长度至少 %d 位", policy.MinLength)
	}

	// 2. 不能包含用户名
	if username != "" && strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
		return fmt.Errorf("密码不能包含用户名")
	}

	// 3. 统计字符类型
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	// 4. 按配置检查必须包含的字符类型
	if policy.RequireUppercase && !hasUpper {
		return fmt.Errorf("密码必须包含大写字母")
	}
	if policy.RequireLowercase && !hasLower {
		return fmt.Errorf("密码必须包含小写字母")
	}
	if policy.RequireDigit && !hasDigit {
		return fmt.Errorf("密码必须包含数字")
	}
	if policy.RequireSpecial && !hasSpecial {
		return fmt.Errorf("密码必须包含特殊字符")
	}

	// 5. 至少满足 3 类字符
	categoryCount := 0
	if hasUpper {
		categoryCount++
	}
	if hasLower {
		categoryCount++
	}
	if hasDigit {
		categoryCount++
	}
	if hasSpecial {
		categoryCount++
	}
	if categoryCount < 3 {
		return fmt.Errorf("密码必须至少包含 3 种字符类型（大写、小写、数字、特殊字符）")
	}

	return nil
}

// ValidatePasswordDifferent 校验新旧密码不同
func ValidatePasswordDifferent(oldPwd, newPwd string) error {
	if oldPwd == newPwd {
		return fmt.Errorf("新密码不能与旧密码相同")
	}
	return nil
}

// GetPasswordStrength 返回密码强度等级（1-4）
// 1=弱, 2=中, 3=强, 4=很强
func GetPasswordStrength(password string) int {
	if len(password) == 0 {
		return 0
	}

	score := 0

	// 长度评分
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	// 字符类型评分
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	categoryCount := 0
	if hasUpper {
		categoryCount++
	}
	if hasLower {
		categoryCount++
	}
	if hasDigit {
		categoryCount++
	}
	if hasSpecial {
		categoryCount++
	}

	// 综合评分
	strength := 1
	if len(password) >= 8 && categoryCount >= 2 {
		strength = 2
	}
	if len(password) >= 12 && categoryCount >= 3 {
		strength = 3
	}
	if len(password) >= 14 && categoryCount >= 4 {
		strength = 4
	}

	return strength
}

// GetPasswordStrengthLabel 返回密码强度标签
func GetPasswordStrengthLabel(strength int) string {
	switch strength {
	case 1:
		return "弱"
	case 2:
		return "中"
	case 3:
		return "强"
	case 4:
		return "很强"
	default:
		return "未知"
	}
}
