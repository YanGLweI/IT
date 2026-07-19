<template>
  <div class="placeholder-page">
    <div class="placeholder-card">
      <div class="placeholder-icon">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          <rect x="9" y="10" width="6" height="5" rx="1"/>
          <path d="M10 10V8a2 2 0 0 1 4 0v2"/>
        </svg>
      </div>
      <h2 class="placeholder-title">IPsec VPN</h2>
      <p class="placeholder-desc">该模块正在开发中，敬请期待...</p>
      <div class="placeholder-badge">Coming Soon</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'IPsecVpnPlaceholder'
}
</script>

<style scoped>
.placeholder-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: #f8fafc;
  padding: 24px;
}
.placeholder-card {
  text-align: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 60px 80px;
}
.placeholder-icon {
  color: #94a3b8;
  margin-bottom: 20px;
}
.placeholder-title {
  font-size: 20px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 8px;
}
.placeholder-desc {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 20px;
}
.placeholder-badge {
  display: inline-block;
  padding: 5px 14px;
  background: rgba(59, 130, 246, 0.08);
  color: #3b82f6;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}
</style>
