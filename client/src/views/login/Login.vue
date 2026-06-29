<template>
  <div class="login-container">
    <!-- 背景动画 -->
    <div class="login-bg">
      <div class="circuit-lines"></div>
      <div class="floating-particles">
        <span v-for="n in 20" :key="n" :style="particleStyle(n)"></span>
      </div>
    </div>

    <!-- 鼠标轨迹粒子画布 -->
    <canvas ref="trailCanvas" class="trail-canvas"></canvas>

    <!-- 登录卡片 -->
    <div class="login-card">
      <div class="login-header">
        <div class="logo-icon">
          <i class="el-icon-monitor"></i>
        </div>
        <h1 ref="titleText">IT 管理平台</h1>
        <p ref="subtitleText" class="subtitle">IT Management Platform</p>
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
import { animate, stagger, splitText, scrambleText } from 'animejs'

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
      loading: false,
      titleAnimation: null,
      subtitleAnimation: null,
      // 鼠标轨迹粒子相关
      trailParticles: [],
      trailLastX: 0,
      trailLastY: 0,
      trailAnimId: null,
      // 点击爆发范围（持续点击会递增）
      clickBurstRadius: 60,
      clickDecayTimer: null
    }
  },
  mounted() {
    this.$nextTick(() => {
      if (this.$refs.usernameInput) this.$refs.usernameInput.focus()
    })
    this.initTrailCanvas()
    this.initTitleAnimation()
    this.initSubtitleAnimation()
  },
  beforeDestroy() {
    if (this.titleAnimation) {
      this.titleAnimation.pause()
    }
    if (this.subtitleAnimation) {
      this.subtitleAnimation.pause()
    }
    if (this.trailAnimId) {
      cancelAnimationFrame(this.trailAnimId)
    }
    window.removeEventListener('resize', this.resizeTrailCanvas)
    window.removeEventListener('mousemove', this.onMouseMove)
    window.removeEventListener('click', this.onMouseClick)
    if (this.clickDecayTimer) clearTimeout(this.clickDecayTimer)
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
    },
    initTitleAnimation() {
      const el = this.$refs.titleText
      if (!el) return
      const { chars } = splitText(el, { words: false, chars: true })
      this.titleAnimation = animate(chars, {
        y: [
          { to: '-1.5rem', ease: 'outExpo', duration: 500 },
          { to: 0, ease: 'outBounce', duration: 700, delay: 80 }
        ],
        rotate: {
          from: '-1turn',
          delay: 0
        },
        delay: stagger(50),
        ease: 'inOutCirc',
        loopDelay: 1500,
        loop: true
      })
    },
    initSubtitleAnimation() {
      const el = this.$refs.subtitleText
      if (!el) return
      this.subtitleAnimation = animate(el, {
        innerHTML: scrambleText(),
        loop: true,
        loopDelay: 2000
      })
    },
    // ---- 鼠标轨迹粒子效果 ----
    initTrailCanvas() {
      const canvas = this.$refs.trailCanvas
      if (!canvas) return
      this.trailCtx = canvas.getContext('2d')
      this.resizeTrailCanvas()
      window.addEventListener('resize', this.resizeTrailCanvas)
      window.addEventListener('mousemove', this.onMouseMove)
      window.addEventListener('click', this.onMouseClick)
      this.animateTrail()
    },
    resizeTrailCanvas() {
      const canvas = this.$refs.trailCanvas
      if (!canvas) return
      canvas.width = window.innerWidth
      canvas.height = window.innerHeight
    },
    onMouseMove(e) {
      const charList = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ<>{}/_+=[];:.,@#$&'
      const minDistance = 12
      const dx = e.clientX - this.trailLastX
      const dy = e.clientY - this.trailLastY
      const distance = Math.sqrt(dx * dx + dy * dy)
      if (distance > minDistance) {
        // 每次鼠标移动生成一圈 3~5 个字符围绕鼠标
        const count = 3 + Math.floor(Math.random() * 3)
        const radius = 15 + Math.random() * 20
        for (let i = 0; i < count; i++) {
          const angle = (Math.PI * 2 / count) * i + Math.random() * 0.5
          const px = e.clientX + Math.cos(angle) * radius
          const py = e.clientY + Math.sin(angle) * radius
          const text = charList.charAt(Math.floor(Math.random() * charList.length))
          this.trailParticles.push({
            x: px,
            y: py,
            text: text,
            alpha: 0.7 + Math.random() * 0.3,
            fadeSpeed: 0.012 + Math.random() * 0.008,
            size: Math.random() * 4 + 12,
            velocityY: -0.2 - Math.random() * 0.3,
            velocityX: (Math.random() - 0.5) * 0.2
          })
        }
        this.trailLastX = e.clientX
        this.trailLastY = e.clientY
      }
    },
    onMouseClick(e) {
      const charList = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ<>{}/_+=[];:.,@#$&*()'
      const radius = this.clickBurstRadius
      // 在鼠标为圆心、当前半径的圆形区域内随机生成字符
      const count = Math.floor(radius / 3) + 8
      for (let i = 0; i < count; i++) {
        // 极坐标随机：角度随机，半径取 sqrt 使分布更均匀
        const angle = Math.random() * Math.PI * 2
        const r = Math.sqrt(Math.random()) * radius
        const px = e.clientX + Math.cos(angle) * r
        const py = e.clientY + Math.sin(angle) * r
        const text = charList.charAt(Math.floor(Math.random() * charList.length))
        this.trailParticles.push({
          x: px,
          y: py,
          text: text,
          alpha: 0.9 + Math.random() * 0.1,
          fadeSpeed: 0.008 + Math.random() * 0.008,
          size: Math.random() * 5 + 12,
          velocityY: -0.15 - Math.random() * 0.25,
          velocityX: (Math.random() - 0.5) * 0.15
        })
      }
      // 每次点击扩大范围，上限 300px
      this.clickBurstRadius = Math.min(this.clickBurstRadius + 20, 300)
      // 重置衰减计时器：2秒无点击则恢复初始范围
      if (this.clickDecayTimer) clearTimeout(this.clickDecayTimer)
      this.clickDecayTimer = setTimeout(() => {
        this.clickBurstRadius = 60
      }, 2000)
    },
    animateTrail() {
      const ctx = this.trailCtx
      const canvas = this.$refs.trailCanvas
      if (!ctx || !canvas) return
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      for (let i = this.trailParticles.length - 1; i >= 0; i--) {
        const p = this.trailParticles[i]
        ctx.fillStyle = `rgba(0, 200, 255, ${p.alpha})`
        ctx.font = `${p.size}px 'Courier New', monospace`
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        ctx.fillText(p.text, p.x, p.y)
        p.x += (p.velocityX || 0)
        p.y += p.velocityY
        p.alpha -= p.fadeSpeed
        if (p.alpha <= 0) {
          this.trailParticles.splice(i, 1)
        }
      }
      this.trailAnimId = requestAnimationFrame(this.animateTrail)
    }
  }
}
</script>

<style scoped>
.trail-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
  pointer-events: none;
}

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
  z-index: 2;
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
  position: relative;
  width: 100%;
  height: 46px;
  font-size: 16px;
  border-radius: 8px;
  letter-spacing: 4px;
  background: linear-gradient(135deg, #1890ff, #0050b3);
  border: none;
  cursor: pointer;
  overflow: visible;
  transition: all 0.2s ease;
}

.login-btn:active {
  transform: scale(0.96);
}

.login-btn::before,
.login-btn::after {
  position: absolute;
  content: "";
  width: 150%;
  left: 50%;
  height: 100%;
  transform: translateX(-50%);
  z-index: -1;
  background-repeat: no-repeat;
  pointer-events: none;
}

.login-btn:hover::before {
  top: -70%;
  background-image:
    radial-gradient(circle, #a89215 20%, transparent 20%),
    radial-gradient(circle, transparent 20%, #13a5be 20%, transparent 30%),
    radial-gradient(circle, #a3b82d 20%, transparent 20%),
    radial-gradient(circle, #590cbe 20%, transparent 20%),
    radial-gradient(circle, transparent 10%, #bd1717 15%, transparent 20%),
    radial-gradient(circle, #2a7ce8 20%, transparent 20%),
    radial-gradient(circle, #30e82a 20%, transparent 20%),
    radial-gradient(circle, #e92c75 20%, transparent 20%),
    radial-gradient(circle, #914fe7 20%, transparent 20%);
  background-size: 10% 10%, 20% 20%, 15% 15%, 20% 20%, 18% 18%, 10% 10%, 15% 15%, 10% 10%, 18% 18%;
  background-position: 50% 120%;
  animation: topBubbles 0.6s ease;
}

@keyframes topBubbles {
  0% {
    background-position: 5% 90%, 10% 90%, 10% 90%, 15% 90%, 25% 90%, 25% 90%, 40% 90%, 55% 90%, 70% 90%;
  }
  50% {
    background-position: 0% 80%, 0% 20%, 10% 40%, 20% 0%, 30% 30%, 22% 50%, 50% 50%, 65% 20%, 90% 30%;
  }
  100% {
    background-position: 0% 70%, 0% 10%, 10% 30%, 20% -10%, 30% 20%, 22% 40%, 50% 40%, 65% 10%, 90% 20%;
    background-size: 0% 0%, 0% 0%, 0% 0%, 0% 0%, 0% 0%, 0% 0%;
  }
}

.login-btn:hover::after {
  bottom: -70%;
  background-image:
    radial-gradient(circle, #ff93db 20%, transparent 20%),
    radial-gradient(circle, #2ae8df 20%, transparent 20%),
    radial-gradient(circle, transparent 10%, #71ffbd 15%, transparent 20%),
    radial-gradient(circle, #2a9ce8 20%, transparent 20%),
    radial-gradient(circle, #7814fc 20%, transparent 20%),
    radial-gradient(circle, #73e4f8 20%, transparent 20%),
    radial-gradient(circle, #f8d3a9 20%, transparent 20%);
  background-size: 15% 15%, 20% 20%, 18% 18%, 20% 20%, 15% 15%, 20% 20%, 18% 18%;
  background-position: 50% 0%;
  animation: bottomBubbles 0.6s ease;
}

@keyframes bottomBubbles {
  0% {
    background-position: 10% -10%, 30% 10%, 55% -10%, 70% -10%, 85% -10%, 70% -10%, 70% 0%;
  }
  50% {
    background-position: 0% 80%, 20% 80%, 45% 60%, 60% 100%, 75% 70%, 95% 60%, 105% 0%;
  }
  100% {
    background-position: 0% 90%, 20% 90%, 45% 70%, 60% 110%, 75% 80%, 95% 70%, 110% 10%;
    background-size: 0% 0%, 0% 0%, 0% 0%, 0% 0%, 0% 0%, 0% 0%;
  }
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
