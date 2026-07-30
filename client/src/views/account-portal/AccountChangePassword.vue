<template>
  <div class="change-password-page">
    <!-- Background -->
    <div class="page-bg">
      <div class="bg-blob bg-blob-1"></div>
      <div class="bg-blob bg-blob-2"></div>
    </div>

    <!-- Content -->
    <div class="page-content">
      <div class="change-card">
        <!-- Header -->
        <div class="card-header">
          <div class="header-icon">
            <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
              <rect width="40" height="40" rx="10" fill="#007AFF"/>
              <path d="M20 10C17.24 10 15 12.24 15 15C15 17.76 17.24 20 20 20C22.76 20 25 17.76 25 15C25 12.24 22.76 10 20 10ZM20 18C18.34 18 17 16.66 17 15C17 13.34 18.34 12 20 12C21.66 12 23 13.34 23 15C23 16.66 21.66 18 20 18ZM20 22C16.67 22 10 23.67 10 27V29H30V27C30 23.67 23.33 22 20 22Z" fill="white"/>
            </svg>
          </div>
          <div class="header-text">
            <h1>修改密码</h1>
            <p v-if="isExpired" class="header-desc">密码已过期，请立即修改</p>
            <p v-else class="header-desc">更新您的域账号密码</p>
          </div>
          <el-tag v-if="isExpired" type="danger" class="expired-tag" effect="dark">
            已过期
          </el-tag>
        </div>

        <!-- Expired Alert -->
        <div v-if="isExpired" class="alert-wrapper">
          <div class="custom-alert alert-error">
            <svg class="alert-icon" width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
            </svg>
            <div>
              <p class="alert-title">密码已过期</p>
              <p class="alert-message">请修改密码后才能继续使用系统</p>
            </div>
          </div>
        </div>

        <!-- Form -->
        <el-form
          ref="formRef"
          :model="passwordForm"
          :rules="rules"
          label-position="top"
          class="password-form"
        >
          <el-form-item label="账号" prop="username">
            <el-input
              v-model="passwordForm.username"
              placeholder="域账号"
              disabled
            />
          </el-form-item>

          <el-form-item label="旧密码" prop="oldPassword">
            <el-input
              ref="oldPasswordInput"
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
            <!-- Password Strength -->
            <div v-if="passwordForm.newPassword" class="strength-indicator">
              <div class="strength-bars">
                <div class="strength-bar" :class="{ active: strengthLevel >= 1 }" style="--delay: 0s"></div>
                <div class="strength-bar" :class="{ active: strengthLevel >= 2 }" style="--delay: 0.1s"></div>
                <div class="strength-bar" :class="{ active: strengthLevel >= 3 }" style="--delay: 0.2s"></div>
                <div class="strength-bar" :class="{ active: strengthLevel >= 4 }" style="--delay: 0.3s"></div>
              </div>
              <span class="strength-text" :style="{ color: strengthColor }">{{ strengthText }}</span>
            </div>
            <!-- Password Tips -->
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
            <el-button class="btn-secondary" @click="handleBackToLogin">
              返回登录
            </el-button>
          </div>
        </el-form>

        <!-- 返回 IT 管理平台链接 -->
        <div class="back-platform-link">
          <router-link to="/login">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="19" y1="12" x2="5" y2="12"/>
              <polyline points="12 19 5 12 12 5"/>
            </svg>
            返回 IT 管理平台登录
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { changeDomainPassword } from '@/api/domain_account'

export default {
  name: 'AccountChangePassword',
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
      isExpired: false,
      passwordForm: {
        username: '',
        oldPassword: '',
        newPassword: '',
        confirmPassword: ''
      },
      rules: {
        username: [
          { required: true, message: '请输入域账号', trigger: 'blur' }
        ],
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
  mounted() {
    this.isExpired = this.$route.query.expired === 'true'
    // 优先从 sessionStorage 读取，fallback 到 localStorage
    const tempUsername = sessionStorage.getItem('temp_username') || localStorage.getItem('temp_username') || ''
    if (tempUsername) {
      this.passwordForm.username = tempUsername
    }
    // 自动聚焦到旧密码输入框
    this.$nextTick(() => {
      this.$refs.oldPasswordInput && this.$refs.oldPasswordInput.focus()
    })
  },
  methods: {
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
            this.passwordForm.username,
            this.passwordForm.oldPassword,
            this.passwordForm.newPassword
          )

          this.$message.success('密码修改成功，请重新登录')
          sessionStorage.removeItem('temp_username')
          localStorage.removeItem('temp_username')
          this.$router.push('/account-portal/login')
        } catch (error) {
          const message = error.response?.data?.message || error.message || '密码修改失败'
          this.$message.error(message)
        } finally {
          this.loading = false
        }
      })
    },

    handleBackToLogin() {
      sessionStorage.removeItem('temp_username')
      localStorage.removeItem('temp_username')
      this.$router.push('/account-portal/login')
    }
  }
}
</script>

<style scoped>
.change-password-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 16px;
  box-sizing: border-box;
}

/* Background */
.page-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  z-index: 0;
}

.bg-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.5;
  animation: float 20s infinite ease-in-out;
  min-width: 40vmax;
  min-height: 40vmax;
}

.bg-blob-1 {
  width: 55vmax;
  height: 55vmax;
  background: rgba(0, 122, 255, 0.4);
  top: -15vmax;
  right: -15vmax;
}

.bg-blob-2 {
  width: 45vmax;
  height: 45vmax;
  background: rgba(52, 199, 89, 0.3);
  bottom: -10vmax;
  left: -10vmax;
  animation-delay: -10s;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(3vmax, -3vmax) scale(1.05); }
  66% { transform: translate(-2vmax, 2vmax) scale(0.95); }
}

/* Content */
.page-content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
}

.change-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 16px;
  padding: 20px;
  box-shadow:
    0 20px 60px rgba(0, 0, 0, 0.15),
    0 8px 20px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  animation: slideUp 0.5s ease-out;
  max-height: calc(100vh - 32px);
  overflow-y: auto;
}

.change-card::-webkit-scrollbar {
  width: 4px;
}

.change-card::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 2px;
}

.change-card::-webkit-scrollbar-track {
  background: transparent;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Header */
.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #E2E8F0;
}

.header-icon {
  flex-shrink: 0;
}

.header-text {
  flex: 1;
}

.header-text h1 {
  font-size: 20px;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
  letter-spacing: -0.3px;
}

.header-desc {
  font-size: 13px;
  color: #64748B;
  margin: 4px 0 0;
}

.expired-tag {
  flex-shrink: 0;
  border-radius: 20px !important;
  font-weight: 600 !important;
}

/* Alert */
.alert-wrapper {
  margin-bottom: 12px;
}

.custom-alert {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  border-radius: 10px;
}

.alert-error {
  background: rgba(255, 59, 48, 0.1);
  border: 1px solid rgba(255, 59, 48, 0.2);
}

.alert-icon {
  flex-shrink: 0;
  color: #FF3B30;
  margin-top: 2px;
}

.alert-title {
  font-size: 13px;
  font-weight: 600;
  color: #FF3B30;
  margin: 0;
}

.alert-message {
  font-size: 12px;
  color: #64748B;
  margin: 2px 0 0;
}

/* Form */
.password-form ::v-deep .el-form-item {
  margin-bottom: 12px !important;
}

.password-form ::v-deep .el-form-item__label {
  font-weight: 500;
  color: #1E293B;
  padding-bottom: 4px;
}

.password-form ::v-deep .el-input__inner {
  padding: 8px 12px;
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
  animation: barFill 0.3s ease-out var(--delay);
}

@keyframes barFill {
  from { transform: scaleX(0); transform-origin: left; }
  to { transform: scaleX(1); transform-origin: left; }
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

/* Actions */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.btn-primary {
  flex: 1;
  height: 38px;
  font-weight: 600 !important;
  border-radius: 8px !important;
}

.btn-secondary {
  height: 38px;
  border-radius: 8px !important;
}

/* Responsive */
@media (max-width: 480px) {
  .change-card {
    padding: 20px;
    border-radius: 16px;
  }

  .header-text h1 {
    font-size: 18px;
  }

  .form-actions {
    flex-direction: column;
  }

  .btn-primary,
  .btn-secondary {
    width: 100%;
  }
}

/* Back Platform Link */
.back-platform-link {
  text-align: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #E2E8F0;
}

.back-platform-link a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #94A3B8;
  text-decoration: none;
  transition: color 0.2s;
}

.back-platform-link a:hover {
  color: #007AFF;
}
</style>
