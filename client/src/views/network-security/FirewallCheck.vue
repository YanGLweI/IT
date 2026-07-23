<template>
  <div class="firewall-check">
    <el-card>
      <div slot="header" class="page-header">
        <span>防火墙检查</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-plus" @click="openCreate">新增记录</el-button>
          <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="全部年份" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="y + '年'" :value="y" />
        </el-select>
        <el-select v-model="filterQuarter" placeholder="全部季度" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option v-for="q in 4" :key="q" :label="'Q' + q" :value="q" />
        </el-select>
        <el-select v-model="filterResult" placeholder="检查结果" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option label="合规" value="compliant" />
          <el-option label="不合规" value="non_compliant" />
        </el-select>
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleFilterChange">搜索</el-button>
      </div>

      <!-- 数据表格 -->
      <div class="table-card" style="margin-top: 12px">
        <el-table :data="records" stripe v-loading="loading">
          <el-table-column type="index" label="#" width="70" align="center" />
          <el-table-column prop="year" label="年份" width="85" align="center" />
          <el-table-column label="季度" width="85" align="center">
            <template slot-scope="{ row }">Q{{ row.quarter }}</template>
          </el-table-column>
          <el-table-column prop="report_date" label="报告日期" width="150" align="center" />
          <el-table-column label="防火墙" min-width="150" show-overflow-tooltip>
            <template slot-scope="{ row }">
              <span v-if="row.asset">{{ row.asset.computer_name }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="检查结果" width="150" align="center">
            <template slot-scope="{ row }">
              <el-tag :type="row.check_result === 'compliant' ? 'success' : 'danger'" size="small">
                {{ row.check_result === 'compliant' ? '合规' : '不合规' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="file_name" label="检查报告" min-width="150" show-overflow-tooltip>
            <template slot-scope="{ row }">
              <span v-if="row.file_name">{{ row.file_name }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="整改报告" width="250" align="center">
            <template slot-scope="{ row }">
              <template v-if="row.check_result === 'non_compliant'">
                <template v-if="row.rect_file_name">
                  <div class="op-btns" style="justify-content: center">
                    <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreviewRect(row)">预览</el-button>
                    <el-button size="mini" type="danger" icon="el-icon-delete" @click="handleDeleteRect(row)">删除</el-button>
                  </div>
                </template>
                <el-button v-else size="mini" type="text" icon="el-icon-document-checked" style="color: #E6A23C" @click="openRectUpload(row)">整改</el-button>
              </template>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="350" fixed="right" align="center">
            <template slot-scope="{ row }">
              <div class="op-btns">
                <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreview(row)">预览</el-button>
                <el-button size="mini" type="text" icon="el-icon-download" @click="handleDownload(row)">下载</el-button>
                <el-button size="mini" type="text" icon="el-icon-edit" @click="openEdit(row)">编辑</el-button>
                <el-button size="mini" type="danger" icon="el-icon-delete" @click="handleDelete(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
      

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

    <!-- 新增/编辑弹窗 -->
    <el-dialog :title="isEdit ? '编辑检查记录' : '新增检查记录'" :visible.sync="showForm" width="580px" :close-on-click-modal="false">
      <el-form :model="form" ref="formRef" :rules="formRules" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="年份" prop="year">
              <el-input-number v-model="form.year" :min="2020" :max="2100" :step="1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="季度" prop="quarter">
              <el-select v-model="form.quarter" placeholder="选择季度" style="width: 100%">
                <el-option v-for="q in 4" :key="q" :label="'Q' + q" :value="q" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="报告日期" prop="report_date">
          <el-date-picker v-model="form.report_date" type="date" value-format="yyyy-MM-dd" placeholder="选择日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="防火墙" prop="asset_id">
          <el-select v-model="form.asset_id" placeholder="选择防火墙设备" filterable style="width: 100%">
            <el-option v-for="a in assetOptions" :key="a.id" :label="`${a.computer_name} (${a.ip_address || '无IP'})`" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="检查结果" prop="check_result">
          <el-radio-group v-model="form.check_result">
            <el-radio label="compliant">合规</el-radio>
            <el-radio label="non_compliant">不合规</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="检查报告" v-if="!isEdit">
          <el-upload
            ref="uploader"
            action=""
            :auto-upload="false"
            :limit="1"
            accept=".pdf,.docx"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            drag
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
            <div slot="tip" class="el-upload__tip">支持 PDF / DOCX 格式文件</div>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可更换文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showForm = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- 整改报告上传弹窗 -->
    <el-dialog title="上传整改报告" :visible.sync="showRectUpload" width="480px" :close-on-click-modal="false">
      <el-upload
        ref="rectUploader"
        action=""
        :auto-upload="false"
        :limit="1"
        accept=".pdf,.docx"
        :on-change="handleRectFileChange"
        :on-remove="handleRectFileRemove"
        :file-list="rectFileList"
        drag
      >
        <i class="el-icon-upload"></i>
        <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
        <div slot="tip" class="el-upload__tip">支持 PDF / DOCX 格式文件</div>
      </el-upload>
      <span slot="footer">
        <el-button @click="showRectUpload = false">取消</el-button>
        <el-button type="primary" :loading="rectUploading" @click="handleRectUpload">确定上传</el-button>
      </span>
    </el-dialog>

    <!-- 检查报告预览弹窗 -->
    <el-dialog title="检查报告预览" :visible.sync="previewVisible" width="80%" top="3vh" :close-on-click-modal="true" @close="clearPreview">
      <iframe v-if="previewUrl && isPdf && pdfBlobUrl" :src="pdfBlobUrl" style="width: 100%; height: 70vh; border: none;" />
      <div v-else-if="!isPdf" ref="docxScrollContainer" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 20px">
        <div ref="docxContainer" class="docx-preview-container"></div>
      </div>
      <span slot="footer">
        <el-button type="primary" size="small" icon="el-icon-download" @click="handleDownloadFromPreview">下载</el-button>
      </span>
    </el-dialog>

    <!-- 整改报告预览弹窗 -->
    <el-dialog title="整改报告预览" :visible.sync="rectPreviewVisible" width="80%" top="3vh" :close-on-click-modal="true" @close="clearRectPreview">
      <iframe v-if="rectPreviewUrl && rectIsPdf && rectPdfBlobUrl" :src="rectPdfBlobUrl" style="width: 100%; height: 70vh; border: none;" />
      <div v-else-if="!rectIsPdf" ref="rectDocxScrollContainer" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 20px">
        <div ref="rectDocxContainer" class="docx-preview-container"></div>
      </div>
      <span slot="footer">
        <el-button type="primary" size="small" icon="el-icon-download" @click="handleDownloadRect">下载</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getFirewallChecks, createFirewallCheck, updateFirewallCheck, deleteFirewallCheck,
  uploadFirewallRectReport, deleteFirewallRectReport,
  getFirewallCheckPreviewUrl, getFirewallCheckDownloadUrl,
  getFirewallRectPreviewUrl, getFirewallRectDownloadUrl
} from '@/api/firewall_check'
import { getAssets } from '@/api/asset'
import { getRegions } from '@/api/region'
import DualControlDialog from '@/components/DualControlDialog.vue'
import { renderAsync } from 'docx-preview'

export default {
  name: 'FirewallCheck',
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
      filterQuarter: '',
      filterResult: '',
      yearOptions: Array.from({ length: 10 }, (_, i) => now.getFullYear() - i),
      // 新增/编辑
      showForm: false,
      isEdit: false,
      editingId: null,
      submitting: false,
      form: {
        year: now.getFullYear(),
        quarter: Math.ceil((now.getMonth() + 1) / 3),
        report_date: '',
        asset_id: null,
        check_result: 'compliant'
      },
      formRules: {
        year: [{ required: true, message: '请选择年份', trigger: 'change' }],
        quarter: [{ required: true, message: '请选择季度', trigger: 'change' }],
        report_date: [{ required: true, message: '请选择报告日期', trigger: 'change' }],
        asset_id: [{ required: true, message: '请选择防火墙', trigger: 'change' }]
      },
      assetOptions: [],
      selectedFile: null,
      fileList: [],
      // 整改报告上传
      showRectUpload: false,
      rectRowId: null,
      rectUploading: false,
      rectSelectedFile: null,
      rectFileList: [],
      // 检查报告预览
      previewVisible: false,
      previewUrl: '',
      previewDownloadUrl: '',
      previewFileName: '',
      isPdf: true,
      pdfBlobUrl: '',
      // 整改报告预览
      rectPreviewVisible: false,
      rectPreviewUrl: '',
      rectDownloadUrl: '',
      rectFileName: '',
      rectIsPdf: true,
      rectPdfBlobUrl: ''
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
        if (this.filterQuarter) params.quarter = this.filterQuarter
        if (this.filterResult) params.check_result = this.filterResult
        const res = await getFirewallChecks(params)
        this.records = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    async fetchAssets() {
      try {
        // 先获取"Firewall"区域ID
        const regionsRes = await getRegions()
        const networkRegion = (regionsRes.data || []).find(r => r.name === 'Firewall')
        if (!networkRegion) {
          this.$message.warning('未找到"Firewall"区域，请先创建区域')
          return
        }
        // 只获取该区域下的资产
        const res = await getAssets({ page: 1, page_size: 500, region_id: networkRegion.id })
        this.assetOptions = res.data || []
      } catch (e) {
        console.error(e)
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
      this.fetchAssets()
    },
    // 编辑
    openEdit(row) {
      this.isEdit = true
      this.editingId = row.id
      this.form = {
        year: row.year,
        quarter: row.quarter,
        report_date: row.report_date || '',
        asset_id: row.asset_id,
        check_result: row.check_result || 'compliant'
      }
      this.showForm = true
      this.fetchAssets()
    },
    resetForm() {
      const now = new Date()
      this.isEdit = false
      this.editingId = null
      this.selectedFile = null
      this.fileList = []
      this.form = {
        year: now.getFullYear(),
        quarter: Math.ceil((now.getMonth() + 1) / 3),
        report_date: '',
        asset_id: null,
        check_result: 'compliant'
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
          this.$message.warning('请选择检查报告文件')
          return
        }
        this.submitting = true
        try {
          const dualToken = await this.$refs.dualControl.open()
          if (this.isEdit) {
            const formData = new FormData()
            formData.append('year', this.form.year)
            formData.append('quarter', this.form.quarter)
            formData.append('report_date', this.form.report_date || '')
            formData.append('asset_id', this.form.asset_id)
            formData.append('check_result', this.form.check_result)
            await updateFirewallCheck(this.editingId, formData, dualToken)
            this.$message.success('更新成功')
          } else {
            const formData = new FormData()
            formData.append('year', this.form.year)
            formData.append('quarter', this.form.quarter)
            formData.append('report_date', this.form.report_date || '')
            formData.append('asset_id', this.form.asset_id)
            formData.append('check_result', this.form.check_result)
            formData.append('file', this.selectedFile)
            await createFirewallCheck(formData, dualToken)
            this.$message.success('创建成功')
          }
          this.showForm = false
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') {
            console.error(e)
            this.$message.error(this.isEdit ? '更新失败' : '创建失败')
          }
        } finally {
          this.submitting = false
        }
      })
    },
    // 删除
    async handleDelete(row) {
      try {
        await this.$confirm('确定要删除该检查记录吗？此操作不可恢复。', '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteFirewallCheck(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    // 检查报告预览
    async handlePreview(row) {
      this.clearPreview()
      this.previewUrl = getFirewallCheckPreviewUrl(row.id)
      this.previewDownloadUrl = getFirewallCheckDownloadUrl(row.id)
      this.previewFileName = row.file_name
      this.isPdf = row.file_name && row.file_name.toLowerCase().endsWith('.pdf')
      if (this.pdfBlobUrl) {
        URL.revokeObjectURL(this.pdfBlobUrl)
        this.pdfBlobUrl = ''
      }
      this.previewVisible = true
      if (this.isPdf) {
        await this.fetchPdfAsBlob(this.previewUrl)
      } else {
        this.$nextTick(() => { this.renderDocx(this.previewUrl) })
      }
    },
    async fetchPdfAsBlob(url) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
        this.pdfBlobUrl = URL.createObjectURL(blob)
      } catch (e) {
        console.error('PDF加载失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    async renderDocx(url) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
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
      if (this.$refs.docxContainer) this.$refs.docxContainer.innerHTML = ''
      if (this.pdfBlobUrl) {
        URL.revokeObjectURL(this.pdfBlobUrl)
        this.pdfBlobUrl = ''
      }
    },
    async handleDownload(row) {
      await this.downloadWithAuth(getFirewallCheckDownloadUrl(row.id), row.file_name)
    },
    async handleDownloadFromPreview() {
      await this.downloadWithAuth(this.previewDownloadUrl, this.previewFileName)
    },
    // 通用下载方法（携带JWT token）
    async downloadWithAuth(url, fileName) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = fileName || '下载文件'
        link.click()
        URL.revokeObjectURL(link.href)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('文件下载失败')
      }
    },
    // 整改报告上传
    openRectUpload(row) {
      this.rectRowId = row.id
      this.rectSelectedFile = null
      this.rectFileList = []
      this.showRectUpload = true
    },
    handleRectFileChange(file) {
      this.rectSelectedFile = file.raw
    },
    handleRectFileRemove() {
      this.rectSelectedFile = null
    },
    async handleRectUpload() {
      if (!this.rectSelectedFile) {
        this.$message.warning('请选择文件')
        return
      }
      this.rectUploading = true
      try {
        const formData = new FormData()
        formData.append('file', this.rectSelectedFile)
        const dualToken = await this.$refs.dualControl.open()
        await uploadFirewallRectReport(this.rectRowId, formData, dualToken)
        this.$message.success('上传成功')
        this.showRectUpload = false
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') {
          console.error(e)
          this.$message.error('上传失败')
        }
      } finally {
        this.rectUploading = false
      }
    },
    // 删除整改报告
    async handleDeleteRect(row) {
      try {
        await this.$confirm('确定要删除该整改报告吗？', '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteFirewallRectReport(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    // 预览整改报告
    async handlePreviewRect(row) {
      this.clearRectPreview()
      this.rectPreviewUrl = getFirewallRectPreviewUrl(row.id)
      this.rectDownloadUrl = getFirewallRectDownloadUrl(row.id)
      this.rectFileName = row.rect_file_name
      this.rectIsPdf = row.rect_file_name && row.rect_file_name.toLowerCase().endsWith('.pdf')
      if (this.rectPdfBlobUrl) {
        URL.revokeObjectURL(this.rectPdfBlobUrl)
        this.rectPdfBlobUrl = ''
      }
      this.rectPreviewVisible = true
      if (this.rectIsPdf) {
        await this.fetchRectPdfAsBlob(this.rectPreviewUrl)
      } else {
        this.$nextTick(() => { this.renderRectDocx(this.rectPreviewUrl) })
      }
    },
    async fetchRectPdfAsBlob(url) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
        this.rectPdfBlobUrl = URL.createObjectURL(blob)
      } catch (e) {
        console.error('PDF加载失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    async renderRectDocx(url) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
        const arrayBuffer = await blob.arrayBuffer()
        const container = this.$refs.rectDocxContainer
        if (container) {
          container.innerHTML = ''
          await renderAsync(arrayBuffer, container)
        }
      } catch (e) {
        console.error('docx渲染失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    clearRectPreview() {
      if (this.$refs.rectDocxContainer) this.$refs.rectDocxContainer.innerHTML = ''
      if (this.rectPdfBlobUrl) {
        URL.revokeObjectURL(this.rectPdfBlobUrl)
        this.rectPdfBlobUrl = ''
      }
    },
    async handleDownloadRect() {
      await this.downloadWithAuth(this.rectDownloadUrl, this.rectFileName)
    }
  }
}
</script>

<style scoped>
.firewall-check {
  margin: 20px;
  padding: 0;
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
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}
.op-btns {
  display: flex;
  gap: 4px;
  justify-content: center;
}
/* DOCX 预览容器样式 */
.docx-preview-container >>> .docx-wrapper {
  background: #fff;
}
.docx-preview-container >>> .docx table {
  border-collapse: collapse;
  width: 100%;
}
.docx-preview-container >>> .docx table td,
.docx-preview-container >>> .docx table th {
  border: 1px solid #ddd;
  padding: 8px;
}
</style>
