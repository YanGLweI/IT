<template>
  <div class="notfound-container">
    <!-- 像素网格背景 -->
    <div class="pixel-grid-bg"></div>

    <!-- 浮动像素方块装饰 -->
    <div class="floating-pixels">
      <span class="px-block px-blue" style="top:12%;left:8%"></span>
      <span class="px-block px-purple" style="top:20%;right:12%"></span>
      <span class="px-block px-green" style="bottom:28%;left:14%"></span>
      <span class="px-block px-blue" style="bottom:18%;right:8%"></span>
      <span class="px-block px-purple" style="top:45%;left:5%"></span>
      <span class="px-block px-green" style="top:35%;right:6%"></span>
    </div>

    <!-- 主卡片 -->
    <div class="pixel-card">
      <!-- 404 大标题 -->
      <div class="error-code">
        <span class="digit">4</span>
        <span class="digit">0</span>
        <span class="digit">4</span>
      </div>

      <!-- 像素走丢小人插图 -->
      <div class="pixel-person">
        <div class="person-sprite"></div>
        <div class="question-mark">?</div>
      </div>

      <!-- 文案 -->
      <h2 class="error-title">哎呀，这个页面走丢了</h2>
      <p class="error-desc">目标区域尚未解锁，请检查地址或返回安全区域</p>

      <!-- 操作按钮 -->
      <div class="btn-group">
        <button class="pixel-btn pixel-btn-primary" @click="goHome">
          <span class="btn-icon">&#9632;</span> 返回首页
        </button>
        <button class="pixel-btn pixel-btn-secondary" @click="goBack">
          <span class="btn-icon">&#9633;</span> 返回上页
        </button>
      </div>

      <!-- 终端信息区 -->
      <div class="terminal-info">
        <p><span class="prompt">&gt;</span> route: <span class="val">{{ currentPath }}</span></p>
        <p><span class="prompt">&gt;</span> status: <span class="val val-warn">404 NOT FOUND</span></p>
        <p><span class="prompt">&gt;</span> time: <span class="val">{{ timestamp }}</span></p>
        <p class="cursor-line"><span class="prompt">&gt;</span> <span class="pixel-cursor"></span></p>
      </div>
    </div>

    <!-- 底部像素地面 -->
    <div class="pixel-ground"></div>
  </div>
</template>

<script>
export default {
  name: 'NotFound',
  data() {
    return {
      timestamp: ''
    }
  },
  computed: {
    currentPath() {
      return this.$route.path
    }
  },
  created() {
    this.timestamp = new Date().toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false
    })
  },
  methods: {
    goHome() {
      const token = localStorage.getItem('token')
      const target = token ? '/' : '/login'
      this.$router.push(target).catch(() => {})
    },
    goBack() {
      this.$router.go(-1)
    }
  }
}
</script>

<style scoped>
/* ===== 容器 ===== */
.notfound-container {
  width: 100%;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: #F8FAFC;
}

/* ===== 像素网格背景 ===== */
.pixel-grid-bg {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 0, 0, 0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.02) 1px, transparent 1px);
  background-size: 4px 4px;
  pointer-events: none;
}

/* ===== 浮动像素方块 ===== */
.floating-pixels .px-block {
  position: absolute;
  width: 10px;
  height: 10px;
  image-rendering: pixelated;
  animation: pixelFloat 3s steps(4) infinite;
}
.px-blue { background: #409EFF; box-shadow: 2px 2px 0 rgba(64, 158, 255, 0.3); }
.px-purple { background: #a855f7; box-shadow: 2px 2px 0 rgba(168, 85, 247, 0.3); }
.px-green { background: #22C55E; box-shadow: 2px 2px 0 rgba(34, 197, 94, 0.3); }

.floating-pixels .px-block:nth-child(1) { animation-delay: 0s; }
.floating-pixels .px-block:nth-child(2) { animation-delay: 0.5s; }
.floating-pixels .px-block:nth-child(3) { animation-delay: 1s; }
.floating-pixels .px-block:nth-child(4) { animation-delay: 1.5s; }
.floating-pixels .px-block:nth-child(5) { animation-delay: 0.8s; }
.floating-pixels .px-block:nth-child(6) { animation-delay: 1.2s; }

@keyframes pixelFloat {
  0%, 100% { transform: translateY(0); }
  25% { transform: translateY(-6px); }
  50% { transform: translateY(-3px); }
  75% { transform: translateY(-8px); }
}

/* ===== 像素卡片 ===== */
.pixel-card {
  position: relative;
  z-index: 2;
  background: #FFFFFF;
  border: 4px solid #1E293B;
  box-shadow:
    4px 4px 0 0 #1E293B,
    -4px 4px 0 0 #1E293B,
    4px -4px 0 0 #1E293B,
    -4px -4px 0 0 #1E293B;
  padding: 48px 56px 40px;
  text-align: center;
  max-width: 520px;
  width: 90%;
  animation: cardEnter 0.5s steps(6) both;
}

@keyframes cardEnter {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

/* ===== 404 大标题 ===== */
.error-code {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 24px;
}

.error-code .digit {
  font-family: 'Courier New', 'Consolas', monospace;
  font-size: 110px;
  font-weight: 900;
  color: #1E293B;
  line-height: 1;
  text-shadow:
    4px 4px 0 #409EFF,
    8px 8px 0 rgba(168, 85, 247, 0.3);
  animation: glitchShake 3s steps(1) infinite;
}

.error-code .digit:nth-child(2) {
  color: #409EFF;
  text-shadow:
    4px 4px 0 #1E293B,
    8px 8px 0 rgba(34, 197, 94, 0.3);
  animation-delay: 0.1s;
}

@keyframes glitchShake {
  0%, 90%, 100% { transform: translate(0, 0); }
  92% { transform: translate(-3px, 2px); }
  94% { transform: translate(3px, -2px); }
  96% { transform: translate(-2px, -1px); }
  98% { transform: translate(2px, 1px); }
}

/* ===== 像素走丢小人 ===== */
.pixel-person {
  position: relative;
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
  height: 72px;
}

.question-mark {
  position: absolute;
  top: -6px;
  left: 50%;
  transform: translateX(-50%);
  font-family: 'Courier New', monospace;
  font-size: 22px;
  font-weight: 900;
  color: #409EFF;
  animation: questionBounce 1.2s steps(3) infinite;
}

@keyframes questionBounce {
  0%, 100% { transform: translateX(-50%) translateY(0); opacity: 1; }
  50% { transform: translateX(-50%) translateY(-5px); opacity: 0.6; }
}

.person-sprite {
  width: 6px;
  height: 6px;
  background: transparent;
  margin-top: 16px;
  animation: personIdle 1s steps(2) infinite;
  /* 纯CSS box-shadow 绘制走丢的小人 */
  box-shadow:
    /* 头发 */
    12px 0px 0 #4A3728,
    18px 0px 0 #4A3728,
    24px 0px 0 #4A3728,
    30px 0px 0 #4A3728,
    6px 6px 0 #4A3728,
    12px 6px 0 #4A3728,
    18px 6px 0 #4A3728,
    24px 6px 0 #4A3728,
    30px 6px 0 #4A3728,
    36px 6px 0 #4A3728,
    /* 脸 */
    6px 12px 0 #FDDCB5,
    12px 12px 0 #FDDCB5,
    18px 12px 0 #FDDCB5,
    24px 12px 0 #FDDCB5,
    30px 12px 0 #FDDCB5,
    36px 12px 0 #FDDCB5,
    6px 18px 0 #FDDCB5,
    12px 18px 0 #1E293B,
    18px 18px 0 #FDDCB5,
    24px 18px 0 #FDDCB5,
    30px 18px 0 #1E293B,
    36px 18px 0 #FDDCB5,
    6px 24px 0 #FDDCB5,
    12px 24px 0 #FDDCB5,
    18px 24px 0 #FDDCB5,
    24px 24px 0 #FDDCB5,
    30px 24px 0 #FDDCB5,
    36px 24px 0 #FDDCB5,
    /* 嘴巴（困惑的O形） */
    18px 30px 0 #FDDCB5,
    24px 30px 0 #E8A87C,
    30px 30px 0 #FDDCB5,
    /* 身体（T恤） */
    6px 36px 0 #409EFF,
    12px 36px 0 #409EFF,
    18px 36px 0 #409EFF,
    24px 36px 0 #409EFF,
    30px 36px 0 #409EFF,
    36px 36px 0 #409EFF,
    /* 手臂 + 身体 */
    0px 42px 0 #FDDCB5,
    6px 42px 0 #409EFF,
    12px 42px 0 #409EFF,
    18px 42px 0 #409EFF,
    24px 42px 0 #409EFF,
    30px 42px 0 #409EFF,
    36px 42px 0 #409EFF,
    42px 42px 0 #FDDCB5,
    0px 48px 0 #FDDCB5,
    6px 48px 0 #409EFF,
    12px 48px 0 #409EFF,
    18px 48px 0 #409EFF,
    24px 48px 0 #409EFF,
    30px 48px 0 #409EFF,
    36px 48px 0 #409EFF,
    42px 48px 0 #FDDCB5,
    /* 裤子 */
    12px 54px 0 #334155,
    18px 54px 0 #334155,
    24px 54px 0 #334155,
    30px 54px 0 #334155,
    /* 腿 */
    12px 60px 0 #334155,
    18px 60px 0 #FDDCB5,
    24px 60px 0 #FDDCB5,
    30px 60px 0 #334155,
    /* 鞋子 */
    6px 66px 0 #EF4444,
    12px 66px 0 #EF4444,
    30px 66px 0 #EF4444,
    36px 66px 0 #EF4444;
}

@keyframes personIdle {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

/* ===== 文案 ===== */
.error-title {
  font-size: 22px;
  font-weight: 700;
  color: #1E293B;
  margin: 0 0 8px;
  font-family: 'Courier New', monospace;
}

.error-desc {
  font-size: 14px;
  color: #64748B;
  margin: 0 0 32px;
  line-height: 1.6;
}

/* ===== 像素按钮 ===== */
.btn-group {
  display: flex;
  gap: 16px;
  justify-content: center;
  margin-bottom: 32px;
  flex-wrap: wrap;
}

.pixel-btn {
  font-family: 'Courier New', monospace;
  font-size: 14px;
  font-weight: 700;
  padding: 12px 24px;
  cursor: pointer;
  border: 3px solid #1E293B;
  box-shadow: 0 4px 0 0 #1E293B;
  transition: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  user-select: none;
}

.pixel-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 0 0 #1E293B;
}

.pixel-btn:active {
  transform: translateY(2px);
  box-shadow: 0 2px 0 0 #1E293B;
}

.pixel-btn-primary {
  background: #409EFF;
  color: #FFFFFF;
}
.pixel-btn-primary:hover {
  background: #66B1FF;
}

.pixel-btn-secondary {
  background: #FFFFFF;
  color: #1E293B;
}
.pixel-btn-secondary:hover {
  background: #F1F5F9;
}

.btn-icon {
  font-size: 10px;
}

/* ===== 终端信息区 ===== */
.terminal-info {
  background: #F1F5F9;
  border: 2px solid #E2E8F0;
  padding: 16px 20px;
  text-align: left;
  font-family: 'Courier New', 'Consolas', monospace;
  font-size: 12px;
  line-height: 2;
  color: #64748B;
}

.terminal-info p {
  margin: 0;
}

.terminal-info .prompt {
  color: #22C55E;
  font-weight: 700;
  margin-right: 6px;
}

.terminal-info .val {
  color: #1E293B;
}

.terminal-info .val-warn {
  color: #EF4444;
  font-weight: 700;
}

.cursor-line {
  display: flex;
  align-items: center;
}

.pixel-cursor {
  display: inline-block;
  width: 8px;
  height: 14px;
  background: #1E293B;
  margin-left: 4px;
  animation: cursorBlink 1s steps(1) infinite;
}

@keyframes cursorBlink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* ===== 底部像素地面 ===== */
.pixel-ground {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 24px;
  background:
    repeating-conic-gradient(#E2E8F0 0% 25%, #F1F5F9 0% 50%) 0 0 / 24px 24px;
  border-top: 4px solid #1E293B;
}

/* ===== 响应式 ===== */
@media (max-width: 640px) {
  .pixel-card {
    padding: 32px 24px 28px;
    width: 92%;
  }

  .error-code .digit {
    font-size: 64px;
    text-shadow:
      3px 3px 0 #409EFF,
      6px 6px 0 rgba(168, 85, 247, 0.3);
  }

  .error-code .digit:nth-child(2) {
    text-shadow:
      3px 3px 0 #1E293B,
      6px 6px 0 rgba(34, 197, 94, 0.3);
  }

  .person-sprite {
    transform: scale(0.85);
  }

  @keyframes personIdle {
    0%, 100% { transform: scale(0.85) translateY(0); }
    50% { transform: scale(0.85) translateY(-3px); }
  }

  .btn-group {
    flex-direction: column;
    align-items: center;
  }

  .pixel-btn {
    width: 100%;
    justify-content: center;
  }

  .terminal-info {
    font-size: 11px;
    padding: 12px 14px;
  }
}
</style>
