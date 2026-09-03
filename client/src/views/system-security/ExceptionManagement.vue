<template>
  <div class="exception-management">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">例外管理</h2>
        <p class="page-subtitle">管理电脑补丁升级例外的授权文件存档</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="openCreate">上传授权文件</el-button>
        <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-input v-model="keyword" placeholder="搜索说明..." size="small" clearable @keyup.enter.native="handleFilterChange" @clear="handleFilterChange" style="width: 200px" />
      <el-button size="small" type="primary" icon="el-icon-search" @click="handleFilterChange">搜索</el-button>
    </div>

    <!-- 数据表格 -->
    <div class="table-card" ref="tableCard" style="margin-top: 12px">
      <div class="table-wrapper">
        <el-table :data="records" stripe v-loading="loading" :max-height="tableMaxHeight">
          <el-table-column type="index" label="#" width="70" align="center" />
          <el-table-column prop="apply_date" label="申请日期" width="130" align="center">
            <template slot-scope="{ row }">{{ formatDateOnly(row.apply_date) }}</template>
          </el-table-column>
          <el-table-column prop="applicant" label="申请人" width="120" align="center" />
          <el-table-column prop="reason" label="例外说明" min-width="250" show-overflow-tooltip />
          <el-table-column prop="end_date" label="持续到" width="130" align="center">
            <template slot-scope="{ row }">{{ formatDateOnly(row.end_date) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="350" fixed="right" align="center">
            <template slot-scope="{ row }">
              <div class="op-btns">
                <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreview(row)">预览</el-button>
                <el-button size="mini" type="text" icon="el-icon-download" @click="handleDownload(row)">下载</el-button>
                <el-button size="mini" type="text" icon="el-icon-edit" @click="handleEdit(row)">编辑</el-button>
                <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog class="vault-dialog" :title="isEdit ? '编辑授权文件' : '上传授权文件'" :visible.sync="showForm" width="560px" :close-on-click-modal="false">
      <el-form :model="form" ref="formRef" :rules="formRules" label-width="110px">
        <el-form-item label="申请日期" prop="apply_date">
          <el-date-picker 
            v-model="form.apply_date" 
            type="date" 
            value-format="yyyy-MM-dd" 
            placeholder="选择申请日期" 
            style="width: 100%" 
          />
        </el-form-item>
        <el-form-item label="申请人" prop="applicant">
          <el-input v-model="form.applicant" placeholder="请输入申请人" />
        </el-form-item>
        <el-form-item label="例外说明" prop="reason">
          <el-input v-model="form.reason" type="textarea" :rows="3" placeholder="输入例外情况说明" />
        </el-form-item>
        <el-form-item label="持续到" prop="end_date">
          <el-date-picker 
            v-model="form.end_date" 
            type="date" 
            value-format="yyyy-MM-dd" 
            placeholder="选择持续时间" 
            style="width: 100%" 
          />
        </el-form-item>
        <el-form-item label="上传扫描件" prop="file" v-if="!isEdit">
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
            <div slot="tip" class="el-upload__tip">仅支持 PDF 格式文件（领导签字授权单）</div>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可更换文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showForm = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- 预览弹窗（使用 file-viewer） -->
    <el-dialog class="vault-dialog preview-dialog fv-preview-dialog" title="文件预览" :visible.sync="previewVisible" width="80%" top="3vh" :close-on-click-modal="true">
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
          <el-button type="primary" @click="handleDownloadFromPreview">下载文件</el-button>
        </div>
      </div>
      <span slot="footer">
        <el-button type="primary" size="small" icon="el-icon-download" @click="handleDownloadFromPreview">下载</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getExceptionManagementList, createExceptionManagement, updateExceptionManagement, deleteExceptionManagement,
  previewExceptionManagement, previewExceptionManagementUrl, downloadExceptionManagementUrl
} from '@/api/exception_management'
import DualControlDialog from '@/components/DualControlDialog.vue'
import tableHeightMixin from '@/mixins/table-height'
import { FileViewer } from '@file-viewer/vue2.7'
import officePreset from '@file-viewer/preset-office'
import fvFeature from '@/config/fv-feature'

export default {
  name: 'ExceptionManagement',
  components: { DualControlDialog, FileViewer },
  mixins: [tableHeightMixin],
  data() {
    return {
      records: [],
      loading: false,
      page: 1,
      pageSize: 10,
      total: 0,
      keyword: '',
      // 新增/编辑
      showForm: false,
      isEdit: false,
      editingId: null,
      submitting: false,
      form: {
        apply_date: '',
        applicant: '',
        reason: '',
        end_date: ''
      },
      formRules: {
        apply_date: [{ required: true, message: '请选择申请日期', trigger: 'change' }],
        applicant: [{ required: true, message: '请输入申请人', trigger: 'blur' }],
        reason: [{ required: true, message: '请输入例外说明', trigger: 'blur' }],
        end_date: [{ required: true, message: '请选择截止时间', trigger: 'change' }]
      },
      selectedFile: null,
      fileList: [],
      // 预览
      previewVisible: false,
      currentFileUrl: '',
      currentFileName: '',
      currentFileType: '',
      fvFeature: fvFeature
    }
  },
  computed: {
    fvOptions() {
      return {
        preset: officePreset,
        fetchFile: this.fetchFileWithAuth
      }
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const params = { 
          page: this.page, 
          page_size: this.pageSize,
          keyword: this.keyword
        }
        const res = await getExceptionManagementList(params)
        // 确保数据是数组格式
        this.records = Array.isArray(res.data) ? res.data : (res.data && res.data.rows ? res.data.rows : [])
        this.total = res.total || res.count || 0
      } catch (e) {
        console.error('获取数据失败:', e)
        this.records = [] // 确保 records 始终是数组
        this.total = 0
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
    // 新增
    openCreate() {
      this.resetForm()
      this.showForm = true
    },
    // 编辑
    handleEdit(row) {
      this.isEdit = true
      this.editingId = row.id
      this.form = {
        apply_date: row.apply_date || '',
        applicant: row.applicant || '',
        reason: row.reason || '',
        end_date: row.end_date || ''
      }
      this.showForm = true
    },
    resetForm() {
      this.isEdit = false
      this.editingId = null
      this.selectedFile = null
      this.fileList = []
      this.form = {
        apply_date: '',
        applicant: '',
        reason: '',
        end_date: ''
      }
    },
    handleFileChange(file) {
      if (!this.validateFileSize(file)) {
        this.selectedFile = null
        return
      }
      this.selectedFile = file.raw
    },
    handleFileRemove() {
      this.selectedFile = null
    },
    async handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        if (!this.isEdit && !this.selectedFile) {
          this.$message.warning('请选择 PDF 文件')
          return
        }
        this.submitting = true
        try {
          const dualToken = await this.$refs.dualControl.open()
          const formData = new FormData()
          formData.append('apply_date', this.form.apply_date)
          formData.append('applicant', this.form.applicant)
          formData.append('reason', this.form.reason)
          formData.append('endDate', this.form.end_date)

          if (this.isEdit) {
            await updateExceptionManagement(this.editingId, formData, dualToken)
            this.$message.success('更新成功')
          } else {
            formData.append('file', this.selectedFile)
            await createExceptionManagement(formData, dualToken)
            this.$message.success('上传成功')
          }
          this.showForm = false
          this.resetForm()
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        } finally {
          this.submitting = false
        }
      })
    },
    // 删除记录
    async handleDelete(row) {
      try {
        await this.$confirm(`确定要删除该条例外管理记录吗？`, '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteExceptionManagement(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    // 预览文件
    async handlePreview(row) {
      const url = previewExceptionManagementUrl(row.id)
      const fileExtension = row.file_name ? row.file_name.split('.').pop().toLowerCase() : ''
      
      this.currentFileName = row.file_name || 'unknown_file'
      this.currentFileType = fileExtension
      this.currentFileUrl = url
      
      // 先关闭再打开，确保 FileViewer 组件重新渲染
      this.previewVisible = false
      await this.$nextTick()
      this.previewVisible = true
      await this.$nextTick()
    },
    async fetchFileWithAuth({ url }) {
      const token = localStorage.getItem('token')
      const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
      if (!response.ok) {
        throw new Error(`Failed to fetch file: ${response.status} ${response.statusText}`)
      }
      return response.arrayBuffer()
    },
    async handleDownloadFromPreview() {
      if (!this.currentFileUrl) return
      let link = null
      try {
        const response = await fetch(this.currentFileUrl, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = this.currentFileName || 'download'
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      } finally {
        if (link) URL.revokeObjectURL(link.href) // 确保释放
      }
    },
    async handleDownload(row) {
      let link = null
      try {
        const url = downloadExceptionManagementUrl(row.id)
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = row.file_name
        link.click()
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      } finally {
        if (link) URL.revokeObjectURL(link.href)
      }
    },
    formatDateOnly(dateStr) {
      if (!dateStr) return '-'
      return dateStr.substring(0, 10)
    },
    // 转义 HTML 防止 XSS
    escapeHtml(text) {
      if (!text) return ''
      const div = document.createElement('div')
      div.textContent = String(text)
      return div.innerHTML
    },
    // 验证文件大小（最大 10MB）
    validateFileSize(file) {
      const maxSize = 10 * 1024 * 1024 // 10MB
      if (file.size > maxSize) {
        this.$message.warning('文件大小不能超过 10MB')
        return false
      }
      return true
    },
  }
}
</script>

<style scoped>
.exception-management {
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


.op-btns {
  display: flex;
  gap: 6px;
}

/* 主按钮 */
.header-actions .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 500;
  color: #fff;
}
.header-actions .el-button--primary:hover {
  background: #2563eb;
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

/* 筛选栏搜索按钮 */
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
