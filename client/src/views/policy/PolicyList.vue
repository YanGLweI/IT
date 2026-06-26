<template>
  <div class="policy-list">
    <el-card>
      <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
        <span>IT政策管理</span>
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="uploadVisible = true">上传政策</el-button>
      </div>
      <el-table :data="policies" border stripe>
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="file_name" label="文件名" />
        <el-table-column label="文件大小" width="120">
          <template slot-scope="scope">{{ formatSize(scope.row.file_size) }}</template>
        </el-table-column>
        <el-table-column label="上传时间" width="180">
          <template slot-scope="scope">{{ formatDate(scope.row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280">
          <template slot-scope="scope">
            <el-button size="mini" @click="handlePreview(scope.row)">预览</el-button>
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 上传弹窗 -->
    <el-dialog title="上传政策文件" :visible.sync="uploadVisible" width="500px">
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
    <el-dialog title="编辑政策" :visible.sync="editVisible" width="500px">
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

    <!-- 预览弹窗 -->
    <el-dialog title="文件预览" :visible.sync="previewVisible" width="80%" top="5vh" @closed="clearDocxPreview">
      <!-- 工具栏：搜索 + 下载 -->
      <div class="preview-toolbar" slot="title">
        <span>文件预览</span>
        <div class="preview-toolbar-right">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索关键词"
            size="small"
            prefix-icon="el-icon-search"
            clearable
            style="width: 320px; margin-right: 10px"
            @keyup.enter.native="searchNext"
            @clear="clearSearch"
          >
            <el-button slot="append" icon="el-icon-bottom" @click="searchNext">下一个</el-button>
          </el-input>
          <span v-if="searchKeyword" class="search-info">{{ searchIndex + 1 }} / {{ searchTotal }}</span>
          <el-button type="primary" size="small" icon="el-icon-download" @click="downloadFile">下载</el-button>
        </div>
      </div>
      <div v-if="previewType === 'pdf'" style="height: 70vh">
        <iframe ref="pdfFrame" :src="previewUrl" style="width: 100%; height: 100%; border: none"></iframe>
      </div>
      <div v-else-if="previewType === 'image'" style="text-align: center">
        <img :src="previewUrl" style="max-width: 100%; max-height: 70vh" />
      </div>
      <div v-else-if="previewType === 'docx'" ref="docxScrollContainer" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 20px">
        <div ref="docxContainer" class="docx-preview-container"></div>
      </div>
      <div v-else style="text-align: center; padding: 40px">
        <p>该文件格式不支持在线预览</p>
        <el-button type="primary" @click="downloadFile">下载文件</el-button>
      </div>
    </el-dialog>
    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getPolicies, createPolicy, updatePolicy, deletePolicy, getPolicyPreviewUrl, getPolicyDownloadUrl } from '@/api/policy'
import { renderAsync } from 'docx-preview'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  components: { DualControlDialog },
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
      previewUrl: '',
      previewType: '',
      previewFileName: '',
      previewId: null,
      searchKeyword: '',
      searchIndex: 0,
      searchTotal: 0,
      searchMarks: []
    }
  },
  mounted() {
    this.fetchData()
  },
  watch: {
    searchKeyword(val) {
      clearTimeout(this._searchTimer)
      if (!val.trim()) {
        this.clearSearch()
        return
      }
      this._searchTimer = setTimeout(() => {
        this.highlightAll(val)
      }, 300)
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
      const fileType = row.file_type || ''
      const fileName = (row.file_name || '').toLowerCase()
      this.previewFileName = row.file_name || '文件'
      this.previewId = row.id
      // 重置搜索
      this.searchKeyword = ''
      this.searchIndex = 0
      this.searchTotal = 0
      this.searchMarks = []
      if (fileType.includes('pdf')) {
        this.previewType = 'pdf'
      } else if (fileType.startsWith('image/')) {
        this.previewType = 'image'
      } else if (fileType.includes('wordprocessingml') || fileType.includes('msword') || fileName.endsWith('.docx') || fileName.endsWith('.doc')) {
        this.previewType = 'docx'
      } else {
        this.previewType = 'other'
      }
      this.previewVisible = true

      // 使用 fetch 带 token 获取文件并创建 blob URL
      try {
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.previewUrl = URL.createObjectURL(blob)

        if (this.previewType === 'docx') {
          this.$nextTick(() => {
            this.renderDocxFromBlob(blob)
          })
        }
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('文件预览失败')
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
      this.searchMarks = []
      this.searchTotal = 0
      // 释放 blob URL
      if (this.previewUrl) {
        URL.revokeObjectURL(this.previewUrl)
        this.previewUrl = ''
      }
    },
    // 关键词搜索：跳转到下一个匹配
    searchNext() {
      if (!this.searchKeyword.trim()) return
      // 关键词变化时，重新高亮（已在 watch 中处理，这里作为保底）
      if (this.searchMarks.length === 0 || this._lastKeyword !== this.searchKeyword) {
        this._lastKeyword = this.searchKeyword
        this.highlightAll(this.searchKeyword)
        return // highlightAll 已经定位到第一个，不需要再 +1
      }
      if (this.searchTotal === 0) {
        this.$message.info('未找到匹配内容')
        return
      }
      // 跳转到下一个匹配
      this.searchIndex = (this.searchIndex + 1) % this.searchTotal
      this.scrollToMark(this.searchIndex)
    },
    highlightAll(keyword) {
      // 清除之前的高亮
      this.clearHighlights()
      if (!keyword.trim()) {
        this.searchTotal = 0
        this.searchIndex = 0
        return
      }
      const container = this.previewType === 'docx'
        ? this.$refs.docxContainer
        : (this.previewType === 'pdf' ? null : null)
      if (!container) {
        // PDF iframe 内无法直接操作，提示用户使用浏览器自带搜索
        this.$message.info('PDF 文件请使用 Ctrl+F / Cmd+F 进行搜索')
        return
      }
      const regex = new RegExp(this.escapeRegExp(keyword), 'gi')
      const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, null, false)
      const textNodes = []
      while (walker.nextNode()) {
        if (walker.currentNode.textContent.trim()) textNodes.push(walker.currentNode)
      }
      let total = 0
      this.searchMarks = []
      textNodes.forEach(node => {
        const text = node.textContent
        let match
        const matches = []
        while ((match = regex.exec(text)) !== null) {
          matches.push({ index: match.index, length: match[0].length })
        }
        if (matches.length === 0) return
        // 从后往前替换，避免索引偏移
        const frag = document.createDocumentFragment()
        let lastIdx = 0
        matches.forEach(m => {
          if (m.index > lastIdx) {
            frag.appendChild(document.createTextNode(text.slice(lastIdx, m.index)))
          }
          const mark = document.createElement('mark')
          mark.className = 'search-highlight'
          mark.dataset.searchIndex = total
          mark.textContent = text.slice(m.index, m.index + m.length)
          this.searchMarks.push(mark)
          frag.appendChild(mark)
          lastIdx = m.index + m.length
          total++
        })
        if (lastIdx < text.length) {
          frag.appendChild(document.createTextNode(text.slice(lastIdx)))
        }
        node.parentNode.replaceChild(frag, node)
      })
      this.searchTotal = total
      this.searchIndex = 0
      this._lastKeyword = keyword
      if (total > 0) {
        this.scrollToMark(0)
      } else {
        this.$message.info('未找到匹配内容')
      }
    },
    scrollToMark(index) {
      // 移除之前的 active
      this.searchMarks.forEach(m => m.classList.remove('search-highlight-active'))
      const mark = this.searchMarks[index]
      if (mark) {
        mark.classList.add('search-highlight-active')
        mark.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    },
    clearHighlights() {
      const container = this.$refs.docxContainer
      if (!container) return
      const marks = container.querySelectorAll('mark.search-highlight')
      marks.forEach(mark => {
        const parent = mark.parentNode
        parent.replaceChild(document.createTextNode(mark.textContent), mark)
        parent.normalize()
      })
      this.searchMarks = []
      this.searchTotal = 0
    },
    clearSearch() {
      this.clearHighlights()
      this.searchIndex = 0
      this._lastKeyword = ''
    },
    escapeRegExp(str) {
      return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
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
.search-info {
  font-size: 12px;
  color: #909399;
  margin-right: 10px;
  white-space: nowrap;
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
/* 搜索高亮样式 - 不能用 scoped，需要用 >>> */
.docx-preview-container >>> mark.search-highlight {
  background-color: #fff3cd;
  color: inherit;
  padding: 1px 2px;
  border-radius: 2px;
}
.docx-preview-container >>> mark.search-highlight-active {
  background-color: #ff9632;
  color: #fff;
  padding: 1px 2px;
  border-radius: 2px;
}
</style>
