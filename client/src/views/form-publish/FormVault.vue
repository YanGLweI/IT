<template>
  <div class="form-vault-page">
    <el-card>
      <div slot="header" style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 10px;">
        <span>表单发布 — 保管区</span>
        <div style="display: flex; gap: 8px;">
          <el-button type="primary" size="small" icon="el-icon-upload2" @click="showUploadDialog">上传表单</el-button>
          <el-button size="small" icon="el-icon-link" @click="showCrossModuleDialog">跨模块引用</el-button>
        </div>
      </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-input v-model="filters.keyword" placeholder="搜索标题或描述..." prefix-icon="el-icon-search" clearable style="width: 220px;" @input="fetchItems" />
        <el-select v-model="filters.category" placeholder="全部分类" clearable style="width: 130px;" @change="fetchItems">
          <el-option v-for="cat in categoryList" :key="cat" :label="cat" :value="cat" />
        </el-select>
        <el-select v-model="filters.source_type" placeholder="全部来源" clearable style="width: 130px;" @change="fetchItems">
          <el-option label="直接上传" value="upload" />
          <el-option label="静态引用" value="static" />
          <el-option label="动态生成" value="dynamic" />
        </el-select>
        <el-select v-model="filters.is_published" placeholder="全部状态" clearable style="width: 130px;" @change="fetchItems">
          <el-option label="已发布" value="true" />
          <el-option label="未发布" value="false" />
        </el-select>
      </div>

      <!-- 表格 -->
      <el-table :data="items" border stripe v-loading="loading">
        <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column label="来源" width="90" align="center">
          <template slot-scope="scope">
            <el-tag :type="sourceTagType(scope.row.source_type)" size="mini">
              {{ sourceLabel(scope.row.source_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template slot-scope="scope">
            <el-tag :type="scope.row.is_published ? 'success' : 'info'" size="mini">
              {{ scope.row.is_published ? '已发布' : '未发布' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="文件" width="160" show-overflow-tooltip>
          <template slot-scope="scope">
            <span v-if="scope.row.source_type !== 'dynamic'">{{ scope.row.file_name || '-' }}</span>
            <span v-else style="color: #94A3B8; font-size: 12px;">动态生成</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template slot-scope="scope">{{ formatDate(scope.row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="380" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button size="mini" @click="handlePreview(scope.row)">预览</el-button>
            <el-button size="mini" @click="handleDownload(scope.row)">下载</el-button>
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button
              size="mini"
              :type="scope.row.is_published ? 'warning' : 'success'"
              @click="handleTogglePublish(scope.row)"
            >
              {{ scope.row.is_published ? '取消发布' : '发布' }}
            </el-button>
            <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap" v-if="total > pageSize">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page.sync="currentPage"
          @current-change="fetchItems"
        />
      </div>
    </el-card>

    <!-- 上传弹窗 -->
    <el-dialog title="上传新表单" :visible.sync="uploadVisible" width="520px">
      <el-form :model="uploadForm" :rules="uploadRules" ref="uploadFormRef" label-width="80px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="uploadForm.title" placeholder="请输入表单标题" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="uploadForm.description" type="textarea" :rows="2" placeholder="请输入描述（可选）" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="uploadForm.category" placeholder="如：模板、政策、检查表" />
        </el-form-item>
        <el-form-item label="文件" prop="file">
          <el-upload
            ref="uploader"
            action=""
            :auto-upload="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="uploadFileList"
            :limit="1"
            drag
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">将文件拖到此处，或<em>点击选择</em></div>
            <div class="el-upload__tip" slot="tip">支持 DOCX、PDF、XLSX、XLS、DOC 格式</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </span>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog title="编辑表单信息" :visible.sync="editVisible" width="500px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="editForm.category" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSaving" @click="handleEditSubmit">保存</el-button>
      </span>
    </el-dialog>

    <!-- 跨模块引用弹窗 -->
    <el-dialog title="从其他模块引用" :visible.sync="crossModuleVisible" width="560px">
      <el-form :model="crossForm" label-width="90px">
        <el-form-item label="引用方式">
          <el-radio-group v-model="crossForm.source_type" @change="onCrossSourceTypeChange">
            <el-radio label="static">静态快照（复制文件）</el-radio>
            <el-radio label="dynamic">动态实时（实时生成）</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择模块">
          <el-select v-model="crossForm.module_key" placeholder="请选择模块" @change="onCrossModuleChange" style="width: 100%;">
            <el-option
              v-for="src in filteredSources"
              :key="src.module_key"
              :label="src.module_name"
              :value="src.module_key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="选择文件" v-if="crossForm.source_type === 'static' && crossFiles.length > 0">
          <el-table :data="crossFiles" highlight-current-row @current-change="onCrossFileSelect" size="small" max-height="200">
            <el-table-column prop="name" label="名称" show-overflow-tooltip />
            <el-table-column prop="file_name" label="文件名" width="180" show-overflow-tooltip />
          </el-table>
        </el-form-item>
        
        <!-- 动态类型的参数输入 -->
        <div v-if="crossForm.source_type === 'dynamic' && generatorParams.length > 0" style="margin-top: 10px;">
          <el-divider content-position="left">参数配置</el-divider>
          <el-form-item 
            v-for="param in generatorParams" 
            :key="param.name" 
            :label="param.label"
          >
            <!-- select 类型 -->
            <el-select 
              v-if="param.type === 'select'" 
              v-model="paramValues[param.name]" 
              placeholder="请选择" 
              style="width: 100%"
              @change="onParamValueChange"
            >
              <el-option 
                v-for="opt in paramOptions[param.name]" 
                :key="opt.value" 
                :label="opt.label" 
                :value="opt.value" 
              />
            </el-select>
            
            <!-- input 类型 -->
            <el-input 
              v-else-if="param.type === 'input'" 
              v-model="paramValues[param.name]" 
              :placeholder="'请输入' + param.label" 
            />
          </el-form-item>
        </div>
        
        <el-form-item label="标题">
          <el-input v-model="crossForm.title" placeholder="表单标题" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="crossForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="crossForm.category" placeholder="如：模板" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="crossModuleVisible = false">取消</el-button>
        <el-button type="primary" :loading="crossSubmitting" @click="handleCrossModuleSubmit">确认引用</el-button>
      </span>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog title="文件预览" :visible.sync="previewVisible" width="80%" top="5vh" @closed="clearDocxPreview">
      <div class="preview-toolbar" slot="title">
        <span>文件预览</span>
        <div class="preview-toolbar-right">
          <el-button type="primary" size="small" icon="el-icon-download" @click="downloadFromPreview">下载</el-button>
        </div>
      </div>
      <div v-if="previewType === 'pdf'" style="height: 70vh">
        <iframe v-if="previewUrl" :src="previewUrl" style="width: 100%; height: 100%; border: none;"></iframe>
      </div>
      <div v-else-if="previewType === 'image'" style="text-align: center">
        <img v-if="previewUrl" :src="previewUrl" style="max-width: 100%; max-height: 70vh;" />
      </div>
      <div v-else-if="previewType === 'docx'" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 20px;">
        <div ref="docxContainer" class="docx-preview-container"></div>
      </div>
      <div v-else style="text-align: center; padding: 40px;">
        <p>该文件格式不支持在线预览</p>
        <el-button type="primary" @click="downloadFromPreview">下载文件</el-button>
      </div>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getFormVaultItems,
  uploadFormVaultItem,
  updateFormVaultItem,
  deleteFormVaultItem,
  publishFormVaultItem,
  unpublishFormVaultItem,
  getCrossModuleSources,
  getCrossModuleFiles,
  getGeneratorParams,
  createCrossModuleRef
} from '@/api/form_vault'
import { getDepartments } from '@/api/department'
import { renderAsync } from 'docx-preview'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'FormVault',
  components: { DualControlDialog },
  data() {
    return {
      items: [],
      total: 0,
      currentPage: 1,
      pageSize: 20,
      loading: false,
      categoryList: [],
      filters: {
        keyword: '',
        category: '',
        source_type: '',
        is_published: ''
      },
      // 上传
      uploadVisible: false,
      uploadForm: { title: '', description: '', category: '' },
      uploadRules: {
        title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
      },
      uploadFileList: [],
      uploadFile: null,
      uploading: false,
      // 编辑
      editVisible: false,
      editForm: { id: null, title: '', description: '', category: '' },
      editSaving: false,
      // 跨模块引用
      crossModuleVisible: false,
      crossForm: {
        source_type: 'static',
        module_key: '',
        ref_id: null,
        title: '',
        description: '',
        category: ''
      },
      crossSources: [],
      crossFiles: [],
      crossSelectedFile: null,
      crossSubmitting: false,
      // 动态生成器参数
      generatorParams: [],      // 当前选中生成器的参数定义
      paramValues: {},          // 用户填写的参数值
      paramOptions: {},         // 存储各参数的下拉选项 { paramName: [{label, value}] }
      // 预览
      previewVisible: false,
      previewUrl: '',
      previewType: '',
      previewFileName: '',
      previewRowId: null,
      previewBlob: null
    }
  },
  computed: {
    filteredSources() {
      return this.crossSources.filter(s => s.source_type === this.crossForm.source_type)
    }
  },
  created() {
    this.fetchItems()
    this.fetchCrossSources()
  },
  methods: {
    async fetchItems() {
      this.loading = true
      try {
        const params = {
          page: this.currentPage,
          page_size: this.pageSize
        }
        if (this.filters.keyword) params.keyword = this.filters.keyword
        if (this.filters.category) params.category = this.filters.category
        if (this.filters.source_type) params.source_type = this.filters.source_type
        if (this.filters.is_published) params.is_published = this.filters.is_published

        const res = await getFormVaultItems(params)
        // 后端返回结构: { code, data: [...], page_size, total }
        this.items = res.data || []
        this.total = res.total || 0

        // 收集分类
        const cats = new Set()
        this.items.forEach(i => { if (i.category) cats.add(i.category) })
        this.categoryList = Array.from(cats)
      } catch (e) {
        console.error('获取表单列表失败:', e)
      } finally {
        this.loading = false
      }
    },
    async fetchCrossSources() {
      try {
        const res = await getCrossModuleSources()
        this.crossSources = res.data || []
      } catch (e) {
        console.error('获取跨模块源失败:', e)
      }
    },

    // ---- 上传 ----
    showUploadDialog() {
      this.uploadForm = { title: '', description: '', category: '' }
      this.uploadFileList = []
      this.uploadFile = null
      this.uploadVisible = true
    },
    handleFileChange(file) {
      this.uploadFile = file.raw
      this.uploadFileList = [file]
    },
    handleFileRemove() {
      this.uploadFile = null
      this.uploadFileList = []
    },
    handleUpload() {
      this.$refs.uploadFormRef.validate(async valid => {
        if (!valid) return
        if (!this.uploadFile) {
          this.$message.warning('请选择文件')
          return
        }

        this.uploading = true
        try {
          const fd = new FormData()
          fd.append('title', this.uploadForm.title)
          fd.append('description', this.uploadForm.description)
          fd.append('category', this.uploadForm.category)
          fd.append('file', this.uploadFile)

          const dualToken = await this.$refs.dualControl.open()
          await uploadFormVaultItem(fd, dualToken)
          this.$message.success('上传成功')
          this.uploadVisible = false
          this.fetchItems()
        } catch (e) {
          // 双控取消或上传失败
        } finally {
          this.uploading = false
        }
      })
    },

    // ---- 编辑 ----
    handleEdit(row) {
      this.editForm = {
        id: row.id,
        title: row.title,
        description: row.description,
        category: row.category
      }
      this.editVisible = true
    },
    async handleEditSubmit() {
      this.editSaving = true
      try {
        const params = new URLSearchParams()
        params.append('title', this.editForm.title)
        params.append('description', this.editForm.description || '')
        params.append('category', this.editForm.category || '')

        const dualToken = await this.$refs.dualControl.open()
        await updateFormVaultItem(this.editForm.id, params, dualToken)
        this.$message.success('更新成功')
        this.editVisible = false
        this.fetchItems()
      } catch (e) {
        // 双控取消或更新失败
      } finally {
        this.editSaving = false
      }
    },

    // ---- 删除 ----
    async handleDelete(row) {
      try {
        await this.$confirm(`确定要删除表单「${row.title}」吗？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteFormVaultItem(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchItems()
      } catch (e) {
        // 双控取消或删除失败
      }
    },

    // ---- 发布/取消发布 ----
    async handleTogglePublish(row) {
      const action = row.is_published ? '取消发布' : '发布'
      try {
        await this.$confirm(`确定要${action}表单「${row.title}」吗？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        if (row.is_published) {
          await unpublishFormVaultItem(row.id, dualToken)
        } else {
          await publishFormVaultItem(row.id, dualToken)
        }
        this.$message.success(`${action}成功`)
        this.fetchItems()
      } catch (e) {
        // 双控取消或操作失败
      }
    },

    // ---- 预览/下载 ----
    async handlePreview(row) {
      const fileName = (row.file_name || '').toLowerCase()
      this.previewFileName = row.file_name || '文件'
      this.previewRowId = row.id
      // 检测文件类型
      if (fileName.endsWith('.pdf')) {
        this.previewType = 'pdf'
      } else if (fileName.endsWith('.png') || fileName.endsWith('.jpg') || fileName.endsWith('.jpeg') || fileName.endsWith('.gif') || fileName.endsWith('.bmp')) {
        this.previewType = 'image'
      } else if (fileName.endsWith('.docx') || fileName.endsWith('.doc')) {
        this.previewType = 'docx'
      } else {
        this.previewType = 'other'
      }
      this.previewVisible = true
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(`/api/form-vault/${row.id}/preview`, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.previewBlob = blob
        if (this.previewUrl) URL.revokeObjectURL(this.previewUrl)
        this.previewUrl = URL.createObjectURL(blob)
        if (this.previewType === 'docx') {
          this.$nextTick(() => {
            this.renderDocxFromBlob(blob)
          })
        }
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('预览失败')
      }
    },
    async renderDocxFromBlob(blob) {
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
    clearDocxPreview() {
      if (this.$refs.docxContainer) {
        this.$refs.docxContainer.innerHTML = ''
      }
      if (this.previewUrl) {
        URL.revokeObjectURL(this.previewUrl)
        this.previewUrl = ''
      }
      this.previewBlob = null
    },
    async downloadFromPreview() {
      if (!this.previewRowId) return
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(`/api/form-vault/${this.previewRowId}/download`, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = this.previewFileName
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(link.href)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      }
    },
    async handleDownload(row) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(`/api/form-vault/${row.id}/download`, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = row.file_name || 'download'
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(link.href)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      }
    },

    // ---- 跨模块引用 ----
    showCrossModuleDialog() {
      this.crossForm = {
        source_type: 'static',
        module_key: '',
        ref_id: null,
        title: '',
        description: '',
        category: ''
      }
      this.crossFiles = []
      this.crossSelectedFile = null
      this.crossModuleVisible = true
    },
    onCrossSourceTypeChange() {
      this.crossForm.module_key = ''
      this.crossFiles = []
      this.crossSelectedFile = null
    },
    async onCrossModuleChange(moduleKey) {
      this.crossFiles = []
      this.crossSelectedFile = null
      this.generatorParams = []
      this.paramValues = {}
      if (!moduleKey) return
      
      // 如果是动态类型，获取生成器参数定义
      if (this.crossForm.source_type === 'dynamic') {
        await this.fetchGeneratorParams(moduleKey)
      } else {
        try {
          const res = await getCrossModuleFiles(moduleKey)
          this.crossFiles = res.data || []
          // 动态类型自动填充标题
          if (this.crossForm.source_type === 'dynamic' && this.crossFiles.length > 0) {
            this.crossForm.title = this.crossForm.title || this.crossFiles[0].name
          }
        } catch (e) {
          console.error('获取文件列表失败:', e)
        }
      }
    },
    
    // 获取动态生成器的参数定义
    async fetchGeneratorParams(moduleKey) {
      // 映射 moduleKey 到 generatorName
      const moduleToGeneratorMap = {
        'user_change_record': 'export_user_change_record',
        'department_confirmation': 'export_department_confirmation'
      }
      
      const generatorName = moduleToGeneratorMap[moduleKey]
      if (!generatorName) {
        console.warn(`未找到模块 ${moduleKey} 对应的生成器`)
        return
      }
      
      try {
        const res = await getGeneratorParams(generatorName)
        this.generatorParams = res.data || []
        
        // 为 select 类型的参数加载选项
        for (const param of this.generatorParams) {
          if (param.type === 'select' && param.source) {
            await this.loadParamOptions(param)
          }
        }
      } catch (e) {
        console.error('获取生成器参数失败:', e)
      }
    },
    
    // 加载参数选项（如部门列表）
    async loadParamOptions(param) {
      if (param.source === '/api/departments') {
        try {
          const res = await getDepartments()
          this.$set(this.paramOptions, param.name, (res.data || []).map(d => ({
            label: d.name,
            value: d.id
          })))
        } catch (e) {
          console.error(`加载参数 ${param.name} 的选项失败:`, e)
        }
      }
    },
    
    // 参数值变化时的处理
    onParamValueChange() {
      // 可选：根据参数值自动更新标题
      // 例如：选择部门后，标题变为 "部门用户确认表 - {部门名}"
    },
    onCrossFileSelect(row) {
      this.crossSelectedFile = row
      if (row) {
        this.crossForm.ref_id = row.id
        if (!this.crossForm.title) {
          this.crossForm.title = row.name || row.file_name
        }
      }
    },
    async handleCrossModuleSubmit() {
      if (!this.crossForm.module_key) {
        this.$message.warning('请选择模块')
        return
      }
      if (!this.crossForm.title) {
        this.$message.warning('请输入标题')
        return
      }
      if (this.crossForm.source_type === 'static' && !this.crossSelectedFile) {
        this.$message.warning('请选择文件')
        return
      }
      
      // 验证动态类型的必填参数
      if (this.crossForm.source_type === 'dynamic') {
        for (const param of this.generatorParams) {
          if (param.required && !this.paramValues[param.name]) {
            this.$message.warning(`请填写必填参数：${param.label}`)
            return
          }
        }
      }

      this.crossSubmitting = true
      try {
        const fd = new FormData()
        fd.append('title', this.crossForm.title)
        fd.append('description', this.crossForm.description || '')
        fd.append('category', this.crossForm.category || '')
        fd.append('module_key', this.crossForm.module_key)
        fd.append('source_type', this.crossForm.source_type)
        if (this.crossForm.source_type === 'static' && this.crossSelectedFile) {
          fd.append('ref_id', this.crossSelectedFile.id)
        }
        
        // 添加动态生成器参数
        if (this.crossForm.source_type === 'dynamic' && Object.keys(this.paramValues).length > 0) {
          fd.append('ref_params', JSON.stringify(this.paramValues))
        }

        const dualToken = await this.$refs.dualControl.open()
        await createCrossModuleRef(fd, dualToken)
        this.$message.success('创建引用成功')
        this.crossModuleVisible = false
        this.fetchItems()
      } catch (e) {
        // 双控取消或创建失败
      } finally {
        this.crossSubmitting = false
      }
    },

    // ---- 工具方法 ----
    sourceLabel(type) {
      return { upload: '上传', static: '静态', dynamic: '动态' }[type] || type
    },
    sourceTagType(type) {
      return { upload: '', static: 'warning', dynamic: 'danger' }[type] || ''
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const d = new Date(dateStr)
      return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
    }
  }
}
</script>

<style scoped>
.form-vault-page {
  /* 使用全局 padding */
}

.filter-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.pagination-wrap {
  margin-top: 16px;
  text-align: right;
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
