<template>
  <div class="system-hardening">
    <el-card>
      <template #header>
        <div class="page-header">
        <span>系统加固</span>
        <div class="page-header-right">
          <el-button type="success" size="small" :icon="Download" @click="handleExportChecklist">导出检查表</el-button>
          <el-button type="primary" size="small" :icon="Upload" @click="showUpload = true">上传记录</el-button>
          <el-button type="default" size="small" :icon="Refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>
      </template>
      

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="全部年份" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="y + '年'" :value="y" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索描述..." size="small" clearable @keyup.enter="handleFilterChange" @clear="handleFilterChange" style="width: 200px" />
        <el-button size="small" type="primary" :icon="Search" @click="handleFilterChange">搜索</el-button>
      </div>

      <!-- 数据表格 -->
      <el-table :data="records" border stripe v-loading="loading" style="margin-top: 12px">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="year" label="年份" width="80" align="center" />
        <el-table-column label="季度" width="80" align="center">
          <template v-slot="{ row }">Q{{ row.quarter }}</template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
        <el-table-column label="文件大小" width="100" align="center">
          <template v-slot="{ row }">{{ formatSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column label="上传时间" width="180" align="center">
          <template v-slot="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right" align="center">
          <template v-slot="{ row }">
            <div class="op-btns">
              <el-button size="small" text :icon="View" @click="handlePreview(row)">预览</el-button>
              <el-button size="small" text :icon="Download" @click="handleDownload(row)">下载</el-button>
              <el-button size="small" text :icon="Edit" @click="handleEdit(row)">编辑</el-button>
              <el-button size="small" text :icon="Delete" style="color: #F56C6C" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        style="margin-top: 16px; text-align: right"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        :page-sizes="[10, 20, 50]"
        @size-change="handleSizeChange"
        @current-change="fetchData"
      />
    </el-card>

    <!-- 上传/编辑弹窗 -->
    <el-dialog :title="isEdit ? '编辑系统加固检查记录' : '上传系统加固检查记录'" v-model="showUpload" width="520px" :close-on-click-modal="false">
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
            <el-icon><Upload /></el-icon>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
            <template #tip><div  class="el-upload__tip">仅支持 PDF 格式文件</div></template>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可修改文件" type="info" :closable="false" show-icon />
      </el-form>
      <template #footer>
        <el-button @click="showUpload = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">{{ isEdit ? '保存' : '确定上传' }}</el-button>
      </template>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog title="文件预览" v-model="previewVisible" width="80%" top="3vh" :close-on-click-modal="true">
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

export default {
  name: 'SystemHardening',
  components: { DualControlDialog },
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
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header-right {
  display: flex;
  gap: 8px;
}

.filter-bar {
  display: flex;
  gap: 8px;
  align-items: center;
}

.op-btns {
  display: flex;
  gap: 4px;
}
</style>
