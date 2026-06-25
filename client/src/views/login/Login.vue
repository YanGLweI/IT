<template>
  <div class="login-container">
    <!-- 背景动画 -->
    <div class="login-bg">
      <div class="circuit-lines"></div>
      <div class="floating-particles">
        <span v-for="n in 20" :key="n" :style="particleStyle(n)"></span>
      </div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-card">
      <div class="login-header">
        <div class="logo-icon">
          <i class="el-icon-monitor"></i>
        </div>
        <h1>IT 管理平台</h1>
        <p class="subtitle">IT Management Platform</p>
      </div>

      <el-form :model="loginForm" :rules="loginRules" ref="loginFormRef" class="login-form">
        <el-form-item prop="username">
          <el-input
            ref="usernameInput"
            v-model="loginForm.username"
            placeholder="请输入域账号"
            prefix-icon="el-icon-user"
            @keyup.enter.native="handleLogin"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="el-icon-lock"
            show-password
            @keyup.enter.native="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登 录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <p>&copy; IT Department - LDAP Authentication</p>
      </div>
    </div>
  </div>
</template>

<script>
import { login } from '@/api/auth'

export default {
  name: 'Login',
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
    this.$nextTick(() => {
      if (this.$refs.usernameInput) this.$refs.usernameInput.focus()
    })
  },
  methods: {
    handleLogin() {
      this.$refs.loginFormRef.validate(async valid => {
        if (!valid) return
        this.loading = true
        try {
          const res = await login(this.loginForm.username, this.loginForm.password)
          const data = res.data?.data || res.data
          if (data && data.token) {
            // 存储 token 和用户信息
            localStorage.setItem('token', data.token)
            localStorage.setItem('username', data.username)
            localStorage.setItem('display_name', data.display_name)
            this.$message.success('登录成功')
            this.$router.push('/')
          } else {
            this.$message.error('登录失败')
          }
        } catch (e) {
          const msg = e.response?.data?.message || '登录失败，请检查账号密码'
          this.$message.error(msg)
        } finally {
          this.loading = false
        }
      })
    },
    particleStyle(n) {
      const left = Math.random() * 100
      const delay = Math.random() * 20
      const duration = 15 + Math.random() * 10
      const size = 2 + Math.random() * 4
      return {
        left: `${left}%`,
        animationDelay: `${delay}s`,
        animationDuration: `${duration}s`,
        width: `${size}px`,
        height: `${size}px`
      }
    }
  }
}
</script>

<style scoped>
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #0c1e35 0%, #1a3a5c 50%, #0d2b45 100%);
}

.login-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
}

.circuit-lines {
  position: absolute;
  width: 100%;
  height: 100%;
  background-image:
    linear-gradient(rgba(0, 150, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 150, 255, 0.03) 1px, transparent 1px);
  background-size: 50px 50px;
}

.floating-particles span {
  position: absolute;
  bottom: -10px;
  background: rgba(0, 150, 255, 0.4);
  border-radius: 50%;
  animation: float-up linear infinite;
}

@keyframes float-up {
  0% {
    transform: translateY(0) scale(1);
    opacity: 0;
  }
  10% {
    opacity: 1;
  }
  90% {
    opacity: 1;
  }
  100% {
    transform: translateY(-100vh) scale(0.5);
    opacity: 0;
  }
}

.login-card {
  width: 420px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  padding: 50px 40px 35px;
  box-shadow:
    0 20px 60px rgba(0, 0, 0, 0.3),
    0 0 40px rgba(0, 120, 255, 0.1);
  position: relative;
  z-index: 1;
  backdrop-filter: blur(10px);
}

.login-header {
  text-align: center;
  margin-bottom: 35px;
}

.logo-icon {
  width: 70px;
  height: 70px;
  margin: 0 auto 15px;
  background: linear-gradient(135deg, #1890ff, #0050b3);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 24px rgba(24, 144, 255, 0.3);
}

.logo-icon i {
  font-size: 36px;
  color: #fff;
}

.login-header h1 {
  font-size: 24px;
  color: #1a1a2e;
  margin: 0 0 5px 0;
  font-weight: 600;
  letter-spacing: 2px;
}

.subtitle {
  font-size: 12px;
  color: #8c8c8c;
  letter-spacing: 1px;
  margin: 0;
}

.login-form {
  margin-top: 10px;
}

.login-form .el-input__inner {
  height: 46px;
  line-height: 46px;
  border-radius: 8px;
  font-size: 14px;
}

.login-form .el-input__prefix {
  left: 10px;
}

.login-form .el-input__icon {
  font-size: 18px;
  color: #1890ff;
}

.login-btn {
  width: 100%;
  height: 46px;
  font-size: 16px;
  border-radius: 8px;
  letter-spacing: 4px;
  background: linear-gradient(135deg, #1890ff, #0050b3);
  border: none;
  transition: all 0.3s;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(24, 144, 255, 0.4);
}

.login-footer {
  text-align: center;
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
}

.login-footer p {
  font-size: 11px;
  color: #bfbfbf;
  margin: 0;
  letter-spacing: 0.5px;
}
</style>
