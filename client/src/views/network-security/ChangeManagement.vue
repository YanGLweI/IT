<template>
  <div class="change-management">

    <!-- ==================== 区块一：模板管理 ==================== -->
    <el-card style="margin-bottom: 20px">
      <div slot="header" class="page-header">
        <span>变更记录表模板</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-upload2" @click="showTemplateUpload = true">上传新版本</el-button>
        </div>
      </div>

      <!-- 当前版本信息 -->
      <div v-if="currentTemplate" class="current-template-info">
        <div class="current-template-row">
          <span class="label">当前版本：</span>
          <el-tag type="success" size="medium">{{ currentTemplate.version }}</el-tag>
          <span style="margin-left: 12px; color: #909399; font-size: 13px">
            {{ currentTemplate.file_name }} · {{ formatSize(currentTemplate.file_size) }} · {{ formatDate(currentTemplate.created_at) }}
          </span>
          <span v-if="currentTemplate.description" style="margin-left: 12px; color: #606266; font-size: 13px">
            （{{ currentTemplate.description }}）
          </span>
          <el-button type="primary" size="mini" icon="el-icon-view" style="margin-left: 16px" @click="previewTemplate(currentTemplate)">预览</el-button>
          <el-button type="default" size="mini" icon="el-icon-download" @click="downloadTemplate(currentTemplate)">下载当前模板</el-button>
        </div>
      </div>
      <el-empty v-else description="暂无模板，请上传第一个版本" :image-size="60" style="padding: 16px 0" />

      <!-- 历史版本折叠面板 -->
      <el-collapse v-if="templateHistory.length > 0" v-model="templateCollapseActive" style="margin-top: 16px">
        <el-collapse-item title="历史版本" name="history">
          <el-table :data="templateHistory" border size="small">
            <el-table-column type="index" label="序号" width="56" align="center" />
            <el-table-column prop="version" label="版本号" width="110" align="center" />
            <el-table-column prop="description" label="版本说明" min-width="160" show-overflow-tooltip>
              <template slot-scope="{ row }">{{ row.description || '-' }}</template>
            </el-table-column>
            <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
            <el-table-column label="文件大小" width="100" align="center">
              <template slot-scope="{ row }">{{ formatSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column label="上传时间" width="180" align="center">
              <template slot-scope="{ row }">{{ formatDate(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="200" align="center">
              <template slot-scope="{ row }">
                <el-button size="mini" type="text" icon="el-icon-view" @click="previewTemplate(row)">预览</el-button>
                <el-button size="mini" type="text" icon="el-icon-download" @click="downloadTemplate(row)">下载</el-button>
                <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="deleteTemplate(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <!-- ==================== 区块二：扫描件存档 ==================== -->
    <el-card>
      <div slot="header" class="page-header">
        <span>变更记录扫描件</span>
        <div class="page-header-right">
          <el-button type="default" size="small" icon="el-icon-s-operation" @click="showTypeManager = true">类型管理</el-button>
          <el-button type="primary" size="small" icon="el-icon-upload2" @click="openRecordUpload">上传记录</el-button>
          <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchRecords" :loading="recordsLoading">刷新</el-button>
        </div>
      </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="全部年份" size="small" clearable @change="handleRecordFilterChange" style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="y + '年'" :value="y" />
        </el-select>
        <el-select v-model="filterTypeId" placeholder="全部类型" size="small" clearable @change="handleRecordFilterChange" style="width: 150px" multiple collapse-tags>
          <el-option v-for="t in changeTypes" :key="t.id" :label="t.name" :value="t.id" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索描述..." size="small" clearable @keyup.enter.native="handleRecordFilterChange" @clear="handleRecordFilterChange" style="width: 200px" />
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleRecordFilterChange">搜索</el-button>
      </div>

      <!-- 数据表格 -->
      <el-table :data="records" border stripe v-loading="recordsLoading" style="margin-top: 12px">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="year" label="年份" width="80" align="center" />
        <el-table-column prop="month" label="月份" width="80" align="center">
          <template slot-scope="{ row }">{{ row.month }}月</template>
        </el-table-column>
        <el-table-column label="变更类型" width="200" align="center">
          <template slot-scope="{ row }">
            <el-tag v-for="t in (row.change_types || [])" :key="t.id" size="mini" style="margin: 2px">{{ t.name }}</el-tag>
            <span v-if="!row.change_types || row.change_types.length === 0">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
        <el-table-column label="文件大小" width="100" align="center">
          <template slot-scope="{ row }">{{ formatSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column label="申请日期" width="120" align="center">
          <template slot-scope="{ row }">{{ formatDateOnly(row.apply_date) }}</template>
        </el-table-column>
        <el-table-column label="实施日期" width="120" align="center">
          <template slot-scope="{ row }">{{ formatDateOnly(row.implement_date) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template slot-scope="{ row }">
            <div class="op-btns">
              <el-button size="mini" type="text" icon="el-icon-view" @click="previewRecord(row)">预览</el-button>
              <el-button size="mini" type="text" icon="el-icon-download" @click="downloadRecord(row)">下载</el-button>
              <el-button size="mini" type="text" icon="el-icon-edit" @click="editRecord(row)">编辑</el-button>
              <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="deleteRecord(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        style="margin-top: 16px; text-align: right"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="recordsTotal"
        :page-size.sync="recordsPageSize"
        :current-page.sync="recordsPage"
        :page-sizes="[10, 20, 50]"
        @size-change="handleRecordSizeChange"
        @current-change="fetchRecords"
      />
    </el-card>

    <!-- ==================== 弹窗：上传新版本模板 ==================== -->
    <el-dialog title="上传新版本模板" :visible.sync="showTemplateUpload" width="520px" :close-on-click-modal="false">
      <el-form :model="templateForm" ref="templateFormRef" :rules="templateRules" label-width="90px">
        <el-form-item label="版本号" prop="version">
          <el-input v-model="templateForm.version" placeholder="如：IT02-3.0" />
        </el-form-item>
        <el-form-item label="版本说明">
          <el-input v-model="templateForm.description" type="textarea" :rows="2" placeholder="简要说明本次变更内容" />
        </el-form-item>
        <el-form-item label="模板文件" prop="file">
          <el-upload ref="templateUploader" action="" :auto-upload="false" :limit="1" accept=".docx,.pdf"
            :on-change="handleTemplateFileChange" :on-remove="handleTemplateFileRemove" :file-list="templateFileList" drag>
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击选择</em></div>
            <div slot="tip" class="el-upload__tip">支持 DOCX、PDF 格式</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showTemplateUpload = false">取消</el-button>
        <el-button type="primary" :loading="templateUploading" @click="submitTemplateUpload">确定上传</el-button>
      </span>
    </el-dialog>

    <!-- ==================== 弹窗：上传/编辑扫描件 ==================== -->
    <el-dialog :title="recordIsEdit ? '编辑变更记录' : '上传变更记录'" :visible.sync="showRecordUpload" width="520px" :close-on-click-modal="false">
      <el-form :model="recordForm" ref="recordFormRef" :rules="recordRules" label-width="80px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="年份" prop="year">
              <el-input-number v-model="recordForm.year" :min="2020" :max="2100" :step="1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="月份" prop="month">
              <el-select v-model="recordForm.month" placeholder="请选择" style="width: 100%">
                <el-option v-for="m in 12" :key="m" :label="m + '月'" :value="m" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="申请日期" prop="apply_date">
              <el-date-picker v-model="recordForm.apply_date" type="date" value-format="yyyy-MM-dd" placeholder="选择申请日期" :picker-options="applyDatePickerOptions" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="实施日期" prop="implement_date">
              <el-date-picker v-model="recordForm.implement_date" type="date" value-format="yyyy-MM-dd" placeholder="选择实施日期" :picker-options="implementDatePickerOptions" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="变更类型" prop="typeIds">
          <el-select v-model="recordForm.typeIds" multiple collapse-tags placeholder="请选择变更类型" style="width: 100%">
            <el-option v-for="t in changeTypes" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="recordForm.description" type="textarea" :rows="2" placeholder="输入描述，便于后续搜索" />
        </el-form-item>
        <el-form-item label="文件" prop="file" v-if="!recordIsEdit">
          <el-upload ref="recordUploader" action="" :auto-upload="false" :limit="1" accept=".pdf"
            :on-change="handleRecordFileChange" :on-remove="handleRecordFileRemove" :file-list="recordFileList" drag>
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
            <div slot="tip" class="el-upload__tip">仅支持 PDF 格式文件</div>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可修改文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showRecordUpload = false">取消</el-button>
        <el-button type="primary" :loading="recordUploading" @click="submitRecordUpload">{{ recordIsEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- ==================== 弹窗：类型管理 ==================== -->
    <el-dialog title="变更类型管理" :visible.sync="showTypeManager" width="600px">
      <div style="margin-bottom: 12px">
        <el-button type="primary" size="small" icon="el-icon-plus" @click="openTypeDialog()">新增类型</el-button>
      </div>
      <el-table :data="changeTypes" border size="small">
        <el-table-column label="#" width="64" align="center">
          <template slot-scope="{ $index }">
            <div class="sort-btns">
              <el-button size="mini" type="text" icon="el-icon-arrow-up"
                :disabled="$index === 0" @click="moveType(changeTypes[$index], 'up')" />
              <el-button size="mini" type="text" icon="el-icon-arrow-down"
                :disabled="$index === changeTypes.length - 1" @click="moveType(changeTypes[$index], 'down')" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="类型名称" min-width="200" />
        <el-table-column label="操作" width="160" align="center">
          <template slot-scope="{ row }">
            <el-button size="mini" type="text" icon="el-icon-edit" @click="openTypeDialog(row)">编辑</el-button>
            <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="deleteType(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- ==================== 弹窗：新增/编辑类型 ==================== -->
    <el-dialog :title="typeIsEdit ? '编辑变更类型' : '新增变更类型'" :visible.sync="showTypeDialog" width="420px" :close-on-click-modal="false">
      <el-form :model="typeForm" ref="typeFormRef" :rules="typeRules" label-width="80px">
        <el-form-item label="类型名称" prop="name">
          <el-input v-model="typeForm.name" placeholder="请输入类型名称" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showTypeDialog = false">取消</el-button>
        <el-button type="primary" @click="submitType">确定</el-button>
      </span>
    </el-dialog>

    <!-- ==================== 弹窗：模板预览（支持docx/pdf） ==================== -->
    <el-dialog :visible.sync="templatePreviewVisible" width="80%" top="3vh" @closed="clearTemplatePreview">
      <div class="preview-toolbar" slot="title">
        <span>模板预览</span>
        <div class="preview-toolbar-right">
          <el-button type="primary" size="small" icon="el-icon-download" @click="downloadTemplate(templatePreviewRow)">下载</el-button>
        </div>
      </div>
      <div v-if="templatePreviewType === 'pdf'" style="height: 70vh">
        <iframe :src="templatePreviewUrl" style="width: 100%; height: 100%; border: none" />
      </div>
      <div v-else-if="templatePreviewType === 'docx'" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 20px">
        <div ref="templateDocxContainer" class="docx-preview-container"></div>
      </div>
      <div v-else style="text-align: center; padding: 40px">
        <p>该文件格式不支持在线预览</p>
        <el-button type="primary" @click="downloadTemplate(templatePreviewRow)">下载文件</el-button>
      </div>
    </el-dialog>

    <!-- ==================== 弹窗：PDF预览 ==================== -->
    <el-dialog title="文件预览" :visible.sync="previewVisible" width="80%" top="3vh" :close-on-click-modal="true">
      <iframe v-if="previewUrl" :src="previewUrl" style="width: 100%; height: 70vh; border: none;" />
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getChangeTypes, createChangeType, updateChangeType, deleteChangeType, reorderChangeType,
  getChangeRecordTemplates, uploadChangeRecordTemplate, deleteChangeRecordTemplate, getChangeRecordTemplateDownloadUrl, getChangeRecordTemplatePreviewUrl,
  getChangeRecords, createChangeRecord, updateChangeRecord, deleteChangeRecord, getChangeRecordPreviewUrl, getChangeRecordDownloadUrl
} from '@/api/change_record'
import { renderAsync } from 'docx-preview'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'ChangeManagement',
  components: { DualControlDialog },
  data() {
    const now = new Date()
    return {
      // 模板相关
      currentTemplate: null,
      templateHistory: [],
      templateCollapseActive: [],
      showTemplateUpload: false,
      templateUploading: false,
      templateForm: { version: '', description: '' },
      templateRules: {
        version: [{ required: true, message: '请输入版本号', trigger: 'blur' }]
      },
      templateSelectedFile: null,
      templateFileList: [],
      // 模板预览
      templatePreviewVisible: false,
      templatePreviewUrl: '',
      templatePreviewType: '',
      templatePreviewRow: null,
      // 变更类型管理
      changeTypes: [],
      showTypeManager: false,
      showTypeDialog: false,
      typeIsEdit: false,
      editingTypeId: null,
      typeForm: { name: '' },
      typeRules: {
        name: [{ required: true, message: '请输入类型名称', trigger: 'blur' }]
      },
      // 扫描件相关
      records: [],
      recordsLoading: false,
      recordsPage: 1,
      recordsPageSize: 10,
      recordsTotal: 0,
      filterYear: '',
      filterTypeId: [],
      keyword: '',
      yearOptions: Array.from({ length: 10 }, (_, i) => now.getFullYear() - i),
      showRecordUpload: false,
      recordIsEdit: false,
      editingRecordId: null,
      recordUploading: false,
      recordForm: { year: now.getFullYear(), month: now.getMonth() + 1, description: '', typeIds: [], apply_date: '', implement_date: '' },
      recordRules: {
        year: [{ required: true, message: '请选择年份', trigger: 'change' }],
        month: [{ required: true, message: '请选择月份', trigger: 'change' }],
        apply_date: [{ required: true, message: '请选择申请日期', trigger: 'change' }],
        implement_date: [{ required: true, message: '请选择实施日期', trigger: 'change' }],
        typeIds: [{ required: true, message: '请选择变更类型', trigger: 'change', type: 'array', min: 1 }]
      },
      recordSelectedFile: null,
      recordFileList: [],
      // 预览
      previewVisible: false,
      previewUrl: ''
    }
  },
  computed: {
    // 申请日期：只能选择当月日期
    applyDatePickerOptions() {
      const year = this.recordForm.year
      const month = this.recordForm.month
      return {
        disabledDate(date) {
          return date.getFullYear() !== year || date.getMonth() !== month - 1
        }
      }
    },
    // 实施日期：不能早于所选月份（从当月1号起可选）
    implementDatePickerOptions() {
      const year = this.recordForm.year
      const month = this.recordForm.month
      const minDate = new Date(year, month - 1, 1)
      minDate.setHours(0, 0, 0, 0)
      return {
        disabledDate(date) {
          return date < minDate
        }
      }
    }
  },
  watch: {
    'recordForm.year'() {
      this.clearInvalidRecordDates()
    },
    'recordForm.month'() {
      this.clearInvalidRecordDates()
    }
  },
  mounted() {
    this.fetchChangeTypes()
    this.fetchTemplates()
    this.fetchRecords()
  },
  methods: {
    // ============ 变更类型管理 ============
    async fetchChangeTypes() {
      try {
        const res = await getChangeTypes()
        this.changeTypes = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    openTypeDialog(row) {
      if (row) {
        this.typeIsEdit = true
        this.editingTypeId = row.id
        this.typeForm = { name: row.name }
      } else {
        this.typeIsEdit = false
        this.editingTypeId = null
        this.typeForm = { name: '' }
      }
      this.showTypeDialog = true
    },
    submitType() {
      this.$refs.typeFormRef.validate(async valid => {
        if (!valid) return
        try {
          const dualToken = await this.$refs.dualControl.open()
          if (this.typeIsEdit) {
            await updateChangeType(this.editingTypeId, this.typeForm, dualToken)
            this.$message.success('更新成功')
          } else {
            await createChangeType(this.typeForm, dualToken)
            this.$message.success('创建成功')
          }
          this.showTypeDialog = false
          this.fetchChangeTypes()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        }
      })
    },
    async deleteType(row) {
      try {
        await this.$confirm(`确定要删除类型“${row.name}”吗？`, '删除确认', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteChangeType(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchChangeTypes()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    async moveType(row, direction) {
      try {
        await reorderChangeType({ id: row.id, direction })
        await this.fetchChangeTypes()
      } catch (e) {
        this.$message.error(e.response?.data?.message || '移动失败')
      }
    },

    // ============ 模板管理 ============
    async fetchTemplates() {
      try {
        const res = await getChangeRecordTemplates()
        const list = res.data || []
        this.currentTemplate = list.find(t => t.is_current) || null
        this.templateHistory = list.filter(t => !t.is_current)
      } catch (e) {
        console.error(e)
      }
    },
    handleTemplateFileChange(file) {
      this.templateSelectedFile = file.raw
    },
    handleTemplateFileRemove() {
      this.templateSelectedFile = null
    },
    submitTemplateUpload() {
      this.$refs.templateFormRef.validate(async valid => {
        if (!valid) return
        if (!this.templateSelectedFile) {
          this.$message.warning('请选择模板文件')
          return
        }
        this.templateUploading = true
        try {
          const formData = new FormData()
          formData.append('version', this.templateForm.version)
          formData.append('description', this.templateForm.description || '')
          formData.append('file', this.templateSelectedFile)
          const dualToken = await this.$refs.dualControl.open()
          await uploadChangeRecordTemplate(formData, dualToken)
          this.$message.success('上传成功')
          this.showTemplateUpload = false
          this.templateForm = { version: '', description: '' }
          this.templateSelectedFile = null
          this.templateFileList = []
          if (this.$refs.templateUploader) this.$refs.templateUploader.clearFiles()
          this.fetchTemplates()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        } finally {
          this.templateUploading = false
        }
      })
    },
    async downloadTemplate(row) {
      const url = getChangeRecordTemplateDownloadUrl(row.id)
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
    async previewTemplate(row) {
      const url = getChangeRecordTemplatePreviewUrl(row.id)
      const fileName = (row.file_name || '').toLowerCase()
      this.templatePreviewRow = row
      this.templatePreviewUrl = ''

      if (fileName.endsWith('.pdf')) {
        this.templatePreviewType = 'pdf'
      } else if (fileName.endsWith('.docx')) {
        this.templatePreviewType = 'docx'
      } else {
        this.templatePreviewType = 'other'
      }

      this.templatePreviewVisible = true

      try {
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.templatePreviewUrl = URL.createObjectURL(blob)

        if (this.templatePreviewType === 'docx') {
          this.$nextTick(() => {
            this.renderTemplateDocx(blob)
          })
        }
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('模板预览失败')
      }
    },
    async renderTemplateDocx(blob) {
      try {
        const arrayBuffer = await blob.arrayBuffer()
        const container = this.$refs.templateDocxContainer
        if (container) {
          container.innerHTML = ''
          await renderAsync(arrayBuffer, container)
        }
      } catch (e) {
        console.error('docx渲染失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    clearTemplatePreview() {
      if (this.$refs.templateDocxContainer) {
        this.$refs.templateDocxContainer.innerHTML = ''
      }
      if (this.templatePreviewUrl) {
        URL.revokeObjectURL(this.templatePreviewUrl)
        this.templatePreviewUrl = ''
      }
      this.templatePreviewRow = null
    },
    async deleteTemplate(row) {
      try {
        await this.$confirm(`确定要删除版本 ${row.version} 吗？`, '删除确认', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteChangeRecordTemplate(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchTemplates()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },

    // ============ 扫描件存档 ============
    async fetchRecords() {
      this.recordsLoading = true
      try {
        const params = { page: this.recordsPage, page_size: this.recordsPageSize }
        if (this.filterYear) params.year = this.filterYear
        if (this.filterTypeId && this.filterTypeId.length > 0) params.type_id = this.filterTypeId.join(',')
        if (this.keyword) params.keyword = this.keyword
        const res = await getChangeRecords(params)
        this.records = res.data || []
        this.recordsTotal = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.recordsLoading = false
      }
    },
    handleRecordSizeChange() {
      this.recordsPage = 1
      this.fetchRecords()
    },
    handleRecordFilterChange() {
      this.recordsPage = 1
      this.fetchRecords()
    },
    openRecordUpload() {
      this.recordIsEdit = false
      this.editingRecordId = null
      const now = new Date()
      this.recordForm = { year: now.getFullYear(), month: now.getMonth() + 1, description: '', typeIds: [], apply_date: '', implement_date: '' }
      this.recordSelectedFile = null
      this.recordFileList = []
      this.showRecordUpload = true
    },
    handleRecordFileChange(file) {
      this.recordSelectedFile = file.raw
    },
    handleRecordFileRemove() {
      this.recordSelectedFile = null
    },
    editRecord(row) {
      this.recordIsEdit = true
      this.editingRecordId = row.id
      const typeIds = (row.change_types || []).map(t => t.id)
      this.recordForm = { year: row.year, month: row.month, description: row.description || '', typeIds, apply_date: row.apply_date ? row.apply_date.substring(0, 10) : '', implement_date: row.implement_date ? row.implement_date.substring(0, 10) : '' }
      this.showRecordUpload = true
    },
    submitRecordUpload() {
      this.$refs.recordFormRef.validate(async valid => {
        if (!valid) return
        if (!this.recordIsEdit && !this.recordSelectedFile) {
          this.$message.warning('请选择PDF文件')
          return
        }
        this.recordUploading = true
        try {
          const formData = new FormData()
          formData.append('year', this.recordForm.year)
          formData.append('month', this.recordForm.month)
          formData.append('description', this.recordForm.description || '')
          formData.append('apply_date', this.recordForm.apply_date || '')
          formData.append('implement_date', this.recordForm.implement_date || '')
          formData.append('type_ids', (this.recordForm.typeIds || []).join(','))
          if (this.recordSelectedFile) formData.append('file', this.recordSelectedFile)

          const dualToken = await this.$refs.dualControl.open()
          if (this.recordIsEdit) {
            await updateChangeRecord(this.editingRecordId, formData, dualToken)
            this.$message.success('更新成功')
          } else {
            await createChangeRecord(formData, dualToken)
            this.$message.success('上传成功')
          }
          this.showRecordUpload = false
          this.fetchRecords()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        } finally {
          this.recordUploading = false
        }
      })
    },
    async previewRecord(row) {
      const url = getChangeRecordPreviewUrl(row.id)
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
    async downloadRecord(row) {
      const url = getChangeRecordDownloadUrl(row.id)
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
    async deleteRecord(row) {
      try {
        await this.$confirm(`确定要删除 ${row.year}年${row.month}月 的变更记录吗？`, '删除确认', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteChangeRecord(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchRecords()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },

    // ============ 工具函数 ============
    formatSize(bytes) {
      if (!bytes || bytes === 0) return '0 B'
      const units = ['B', 'KB', 'MB', 'GB']
      let i = 0, size = bytes
      while (size >= 1024 && i < units.length - 1) { size /= 1024; i++ }
      return size.toFixed(i === 0 ? 0 : 1) + ' ' + units[i]
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      return dateStr.replace('T', ' ').substring(0, 19)
    },
    formatDateOnly(dateStr) {
      if (!dateStr) return '-'
      return dateStr.substring(0, 10)
    },
    // 年月变化时清除不在范围内的日期
    clearInvalidRecordDates() {
      const year = this.recordForm.year
      const month = this.recordForm.month
      // 清除不在当月的申请日期
      if (this.recordForm.apply_date) {
        const d = new Date(this.recordForm.apply_date)
        if (d.getFullYear() !== year || d.getMonth() !== month - 1) {
          this.recordForm.apply_date = ''
        }
      }
      // 清除早于当月的实施日期
      if (this.recordForm.implement_date) {
        const d = new Date(this.recordForm.implement_date)
        const minDate = new Date(year, month - 1, 1)
        if (d < minDate) {
          this.recordForm.implement_date = ''
        }
      }
    }
  }
}
</script>

<style scoped>
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
.sort-btns {
  display: flex;
  gap: 2px;
  justify-content: center;
}
.current-template-info {
  background: #f0f9eb;
  border: 1px solid #e1f3d8;
  border-radius: 4px;
  padding: 12px 16px;
}
.current-template-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}
.current-template-row .label {
  font-size: 14px;
  color: #606266;
  font-weight: bold;
}
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
</style>
