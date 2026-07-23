<template>
  <div class="virus-control">
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">病毒控制</h2>
        <p class="page-subtitle">管理病毒防护策略与查杀记录</p>
      </div>
    </div>
    <el-empty description="功能开发中，敬请期待..."></el-empty>
  </div>
</template>

<script>
export default {
  name: 'VirusControl'
}
</script>

<style scoped>
.virus-control {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
  margin: 0;
}
</style>
