<template>
  <div class="login-container">
    <!-- 背景层 -->
    <div class="login-bg">
      <div class="grid-overlay"></div>
      <div class="code-scroll"></div>
      <div class="floating-particles">
        <span v-for="n in 18" :key="n" :style="particleStyle(n)"></span>
      </div>
    </div>

    <!-- 鼠标轨迹粒子画布 -->
    <canvas ref="trailCanvas" class="trail-canvas"></canvas>

    <!-- 登录卡片 -->
    <div class="login-card">
      <!-- 头部 -->
      <div class="login-header">
        <div class="logo-wrap">
          <div class="logo-box">
            <svg viewBox="0 0 24 24" width="34" height="34">
              <rect x="2" y="3" width="20" height="14" rx="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
              <polyline points="7 8 10 11 7 14"/>
              <line x1="13" y1="14" x2="17" y2="14"/>
            </svg>
          </div>
        </div>
        <h1 class="login-title">
          <span class="title-prefix">&gt;</span>
          IT 管理平台
          <span class="title-cursor"></span>
        </h1>
        <p ref="subtitleText" class="login-subtitle">IT Management Platform</p>
      </div>

      <!-- 状态指示 -->
      <div class="status-bar">
        <span class="status-dot"></span>
        <span class="status-text">SYSTEM ONLINE</span>
      </div>

      <!-- 表单 -->
      <el-form :model="loginForm" :rules="loginRules" ref="loginFormRef" class="login-form">
        <el-form-item prop="username" class="form-group">
          <div class="input-wrap">
            <span class="input-icon">
              <svg viewBox="0 0 24 24"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            </span>
            <el-input
              ref="usernameInput"
              v-model="loginForm.username"
              placeholder="请输入域账号"
              class="dark-input"
              @keyup.enter.native="handleLogin"
            />
          </div>
        </el-form-item>
        <el-form-item prop="password" class="form-group">
          <div class="input-wrap">
            <span class="input-icon">
              <svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            </span>
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              show-password
              class="dark-input"
              @keyup.enter.native="handleLogin"
            />
          </div>
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

      <!-- 底部 -->
      <div class="login-footer">
        <p>
          <svg class="lock-icon" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          &copy; IT Department — LDAP Authentication
        </p>
      </div>
    </div>
  </div>
</template>

<script>
import { login } from '@/api/auth'
import { animate, scrambleText } from 'animejs'

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
      subtitleAnimation: null,
      trailParticles: [],
      trailLastX: 0,
      trailLastY: 0,
      trailAnimId: null,
      clickBurstRadius: 60,
      clickDecayTimer: null
    }
  },
  mounted() {
    this.$nextTick(() => {
      if (this.$refs.usernameInput && this.$refs.usernameInput.$el) {
        const input = this.$refs.usernameInput.$el.querySelector('input')
        if (input) input.focus()
      }
    })
    this.initTrailCanvas()
    this.initSubtitleAnimation()
  },
  beforeDestroy() {
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
      const duration = 15 + Math.random() * 12
      const size = 3 + Math.random() * 5
      return {
        left: `${left}%`,
        animationDelay: `${delay}s`,
        animationDuration: `${duration}s`,
        width: `${size}px`,
        height: `${size}px`
      }
    },
    // ---- 副标题 scrambleText 动画 ----
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
        const count = 2 + Math.floor(Math.random() * 3)
        const radius = 12 + Math.random() * 18
        for (let i = 0; i < count; i++) {
          const angle = (Math.PI * 2 / count) * i + Math.random() * 0.5
          const px = e.clientX + Math.cos(angle) * radius
          const py = e.clientY + Math.sin(angle) * radius
          const text = charList.charAt(Math.floor(Math.random() * charList.length))
          this.trailParticles.push({
            x: px, y: py, text: text,
            alpha: 0.8 + Math.random() * 0.2,
            fadeSpeed: 0.008 + Math.random() * 0.006,
            size: 12 + Math.random() * 5,
            velocityY: -0.3 - Math.random() * 0.4,
            velocityX: (Math.random() - 0.5) * 0.3,
            hue: Math.random() > 0.5 ? 210 : 270
          })
        }
        this.trailLastX = e.clientX
        this.trailLastY = e.clientY
      }
    },
    onMouseClick(e) {
      const charList = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ<>{}/_+=[];:.,@#$&*()'
      const radius = this.clickBurstRadius
      const count = Math.floor(radius / 3) + 8
      for (let i = 0; i < count; i++) {
        const angle = Math.random() * Math.PI * 2
        const r = Math.sqrt(Math.random()) * radius
        const px = e.clientX + Math.cos(angle) * r
        const py = e.clientY + Math.sin(angle) * r
        const text = charList.charAt(Math.floor(Math.random() * charList.length))
        this.trailParticles.push({
          x: px, y: py, text: text,
          alpha: 0.9 + Math.random() * 0.1,
          fadeSpeed: 0.006 + Math.random() * 0.006,
          size: 12 + Math.random() * 6,
          velocityY: -0.2 - Math.random() * 0.3,
          velocityX: (Math.random() - 0.5) * 0.2,
          hue: Math.random() > 0.5 ? 210 : 270
        })
      }
      this.clickBurstRadius = Math.min(this.clickBurstRadius + 20, 300)
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
        ctx.fillStyle = `hsla(${p.hue}, 80%, 65%, ${p.alpha})`
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
/* ===== 容器 ===== */
.login-container {
  width: 100%; height: 100vh;
  display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden;
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #0F172A 100%);
}

/* ===== 背景层 ===== */
.login-bg {
  position: absolute; inset: 0; z-index: 0;
}

.grid-overlay {
  position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(64, 158, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(64, 158, 255, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
}

.code-scroll {
  position: absolute; inset: 0;
  overflow: hidden; pointer-events: none;
}
.code-scroll::before {
  content: 'function deploy() { return system.init(); }  const config = { mode: "production", secure: true };  if (status === "active") { console.log("system online"); }  class Server extends Node { constructor() { super(); this.port = 8080; } }  import { Router, Controller } from "@core";  export default defineConfig({ plugins: [vue()] });  SELECT * FROM assets WHERE status = \'active\' ORDER BY created_at DESC;  app.use(cors()); app.listen(3000, () => console.log("running")); const token = getAuthToken(); docker run -d --name gateway -p 80:80 nginx:alpine; git pull origin main && npm run build; const store = createPinia(); app.use(store); ';
  position: absolute; inset: 0;
  color: rgba(211, 219, 230, 0.05);
  font-family: 'Maple Mono NF', 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.8;
  white-space: pre;
  animation: codeDrift 60s linear infinite;
}
@keyframes codeDrift {
  0%   { transform: translateX(0) translateY(0); }
  100% { transform: translateX(-30%) translateY(-20px); }
}

.floating-particles span {
  position: absolute; bottom: -10px;
  background: rgba(64, 158, 255, 0.5);
  border-radius: 50%;
  box-shadow: 0 0 8px rgba(64, 158, 255, 0.4), 0 0 15px rgba(64, 158, 255, 0.2);
  animation: floatUp linear infinite;
}
@keyframes floatUp {
  0%   { transform: translateY(0) scale(1); opacity: 0; }
  10%  { opacity: 1; }
  90%  { opacity: 0.8; }
  100% { transform: translateY(-100vh) scale(0.4); opacity: 0; }
}

/* ===== 轨迹画布 ===== */
.trail-canvas {
  position: absolute; inset: 0; z-index: 1;
  pointer-events: none;
}

/* ===== 登录卡片 ===== */
.login-card {
  width: 460px;
  position: relative; z-index: 2;
  background: rgba(15, 23, 42, 0.82);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px;
  padding: 48px 40px 36px;
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.5),
    0 0 60px rgba(64, 158, 255, 0.06),
    inset 0 1px 0 rgba(255, 255, 255, 0.04);
  animation: cardFadeIn 0.8s ease-out;
}
@keyframes cardFadeIn {
  from { opacity: 0; transform: translateY(20px) scale(0.98); }
  to   { opacity: 1; transform: translateY(0) scale(1); }
}

/* 顶部光效线 */
.login-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 2px;
  border-radius: 16px 16px 0 0;
  background: linear-gradient(90deg, transparent 0%, #409EFF 20%, #a855f7 50%, #409EFF 80%, transparent 100%);
  background-size: 200% 100%;
  animation: topLineFlow 4s linear infinite;
  overflow: hidden;
}
.login-card::before {
  clip-path: inset(0 round 16px 16px 0 0);
}
@keyframes topLineFlow {
  0%   { background-position: 0% 50%; }
  100% { background-position: 200% 50%; }
}

/* 扫描线纹理 */
.login-card::after {
  content: '';
  position: absolute; inset: 0;
  border-radius: 16px;
  background: repeating-linear-gradient(
    0deg, transparent, transparent 2px,
    rgba(148, 163, 184, 0.008) 2px, rgba(148, 163, 184, 0.008) 4px
  );
  pointer-events: none;
}

/* ===== 头部 ===== */
.login-header {
  text-align: center;
  margin-bottom: 28px;
  position: relative; z-index: 1;
}

.logo-wrap {
  width: 68px; height: 68px;
  margin: 0 auto 18px;
  position: relative;
}
.logo-box {
  width: 100%; height: 100%;
  background: linear-gradient(135deg, #409EFF 0%, #a855f7 100%);
  border-radius: 16px;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 8px 32px rgba(64, 158, 255, 0.3);
  position: relative; z-index: 1;
}
.logo-box svg {
  fill: none; stroke: #fff; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round;
}
.logo-wrap::after {
  content: '';
  position: absolute; inset: -6px;
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.2), rgba(168, 85, 247, 0.2));
  filter: blur(12px);
  animation: breathe 3s ease-in-out infinite;
  z-index: 0;
}
@keyframes breathe {
  0%, 100% { opacity: 0.4; transform: scale(0.95); }
  50%      { opacity: 0.8; transform: scale(1.05); }
}

.login-title {
  font-family: 'Maple Mono NF', 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 22px; font-weight: 600;
  color: #E2E8F0;
  letter-spacing: 1px;
  margin: 0 0 6px 0;
  display: flex; align-items: center; justify-content: center; gap: 8px;
}
.title-prefix {
  color: #409EFF;
  font-weight: 700;
  font-size: 20px;
  text-shadow: 0 0 10px rgba(64, 158, 255, 0.5);
}
.title-cursor {
  display: inline-block;
  width: 2px; height: 18px;
  background: #409EFF;
  vertical-align: middle;
  animation: cursorBlink 1.2s step-end infinite;
  margin-left: 2px;
}
@keyframes cursorBlink {
  0%, 100% { opacity: 1; }
  50%      { opacity: 0; }
}

.login-subtitle {
  font-family: 'Maple Mono NF', 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  color: rgba(148, 163, 184, 0.5);
  letter-spacing: 2px;
  margin: 0;
}

/* ===== 状态栏 ===== */
.status-bar {
  display: flex; align-items: center; justify-content: center; gap: 6px;
  margin-bottom: 28px;
  position: relative; z-index: 1;
}
.status-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 8px rgba(34, 197, 94, 0.5);
  animation: statusPulse 2s ease-in-out infinite;
}
@keyframes statusPulse {
  0%, 100% { opacity: 0.7; }
  50%      { opacity: 1; }
}
.status-text {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.45);
  font-family: 'Maple Mono NF', monospace;
  letter-spacing: 1px;
}

/* ===== 表单 ===== */
.login-form {
  position: relative; z-index: 1;
}
.login-form .form-group {
  margin-bottom: 24px;
}

.input-wrap {
  position: relative;
}
.input-icon {
  position: absolute;
  left: 14px; top: 50%; transform: translateY(-50%);
  width: 18px; height: 18px;
  color: #409EFF;
  pointer-events: none;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.input-icon svg {
  width: 18px; height: 18px;
  fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round;
}

/* Element UI 输入框暗色覆盖 */
.login-form ::v-deep .dark-input .el-input__inner {
  height: 46px;
  line-height: 46px;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 10px;
  color: #E2E8F0;
  font-size: 14px;
  font-family: 'Inter', 'Noto Sans SC', -apple-system, sans-serif;
  padding-left: 42px;
  transition: all 0.25s ease;
}
.login-form ::v-deep .dark-input .el-input__inner::placeholder {
  color: rgba(148, 163, 184, 0.35);
}
.login-form ::v-deep .dark-input .el-input__inner:focus {
  background: rgba(30, 41, 59, 1);
  border-color: #409EFF;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.12), 0 0 20px rgba(64, 158, 255, 0.06);
}

/* 密码眼睛图标 */
.login-form ::v-deep .el-input__suffix {
  right: 12px;
}
.login-form ::v-deep .el-input__suffix-inner .el-icon--hide,
.login-form ::v-deep .el-input__suffix-inner .el-icon--view {
  color: #64748B;
  font-size: 18px;
}
.login-form ::v-deep .el-input__suffix-inner .el-icon--hide:hover,
.login-form ::v-deep .el-input__suffix-inner .el-icon--view:hover {
  color: #94A3B8;
}

/* 表单验证错误 */
.login-form ::v-deep .el-form-item__error {
  color: #f87171;
  font-size: 12px;
  padding-top: 4px;
}

/* ===== 登录按钮 ===== */
.login-btn {
  width: 100%; height: 48px;
  border: none; border-radius: 10px;
  background: linear-gradient(135deg, #409EFF 0%, #a855f7 100%);
  color: #fff;
  font-size: 15px; font-weight: 600;
  font-family: 'Inter', 'Noto Sans SC', sans-serif;
  letter-spacing: 4px;
  cursor: pointer;
  position: relative; z-index: 1;
  transition: all 0.3s ease;
  margin-top: 8px;
  overflow: hidden;
}
.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 28px rgba(64, 158, 255, 0.35), 0 4px 16px rgba(168, 85, 247, 0.2);
}
.login-btn:active {
  transform: translateY(0);
}
.login-btn::before {
  content: '';
  position: absolute;
  top: 0; left: -100%; width: 60%; height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.15), transparent);
  transition: left 0.5s ease;
}
.login-btn:hover::before {
  left: 120%;
}

/* ===== 底部 ===== */
.login-footer {
  text-align: center;
  margin-top: 28px;
  padding-top: 18px;
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  position: relative; z-index: 1;
}
.login-footer p {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.3);
  letter-spacing: 0.5px;
  font-family: 'Maple Mono NF', monospace;
  margin: 0;
  display: flex; align-items: center; justify-content: center; gap: 4px;
}
.lock-icon {
  width: 12px; height: 12px;
  fill: none; stroke: rgba(64, 158, 255, 0.35); stroke-width: 2; stroke-linecap: round; stroke-linejoin: round;
}

/* ===== 响应式 ===== */
@media (max-width: 520px) {
  .login-card { width: 92%; padding: 36px 24px 28px; }
}
</style>
