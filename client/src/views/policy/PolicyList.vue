<template>
  <div class="policy-list">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">IT政策管理</h2>
        <p class="page-subtitle">管理公司IT政策文件，支持在线预览与下载</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="uploadVisible = true">上传政策</el-button>
      </div>
    </div>

      <div class="table-card">
        <el-table :data="policies"  stripe>
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="description" label="描述" show-overflow-tooltip />
          <el-table-column prop="file_name" label="文件名" />
          <el-table-column label="文件大小" width="120">
            <template slot-scope="scope">{{ formatSize(scope.row.file_size) }}</template>
          </el-table-column>
          <el-table-column label="上传时间" width="180">
            <template slot-scope="scope">{{ formatDate(scope.row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="280" align="center" fixed="right">
            <template slot-scope="scope">
              <el-button size="mini" @click="handlePreview(scope.row)">预览</el-button>
              <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
              <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

    <!-- 上传弹窗 -->
    <el-dialog title="上传政策文件" :visible.sync="uploadVisible" width="500px" class="vault-dialog">
      <el-form :model="uploadForm" :rules="uploadRules" ref="uploadFormRef" label-width="80px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="uploadForm.title" placeholder="请输入政策标题" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="uploadForm.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="文件">
          <el-upload
            ref="uploader"
            action=""
            :auto-upload="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            drag
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">将文件拖到此处，或<em>点击选择</em></div>
            <div class="el-upload__tip" slot="tip">支持 PDF、DOC、DOCX、XLS、XLSX 等格式</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </span>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog title="编辑政策" :visible.sync="editVisible" width="500px" class="vault-dialog">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditSubmit">保存</el-button>
      </span>
    </el-dialog>

    <!-- 预览弹窗 - 使用 file-viewer -->
    <el-dialog title="文件预览" :visible.sync="previewVisible" width="90%" top="2vh" class="vault-dialog preview-dialog fv-preview-dialog">
      <div class="fv-container" style="height: 80vh">
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
          <el-button type="primary" @click="downloadFile">下载文件</el-button>
        </div>
      </div>
    </el-dialog>
    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getPolicies, createPolicy, updatePolicy, deletePolicy, getPolicyPreviewUrl, getPolicyDownloadUrl } from '@/api/policy'
import { FileViewer } from '@file-viewer/vue2.7'
import officePreset from '@file-viewer/preset-office'
import DualControlDialog from '@/components/DualControlDialog.vue'
import fvFeature from '@/config/fv-feature'

export default {
  components: { DualControlDialog, FileViewer },
  name: 'PolicyList',
  data() {
    return {
      policies: [],
      uploadVisible: false,
      uploading: false,
      uploadForm: { title: '', description: '' },
      uploadRules: {
        title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
      },
      fileList: [],
      selectedFile: null,
      editVisible: false,
      editForm: { id: null, title: '', description: '' },
      previewVisible: false,
      currentFileUrl: '',
      currentFileName: '',
      currentFileType: '',
      fvContainerHeight: '80vh',
      previewFileName: '',
      previewId: null,
      officePreset: officePreset,
      fvFeature: fvFeature
    }
  },
  mounted() {
    this.fetchData()
  },
  computed: {
    fvOptions() {
      return {
        preset: officePreset,
        fetchFile: this.fetchFileWithAuth
      }
    }
  },
  methods: {
    async fetchData() {
      try {
        const res = await getPolicies()
        this.policies = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    handleFileChange(file) {
      this.selectedFile = file.raw
      this.fileList = [file]
    },
    handleFileRemove() {
      this.selectedFile = null
      this.fileList = []
    },
    handleUpload() {
      if (!this.$refs.uploadFormRef) {
        console.error('表单引用未就绪')
        return
      }
      this.$refs.uploadFormRef.validate(valid => {
        if (!valid) {
          console.log('表单验证未通过')
          return
        }
        if (!this.selectedFile) {
          this.$message.warning('请选择文件')
          return
        }
        this.uploading = true
        const formData = new FormData()
        formData.append('title', this.uploadForm.title)
        formData.append('description', this.uploadForm.description)
        formData.append('file', this.selectedFile)
        createPolicy(formData).then(() => {
          this.$message.success('上传成功')
          this.uploadVisible = false
          this.uploadForm = { title: '', description: '' }
          this.selectedFile = null
          this.fileList = []
          this.fetchData()
        }).catch(e => {
          console.error('上传失败:', e)
          this.$message.error('上传失败')
        }).finally(() => {
          this.uploading = false
        })
      })
    },
    handleEdit(row) {
      this.editForm = { id: row.id, title: row.title, description: row.description }
      this.editVisible = true
    },
    async handleEditSubmit() {
      try {
        const dualToken = await this.$refs.dualControl.open()
        await updatePolicy(this.editForm.id, {
          title: this.editForm.title,
          description: this.editForm.description
        }, dualToken)
        this.$message.success('更新成功')
        this.editVisible = false
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    async handleDelete(row) {
      try {
        await this.$confirm('确定要删除该政策吗？', '提示', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deletePolicy(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    async handlePreview(row) {
      const url = getPolicyPreviewUrl(row.id)
      
      // 从文件名中提取扩展名
      const fileExtension = row.file_name ? row.file_name.split('.').pop().toLowerCase() : ''
      
      this.currentFileUrl = url
      this.currentFileName = row.file_name || 'unknown_file'
      this.currentFileType = this.getFileTypeFromExtension(fileExtension)
      
      this.previewFileName = row.file_name || '文件'
      this.previewId = row.id
      
      // 先关闭再打开，确保 FileViewer 组件重新渲染
      this.previewVisible = false
      await this.$nextTick()
      this.previewVisible = true
      await this.$nextTick()
    },
    getFileTypeFromExtension(ext) {
      return ext || ''
    },
    async fetchFileWithAuth({ url }) {
      const token = localStorage.getItem('token')
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (!response.ok) {
        throw new Error(`Failed to fetch file: ${response.status} ${response.statusText}`)
      }
      return response.arrayBuffer()
    },
    async downloadFile() {
      if (this.previewId) {
        const url = getPolicyDownloadUrl(this.previewId)
        try {
          const response = await fetch(url, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
          })
          if (!response.ok) throw new Error('下载失败')
          const blob = await response.blob()
          const downloadUrl = URL.createObjectURL(blob)
          const link = document.createElement('a')
          link.href = downloadUrl
          link.download = this.previewFileName
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          URL.revokeObjectURL(downloadUrl)
        } catch (e) {
          console.error('下载失败:', e)
          this.$message.error('下载失败')
        }
      }
    },
    formatSize(size) {
      if (!size) return '-'
      if (size < 1024) return size + ' B'
      if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
      return (size / (1024 * 1024)).toFixed(1) + ' MB'
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      return new Date(dateStr).toLocaleString('zh-CN')
    }
  }
}
</script>

<style scoped>
/* ========== 页面容器 ========== */
.policy-list {
  padding: 24px;
  height: calc(100% - 85px);
  overflow-y: auto;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  margin: 20px;
}

/* ========== 页面头部 ========== */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
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
  margin: 4px 0 0;
}
.header-actions {
  display: flex;
  gap: 10px;
}

/* 头部按钮 */
.page-header ::v-deep .el-button--primary {
  padding: 9px 18px;
  background: #3b82f6;
  border-color: #3b82f6;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  transition: background 0.2s;
}
.page-header ::v-deep .el-button--primary:hover {
  background: #2563eb;
  border-color: #2563eb;
}

/* ========== 预览弹窗 ========== */
.fv-preview-dialog ::v-deep .el-dialog__body {
  padding: 0;
  overflow: hidden;
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
