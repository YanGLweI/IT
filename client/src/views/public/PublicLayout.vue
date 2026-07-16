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
}

.public-footer {
  flex-shrink: 0;
  padding: 20px 24px;
  text-align: center;
  border-top: 1px solid #E2E8F0;
  background: #fff;
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
</style>
