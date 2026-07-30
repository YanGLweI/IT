<template>
  <div class="account-page">
    <!-- Header -->
    <header class="page-header">
      <div class="header-content">
        <div class="header-left">
          <div class="header-logo">
            <svg width="32" height="32" viewBox="0 0 48 48" fill="none">
              <rect width="48" height="48" rx="12" fill="#007AFF"/>
              <path d="M24 14C19.58 14 16 17.58 16 22C16 26.42 19.58 30 24 30C28.42 30 32 26.42 32 22C32 17.58 28.42 14 24 14ZM24 27C21.24 27 19 24.76 19 22C19 19.24 21.24 17 24 17C26.76 17 29 19.24 29 22C29 24.76 26.76 27 24 27ZM24 32C19.33 32 10 34.34 10 39V41H38V39C38 34.34 28.67 32 24 32Z" fill="white"/>
            </svg>
          </div>
          <h1 class="header-title">域账号自助管理平台</h1>
        </div>
        <div class="header-right">
          <span class="user-badge">
            <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/>
            </svg>
            {{ username }}
          </span>
          <el-button type="danger" plain class="btn-logout" @click="handleLogout">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 4px">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
              <polyline points="16 17 21 12 16 7"/>
              <line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
            退出
          </el-button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <!-- Account Info Card -->
      <section class="info-section">
        <div class="card info-card" v-loading="loading">
          <div class="card-header">
            <div class="card-title">
              <svg width="20" height="20" viewBox="0 0 20 20" fill="#007AFF">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
              </svg>
              <h2>账号信息</h2>
            </div>
          </div>

          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">账号</span>
              <span class="info-value">{{ username }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">密码状态</span>
              <span class="info-value">
                <span class="status-badge" :class="passwordStatusClass">
                  {{ passwordStatusText }}
                </span>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">密码过期时间</span>
              <span class="info-value">
                <template v-if="passwordNeverExpires">
                  <span class="status-badge status-info">永不过期</span>
                </template>
                <template v-else>
                  {{ passwordExpiresAt || '加载中...' }}
                </template>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">剩余天数</span>
              <span class="info-value">
                <template v-if="passwordNeverExpires">
                  <span class="status-badge status-info">-</span>
                </template>
                <template v-else>
                  <span :class="{ 'text-warning': isExpiringSoon, 'text-danger': isExpired }">
                    {{ daysRemaining }} 天
                  </span>
                </template>
              </span>
            </div>
          </div>

          <!-- Status Alerts -->
          <div v-if="passwordNeverExpires" class="status-alert status-alert-info">
            <svg width="18" height="18" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
            </svg>
            <span>当前账号已设置密码永不过期</span>
          </div>

          <div v-else-if="isExpiringSoon && !isExpired" class="status-alert status-alert-warning">
            <svg width="18" height="18" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
            </svg>
            <span>您的密码将在 <strong>{{ daysRemaining }}</strong> 天后过期，请尽快修改</span>
          </div>

          <div v-else-if="isExpired" class="status-alert status-alert-error">
            <svg width="18" height="18" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/>
            </svg>
            <span>您的密码已过期，请立即修改</span>
          </div>
        </div>
      </section>

      <!-- Change Password Card -->
      <section class="password-section">
        <div class="card password-card">
          <div class="card-header">
            <div class="card-title">
              <svg width="20" height="20" viewBox="0 0 20 20" fill="#007AFF">
                <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
              </svg>
              <h2>修改密码</h2>
            </div>
          </div>

          <el-form
            ref="formRef"
            :model="passwordForm"
            :rules="rules"
            label-position="top"
            class="password-form"
          >
            <el-form-item label="旧密码" prop="oldPassword">
              <el-input
                v-model="passwordForm.oldPassword"
                type="password"
                placeholder="请输入当前密码"
                show-password
              />
            </el-form-item>

            <el-form-item label="新密码" prop="newPassword">
              <el-input
                v-model="passwordForm.newPassword"
                type="password"
                placeholder="请输入新密码"
                show-password
                @input="checkPasswordStrength"
              />
              <!-- Strength Indicator -->
              <div v-if="passwordForm.newPassword" class="strength-indicator">
                <div class="strength-bars">
                  <div class="strength-bar" :class="{ active: strengthLevel >= 1 }"></div>
                  <div class="strength-bar" :class="{ active: strengthLevel >= 2 }"></div>
                  <div class="strength-bar" :class="{ active: strengthLevel >= 3 }"></div>
                  <div class="strength-bar" :class="{ active: strengthLevel >= 4 }"></div>
                </div>
                <span class="strength-text" :style="{ color: strengthColor }">{{ strengthText }}</span>
              </div>
              <div class="password-tips">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor" class="tip-icon">
                  <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
                </svg>
                <span>至少14位，包含大写字母、小写字母、数字、特殊字符中的至少3类</span>
              </div>
            </el-form-item>

            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                placeholder="请再次输入新密码"
                show-password
              />
            </el-form-item>

            <div class="form-actions">
              <el-button
                type="primary"
                :loading="loading"
                class="btn-primary"
                @click="handleChangePassword"
              >
                确认修改
              </el-button>
              <el-button class="btn-secondary" @click="resetForm">
                重置
              </el-button>
            </div>
          </el-form>
        </div>
      </section>
    </main>
  </div>
</template>

<script>
import { getDomainAccountInfo, changeDomainPassword } from '@/api/domain_account'

export default {
  name: 'AccountDashboard',
  data() {
    const validateConfirm = (rule, value, callback) => {
      if (value !== this.passwordForm.newPassword) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }
    return {
      loading: false,
      username: '',
      passwordExpiresAt: '',
      daysRemaining: 0,
      passwordNeverExpires: false,
      passwordForm: {
        oldPassword: '',
        newPassword: '',
        confirmPassword: ''
      },
      rules: {
        oldPassword: [
          { required: true, message: '请输入旧密码', trigger: 'blur' }
        ],
        newPassword: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 14, message: '密码长度不能少于14位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请确认新密码', trigger: 'blur' },
          { validator: validateConfirm, trigger: 'blur' }
        ]
      },
      strengthLevel: 0,
      strengthColor: '#909399',
      strengthText: ''
    }
  },
  computed: {
    isExpired() {
      return !this.passwordNeverExpires && this.daysRemaining <= 0
    },
    isExpiringSoon() {
      return !this.passwordNeverExpires && this.daysRemaining > 0 && this.daysRemaining <= 7
    },
    passwordStatusClass() {
      if (this.passwordNeverExpires) return 'status-info'
      if (this.isExpired) return 'status-danger'
      if (this.isExpiringSoon) return 'status-warning'
      return 'status-success'
    },
    passwordStatusText() {
      if (this.passwordNeverExpires) return '永不过期'
      if (this.isExpired) return '已过期'
      if (this.isExpiringSoon) return '即将过期'
      return '正常'
    }
  },
  mounted() {
    // 检查是否已登录
    if (!localStorage.getItem('domain_account_token')) {
      this.$router.push('/account/login')
      return
    }
    this.username = localStorage.getItem('domain_account_username') || ''
    this.fetchInfo()
  },
  methods: {
    async fetchInfo() {
      this.loading = true
      try {
        const res = await getDomainAccountInfo()
        const data = res.data?.data || res.data
        this.passwordExpiresAt = data.password_expires_at || '未知'
        this.daysRemaining = data.days_remaining ?? 0
        this.passwordNeverExpires = data.password_never_expires || false
      } catch (e) {
        // 如果获取失败，尝试从 localStorage 读取
        const stored = localStorage.getItem('domain_account_expiry')
        if (stored) {
          try {
            const data = JSON.parse(stored)
            this.passwordExpiresAt = data.password_expires_at || '未知'
            this.daysRemaining = data.days_remaining ?? 0
            this.passwordNeverExpires = data.password_never_expires || false
          } catch {}
        }
      } finally {
        this.loading = false
      }
    },

    checkPasswordStrength() {
      const pwd = this.passwordForm.newPassword
      if (!pwd) {
        this.strengthLevel = 0
        this.strengthColor = '#909399'
        this.strengthText = ''
        return
      }

      let strength = 0
      if (pwd.length >= 14) strength++
      if (/[A-Z]/.test(pwd)) strength++
      if (/[a-z]/.test(pwd)) strength++
      if (/[0-9]/.test(pwd)) strength++
      if (/[!@#$%^&*()_+\-=\[\]{}|;':",./<>?]/.test(pwd)) strength++

      if (strength <= 2) {
        this.strengthLevel = 1
        this.strengthColor = '#FF3B30'
        this.strengthText = '弱'
      } else if (strength === 3) {
        this.strengthLevel = 2
        this.strengthColor = '#FF9500'
        this.strengthText = '中'
      } else if (strength === 4) {
        this.strengthLevel = 3
        this.strengthColor = '#007AFF'
        this.strengthText = '强'
      } else {
        this.strengthLevel = 4
        this.strengthColor = '#34C759'
        this.strengthText = '很强'
      }
    },

    handleChangePassword() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return

        this.loading = true
        try {
          await changeDomainPassword(
            this.username,
            this.passwordForm.oldPassword,
            this.passwordForm.newPassword
          )

          this.$message.success('密码修改成功')

          this.$confirm('密码修改成功，是否立即退出重新登录？', '提示', {
            confirmButtonText: '退出登录',
            cancelButtonText: '继续操作',
            type: 'success'
          }).then(() => {
            this.handleLogout()
          }).catch(() => {})
        } catch (error) {
          const message = error.response?.data?.message || error.message || '密码修改失败'
          this.$message.error(message)
        } finally {
          this.loading = false
        }
      })
    },

    resetForm() {
      this.$refs.formRef.resetFields()
      this.strengthLevel = 0
      this.strengthColor = '#909399'
      this.strengthText = ''
    },

    handleLogout() {
      this.$confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        localStorage.removeItem('domain_account_token')
        localStorage.removeItem('domain_account_username')
        localStorage.removeItem('domain_account_expiry')
        this.$router.push('/account/login')
      }).catch(() => {})
    }
  }
}
</script>

<style scoped>
.account-page {
  min-height: 100vh;
  background: #F5F7FA;
}

/* Header */
.page-header {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid #E2E8F0;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 960px;
  margin: 0 auto;
  padding: 16px 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-logo {
  flex-shrink: 0;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
  letter-spacing: -0.3px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: rgba(0, 122, 255, 0.1);
  color: #007AFF;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.btn-logout {
  border-radius: 8px !important;
  font-weight: 500 !important;
  padding: 8px 14px !important;
  font-size: 13px !important;
  height: auto !important;
  line-height: 1 !important;
}

.btn-logout ::v-deep span {
  display: flex;
  align-items: center;
}

/* Main Content */
.main-content {
  max-width: 960px;
  margin: 0 auto;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 32px;
}

/* Cards */
.card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.card-header {
  padding: 24px 32px;
  border-bottom: 1px solid #E2E8F0;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-title h2 {
  font-size: 18px;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

/* Info Grid */
.info-grid {
  padding: 32px;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.info-label {
  font-size: 13px;
  color: #64748B;
  font-weight: 500;
}

.info-value {
  font-size: 16px;
  color: #1E293B;
  font-weight: 500;
}

/* Status Badges */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-success {
  background: rgba(52, 199, 89, 0.15);
  color: #248A3D;
}

.status-warning {
  background: rgba(255, 149, 0, 0.15);
  color: #C93400;
}

.status-danger {
  background: rgba(255, 59, 48, 0.15);
  color: #BF2C00;
}

.status-info {
  background: rgba(0, 122, 255, 0.1);
  color: #007AFF;
}

/* Status Alerts */
.status-alert {
  margin: 0 32px 32px;
  padding: 16px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.status-alert-info {
  background: rgba(0, 122, 255, 0.08);
  color: #007AFF;
  border: 1px solid rgba(0, 122, 255, 0.15);
}

.status-alert-warning {
  background: rgba(255, 149, 0, 0.08);
  color: #C93400;
  border: 1px solid rgba(255, 149, 0, 0.15);
}

.status-alert-error {
  background: rgba(255, 59, 48, 0.08);
  color: #BF2C00;
  border: 1px solid rgba(255, 59, 48, 0.15);
}

/* Text Colors */
.text-warning {
  color: #C93400;
  font-weight: 600;
}

.text-danger {
  color: #BF2C00;
  font-weight: 600;
}

/* Password Form */
.password-form {
  padding: 32px;
}

.password-form ::v-deep .el-form-item {
  margin-bottom: 24px;
}

.password-form ::v-deep .el-form-item__label {
  font-weight: 500;
  color: #1E293B;
  padding-bottom: 6px;
}

.password-form ::v-deep .el-input__inner {
  padding: 10px 14px;
  background: #F8FAFC;
  border-radius: 8px;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.password-form ::v-deep .el-input__inner:hover {
  background: #fff;
  border-color: #E2E8F0;
}

.password-form ::v-deep .el-input__inner:focus {
  background: #fff;
  border-color: #007AFF;
  box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.1) !important;
}

/* Strength Indicator */
.strength-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
}

.strength-bars {
  display: flex;
  gap: 4px;
  flex: 1;
}

.strength-bar {
  height: 4px;
  flex: 1;
  background: #E2E8F0;
  border-radius: 2px;
  transition: background 0.3s ease;
}

.strength-bar.active {
  background: currentColor;
}

.strength-text {
  font-size: 12px;
  font-weight: 600;
  min-width: 24px;
}

/* Password Tips */
.password-tips {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-top: 8px;
  font-size: 12px;
  color: #94A3B8;
  line-height: 1.4;
}

.tip-icon {
  flex-shrink: 0;
  margin-top: 1px;
  color: #007AFF;
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 16px;
  margin-top: 32px;
}

.btn-primary {
  height: 44px;
  padding: 0 32px;
  font-weight: 600 !important;
  border-radius: 8px !important;
}

.btn-secondary {
  height: 44px;
  border-radius: 8px !important;
}

/* Responsive */
@media (max-width: 768px) {
  .main-content {
    padding: 24px;
  }

  .info-grid {
    grid-template-columns: 1fr;
    gap: 16px;
    padding: 24px;
  }

  .header-content {
    padding: 16px;
  }

  .header-title {
    font-size: 16px;
  }

  .form-actions {
    flex-direction: column;
  }

  .btn-primary,
  .btn-secondary {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .user-badge span {
    display: none;
  }

  .card {
    border-radius: 12px;
  }
}
</style>
