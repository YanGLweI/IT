<template>
  <div class="form-download-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="#409EFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
          <line x1="12" y1="18" x2="12" y2="12"/>
          <line x1="9" y1="15" x2="15" y2="15"/>
        </svg>
        表单下载中心
      </h1>
      <p class="page-desc">浏览并下载已发布的表单文件，无需登录即可使用</p>
    </div>

    <!-- 搜索和筛选 -->
    <div class="filter-bar">
      <div class="search-wrap">
        <el-input
          v-model="keyword"
          placeholder="搜索表单名称或描述..."
          prefix-icon="el-icon-search"
          clearable
          @input="handleSearch"
        />
      </div>
      <el-select
        v-model="categoryFilter"
        placeholder="全部分类"
        clearable
        @change="handleSearch"
        class="category-select"
      >
        <el-option
          v-for="cat in categories"
          :key="cat"
          :label="cat"
          :value="cat"
        />
      </el-select>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-wrap">
      <el-skeleton :rows="4" animated />
      <el-skeleton :rows="4" animated />
      <el-skeleton :rows="4" animated />
    </div>

    <!-- 表单卡片网格 -->
    <div v-else-if="filteredItems.length > 0" class="form-grid" :key="gridKey">
      <div
        v-for="(item, index) in filteredItems"
        :key="item.id"
        class="form-card card-anim"
        :style="{ animationDelay: index * 0.06 + 's' }"
        @click="handlePreview(item)"
      >
        <!-- 文件图标 -->
        <div class="card-icon" :class="'icon-' + getFileExtClass(item)">
          <svg v-if="getFileExtClass(item) === 'pdf'" viewBox="0 0 24 24" width="36" height="36" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <text x="12" y="17" text-anchor="middle" font-size="7" fill="currentColor" stroke="none" font-weight="bold">PDF</text>
          </svg>
          <svg v-else-if="getFileExtClass(item) === 'doc'" viewBox="0 0 24 24" width="36" height="36" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <text x="12" y="17" text-anchor="middle" font-size="6" fill="currentColor" stroke="none" font-weight="bold">DOC</text>
          </svg>
          <svg v-else viewBox="0 0 24 24" width="36" height="36" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <text x="12" y="17" text-anchor="middle" font-size="6" fill="currentColor" stroke="none" font-weight="bold">XLS</text>
          </svg>
        </div>

        <!-- 卡片内容 -->
        <div class="card-body">
          <h3 class="card-title" :title="item.title">{{ item.title }}</h3>
          <p class="card-desc" v-if="item.description" :title="item.description">{{ item.description }}</p>
          <p class="card-desc" v-else>&nbsp;</p>

          <div class="card-meta">
            <span class="meta-tag" :class="'tag-' + item.source_type">
              {{ getSourceLabel(item.source_type) }}
            </span>
            <span class="meta-ext" v-if="item.file_name">
              {{ getFileExt(item.file_name).toUpperCase() }}
            </span>
          </div>
        </div>

        <!-- 下载按钮 -->
        <div class="card-footer">
          <a
            class="download-link"
            @click.stop="handleDownload(item)"
          >
            <i class="el-icon-download"></i>
            下载
          </a>
          <span class="download-count">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            {{ formatDownloadCount(item.download_count) }}
          </span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="#CBD5E1" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/>
        <polyline points="13 2 13 9 20 9"/>
      </svg>
      <p class="empty-title">暂无可下载的表单</p>
      <p class="empty-desc">管理员尚未发布任何表单，请稍后再来查看</p>
    </div>

    <!-- 预览弹窗（使用 file-viewer，PDF/DOCX/XLSX/图片统一渲染） -->
    <el-dialog
      class="vault-dialog preview-dialog fv-preview-dialog"
      :title="'文件预览 — ' + previewFileName"
      :visible.sync="previewVisible"
      width="80%"
      top="3vh"
      append-to-body
    >
      <div class="fv-container" style="height: 70vh">
        <div v-if="fvFeature.enabled">
          <FileViewer
            v-if="previewVisible"
            :url="currentFileUrl"
            :name="currentFileName"
            :type="currentFileType"
            :options="fvOptions"
          />
        </div>
        <div v-else class="fv-fallback">
          <p>文件预览功能暂不可用，请下载后查看。</p>
          <a :href="publicDownloadUrl">
            <el-button type="primary">下载文件</el-button>
          </a>
        </div>
      </div>
      <span slot="footer">
        <a :href="publicDownloadUrl" class="fv-download-link">
          <el-button type="primary" size="small" icon="el-icon-download">下载</el-button>
        </a>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getPublicForms, getPublicPreviewUrl } from '@/api/public_form'
import { FileViewer } from '@file-viewer/vue2.7'
import officePreset from '@file-viewer/preset-office'
import fvFeature from '@/config/fv-feature'

export default {
  name: 'FormDownload',
  components: { FileViewer },
  data() {
    return {
      items: [],
      categories: [],
      keyword: '',
      categoryFilter: '',
      loading: false,
      gridKey: 0,
      // 预览相关
      previewVisible: false,
      previewFileName: '',
      previewItemId: null,
      currentFileUrl: '',
      currentFileName: '',
      currentFileType: '',
      fvFeature: fvFeature
    }
  },
  computed: {
    filteredItems() {
      return this.items
    },
    fvOptions() {
      return { preset: officePreset, fetchFile: this.fetchFilePublic }
    },
    publicDownloadUrl() {
      return this.previewItemId ? '/api/public/forms/' + this.previewItemId + '/download' : ''
    }
  },
  created() {
    this.fetchForms()
  },
  methods: {
    async fetchForms() {
      this.loading = true
      try {
        const res = await getPublicForms({
          keyword: this.keyword || undefined,
          category: this.categoryFilter || undefined
        })
        // publicRequest 没有响应拦截器，返回完整 axios response
        // res.data = { code, data: [...], categories: [] }
        this.items = res.data.data || []
        this.categories = res.data.categories || []
        this.gridKey++
      } catch (e) {
        console.error('获取表单列表失败:', e)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.fetchForms()
    },
    handleDownload(item) {
      // 创建临时 a 标签触发下载
      const link = document.createElement('a')
      link.href = '/api/public/forms/' + item.id + '/download'
      link.style.display = 'none'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      // 延迟刷新列表以更新下载计数
      setTimeout(() => { this.fetchForms() }, 500)
    },
    getFileExt(fileName) {
      if (!fileName) return ''
      const idx = fileName.lastIndexOf('.')
      return idx > -1 ? fileName.substring(idx + 1) : ''
    },
    getFileExtClass(item) {
      const ext = this.getFileExt(item.file_name).toLowerCase()
      if (ext === 'pdf') return 'pdf'
      if (['doc', 'docx'].includes(ext)) return 'doc'
      return 'xls'
    },
    getSourceLabel(type) {
      const labels = { upload: '文件', static: '模板', dynamic: '动态生成' }
      return labels[type] || type
    },
    formatDownloadCount(n) {
      if (!n) return '0'
      if (n >= 1000) return (n / 1000).toFixed(2) + 'K'
      return String(n)
    },
    // ---- 预览 ----
    async handlePreview(item) {
      const fileName = (item.file_name || '').toLowerCase()
      this.previewFileName = item.title || item.file_name || '文件'
      this.previewItemId = item.id
      this.currentFileName = item.file_name || (item.title || '文件') + '.xlsx'
      // 动态来源由后端生成器实时生成，固定为 xlsx；其余按扩展名识别
      if (item.source_type === 'dynamic') {
        this.currentFileType = 'xlsx'
      } else {
        this.currentFileType = fileName.includes('.') ? fileName.split('.').pop().toLowerCase() : ''
      }
      this.currentFileUrl = getPublicPreviewUrl(item.id)
      this.previewVisible = false
      await this.$nextTick()
      this.previewVisible = true
      await this.$nextTick()
    },
    // 公共接口免登录；若已登录则附带 token（兼容两种场景）
    async fetchFilePublic({ url }) {
      const headers = {}
      const token = localStorage.getItem('token')
      if (token) headers['Authorization'] = `Bearer ${token}`
      const response = await fetch(url, { headers })
      if (!response.ok) {
        throw new Error(`Failed to fetch file: ${response.status} ${response.statusText}`)
      }
      return response.arrayBuffer()
    }
  }
}
</script>

<style scoped>
.form-download-page {
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

.category-select {
  width: 160px;
}

.category-select ::v-deep .el-input__inner {
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
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.form-card {
  background: #fff;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  transition: all 0.25s ease;
  cursor: pointer;
}

.form-card.card-anim {
  opacity: 0;
  animation: cardFadeInUp 0.4s ease-out forwards;
}

@keyframes cardFadeInUp {
  0% {
    opacity: 0;
    transform: translateY(20px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.form-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08), 0 2px 8px rgba(64, 158, 255, 0.06);
  border-color: #BFDBFE;
}

/* 文件图标 */
.card-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 14px;
}

.icon-pdf {
  background: #FEF2F2;
  color: #DC2626;
}

.icon-doc {
  background: #EFF6FF;
  color: #2563EB;
}

.icon-xls {
  background: #F0FDF4;
  color: #16A34A;
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

.tag-upload {
  background: #EFF6FF;
  color: #2563EB;
}

.tag-static {
  background: #F5F3FF;
  color: #7C3AED;
}

.tag-dynamic {
  background: #FFF7ED;
  color: #D97706;
}

.meta-ext {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  background: #F1F5F9;
  color: #64748B;
  font-weight: 500;
  font-family: 'Maple Mono NF', 'SF Mono', monospace;
}

/* 下载按钮 */
.card-footer {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #F1F5F9;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.download-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;
  color: #409EFF;
  text-decoration: none;
  padding: 6px 14px;
  border-radius: 6px;
  background: #EFF6FF;
  transition: all 0.2s ease;
  cursor: pointer;
}

.download-link:hover {
  background: #DBEAFE;
  color: #2563EB;
}

.download-count {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;
  color: #94A3B8;
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
  .form-grid {
    grid-template-columns: 1fr;
  }
  .page-title {
    font-size: 20px;
  }
}

/* 预览弹窗 */
.fv-download-link {
  text-decoration: none;
}

.fv-container {
  background: #f8fafc;
}

/* file-viewer 组件高度修复 - 强制填满容器 */
.fv-container > div {
  height: 100% !important;
}

.fv-container >>> .ff-file-viewer-vue27 {
  height: 100% !important;
}

/* file-viewer 内部工具栏样式覆盖 */
.fv-container >>> .fv-toolbar {
  border-bottom: 1px solid #e2e8f0;
  padding: 8px 12px;
  background: #ffffff;
}

/* 确保滚动正确工作 */
.fv-container >>> .fv-content-wrapper {
  height: calc(100% - 50px);
}

/* 降级方案样式 */
.fv-fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 16px;
  color: #64748b;
}
</style>
