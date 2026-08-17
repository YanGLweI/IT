<template>
  <div class="patch-update">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">补丁更新</h2>
        <p class="page-subtitle">管理月度补丁合规性报表，跟踪修复进度</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="openCreate">上传合规性报表</el-button>
        <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
      </div>
    </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="全部年份" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="y + '年'" :value="y" />
        </el-select>
        <el-select v-model="filterCompliance" placeholder="合规性" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option label="合规" value="compliant" />
          <el-option label="不合规" value="non_compliant" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索文件名..." size="small" clearable @keyup.enter.native="handleFilterChange" @clear="handleFilterChange" style="width: 200px" />
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleFilterChange">搜索</el-button>
      </div>

      <!-- 数据表格 -->
      <div class="table-card" ref="tableCard" style="margin-top: 12px">
      <div class="table-wrapper">
      <el-table :data="records" stripe v-loading="loading" :max-height="tableMaxHeight">
        <el-table-column type="index" label="#" width="70" align="center" />
        <el-table-column prop="year" label="年份" width="85" align="center" />
        <el-table-column prop="month" label="月份" width="85" align="center">
          <template slot-scope="{ row }">{{ row.month }}月</template>
        </el-table-column>
        <el-table-column prop="total_assets" label="资产总数" width="100" align="center" />
        <el-table-column label="合规性" width="120" align="center">
          <template slot-scope="{ row }">
            <el-tag :type="row.compliance === 'compliant' ? 'success' : 'danger'" size="small">
              {{ row.compliance === 'compliant' ? '合规' : '不合规' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="不合规资产数" width="120" align="center">
          <template slot-scope="{ row }">
            <span :style="row.fix_file_name ? 'text-decoration: line-through' : ''">{{ row.non_compliant_assets }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="file_name" label="合规性报表" min-width="200" show-overflow-tooltip />
        <el-table-column label="修复报表" width="180" align="center">
          <template slot-scope="{ row }">
            <template v-if="row.fix_file_name">
              <div class="op-btns" style="justify-content: center">
                <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreviewFix(row)">预览</el-button>
                <el-button size="mini" type="danger" icon="el-icon-delete" @click="handleDeleteFix(row)">删除</el-button>
              </div>
            </template>
            <el-button v-else-if="row.compliance === 'non_compliant'" size="mini" type="text" icon="el-icon-s-check" style="color: #67C23A" @click="openFixUpload(row)">修复</el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="350" fixed="right" align="center">
          <template slot-scope="{ row }">
            <div class="op-btns">
              <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreview(row)">预览</el-button>
              <el-button size="mini" type="text" icon="el-icon-download" @click="handleDownload(row)">下载</el-button>
              <el-button size="mini" type="text" icon="el-icon-edit" @click="handleEdit(row)">编辑</el-button>
              <el-button size="mini" type="danger" icon="el-icon-delete" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog class="vault-dialog" :title="isEdit ? '编辑合规性报表' : '上传合规性报表'" :visible.sync="showForm" width="560px" :close-on-click-modal="false">
      <el-form :model="form" ref="formRef" :rules="formRules" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="年份" prop="year">
              <el-input-number v-model="form.year" :min="2020" :max="2100" :step="1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="月份" prop="month">
              <el-select v-model="form.month" placeholder="请选择" style="width: 100%">
                <el-option v-for="m in 12" :key="m" :label="m + '月'" :value="m" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="资产总数" prop="total_assets">
          <el-input-number v-model="form.total_assets" :min="0" :step="1" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="合规性" prop="compliance">
          <el-radio-group v-model="form.compliance" @change="onComplianceChange">
            <el-radio label="compliant">合规</el-radio>
            <el-radio label="non_compliant">不合规</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="不合规资产数" prop="non_compliant_assets" v-if="form.compliance === 'non_compliant'">
          <el-input-number v-model="form.non_compliant_assets" :min="0" :step="1" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="报表文件" v-if="!isEdit">
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
        <el-alert v-else title="编辑模式下不可更换文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showForm = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- 修复报表上传弹窗 -->
    <el-dialog class="vault-dialog" title="上传修复报表" :visible.sync="showFixUpload" width="480px" :close-on-click-modal="false">
      <el-upload
        ref="fixUploader"
        action=""
        :auto-upload="false"
        :limit="1"
        accept=".pdf"
        :on-change="handleFixFileChange"
        :on-remove="handleFixFileRemove"
        :file-list="fixFileList"
        drag
      >
        <i class="el-icon-upload"></i>
        <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
        <div slot="tip" class="el-upload__tip">仅支持 PDF 格式文件</div>
      </el-upload>
      <span slot="footer">
        <el-button @click="showFixUpload = false">取消</el-button>
        <el-button type="primary" :loading="fixUploading" @click="handleFixUpload">确定上传</el-button>
      </span>
    </el-dialog>

    <!-- 预览弹窗（使用 file-viewer，合规性报表/修复报表共用） -->
    <el-dialog class="vault-dialog preview-dialog fv-preview-dialog" :title="previewTitle" :visible.sync="previewVisible" width="80%" top="3vh" :close-on-click-modal="true">
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
  getPatchUpdates, createPatchUpdate, updatePatchUpdate, deletePatchUpdate,
  uploadPatchFixReport, deletePatchFixReport,
  getPatchUpdatePreviewUrl, getPatchUpdateDownloadUrl,
  getPatchFixPreviewUrl, getPatchFixDownloadUrl
} from '@/api/patch_update'
import DualControlDialog from '@/components/DualControlDialog.vue'
import tableHeightMixin from '@/mixins/table-height'
import { FileViewer } from '@file-viewer/vue2.7'
import officePreset from '@file-viewer/preset-office'
import fvFeature from '@/config/fv-feature'

export default {
  name: 'PatchUpdate',
  components: { DualControlDialog, FileViewer },
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
      filterCompliance: '',
      keyword: '',
      yearOptions: Array.from({ length: 10 }, (_, i) => now.getFullYear() - i),
      // 新增/编辑
      showForm: false,
      isEdit: false,
      editingId: null,
      submitting: false,
      form: {
        year: now.getFullYear(),
        month: now.getMonth() + 1,
        total_assets: 0,
        compliance: 'compliant',
        non_compliant_assets: 0
      },
      formRules: {
        year: [{ required: true, message: '请选择年份', trigger: 'change' }],
        month: [{ required: true, message: '请选择月份', trigger: 'change' }],
        total_assets: [{ required: true, message: '请输入资产总数', trigger: 'blur' }],
        non_compliant_assets: [{
          validator: (rule, value, callback) => {
            if (this.form.compliance === 'non_compliant' && (!value || value <= 0)) {
              callback(new Error('不合规时必须填写不合规资产数'))
            } else {
              callback()
            }
          },
          trigger: 'change'
        }]
      },
      selectedFile: null,
      fileList: [],
      // 修复报表上传
      showFixUpload: false,
      fixRowId: null,
      fixUploading: false,
      fixSelectedFile: null,
      fixFileList: [],
      // 预览（合规性报表/修复报表共用）
      previewVisible: false,
      previewTitle: '文件预览',
      previewDownloadUrl: '',
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
    onComplianceChange() {
      // 切换到合规时清空不合规资产数
      if (this.form.compliance === 'compliant') {
        this.form.non_compliant_assets = 0
      }
      // 触发不合规资产数字段重新验证
      this.$nextTick(() => {
        this.$refs.form && this.$refs.form.validateField('non_compliant_assets')
      })
    },
    async fetchData() {
      this.loading = true
      try {
        const params = { page: this.page, page_size: this.pageSize }
        if (this.filterYear) params.year = this.filterYear
        if (this.filterCompliance) params.compliance = this.filterCompliance
        if (this.keyword) params.keyword = this.keyword
        const res = await getPatchUpdates(params)
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
        year: row.year,
        month: row.month,
        total_assets: row.total_assets,
        compliance: row.compliance || 'compliant',
        non_compliant_assets: row.non_compliant_assets || 0
      }
      this.showForm = true
    },
    resetForm() {
      const now = new Date()
      this.isEdit = false
      this.editingId = null
      this.selectedFile = null
      this.fileList = []
      this.form = {
        year: now.getFullYear(),
        month: now.getMonth() + 1,
        total_assets: 0,
        compliance: 'compliant',
        non_compliant_assets: 0
      }
    },
    handleFileChange(file) {
      this.selectedFile = file.raw
    },
    handleFileRemove() {
      this.selectedFile = null
    },
    async handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        if (!this.isEdit && !this.selectedFile) {
          this.$message.warning('请选择PDF文件')
          return
        }
        this.submitting = true
        try {
          const dualToken = await this.$refs.dualControl.open()
          const formData = new FormData()
          formData.append('year', this.form.year)
          formData.append('month', this.form.month)
          formData.append('total_assets', this.form.total_assets)
          formData.append('compliance', this.form.compliance)
          formData.append('non_compliant_assets', this.form.non_compliant_assets)

          if (this.isEdit) {
            await updatePatchUpdate(this.editingId, formData, dualToken)
            this.$message.success('更新成功')
          } else {
            formData.append('file', this.selectedFile)
            await createPatchUpdate(formData, dualToken)
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
        await this.$confirm(`确定要删除 ${row.year}年${row.month}月 的合规性报表吗？`, '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deletePatchUpdate(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    // 预览合规性报表
    async handlePreview(row) {
      this.previewTitle = '文件预览'
      this.previewDownloadUrl = getPatchUpdateDownloadUrl(row.id)
      this.currentFileName = row.file_name || 'unknown_file'
      this.currentFileType = row.file_name ? row.file_name.split('.').pop().toLowerCase() : ''
      this.currentFileUrl = getPatchUpdatePreviewUrl(row.id)
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
      if (!this.previewDownloadUrl) return
      try {
        const response = await fetch(this.previewDownloadUrl, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = this.currentFileName || 'download'
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(link.href)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      }
    },
    // 下载合规性报表
    async handleDownload(row) {
      const url = getPatchUpdateDownloadUrl(row.id)
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = row.file_name
        link.click()
        URL.revokeObjectURL(link.href)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      }
    },
    // 修复报表上传
    openFixUpload(row) {
      this.fixRowId = row.id
      this.fixSelectedFile = null
      this.fixFileList = []
      this.showFixUpload = true
    },
    handleFixFileChange(file) {
      this.fixSelectedFile = file.raw
    },
    handleFixFileRemove() {
      this.fixSelectedFile = null
    },
    async handleFixUpload() {
      if (!this.fixSelectedFile) {
        this.$message.warning('请选择文件')
        return
      }
      this.fixUploading = true
      try {
        const formData = new FormData()
        formData.append('file', this.fixSelectedFile)
        const dualToken = await this.$refs.dualControl.open()
        await uploadPatchFixReport(this.fixRowId, formData, dualToken)
        this.$message.success('上传成功')
        this.showFixUpload = false
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') {
          console.error(e)
          this.$message.error('上传失败')
        }
      } finally {
        this.fixUploading = false
      }
    },
    // 删除修复报表
    async handleDeleteFix(row) {
      try {
        await this.$confirm('确定要删除该修复报表吗？', '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deletePatchFixReport(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    // 预览修复报表
    async handlePreviewFix(row) {
      this.previewTitle = '修复报表预览'
      this.previewDownloadUrl = getPatchFixDownloadUrl(row.id)
      this.currentFileName = row.fix_file_name || 'unknown_file'
      this.currentFileType = row.fix_file_name ? row.fix_file_name.split('.').pop().toLowerCase() : ''
      this.currentFileUrl = getPatchFixPreviewUrl(row.id)
      this.previewVisible = false
      await this.$nextTick()
      this.previewVisible = true
      await this.$nextTick()
    }
  }
}
</script>

<style scoped>
.patch-update {
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
