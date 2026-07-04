<template>
  <div class="patch-update">
    <el-card>
      <div slot="header" class="page-header">
        <span>补丁更新</span>
        <div class="page-header-right">
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
      <el-table :data="records" border stripe v-loading="loading" style="margin-top: 12px">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="year" label="年份" width="80" align="center" />
        <el-table-column prop="month" label="月份" width="80" align="center">
          <template slot-scope="{ row }">{{ row.month }}月</template>
        </el-table-column>
        <el-table-column prop="total_assets" label="资产总数" width="100" align="center" />
        <el-table-column label="合规性" width="100" align="center">
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
        <el-table-column label="修复报表" width="150" align="center">
          <template slot-scope="{ row }">
            <template v-if="row.fix_file_name">
              <div class="op-btns" style="justify-content: center">
                <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreviewFix(row)">预览</el-button>
                <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="handleDeleteFix(row)">删除</el-button>
              </div>
            </template>
            <el-button v-else-if="row.compliance === 'non_compliant'" size="mini" type="text" icon="el-icon-s-check" style="color: #67C23A" @click="openFixUpload(row)">修复</el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
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

      <!-- 分页 -->
      <el-pagination
        style="margin-top: 16px; text-align: right"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size.sync="pageSize"
        :current-page.sync="page"
        :page-sizes="[10, 20, 50]"
        @size-change="handleSizeChange"
        @current-change="fetchData"
      />
    </el-card>

    <!-- 上传/编辑弹窗 -->
    <el-dialog :title="isEdit ? '编辑合规性报表' : '上传合规性报表'" :visible.sync="showForm" width="560px" :close-on-click-modal="false">
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
    <el-dialog title="上传修复报表" :visible.sync="showFixUpload" width="480px" :close-on-click-modal="false">
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

    <!-- 预览弹窗 -->
    <el-dialog title="文件预览" :visible.sync="previewVisible" width="80%" top="3vh" :close-on-click-modal="true" @close="clearPreview">
      <iframe v-if="pdfBlobUrl" :src="pdfBlobUrl" style="width: 100%; height: 70vh; border: none;" />
    </el-dialog>

    <!-- 修复报表预览弹窗 -->
    <el-dialog title="修复报表预览" :visible.sync="fixPreviewVisible" width="80%" top="3vh" :close-on-click-modal="true" @close="clearFixPreview">
      <iframe v-if="fixPdfBlobUrl" :src="fixPdfBlobUrl" style="width: 100%; height: 70vh; border: none;" />
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

export default {
  name: 'PatchUpdate',
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
      // 合规性报表预览
      previewVisible: false,
      pdfBlobUrl: '',
      // 修复报表预览
      fixPreviewVisible: false,
      fixPdfBlobUrl: ''
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
      const url = getPatchUpdatePreviewUrl(row.id)
      try {
        this.clearPreview()
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.pdfBlobUrl = URL.createObjectURL(blob)
        this.previewVisible = true
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('预览失败')
      }
    },
    clearPreview() {
      if (this.pdfBlobUrl) {
        URL.revokeObjectURL(this.pdfBlobUrl)
        this.pdfBlobUrl = ''
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
      const url = getPatchFixPreviewUrl(row.id)
      try {
        this.clearFixPreview()
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.fixPdfBlobUrl = URL.createObjectURL(blob)
        this.fixPreviewVisible = true
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('预览失败')
      }
    },
    clearFixPreview() {
      if (this.fixPdfBlobUrl) {
        URL.revokeObjectURL(this.fixPdfBlobUrl)
        this.fixPdfBlobUrl = ''
      }
    }
  }
}
</script>

<style scoped>
.patch-update {
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
