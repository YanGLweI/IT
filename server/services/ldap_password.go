package services

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	"it-platform-server/config"

	"github.com/go-ldap/ldap/v3"
)

// PasswordExpiryResult 密码过期查询结果
type PasswordExpiryResult struct {
	PasswordExpiresAt    string `json:"password_expires_at"`
	DaysRemaining        int    `json:"days_remaining"`
	PasswordNeverExpires bool   `json:"password_never_expires"`
}

// AuthWithExpiryResult 认证+过期状态结果
type AuthWithExpiryResult struct {
	Status    string // "ok", "expired", "error"
	AuthResult
}

// AuthResult 认证结果（包含过期信息）
type AuthResult struct {
	PasswordExpiryResult
}

// newLDAPConnection 创建 LDAP 连接（复用现有 TLS 配置逻辑）
func newLDAPConnection() (*ldap.Conn, error) {
	cfg := &config.Cfg.LDAP
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.Insecure,
		ServerName:         "10.60.254.252",
	}
	if cfg.CertPath != "" {
		caCert, err := os.ReadFile(cfg.CertPath)
		if err == nil {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
	}
	conn, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("连接 LDAP 服务器失败: %w", err)
	}
	return conn, nil
}

// serviceBind 使用服务账号绑定 LDAP
func serviceBind() (*ldap.Conn, error) {
	cfg := &config.Cfg.LDAP
	conn, err := newLDAPConnection()
	if err != nil {
		return nil, err
	}
	if err := conn.Bind(cfg.Username, cfg.Password); err != nil {
		conn.Close()
		return nil, fmt.Errorf("服务账号绑定失败: %w", err)
	}
	return conn, nil
}

// AuthenticateWithExpiry 认证用户并返回密码过期状态
func AuthenticateWithExpiry(username, password string) (*AuthWithExpiryResult, error) {
	cfg := &config.Cfg.LDAP
	userDN := fmt.Sprintf("%s@%s", username, cfg.DomainSuffix)

	conn, err := newLDAPConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = conn.Bind(userDN, password)
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			errStr := err.Error()
			// AD 密码过期错误码: data 773
			if strings.Contains(errStr, "773") || strings.Contains(errStr, "data 773") {
				return &AuthWithExpiryResult{Status: "expired"}, nil
			}
			// 账号被锁定: data 775
			if strings.Contains(errStr, "775") || strings.Contains(errStr, "data 775") {
				return &AuthWithExpiryResult{Status: "error"}, fmt.Errorf("账号已被锁定，请联系管理员")
			}
		}
		return &AuthWithExpiryResult{Status: "error"}, fmt.Errorf("账号或密码错误")
	}

	// 认证成功，查询密码过期信息
	expiry, err := GetPasswordExpiry(username)
	if err != nil {
		log.Printf("查询密码过期信息失败: %v", err)
		// 不影响登录，返回空过期信息
		return &AuthWithExpiryResult{
			Status: "ok",
			AuthResult: AuthResult{
				PasswordExpiryResult: PasswordExpiryResult{
					PasswordExpiresAt:    "未知",
					DaysRemaining:        0,
					PasswordNeverExpires: false,
				},
			},
		}, nil
	}

	return &AuthWithExpiryResult{
		Status:     "ok",
		AuthResult: AuthResult{PasswordExpiryResult: *expiry},
	}, nil
}

// GetPasswordExpiry 查询密码过期时间
func GetPasswordExpiry(username string) (*PasswordExpiryResult, error) {
	cfg := &config.Cfg.LDAP

	conn, err := serviceBind()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索用户，包含 userAccountControl 和 pwdLastSet
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(cfg.UserFilter, username),
		[]string{"pwdLastSet", "sAMAccountName", "userAccountControl"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("搜索用户失败: %w", err)
	}
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("用户不存在")
	}

	entry := result.Entries[0]

	// 检查 userAccountControl 中的 DONT_EXPIRE_PASSWORD 标志 (0x10000 = 65536)
	uacStr := entry.GetAttributeValue("userAccountControl")
	var uac int64
	fmt.Sscanf(uacStr, "%d", &uac)
	neverExpires := (uac & 0x10000) != 0

	if neverExpires {
		return &PasswordExpiryResult{
			PasswordExpiresAt:    "永不过期",
			DaysRemaining:        -1,
			PasswordNeverExpires: true,
		}, nil
	}

	// 获取 pwdLastSet (Windows FILETIME 格式)
	pwdLastSetStr := entry.GetAttributeValue("pwdLastSet")
	var pwdLastSetInt int64
	if pwdLastSetStr != "" {
		fmt.Sscanf(pwdLastSetStr, "%d", &pwdLastSetInt)
	} else {
		pwdLastSetBytes := entry.GetRawAttributeValue("pwdLastSet")
		if len(pwdLastSetBytes) == 8 {
			pwdLastSetInt = int64(binary.LittleEndian.Uint64(pwdLastSetBytes))
		}
	}

	if pwdLastSetInt == 0 {
		return &PasswordExpiryResult{
			PasswordExpiresAt:    "未知",
			DaysRemaining:        0,
			PasswordNeverExpires: false,
		}, nil
	}

	pwdLastSet := fileTimeToTime(pwdLastSetInt)
	if pwdLastSet.IsZero() {
		return &PasswordExpiryResult{
			PasswordExpiresAt:    "未知",
			DaysRemaining:        0,
			PasswordNeverExpires: false,
		}, nil
	}

	// 获取域密码最大使用期限
	maxPwdAge := getMaxPwdAge(conn)
	expiresAt := pwdLastSet.Add(maxPwdAge)
	daysRemaining := int(time.Until(expiresAt).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	return &PasswordExpiryResult{
		PasswordExpiresAt:    expiresAt.Format("2006-01-02 15:04:05"),
		DaysRemaining:        daysRemaining,
		PasswordNeverExpires: false,
	}, nil
}

// getMaxPwdAge 从域控获取 maxPwdAge 属性
func getMaxPwdAge(conn *ldap.Conn) time.Duration {
	cfg := &config.Cfg.LDAP

	// 从 BaseDN 搜索 maxPwdAge 属性
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=domain)",
		[]string{"maxPwdAge"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil || len(result.Entries) == 0 {
		// 回退到配置的 max_pwd_age_days
		return time.Duration(cfg.MaxPwdAgeDays) * 24 * time.Hour
	}

	entry := result.Entries[0]
	maxPwdAgeBytes := entry.GetRawAttributeValue("maxPwdAge")
	if len(maxPwdAgeBytes) != 8 {
		return time.Duration(cfg.MaxPwdAgeDays) * 24 * time.Hour
	}

	// maxPwdAge 是负数的 100ns 间隔
	maxPwdAgeInt := int64(binary.LittleEndian.Uint64(maxPwdAgeBytes))
	if maxPwdAgeInt >= 0 {
		return time.Duration(cfg.MaxPwdAgeDays) * 24 * time.Hour
	}

	// 转换为 time.Duration（100ns 间隔 → 正数）
	duration := time.Duration(-maxPwdAgeInt) * 100 * time.Nanosecond
	return duration
}

// fileTimeToTime Windows FILETIME → Go time.Time 转换
func fileTimeToTime(fileTime int64) time.Time {
	seconds := fileTime / 10000000
	unixSeconds := seconds - 11644473600 // Windows epoch(1601) → Unix epoch(1970)
	return time.Unix(unixSeconds, 0).Local()
}

// VerifyPassword 验证密码（data 773 视为通过，因为密码正确但已过期）
func VerifyPassword(username, password string) error {
	cfg := &config.Cfg.LDAP
	userDN := fmt.Sprintf("%s@%s", username, cfg.DomainSuffix)

	conn, err := newLDAPConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Bind(userDN, password)
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			errStr := err.Error()
			// data 773 = 密码正确但已过期，视为验证通过
			if strings.Contains(errStr, "773") || strings.Contains(errStr, "data 773") {
				return nil
			}
		}
		return fmt.Errorf("密码验证失败")
	}
	return nil
}

// ChangePassword 修改域账号密码
func ChangePassword(username, oldPwd, newPwd string) error {
	cfg := &config.Cfg.LDAP

	conn, err := serviceBind()
	if err != nil {
		return err
	}
	defer conn.Close()

	// 搜索用户 DN
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(cfg.UserFilter, username),
		[]string{"dn"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("搜索用户失败: %w", err)
	}
	if len(result.Entries) == 0 {
		return fmt.Errorf("用户不存在")
	}

	userDN := result.Entries[0].DN

	// 编码新密码为 UTF-16LE + 双引号包裹
	newPwdEncoded := encodePassword(newPwd)

	// 使用 Modify 替换 unicodePwd 属性
	modifyRequest := ldap.NewModifyRequest(userDN, nil)
	modifyRequest.Replace("unicodePwd", []string{string(newPwdEncoded)})

	if err := conn.Modify(modifyRequest); err != nil {
		return fmt.Errorf("修改密码失败: %w", err)
	}

	log.Printf("用户 %s 密码修改成功", username)
	return nil
}

// encodePassword AD 密码编码：UTF-16LE + 双引号包裹
func encodePassword(password string) []byte {
	quoted := "\"" + password + "\""
	encoded := utf16.Encode([]rune(quoted))
	utf16Bytes := make([]byte, len(encoded)*2)
	for i, r := range encoded {
		binary.LittleEndian.PutUint16(utf16Bytes[i*2:], r)
	}
	return utf16Bytes
}
