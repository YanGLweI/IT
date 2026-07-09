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
      <p>&copy; IT Department — Public Service Portal</p>
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
