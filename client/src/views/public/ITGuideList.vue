<template>
  <div class="it-guide-public-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="#409EFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
          <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
        </svg>
        IT指南中心
      </h1>
      <p class="page-desc">浏览IT操作指南，图文视频并茂，助力高效工作</p>
    </div>

    <!-- 搜索和筛选 -->
    <div class="filter-bar">
      <div class="search-wrap">
        <el-input v-model="keyword" placeholder="搜索指南名称..." prefix-icon="el-icon-search" clearable @input="handleSearch" />
      </div>
      <el-select v-model="typeFilter" placeholder="全部类型" clearable @change="handleSearch" class="type-select">
        <el-option label="步骤指南" value="step" />
        <el-option label="视频指南" value="video" />
      </el-select>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-wrap">
      <el-skeleton :rows="4" animated />
      <el-skeleton :rows="4" animated />
      <el-skeleton :rows="4" animated />
    </div>

    <!-- 卡片网格 -->
    <div v-else-if="items.length" class="guide-grid">
      <div v-for="item in items" :key="item.id" class="guide-card" @click="goDetail(item)">
        <!-- 图标 -->
        <div class="card-icon" :class="item.guide_type === 'step' ? 'icon-step' : 'icon-video'">
          <svg v-if="item.guide_type === 'step'" viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
            <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
            <line x1="8" y1="7" x2="16" y2="7"/>
            <line x1="8" y1="11" x2="14" y2="11"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
        </div>

        <!-- 卡片内容 -->
        <div class="card-body">
          <h3 class="card-title" :title="item.title">{{ item.title }}</h3>
          <p class="card-desc" :title="item.description">{{ item.description || '暂无描述' }}</p>
          <div class="card-meta">
            <span class="meta-tag" :class="item.guide_type === 'step' ? 'tag-step' : 'tag-video'">
              {{ item.guide_type === 'step' ? '步骤指南' : '视频指南' }}
            </span>
            <span class="meta-tag tag-category" v-if="item.category">{{ item.category }}</span>
          </div>
        </div>

        <!-- 查看详情 -->
        <div class="card-footer">
          <span class="view-link">
            查看详情
            <i class="el-icon-arrow-right"></i>
          </span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="#CBD5E1" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
      </svg>
      <p class="empty-title">暂无IT指南</p>
      <p class="empty-desc">管理员尚未发布任何指南，请稍后再来查看</p>
    </div>
  </div>
</template>

<script>
import { getPublicITGuides } from '@/api/public_form'

export default {
  name: 'PublicITGuideList',
  data() {
    return {
      items: [],
      keyword: '',
      typeFilter: '',
      loading: false
    }
  },
  created() {
    this.fetchGuides()
  },
  methods: {
    async fetchGuides() {
      this.loading = true
      try {
        const params = {}
        if (this.keyword) params.keyword = this.keyword
        if (this.typeFilter) params.guide_type = this.typeFilter
        const res = await getPublicITGuides(params)
        this.items = res.data.data || []
      } catch (e) {
        console.error('获取IT指南列表失败:', e)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.fetchGuides()
    },
    goDetail(item) {
      this.$router.push(`/public/it-guides/${item.id}`)
    }
  }
}
</script>

<style scoped>
.it-guide-public-page {
  max-width: 1200px;
}

/* 页面标题 */
.page-header {
  margin-bottom: 28px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 24px;
  font-weight: 700;
  color: #1E293B;
  margin: 0 0 8px 0;
}

.page-desc {
  font-size: 14px;
  color: #64748B;
  margin: 0;
}

/* 搜索筛选 */
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 28px;
  flex-wrap: wrap;
}

.search-wrap {
  flex: 1;
  min-width: 200px;
}

.search-wrap ::v-deep .el-input__inner {
  border-radius: 8px;
  height: 40px;
  line-height: 40px;
}

.type-select {
  width: 160px;
}

.type-select ::v-deep .el-input__inner {
  border-radius: 8px;
  height: 40px;
  line-height: 40px;
}

/* 加载中 */
.loading-wrap {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

/* 卡片网格 */
.guide-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.guide-card {
  background: #fff;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  transition: all 0.25s ease;
  cursor: pointer;
}

.guide-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08), 0 2px 8px rgba(64, 158, 255, 0.06);
  border-color: #BFDBFE;
}

/* 卡片图标 */
.card-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 14px;
}

.icon-step {
  background: #EFF6FF;
  color: #2563EB;
}

.icon-video {
  background: #FFF7ED;
  color: #D97706;
}

/* 卡片内容 */
.card-body {
  flex: 1;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-desc {
  font-size: 13px;
  color: #64748B;
  margin: 0 0 12px 0;
  line-height: 1.5;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  min-height: 40px;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.meta-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.tag-step {
  background: #EFF6FF;
  color: #2563EB;
}

.tag-video {
  background: #FFF7ED;
  color: #D97706;
}

.tag-category {
  background: #F1F5F9;
  color: #64748B;
}

/* 查看详情 */
.card-footer {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #F1F5F9;
}

.view-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;
  color: #409EFF;
  padding: 6px 14px;
  border-radius: 6px;
  background: #EFF6FF;
  transition: all 0.25s ease;
}

.guide-card:hover .view-link {
  background: #DBEAFE;
  color: #2563EB;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 80px 20px;
}

.empty-state svg {
  margin-bottom: 16px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: #64748B;
  margin: 0 0 6px 0;
}

.empty-desc {
  font-size: 13px;
  color: #94A3B8;
  margin: 0;
}

@media (max-width: 640px) {
  .guide-grid {
    grid-template-columns: 1fr;
  }
  .page-title {
    font-size: 20px;
  }
}
</style>
