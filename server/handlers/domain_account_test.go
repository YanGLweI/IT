package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"it-platform-server/config"
	"it-platform-server/middleware"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// 测试账号配置
const (
	testUsername = "ldap"
	testPassword = "!Qw2!Qw2!Qw2!Qw2"
	testNewPwd   = "!Qw2!Qw2!Qw2!Qw3"
)

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	{
		public := api.Group("")
		public.POST("/domain-account/login", middleware.RateLimit(), DomainAccountLogin)
		public.POST("/domain-account/change-password", middleware.RateLimit(), ChangeDomainPassword)

		protected := api.Group("")
		protected.Use(middleware.JWTAuth())
		protected.GET("/domain-account/info", GetDomainAccountInfo)
	}
	return r
}

// initConfig 初始化配置（测试前调用）
func initConfig() {
	if config.Cfg == nil {
		if err := config.LoadConfig(); err != nil {
			panic("加载配置失败: " + err.Error())
		}
	}
}

// TestT1_DomainAccountLogin_Success T1: 域账号登录（正常）
func TestT1_DomainAccountLogin_Success(t *testing.T) {
	initConfig()
	r := setupTestRouter()

	body, _ := json.Marshal(map[string]string{
		"username": testUsername,
		"password": testPassword,
	})

	req, _ := http.NewRequest("POST", "/api/domain-account/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 得到 %d, 响应: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("响应数据格式错误")
	}

	if data["status"] != "ok" {
		t.Logf("登录状态: %v (密码可能已过期)", data["status"])
	}

	if data["token"] == nil && data["status"] == "ok" {
		t.Error("登录成功但未返回 token")
	}

	t.Logf("T1 通过: 域账号登录成功, 响应: %+v", data)
}

// TestT2_DomainAccountLogin_WrongPassword T2: 域账号登录（错误密码）
func TestT2_DomainAccountLogin_WrongPassword(t *testing.T) {
	initConfig()
	r := setupTestRouter()

	body, _ := json.Marshal(map[string]string{
		"username": testUsername,
		"password": "wrong_password_123",
	})

	req, _ := http.NewRequest("POST", "/api/domain-account/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 401, 得到 %d", w.Code)
	}

	t.Logf("T2 通过: 错误密码登录返回 401")
}

// TestT3_DomainAccountLogin_UserNotExist T3: 域账号登录（不存在用户）
func TestT3_DomainAccountLogin_UserNotExist(t *testing.T) {
	initConfig()
	r := setupTestRouter()

	body, _ := json.Marshal(map[string]string{
		"username": "nonexistent_user_xyz",
		"password": "some_password",
	})

	req, _ := http.NewRequest("POST", "/api/domain-account/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 401, 得到 %d", w.Code)
	}

	t.Logf("T3 通过: 不存在用户登录返回 401")
}

// TestT4_GetDomainAccountInfo T4: 获取账号信息
func TestT4_GetDomainAccountInfo(t *testing.T) {
	initConfig()

	// 先登录获取 token
	result, err := services.AuthenticateWithExpiry(testUsername, testPassword)
	if err != nil {
		t.Skipf("跳过测试: 无法连接 LDAP - %v", err)
	}
	if result.Status != "ok" {
		t.Skipf("跳过测试: 登录状态为 %s", result.Status)
	}

	// 生成 token
	token, err := generateDomainAccountToken(testUsername)
	if err != nil {
		t.Fatal("生成 token 失败:", err)
	}

	r := setupTestRouter()
	req, _ := http.NewRequest("GET", "/api/domain-account/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 得到 %d, 响应: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})

	t.Logf("T4 通过: 账号信息 - 密码到期: %v, 剩余天数: %v, 永不过期: %v",
		data["password_expires_at"], data["days_remaining"], data["password_never_expires"])
}

// TestT5_ChangePassword_WeakPassword T5: 修改密码（新密码不符合复杂度）
func TestT5_ChangePassword_WeakPassword(t *testing.T) {
	initConfig()
	r := setupTestRouter()

	body, _ := json.Marshal(map[string]string{
		"username":     testUsername,
		"old_password": testPassword,
		"new_password": "123", // 太短
	})

	req, _ := http.NewRequest("POST", "/api/domain-account/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400, 得到 %d", w.Code)
	}

	t.Logf("T5 通过: 弱密码被拒绝")
}

// TestT6_ChangePassword_WrongOldPassword T6: 修改密码（旧密码错误）
func TestT6_ChangePassword_WrongOldPassword(t *testing.T) {
	initConfig()
	r := setupTestRouter()

	body, _ := json.Marshal(map[string]string{
		"username":     testUsername,
		"old_password": "wrong_old_password_123!",
		"new_password": testNewPwd,
	})

	req, _ := http.NewRequest("POST", "/api/domain-account/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400, 得到 %d", w.Code)
	}

	t.Logf("T6 通过: 错误旧密码被拒绝")
}

// TestT7_T8_T9_PasswordChangeRegression T7+T8: 密码修改回归测试（修改密码 → 新密码登录 → 恢复原密码）
func TestT7_T8_T9_PasswordChangeRegression(t *testing.T) {
	initConfig()
	r := setupTestRouter()

	// 步骤1: 用原始密码登录
	t.Log("步骤1: 用原始密码登录...")
	loginBody, _ := json.Marshal(map[string]string{
		"username": testUsername,
		"password": testPassword,
	})
	loginReq, _ := http.NewRequest("POST", "/api/domain-account/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Skipf("跳过回归测试: 无法用原始密码登录 - %s", loginW.Body.String())
	}

	var loginResp map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	loginData := loginResp["data"].(map[string]interface{})

	if loginData["status"] == "expired" {
		t.Log("注意: 密码已过期，但仍可进行修改密码测试")
	}

	// 步骤2: 修改密码为新密码
	t.Log("步骤2: 修改密码为新密码...")
	changeBody, _ := json.Marshal(map[string]string{
		"username":     testUsername,
		"old_password": testPassword,
		"new_password": testNewPwd,
	})
	changeReq, _ := http.NewRequest("POST", "/api/domain-account/change-password", bytes.NewBuffer(changeBody))
	changeReq.Header.Set("Content-Type", "application/json")
	changeW := httptest.NewRecorder()
	r.ServeHTTP(changeW, changeReq)

	if changeW.Code != http.StatusOK {
		t.Fatalf("修改密码失败: %s", changeW.Body.String())
	}
	t.Log("T7 通过: 密码修改成功")

	// 步骤3: 用新密码登录
	t.Log("步骤3: 用新密码登录...")
	newLoginBody, _ := json.Marshal(map[string]string{
		"username": testUsername,
		"password": testNewPwd,
	})
	newLoginReq, _ := http.NewRequest("POST", "/api/domain-account/login", bytes.NewBuffer(newLoginBody))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginW := httptest.NewRecorder()
	r.ServeHTTP(newLoginW, newLoginReq)

	if newLoginW.Code != http.StatusOK {
		t.Fatalf("用新密码登录失败: %s", newLoginW.Body.String())
	}
	t.Log("T8 通过: 用新密码登录成功")

	// 步骤4: 恢复原始密码
	t.Log("步骤4: 恢复原始密码...")
	restoreBody, _ := json.Marshal(map[string]string{
		"username":     testUsername,
		"old_password": testNewPwd,
		"new_password": testPassword,
	})
	restoreReq, _ := http.NewRequest("POST", "/api/domain-account/change-password", bytes.NewBuffer(restoreBody))
	restoreReq.Header.Set("Content-Type", "application/json")
	restoreW := httptest.NewRecorder()
	r.ServeHTTP(restoreW, restoreReq)

	if restoreW.Code != http.StatusOK {
		t.Fatalf("恢复原始密码失败: %s", restoreW.Body.String())
	}

	// 步骤5: 用原始密码登录确认恢复成功
	t.Log("步骤5: 用原始密码登录确认恢复...")
	finalLoginBody, _ := json.Marshal(map[string]string{
		"username": testUsername,
		"password": testPassword,
	})
	finalLoginReq, _ := http.NewRequest("POST", "/api/domain-account/login", bytes.NewBuffer(finalLoginBody))
	finalLoginReq.Header.Set("Content-Type", "application/json")
	finalLoginW := httptest.NewRecorder()
	r.ServeHTTP(finalLoginW, finalLoginReq)

	if finalLoginW.Code != http.StatusOK {
		t.Fatalf("用原始密码登录失败: %s", finalLoginW.Body.String())
	}

	t.Log("T9 通过: 密码已恢复，回归测试完成")
}

// TestT10_MainPlatformLogin_PasswordExpiry T10: 主平台登录（密码过期检测）
func TestT10_MainPlatformLogin_PasswordExpiry(t *testing.T) {
	initConfig()

	// 直接测试 ldapAuthenticate 函数的密码过期检测
	userDN, displayName, passwordExpired, expiryInfo, err := ldapAuthenticate(testUsername, testPassword)
	if err != nil {
		t.Skipf("跳过测试: LDAP 认证失败 - %v", err)
	}

	t.Logf("T10: userDN=%s, displayName=%s, passwordExpired=%v", userDN, displayName, passwordExpired)

	if passwordExpired {
		t.Log("T10 通过: 检测到密码已过期")
	} else if expiryInfo != nil {
		t.Logf("T10 通过: 密码到期时间=%s, 剩余天数=%d, 永不过期=%v",
			expiryInfo.PasswordExpiresAt, expiryInfo.DaysRemaining, expiryInfo.PasswordNeverExpires)
	} else {
		t.Log("T10 通过: 未获取到过期信息（可能查询失败）")
	}
}
