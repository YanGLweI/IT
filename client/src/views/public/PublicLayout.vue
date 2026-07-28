<template>
  <div class="public-layout">
    <header class="public-header">
      <div class="header-inner">
        <div class="header-left">
          <div class="header-logo">
            <svg viewBox="0 0 24 24" width="26" height="26">
              <rect x="2" y="3" width="20" height="14" rx="2" fill="none" stroke="#409EFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
              <line x1="8" y1="21" x2="16" y2="21" stroke="#409EFF" stroke-width="1.8" stroke-linecap="round"/>
              <line x1="12" y1="17" x2="12" y2="21" stroke="#409EFF" stroke-width="1.8" stroke-linecap="round"/>
              <polyline points="7 8 10 11 7 14" fill="none" stroke="#409EFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
              <line x1="13" y1="14" x2="17" y2="14" stroke="#409EFF" stroke-width="1.8" stroke-linecap="round"/>
            </svg>
          </div>
          <span class="header-title">IT管理平台 <span class="header-subtitle-tag">· 公共服务</span></span>
        </div>
        <nav class="header-tabs">
          <router-link to="/public/forms" class="tab-item" active-class="tab-active" exact>表单中心</router-link>
          <router-link to="/public/it-guides" class="tab-item" active-class="tab-active">IT指南</router-link>
        </nav>
        <div class="header-right">
          <router-link to="/login" class="back-login-btn">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/>
              <polyline points="10 17 15 12 10 7"/>
              <line x1="15" y1="12" x2="3" y2="12"/>
            </svg>
            返回登录
          </router-link>
        </div>
      </div>
    </header>

    <main class="public-main">
      <!-- 视差背景（仅 IT 指南详情页） -->
      <div v-if="$route.name === 'PublicITGuideDetail'" class="parallax-bg-layer">
        <div class="parallax-bg-base"></div>
        <!-- 太阳 -->
        <div class="sky-sun"></div>
        <!-- 云朵 -->
        <div class="sky-cloud cloud-1"></div>
        <div class="sky-cloud cloud-2"></div>
        <div class="sky-cloud cloud-3"></div>
        <!-- 山峦层 5：最远，最浅，最慢 -->
        <div class="mountain-layer mountain-5"></div>
        <!-- 山峦层 4 -->
        <div class="mountain-layer mountain-4"></div>
        <!-- 山峦层 3 -->
        <div class="mountain-layer mountain-3"></div>
        <!-- 山峦层 2 -->
        <div class="mountain-layer mountain-2"></div>
        <!-- 山峦层 1：最近，最深，最快 -->
        <div class="mountain-layer mountain-1"></div>
      </div>
      <div class="main-inner">
        <router-view />
      </div>
    </main>

    <footer class="public-footer">
      <p class="footer-text">
        &copy; IT Department — Public Service Portal
        <span class="footer-divider">|</span>
        Designed by Yeunglw
        <a class="github-btn Btn" href="https://yanglwei.github.io/" target="_blank" rel="noopener" title="GitHub">
          <span class="svgContainer">
            <svg fill="white" viewBox="0 0 496 512" height="1.4em"><path d="M165.9 397.4c0 2-2.3 3.6-5.2 3.6-3.3.3-5.6-1.3-5.6-3.6 0-2 2.3-3.6 5.2-3.6 3-.3 5.6 1.3 5.6 3.6zm-31.1-4.5c-.7 2 1.3 4.3 4.3 4.9 2.6 1 5.6 0 6.2-2s-1.3-4.3-4.3-5.2c-2.6-.7-5.5.3-6.2 2.3zm44.2-1.7c-2.9.7-4.9 2.6-4.6 4.9.3 2 2.9 3.3 5.9 2.6 2.9-.7 4.9-2.6 4.6-4.6-.3-1.9-3-3.2-5.9-2.9zM244.8 8C106.1 8 0 113.3 0 252c0 110.9 69.8 205.8 169.5 239.2 12.8 2.3 17.3-5.6 17.3-12.1 0-6.2-.3-40.4-.3-61.4 0 0-70 15-84.7-29.8 0 0-11.4-29.1-27.8-36.6 0 0-22.9-15.7 1.6-15.4 0 0 24.9 2 38.6 25.8 21.9 38.6 58.6 27.5 72.9 20.9 2.3-16 8.8-27.1 16-33.7-55.9-6.2-112.3-14.3-112.3-110.5 0-27.5 7.6-41.3 23.6-58.9-2.6-6.5-11.1-33.3 2.6-67.9 20.9-6.5 69 27 69 27 20-5.6 41.5-8.5 62.8-8.5s42.8 2.9 62.8 8.5c0 0 48.1-33.6 69-27 13.7 34.7 5.2 61.4 2.6 67.9 16 17.7 25.8 31.5 25.8 58.9 0 96.5-58.9 104.2-114.8 110.5 9.2 7.9 17 22.9 17 46.4 0 33.7-.3 75.4-.3 83.6 0 6.5 4.6 14.4 17.3 12.1C428.2 457.8 496 362.9 496 252 496 113.3 383.5 8 244.8 8zM97.2 352.9c-1.3 1-1 3.3.7 5.2 1.6 1.6 3.9 2.3 5.2 1 1.3-1 1-3.3-.7-5.2-1.6-1.6-3.9-2.3-5.2-1zm-10.8-8.1c-.7 1.3.3 2.9 2.3 3.9 1.6 1 3.6.7 4.3-.7.7-1.3-.3-2.9-2.3-3.9-2-.6-3.6-.3-4.3.7zm32.4 35.6c-1.6 1.3-1 4.3 1.3 6.2 2.3 2.3 5.2 2.6 6.5 1 1.3-1.3.7-4.3-1.3-6.2-2.2-2.3-5.2-2.6-6.5-1zm-11.4-14.7c-1.6 1-1.6 3.6 0 5.9 1.6 2.3 4.3 3.3 5.6 2.3 1.6-1.3 1.6-3.9 0-6.2-1.4-2.3-4-3.3-5.6-2z"></path></svg>
          </span>
          <span class="BG"></span>
        </a>
      </p>
    </footer>
  </div>
</template>

<script>
export default {
  name: 'PublicLayout'
}
</script>

<style scoped>
.public-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #F8FAFC;
  font-family: 'Inter', 'Noto Sans SC', -apple-system, BlinkMacSystemFont, sans-serif;
}

.public-header {
  background: #fff;
  border-bottom: 1px solid #E2E8F0;
  height: 60px;
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  position: relative;
  z-index: 1;
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  background: #EFF6FF;
  border-radius: 10px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #1E293B;
  letter-spacing: 0.3px;
}

.header-subtitle-tag {
  font-size: 13px;
  font-weight: 400;
  color: #64748B;
}

.header-tabs {
  display: flex;
  gap: 4px;
  align-items: center;
}

.tab-item {
  padding: 8px 20px;
  font-size: 14px;
  font-weight: 500;
  color: #64748B;
  text-decoration: none;
  border-bottom: 2px solid transparent;
  transition: all 0.25s ease;
  border-radius: 6px 6px 0 0;
}

.tab-item:hover {
  color: #409EFF;
  background: #EFF6FF;
}

.tab-active {
  color: #409EFF;
  border-bottom-color: #409EFF;
}

.back-login-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #64748B;
  text-decoration: none;
  padding: 6px 12px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.back-login-btn:hover {
  color: #409EFF;
  background: #EFF6FF;
}

.public-main {
  flex: 1;
  padding: 32px 0;
}

.main-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  position: relative;
  z-index: 1;
}

.public-footer {
  flex-shrink: 0;
  padding: 20px 24px;
  text-align: center;
  border-top: 1px solid #E2E8F0;
  background: #fff;
  position: relative;
  z-index: 1;
}

.public-footer p {
  margin: 0;
  font-size: 12px;
  color: #94A3B8;
  letter-spacing: 0.5px;
}

.footer-divider {
  margin: 0 8px;
  color: #CBD5E1;
}

.github-btn {
  display: inline-flex;
  align-items: center;
  vertical-align: middle;
  margin-left: 4px;
  text-decoration: none;
}

.github-btn.Btn {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  background-color: transparent;
  position: relative;
  z-index: 0;
  border-radius: 7px;
  cursor: pointer;
  transition: all .3s;
}

.github-btn .svgContainer {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  backdrop-filter: blur(0px);
  border-radius: 7px;
  transition: all .3s;
  border: 1px solid rgba(156, 156, 156, 0.466);
}

.github-btn .BG {
  position: absolute;
  content: "";
  width: 100%;
  height: 100%;
  background: #181818;
  z-index: -1;
  border-radius: 7px;
  pointer-events: none;
  transition: all .3s;
}

.github-btn:hover .BG {
  transform: rotate(35deg);
  transform-origin: bottom;
}

.github-btn:hover .svgContainer {
  background-color: rgba(156, 156, 156, 0.466);
  backdrop-filter: blur(4px);
}

@media (max-width: 768px) {
  .header-inner {
    padding: 0 16px;
  }
  .main-inner {
    padding: 0 16px;
  }
  .public-main {
    padding: 20px 0;
  }
}

/* ========== 视差背景（山峦层叠） ========== */
.parallax-bg-layer {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}

/* 底色：天空渐变 */
.parallax-bg-base {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, #EFF6FF 0%, #DBEAFE 40%, #BFDBFE 100%);
  z-index: 0;
}

/* 太阳 */
.sky-sun {
  position: absolute;
  top: 8%;
  right: 15%;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: radial-gradient(circle, #FDE68A 0%, #FCD34D 60%, rgba(252, 211, 77, 0) 100%);
  box-shadow: 0 0 60px 20px rgba(253, 224, 71, 0.4);
  z-index: 1;
  animation: sunMove 1s linear both;
  animation-timeline: scroll(root);
}

@keyframes sunMove {
  from { transform: translateY(15px); }
  to   { transform: translateY(-15px); }
}

/* 云朵基础样式 */
.sky-cloud {
  position: absolute;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 50px;
  z-index: 2;
}

/* 云朵 1：左上 */
.cloud-1 {
  top: 12%;
  left: 8%;
  width: 120px;
  height: 40px;
  border-radius: 40px;
  animation: cloud1Move 1s linear both;
  animation-timeline: scroll(root);
}

.cloud-1::before,
.cloud-1::after {
  content: '';
  position: absolute;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 50%;
}

.cloud-1::before {
  width: 50px;
  height: 50px;
  top: -20px;
  left: 20px;
}

.cloud-1::after {
  width: 35px;
  height: 35px;
  top: -12px;
  right: 20px;
}

@keyframes cloud1Move {
  from { transform: translateY(10px) translateX(-5px); }
  to   { transform: translateY(-10px) translateX(5px); }
}

/* 云朵 2：中上 */
.cloud-2 {
  top: 18%;
  left: 35%;
  width: 100px;
  height: 32px;
  border-radius: 32px;
  animation: cloud2Move 1s linear both;
  animation-timeline: scroll(root);
}

.cloud-2::before,
.cloud-2::after {
  content: '';
  position: absolute;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 50%;
}

.cloud-2::before {
  width: 40px;
  height: 40px;
  top: -16px;
  left: 15px;
}

.cloud-2::after {
  width: 28px;
  height: 28px;
  top: -10px;
  right: 15px;
}

@keyframes cloud2Move {
  from { transform: translateY(8px) translateX(3px); }
  to   { transform: translateY(-8px) translateX(-3px); }
}

/* 云朵 3：右侧 */
.cloud-3 {
  top: 10%;
  right: 30%;
  width: 90px;
  height: 28px;
  border-radius: 28px;
  animation: cloud3Move 1s linear both;
  animation-timeline: scroll(root);
}

.cloud-3::before,
.cloud-3::after {
  content: '';
  position: absolute;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 50%;
}

.cloud-3::before {
  width: 35px;
  height: 35px;
  top: -14px;
  left: 12px;
}

.cloud-3::after {
  width: 25px;
  height: 25px;
  top: -8px;
  right: 12px;
}

@keyframes cloud3Move {
  from { transform: translateY(12px) translateX(4px); }
  to   { transform: translateY(-12px) translateX(-4px); }
}

/* 山峦层基础样式 */
.mountain-layer {
  position: absolute;
  left: -5%;
  width: 110%;
  will-change: transform;
}

/* 山峦层 5：最远，最浅，最慢 */
.mountain-5 {
  bottom: 15%;
  height: 35%;
  background: rgba(147, 197, 253, 0.4);
  clip-path: polygon(
    0% 100%,
    0% 65%,
    8% 45%,
    18% 55%,
    28% 30%,
    38% 50%,
    48% 20%,
    58% 40%,
    68% 25%,
    78% 45%,
    88% 35%,
    100% 55%,
    100% 100%
  );
  animation: mountain5Move 1s linear both;
  animation-timeline: scroll(root);
}

@keyframes mountain5Move {
  from { transform: translateY(5%); }
  to   { transform: translateY(-5%); }
}

/* 山峦层 4 */
.mountain-4 {
  bottom: 10%;
  height: 35%;
  background: rgba(96, 165, 250, 0.35);
  clip-path: polygon(
    0% 100%,
    0% 70%,
    10% 50%,
    20% 60%,
    32% 35%,
    42% 55%,
    55% 25%,
    65% 45%,
    75% 30%,
    85% 50%,
    95% 40%,
    100% 60%,
    100% 100%
  );
  animation: mountain4Move 1s linear both;
  animation-timeline: scroll(root);
}

@keyframes mountain4Move {
  from { transform: translateY(8%); }
  to   { transform: translateY(-8%); }
}

/* 山峦层 3 */
.mountain-3 {
  bottom: 5%;
  height: 35%;
  background: rgba(59, 130, 246, 0.3);
  clip-path: polygon(
    0% 100%,
    0% 75%,
    12% 55%,
    22% 65%,
    35% 40%,
    45% 60%,
    58% 30%,
    68% 50%,
    80% 35%,
    90% 55%,
    100% 45%,
    100% 100%
  );
  animation: mountain3Move 1s linear both;
  animation-timeline: scroll(root);
}

@keyframes mountain3Move {
  from { transform: translateY(12%); }
  to   { transform: translateY(-12%); }
}

/* 山峦层 2 */
.mountain-2 {
  bottom: 0%;
  height: 32%;
  background: rgba(37, 99, 235, 0.25);
  clip-path: polygon(
    0% 100%,
    0% 80%,
    15% 60%,
    25% 70%,
    38% 45%,
    48% 65%,
    60% 35%,
    72% 55%,
    82% 40%,
    92% 60%,
    100% 50%,
    100% 100%
  );
  animation: mountain2Move 1s linear both;
  animation-timeline: scroll(root);
}

@keyframes mountain2Move {
  from { transform: translateY(16%); }
  to   { transform: translateY(-16%); }
}

/* 山峦层 1：最近，最深，最快 */
.mountain-1 {
  bottom: -5%;
  height: 30%;
  background: rgba(29, 78, 216, 0.2);
  clip-path: polygon(
    0% 100%,
    0% 85%,
    18% 65%,
    28% 75%,
    42% 50%,
    52% 70%,
    65% 40%,
    75% 60%,
    85% 45%,
    95% 65%,
    100% 55%,
    100% 100%
  );
  animation: mountain1Move 1s linear both;
  animation-timeline: scroll(root);
}

@keyframes mountain1Move {
  from { transform: translateY(20%); }
  to   { transform: translateY(-20%); }
}

/* 无障碍降级 */
@media (prefers-reduced-motion: reduce) {
  .mountain-layer,
  .sky-sun,
  .sky-cloud {
    animation: none !important;
    transform: none !important;
  }
}

/* 响应式适配 */
@media (max-width: 640px) {
  .mountain-layer {
    left: -10%;
    width: 120%;
  }
}
</style>
