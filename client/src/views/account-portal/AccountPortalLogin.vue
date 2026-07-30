<template>
  <div class="login-page">
    <!-- Background with animated blobs -->
    <div class="login-bg">
      <div class="bg-blob bg-blob-1"></div>
      <div class="bg-blob bg-blob-2"></div>
      <div class="bg-blob bg-blob-3"></div>
    </div>

    <!-- Login Card -->
    <div class="login-content">
      <div class="login-card">
        <!-- Header -->
        <div class="login-header">
          <div class="logo-icon">
            <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
              <rect width="48" height="48" rx="12" fill="#007AFF"/>
              <path d="M24 14C19.58 14 16 17.58 16 22C16 26.42 19.58 30 24 30C28.42 30 32 26.42 32 22C32 17.58 28.42 14 24 14ZM24 27C21.24 27 19 24.76 19 22C19 19.24 21.24 17 24 17C26.76 17 29 19.24 29 22C29 24.76 26.76 27 24 27ZM24 32C19.33 32 10 34.34 10 39V41H38V39C38 34.34 28.67 32 24 32Z" fill="white"/>
            </svg>
          </div>
          <h1 class="login-title">域账号自助管理平台</h1>
          <p class="login-subtitle">请使用域账号登录系统</p>
        </div>

        <!-- Form -->
        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          label-width="0"
          class="login-form"
          @submit.native.prevent="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              ref="usernameInput"
              v-model="loginForm.username"
              placeholder="域账号"
              size="large"
              @keyup.enter.native="handleLogin"
            >
              <i slot="prefix" class="el-icon-user"></i>
            </el-input>
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="密码"
              size="large"
              show-password
              @keyup.enter.native="handleLogin"
            >
              <i slot="prefix" class="el-icon-lock"></i>
            </el-input>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              class="login-button"
              @click="handleLogin"
            >
              登录
            </el-button>
          </el-form-item>
        </el-form>

        <!-- Footer -->
        <div class="login-footer">
          <p class="footer-text">安全连接 · 企业级加密保护</p>
        </div>

        <!-- Back link -->
        <div class="back-link-wrap">
          <router-link to="/login" class="back-link">
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
import { domainAccountLogin } from '@/api/domain_account'

export default {
  name: 'AccountPortalLogin',
  data() {
    return {
      loginForm: {
        username: '',
        password: ''
      },
      loginRules: {
        username: [{ required: true, message: '请输入域账号', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
      },
      loading: false
    }
  },
  mounted() {
    // 如果已登录，直接跳转到 dashboard
    if (localStorage.getItem('domain_account_token')) {
      this.$router.push('/account/dashboard')
      return
    }
    // 自动聚焦到账号输入框
    this.$nextTick(() => {
      this.$refs.usernameInput && this.$refs.usernameInput.focus()
    })
  },
  methods: {
    handleLogin() {
      this.$refs.loginFormRef.validate(async valid => {
        if (!valid) return
        this.loading = true
        try {
          const res = await domainAccountLogin(this.loginForm.username, this.loginForm.password)
          const data = res.data?.data || res.data

          if (data.status === 'expired') {
            // 密码过期，保存用户名并跳转到修改密码页
            const username = this.loginForm.username || data.username || ''
            sessionStorage.setItem('temp_username', username)
            localStorage.setItem('temp_username', username)
            this.$message.warning('密码已过期，请修改密码')
            this.$router.push('/account/change-password?expired=true')
            return
          }

          if (data.token) {
            // 登录成功
            localStorage.setItem('domain_account_token', data.token)
            localStorage.setItem('domain_account_username', data.username)
            localStorage.setItem('domain_account_expiry', JSON.stringify({
              password_expires_at: data.password_expires_at,
              days_remaining: data.days_remaining,
              password_never_expires: data.password_never_expires
            }))
            this.$message.success('登录成功')
            this.$router.push('/account/dashboard')
          }
        } catch (e) {
          const msg = e.response?.data?.message || '登录失败，请检查账号密码'
          this.$message.error(msg)
        } finally {
          this.loading = false
        }
      })
    }
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

/* Animated Background Blobs */
.login-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  z-index: 0;
}

.bg-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.6;
  animation: float 20s infinite ease-in-out;
  min-width: 40vmax;
  min-height: 40vmax;
}

.bg-blob-1 {
  width: 60vmax;
  height: 60vmax;
  background: rgba(0, 122, 255, 0.4);
  top: -20vmax;
  left: -20vmax;
  animation-delay: 0s;
}

.bg-blob-2 {
  width: 50vmax;
  height: 50vmax;
  background: rgba(118, 75, 162, 0.4);
  bottom: -15vmax;
  right: -15vmax;
  animation-delay: -7s;
}

.bg-blob-3 {
  width: 45vmax;
  height: 45vmax;
  background: rgba(52, 199, 89, 0.3);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -14s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(3vmax, -3vmax) scale(1.05);
  }
  66% {
    transform: translate(-2vmax, 2vmax) scale(0.95);
  }
}

/* Content */
.login-content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  padding: 24px;
}

.login-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 48px 40px 36px;
  box-shadow:
    0 20px 60px rgba(0, 0, 0, 0.15),
    0 8px 20px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  animation: cardFadeIn 0.6s ease-out;
}

@keyframes cardFadeIn {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* Header */
.login-header {
  text-align: center;
  margin-bottom: 36px;
}

.logo-icon {
  display: inline-flex;
  margin-bottom: 18px;
  animation: fadeInDown 0.6s ease-out;
}

.login-title {
  font-size: 24px;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 8px 0;
  letter-spacing: -0.5px;
  animation: fadeInDown 0.6s ease-out 0.1s both;
}

.login-subtitle {
  font-size: 14px;
  color: #64748B;
  margin: 0;
  animation: fadeInDown 0.6s ease-out 0.2s both;
}

/* Form */
.login-form {
  animation: fadeInUp 0.6s ease-out 0.3s both;
}

.login-form ::v-deep .el-form-item {
  margin-bottom: 24px;
}

.login-form ::v-deep .el-input__inner {
  height: 48px;
  line-height: 48px;
  padding: 0 16px 0 42px;
  background: #F8FAFC;
  border-radius: 10px;
  border: 1px solid transparent;
  font-size: 14px;
  transition: all 0.25s ease;
}

.login-form ::v-deep .el-input__inner::placeholder {
  color: #94A3B8;
}

.login-form ::v-deep .el-input__inner:hover {
  background: #fff;
  border-color: #E2E8F0;
}

.login-form ::v-deep .el-input__inner:focus {
  background: #fff;
  border-color: #007AFF;
  box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.12) !important;
}

.login-form ::v-deep .el-input__prefix {
  left: 14px;
  color: #007AFF;
  font-size: 18px;
  display: flex;
  align-items: center;
  height: 100%;
}

.login-form ::v-deep .el-input__suffix {
  right: 12px;
  display: flex;
  align-items: center;
  height: 100%;
}

.login-form ::v-deep .el-input__suffix-inner .el-icon--hide,
.login-form ::v-deep .el-input__suffix-inner .el-icon--view {
  color: #94A3B8;
  font-size: 18px;
}

.login-form ::v-deep .el-input__suffix-inner .el-icon--hide:hover,
.login-form ::v-deep .el-input__suffix-inner .el-icon--view:hover {
  color: #64748B;
}

/* Login Button */
.login-button {
  width: 100%;
  height: 48px;
  font-size: 15px;
  font-weight: 600;
  border-radius: 10px !important;
  background: #007AFF !important;
  border: none !important;
  letter-spacing: 2px;
  transition: all 0.3s ease !important;
  margin-top: 8px;
  position: relative;
  overflow: hidden;
}

.login-button:hover {
  background: #0066D6 !important;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 122, 255, 0.35);
}

.login-button:active {
  transform: translateY(0);
}

.login-button::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 60%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
  transition: left 0.5s ease;
}

.login-button:hover::before {
  left: 120%;
}

/* Footer */
.login-footer {
  text-align: center;
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px solid #E2E8F0;
  animation: fadeInUp 0.6s ease-out 0.4s both;
}

.footer-text {
  font-size: 12px;
  color: #94A3B8;
  margin: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.footer-text::before {
  content: '';
  display: inline-block;
  width: 6px;
  height: 6px;
  background: #34C759;
  border-radius: 50%;
}

/* Back Link */
.back-link-wrap {
  text-align: center;
  margin-top: 16px;
  animation: fadeInUp 0.6s ease-out 0.5s both;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #94A3B8;
  text-decoration: none;
  transition: color 0.2s;
}

.back-link:hover {
  color: #007AFF;
}

/* Animations */
@keyframes fadeInDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive */
@media (max-width: 480px) {
  .login-content {
    padding: 16px;
  }

  .login-card {
    padding: 36px 24px 28px;
    border-radius: 16px;
  }

  .login-title {
    font-size: 20px;
  }
}
</style>
