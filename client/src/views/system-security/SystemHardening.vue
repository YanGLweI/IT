<template>
  <div class="system-hardening">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">系统加固</h2>
        <p class="page-subtitle">管理系统加固检查记录，跟踪季度加固执行情况</p>
      </div>
      <div class="header-actions">
        <el-button type="success" size="small" icon="el-icon-download" @click="handleExportChecklist">导出检查表</el-button>
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="showUpload = true">上传记录</el-button>
        <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
      </div>
    </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="全部年份" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="y + '年'" :value="y" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索描述..." size="small" clearable @keyup.enter.native="handleFilterChange" @clear="handleFilterChange" style="width: 200px" />
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleFilterChange">搜索</el-button>
      </div>

      <!-- 数据表格 -->
      <div class="table-card" ref="tableCard" style="margin-top: 12px">
        <div class="table-wrapper">
        <el-table :data="records" stripe v-loading="loading" :max-height="tableMaxHeight">
          <el-table-column type="index" label="#" width="70" align="center" />
          <el-table-column prop="year" label="年份" width="85" align="center" />
          <el-table-column label="季度" width="85" align="center">
            <template slot-scope="{ row }">Q{{ row.quarter }}</template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
          <el-table-column label="文件大小" width="100" align="center">
            <template slot-scope="{ row }">{{ formatSize(row.file_size) }}</template>
          </el-table-column>
          <el-table-column label="上传时间" width="180" align="center">
            <template slot-scope="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="350" fixed="right" align="center">
            <template slot-scope="{ row }">
              <div class="op-btns">
                <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreview(row)">预览</el-button>
                <el-button size="mini" type="text" icon="el-icon-download" @click="handleDownload(row)">下载</el-button>
                <el-button size="mini" type="text" icon="el-icon-edit" @click="handleEdit(row)">编辑</el-button>
                <el-button size="mini" type="danger" icon="el-icon-delete"  @click="handleDelete(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
        </div>
      </div>
      

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size.sync="pageSize"
        :current-page.sync="page"
        :page-sizes="[10, 20, 50]"
        @size-change="handleSizeChange"
        @current-change="fetchData"
      />
    </div>

    <!-- 上传/编辑弹窗 -->
    <el-dialog class="vault-dialog" :title="isEdit ? '编辑系统加固检查记录' : '上传系统加固检查记录'" :visible.sync="showUpload" width="520px" :close-on-click-modal="false">
      <el-form :model="uploadForm" ref="uploadFormRef" :rules="uploadRules" label-width="80px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="年份" prop="year">
              <el-input-number v-model="uploadForm.year" :min="2020" :max="2100" :step="1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="季度" prop="quarter">
              <el-select v-model="uploadForm.quarter" placeholder="请选择" style="width: 100%">
                <el-option v-for="q in 4" :key="q" :label="'Q' + q" :value="q" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述" prop="description">
          <el-input v-model="uploadForm.description" type="textarea" :rows="2" placeholder="输入描述，便于后续搜索" />
        </el-form-item>
        <el-form-item label="文件" prop="file" v-if="!isEdit">
          <el-upload
            ref="uploader"
            action=""
            :auto-upload="false"
            :limit="1"
            accept=".pdf"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            drag
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
            <div slot="tip" class="el-upload__tip">仅支持 PDF 格式文件</div>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可修改文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showUpload = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">{{ isEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog class="vault-dialog preview-dialog" title="文件预览" :visible.sync="previewVisible" width="80%" top="3vh" :close-on-click-modal="true">
      <iframe v-if="previewUrl" :src="previewUrl" style="width: 100%; height: 70vh; border: none;" />
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getSystemHardeningHistories,
  createSystemHardeningHistory,
  updateSystemHardeningHistory,
  deleteSystemHardeningHistory,
  getExportChecklistUrl,
  getSystemHardeningPreviewUrl,
  getSystemHardeningDownloadUrl
} from '@/api/system_hardening'
import DualControlDialog from '@/components/DualControlDialog.vue'
import tableHeightMixin from '@/mixins/table-height'

export default {
  name: 'SystemHardening',
  components: { DualControlDialog },
  mixins: [tableHeightMixin],
  data() {
    const now = new Date()
    return {
      records: [],
      loading: false,
      page: 1,
      pageSize: 10,
      total: 0,
      filterYear: '',
      keyword: '',
      yearOptions: Array.from({ length: 10 }, (_, i) => now.getFullYear() - i),
      // 上传
      showUpload: false,
      isEdit: false,
      editingId: null,
      uploading: false,
      uploadForm: {
        year: now.getFullYear(),
        quarter: Math.ceil((now.getMonth() + 1) / 3),
        description: ''
      },
      uploadRules: {
        year: [{ required: true, message: '请选择年份', trigger: 'change' }],
        quarter: [{ required: true, message: '请选择季度', trigger: 'change' }]
      },
      selectedFile: null,
      fileList: [],
      // 预览
      previewVisible: false,
      previewUrl: ''
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const params = { page: this.page, page_size: this.pageSize }
        if (this.filterYear) params.year = this.filterYear
        if (this.keyword) params.keyword = this.keyword
        const res = await getSystemHardeningHistories(params)
        this.records = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
        this.$nextTick(() => this.calcTableHeight())
      }
    },
    handleSizeChange() {
      this.page = 1
      this.fetchData()
    },
    handleFilterChange() {
      this.page = 1
      this.fetchData()
    },
    handleExportChecklist() {
      const url = getExportChecklistUrl()
      const token = localStorage.getItem('token')
      // 通过创建隐藏链接并附带token来下载
      const link = document.createElement('a')
      // 使用fetch下载并带token
      fetch(url, {
        headers: { 'Authorization': `Bearer ${token}` }
      }).then(response => {
        if (!response.ok) throw new Error('导出失败')
        return response.blob()
      }).then(blob => {
        const downloadUrl = URL.createObjectURL(blob)
        link.href = downloadUrl
        link.download = '系统加固检查表.xlsx'
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(downloadUrl)
      }).catch(e => {
        console.error('导出失败:', e)
        this.$message.error('导出失败')
      })
    },
    handleFileChange(file) {
      this.selectedFile = file.raw
    },
    handleFileRemove() {
      this.selectedFile = null
    },
    handleUpload() {
      this.$refs.uploadFormRef.validate(async valid => {
        if (!valid) return
        if (!this.isEdit && !this.selectedFile) {
          this.$message.warning('请选择PDF文件')
          return
        }
        this.uploading = true
        try {
          const formData = new FormData()
          formData.append('year', this.uploadForm.year)
          formData.append('quarter', this.uploadForm.quarter)
          formData.append('description', this.uploadForm.description || '')
          if (this.selectedFile) {
            formData.append('file', this.selectedFile)
          }

          const dualToken = await this.$refs.dualControl.open()
          if (this.isEdit) {
            await updateSystemHardeningHistory(this.editingId, formData, dualToken)
            this.$message.success('更新成功')
          } else {
            await createSystemHardeningHistory(formData, dualToken)
            this.$message.success('上传成功')
          }
          this.showUpload = false
          this.resetUploadForm()
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        } finally {
          this.uploading = false
        }
      })
    },
    resetUploadForm() {
      const now = new Date()
      this.isEdit = false
      this.editingId = null
      this.uploadForm = {
        year: now.getFullYear(),
        quarter: Math.ceil((now.getMonth() + 1) / 3),
        description: ''
      }
      this.selectedFile = null
      this.fileList = []
      if (this.$refs.uploader) {
        this.$refs.uploader.clearFiles()
      }
    },
    handleEdit(row) {
      this.isEdit = true
      this.editingId = row.id
      this.uploadForm = {
        year: row.year,
        quarter: row.quarter,
        description: row.description || ''
      }
      this.showUpload = true
    },
    async handlePreview(row) {
      const url = getSystemHardeningPreviewUrl(row.id)
      try {
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.previewUrl = URL.createObjectURL(blob)
        this.previewVisible = true
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('预览失败')
      }
    },
    async handleDownload(row) {
      const url = getSystemHardeningDownloadUrl(row.id)
      try {
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        const downloadUrl = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = downloadUrl
        link.download = row.file_name
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(downloadUrl)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      }
    },
    async handleDelete(row) {
      try {
        await this.$confirm(`确定要删除 ${row.year}年Q${row.quarter} 的检查记录吗？`, '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteSystemHardeningHistory(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    formatSize(bytes) {
      if (!bytes || bytes === 0) return '0 B'
      const units = ['B', 'KB', 'MB', 'GB']
      let i = 0
      let size = bytes
      while (size >= 1024 && i < units.length - 1) {
        size /= 1024
        i++
      }
      return size.toFixed(i === 0 ? 0 : 1) + ' ' + units[i]
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      return dateStr.replace('T', ' ').substring(0, 19)
    }
  }
}
</script>

<style scoped>
.system-hardening {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
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

.header-actions {
  display: flex;
  gap: 10px;
}

.table-card {
}

.table-wrapper {
}

.op-btns {
  display: flex;
  gap: 6px;
}

/* 主按钮 */
.header-actions .el-button--primary,
.el-dialog__footer .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 500;
  color: #fff;
}
.header-actions .el-button--primary:hover,
.el-dialog__footer .el-button--primary:hover {
  background: #2563eb;
  color: #fff;
}

/* success 按钮 */
.header-actions .el-button--success {
  background: #10b981;
  border: none;
  border-radius: 10px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 500;
  color: #fff;
}
.header-actions .el-button--success:hover {
  background: #059669;
  color: #fff;
}

/* 次要按钮 */
.header-actions .el-button--default {
  background: transparent;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 9px 18px;
  font-size: 13px;
  color: #64748b;
}
.header-actions .el-button--default:hover {
  border-color: #94a3b8;
  color: #1e293b;
}

/* 筛选栏搜索按钮白色文字 */
.filter-bar .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  color: #fff;
}
.filter-bar .el-button--primary:hover {
  background: #2563eb;
  color: #fff;
}
</style>
