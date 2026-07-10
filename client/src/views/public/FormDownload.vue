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
    <div v-else-if="filteredItems.length > 0" class="form-grid">
      <div
        v-for="item in filteredItems"
        :key="item.id"
        class="form-card"
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
            :href="'/api/public/forms/' + item.id + '/download'"
            class="download-link"
            @click.stop
          >
            <i class="el-icon-download"></i>
            下载
          </a>
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

    <!-- 预览弹窗 -->
    <el-dialog
      :visible.sync="previewVisible"
      width="80%"
      top="5vh"
      @closed="clearPreview"
      v-loading="previewLoading"
    >
      <div class="preview-toolbar" slot="title">
        <span>文件预览 — {{ previewFileName }}</span>
        <div class="preview-toolbar-right">
          <a :href="'/api/public/forms/' + previewItemId + '/download'" class="preview-download-link">
            <el-button type="primary" size="small" icon="el-icon-download">下载</el-button>
          </a>
        </div>
      </div>
      <!-- PDF -->
      <div v-if="previewType === 'pdf'" style="height: 70vh">
        <iframe v-if="previewUrl" :src="previewUrl" style="width: 100%; height: 100%; border: none;"></iframe>
      </div>
      <!-- 图片 -->
      <div v-else-if="previewType === 'image'" style="text-align: center">
        <img v-if="previewUrl" :src="previewUrl" style="max-width: 100%; max-height: 70vh;" />
      </div>
      <!-- DOCX -->
      <div v-else-if="previewType === 'docx'" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 20px;">
        <div ref="docxContainer" class="docx-preview-container"></div>
      </div>
      <!-- XLSX -->
      <div v-else-if="previewType === 'xlsx'" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 10px;">
        <div v-html="xlsxHtml" class="xlsx-preview-container"></div>
      </div>
      <!-- 不支持 -->
      <div v-else style="text-align: center; padding: 40px;">
        <p>该文件格式不支持在线预览</p>
        <a :href="'/api/public/forms/' + previewItemId + '/download'">
          <el-button type="primary">下载文件</el-button>
        </a>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getPublicForms, getPublicPreviewBlob } from '@/api/public_form'
import { renderAsync } from 'docx-preview'
import * as XLSX from 'xlsx'

export default {
  name: 'FormDownload',
  data() {
    return {
      items: [],
      categories: [],
      keyword: '',
      categoryFilter: '',
      loading: false,
      // 预览相关
      previewVisible: false,
      previewUrl: '',
      previewType: '',
      previewFileName: '',
      previewItemId: null,
      previewLoading: false,
      xlsxHtml: ''
    }
  },
  computed: {
    filteredItems() {
      return this.items
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
      } catch (e) {
        console.error('获取表单列表失败:', e)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.fetchForms()
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
    // ---- 预览 ----
    async handlePreview(item) {
      const fileName = (item.file_name || '').toLowerCase()
      this.previewFileName = item.title || item.file_name || '文件'
      this.previewItemId = item.id
      this.xlsxHtml = ''

      // 检测文件类型
      if (fileName.endsWith('.pdf')) {
        this.previewType = 'pdf'
      } else if (['.png', '.jpg', '.jpeg', '.gif', '.bmp'].some(ext => fileName.endsWith(ext))) {
        this.previewType = 'image'
      } else if (fileName.endsWith('.docx') || fileName.endsWith('.doc')) {
        this.previewType = 'docx'
      } else if (fileName.endsWith('.xlsx') || fileName.endsWith('.xls')) {
        this.previewType = 'xlsx'
      } else {
        this.previewType = 'other'
      }

      this.previewVisible = true
      this.previewLoading = true

      try {
        const res = await getPublicPreviewBlob(item.id)
        const blob = res.data

        if (this.previewUrl) URL.revokeObjectURL(this.previewUrl)
        this.previewUrl = URL.createObjectURL(blob)

        if (this.previewType === 'docx') {
          this.$nextTick(() => this.renderDocx(blob))
        } else if (this.previewType === 'xlsx') {
          await this.renderXlsx(blob)
        }
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('预览失败，请稍后重试')
      } finally {
        this.previewLoading = false
      }
    },
    async renderXlsx(blob) {
      try {
        const arrayBuffer = await blob.arrayBuffer()
        const workbook = XLSX.read(arrayBuffer, { type: 'array' })
        const firstSheet = workbook.Sheets[workbook.SheetNames[0]]
        this.xlsxHtml = XLSX.utils.sheet_to_html(firstSheet, { editable: false })
      } catch (e) {
        console.error('xlsx渲染失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    async renderDocx(blob) {
      try {
        const arrayBuffer = await blob.arrayBuffer()
        const container = this.$refs.docxContainer
        if (container) {
          container.innerHTML = ''
          await renderAsync(arrayBuffer, container)
        }
      } catch (e) {
        console.error('docx渲染失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    clearPreview() {
      if (this.previewUrl) {
        URL.revokeObjectURL(this.previewUrl)
        this.previewUrl = ''
      }
      this.xlsxHtml = ''
      if (this.$refs.docxContainer) {
        this.$refs.docxContainer.innerHTML = ''
      }
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
.preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.preview-toolbar-right {
  display: flex;
  align-items: center;
  margin-right: 30px;
}
.preview-download-link {
  text-decoration: none;
}
.docx-preview-container {
  background: #fff;
}
.docx-preview-container >>> .docx-wrapper {
  background: #fff;
  padding: 0;
  width: 100%;
  min-width: 100%;
  overflow-x: auto;
}
.docx-preview-container >>> .docx {
  width: 100%;
  overflow-x: auto;
}
.docx-preview-container >>> .docx table {
  width: 100% !important;
  table-layout: auto;
}
.docx-preview-container >>> .docx table td,
.docx-preview-container >>> .docx table th {
  word-wrap: break-word;
  overflow-wrap: break-word;
  white-space: normal !important;
  min-width: 40px;
}
.xlsx-preview-container {
  font-size: 13px;
}
.xlsx-preview-container >>> table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.xlsx-preview-container >>> table td,
.xlsx-preview-container >>> table th {
  border: 1px solid #E2E8F0;
  padding: 6px 10px;
  text-align: left;
  white-space: nowrap;
  min-width: 60px;
}
.xlsx-preview-container >>> table th {
  background: #F8FAFC;
  font-weight: 600;
  color: #334155;
  position: sticky;
  top: 0;
  z-index: 1;
}
.xlsx-preview-container >>> table tr:hover td {
  background: #F1F5F9;
}
</style>
