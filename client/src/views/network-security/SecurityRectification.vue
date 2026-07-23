<template>
  <div class="security-rectification">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">安全整改记录</h2>
        <p class="page-subtitle">跟踪安全漏洞整改进度，记录整改措施与结果</p>
      </div>
    </div>

    <el-empty description="功能开发中..." />
  </div>
</template>

<script>
export default {
  name: 'SecurityRectification'
}
</script>

<style scoped>
.security-rectification {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow-y: auto;
}

/* --- 页面头部 --- */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}
.page-subtitle {
  font-size: 13px;
  color: #64748b;
  margin: 4px 0 0;
}
</style>
