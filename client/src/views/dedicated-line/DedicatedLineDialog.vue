<template>
  <el-dialog
    :title="isEdit ? '编辑专线信息' : '新增专线信息'"
    :visible.sync="visible"
    width="660px"
    :close-on-click-modal="false"
    :append-to-body="true"
    custom-class="line-dialog"
    @close="handleClose"
  >
    <div class="dialog-body">
      <!-- 基本信息 -->
      <div class="section-label">基本信息</div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">厂区 <span class="required">*</span></label>
          <el-input v-model="form.factory" placeholder="请输入厂区名称" size="small" />
        </div>
        <div class="form-group">
          <label class="form-label">运营商 <span class="required">*</span></label>
          <el-select v-model="form.carrier" placeholder="请选择运营商" size="small" style="width: 100%">
            <el-option label="电信" value="电信" />
            <el-option label="联通" value="联通" />
            <el-option label="移动" value="移动" />
            <el-option label="广电" value="广电" />
          </el-select>
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">上行带宽</label>
          <div class="bandwidth-field">
            <el-input v-model.number="form.bandwidth_up" type="number" placeholder="0" size="small" />
            <span class="bandwidth-unit">Mbps</span>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">下行带宽</label>
          <div class="bandwidth-field">
            <el-input v-model.number="form.bandwidth_down" type="number" placeholder="0" size="small" />
            <span class="bandwidth-unit">Mbps</span>
          </div>
        </div>
      </div>

      <!-- 网络配置 -->
      <div class="section-label">网络配置</div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">IP范围（起始） <span class="required">*</span></label>
          <el-input v-model="form.ip_start" placeholder="如 192.168.1.1" size="small" class="mono-input" @input="calcIP" />
        </div>
        <div class="form-group">
          <label class="form-label">IP范围（结束） <span class="required">*</span></label>
          <el-input v-model="form.ip_end" placeholder="如 192.168.1.254" size="small" class="mono-input" @input="calcIP" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">子网掩码 <span class="required">*</span></label>
          <el-input v-model="form.subnet_mask" placeholder="如 255.255.255.0" size="small" class="mono-input" />
        </div>
        <div class="form-group">
          <label class="form-label">网关 <span class="required">*</span></label>
          <el-input v-model="form.gateway" placeholder="如 192.168.1.1" size="small" class="mono-input" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">DNS</label>
          <el-input v-model="form.dns" placeholder="如 8.8.8.8, 114.114.114.114" size="small" class="mono-input" />
        </div>
        <div class="form-group">
          <label class="form-label">IP数（自动计算）</label>
          <div class="ip-count-box">{{ ipCount > 0 ? ipCount + ' 个可用IP' : '—' }}</div>
        </div>
      </div>

      <!-- 附件与备注 -->
      <div class="section-label">附件与备注</div>
      <div class="form-row full">
        <div class="form-group">
          <label class="form-label">图片附件（可选）</label>
          <el-upload
            ref="upload"
            action=""
            list-type="picture-card"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            accept=".jpg,.jpeg,.png,.gif,.webp"
          >
            <i class="el-icon-plus"></i>
            <div slot="tip" class="el-upload__tip">支持 JPG、PNG、GIF、WebP 格式，限 1 张</div>
          </el-upload>
        </div>
      </div>
      <div class="form-row full">
        <div class="form-group">
          <label class="form-label">备注</label>
          <el-input v-model="form.notes" type="textarea" placeholder="输入备注信息..." :rows="3" size="small" />
        </div>
      </div>
    </div>

    <div slot="footer" class="dialog-footer">
      <button class="btn-cancel" @click="visible = false">取消</button>
      <button class="btn-save" :disabled="saving" @click="handleSave">
        <i v-if="saving" class="el-icon-loading"></i>
        {{ saving ? '保存中...' : '保存' }}
      </button>
    </div>

    <!-- 双控验证 -->
    <DualControlDialog ref="dualControl" />
  </el-dialog>
</template>

<script>
import { createDedicatedLine, updateDedicatedLine } from '@/api/dedicated_line'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'DedicatedLineDialog',
  components: { DualControlDialog },
  data() {
    return {
      visible: false,
      isEdit: false,
      editId: null,
      saving: false,
      ipCount: 0,
      images: [],
      selectedFiles: [],
      form: {
        factory: '',
        carrier: '',
        bandwidth_up: null,
        bandwidth_down: null,
        ip_start: '',
        ip_end: '',
        subnet_mask: '',
        gateway: '',
        dns: '',
        notes: ''
      }
    }
  },
  computed: {
    fileList() {
      const existing = this.images.map(img => ({ name: img.name, url: img.url, isExisting: true, path: img.path }))
      const newFiles = this.selectedFiles.map((item, i) => ({ name: item.file.name, url: item.url, isNew: true, index: i }))
      return [...existing, ...newFiles]
    }
  },
  methods: {
    open(item) {
      this.visible = true
      this.images = []
      this.selectedFiles = []
      if (item) {
        this.isEdit = true
        this.editId = item.id
        this.form = {
          factory: item.factory,
          carrier: item.carrier,
          bandwidth_up: item.bandwidth_up,
          bandwidth_down: item.bandwidth_down,
          ip_start: item.ip_start,
          ip_end: item.ip_end,
          subnet_mask: item.subnet_mask,
          gateway: item.gateway,
          dns: item.dns,
          notes: item.notes
        }
        this.ipCount = item.ip_count || 0
        // 解析已有图片
        if (item.images && item.images !== '[]') {
          try {
            const imgs = JSON.parse(item.images)
            this.images = imgs.map(path => ({ path, url: '/' + path, name: path.split('/').pop() }))
          } catch { this.images = [] }
        }
      } else {
        this.isEdit = false
        this.editId = null
        this.form = {
          factory: '',
          carrier: '',
          bandwidth_up: null,
          bandwidth_down: null,
          ip_start: '',
          ip_end: '',
          subnet_mask: '',
          gateway: '',
          dns: '',
          notes: ''
        }
        this.ipCount = 0
      }
    },
    handleClose() {
      this.form = {
        factory: '', carrier: '', bandwidth_up: null, bandwidth_down: null,
        ip_start: '', ip_end: '', subnet_mask: '', gateway: '', dns: '', notes: ''
      }
      this.images = []
      // 释放所有 Blob URL
      this.selectedFiles.forEach(item => {
        if (item.url) URL.revokeObjectURL(item.url)
      })
      this.selectedFiles = []
      this.ipCount = 0
    },
    // IP计算
    calcIP() {
      const start = this.ipToNum(this.form.ip_start)
      const end = this.ipToNum(this.form.ip_end)
      if (start !== null && end !== null && end >= start) {
        this.ipCount = end - start + 1
      } else {
        this.ipCount = 0
      }
    },
    ipToNum(ip) {
      if (!ip) return null
      const parts = ip.split('.')
      if (parts.length !== 4) return null
      let num = 0
      for (const p of parts) {
        const n = parseInt(p, 10)
        if (isNaN(n) || n < 0 || n > 255) return null
        num = num * 256 + n
      }
      return num
    },
    // 图片选择（不自动上传）
    handleFileChange(file) {
      if (file.raw) {
        this.selectedFiles.push({ file: file.raw, url: URL.createObjectURL(file.raw) })
      }
    },
    handleFileRemove(file) {
      if (file.isExisting) {
        const idx = this.images.findIndex(img => img.path === file.path)
        if (idx !== -1) this.images.splice(idx, 1)
      } else if (file.isNew) {
        const removed = this.selectedFiles.splice(file.index, 1)
        if (removed.length && removed[0].url) URL.revokeObjectURL(removed[0].url)
      }
    },
    // 保存
    async handleSave() {
      // 验证
      if (!this.form.factory.trim()) {
        this.$message.warning('请输入厂区名称')
        return
      }
      if (!this.form.carrier) {
        this.$message.warning('请选择运营商')
        return
      }
      if (!this.form.ip_start.trim() || !this.form.ip_end.trim()) {
        this.$message.warning('请输入IP范围')
        return
      }
      if (!this.form.subnet_mask.trim()) {
        this.$message.warning('请输入子网掩码')
        return
      }
      if (!this.form.gateway.trim()) {
        this.$message.warning('请输入网关')
        return
      }

      this.saving = true
      try {
        const token = await this.$refs.dualControl.open()
        const formData = new FormData()
        formData.append('factory', this.form.factory)
        formData.append('carrier', this.form.carrier)
        formData.append('bandwidth_up', this.form.bandwidth_up || 0)
        formData.append('bandwidth_down', this.form.bandwidth_down || 0)
        formData.append('ip_start', this.form.ip_start)
        formData.append('ip_end', this.form.ip_end)
        formData.append('subnet_mask', this.form.subnet_mask)
        formData.append('gateway', this.form.gateway)
        formData.append('dns', this.form.dns || '')
        formData.append('notes', this.form.notes || '')
        // 保留的已有图片
        formData.append('existing_images', JSON.stringify(this.images.map(i => i.path)))
        // 新上传的图片文件
        this.selectedFiles.forEach(item => {
          formData.append('images', item.file)
        })
        if (this.isEdit) {
          await updateDedicatedLine(this.editId, formData, token)
          this.$message.success('更新成功')
        } else {
          await createDedicatedLine(formData, token)
          this.$message.success('创建成功')
        }
        this.visible = false
        this.$emit('saved')
      } catch (err) {
        if (err.message !== 'canceled') {
          // handled by interceptor
        }
      } finally {
        this.saving = false
      }
    }
  }
}
</script>

<style scoped>
.dialog-body {
  max-height: 65vh;
  overflow-y: auto;
  padding-right: 4px;
}
.section-label {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  margin: 20px 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f1f5f9;
}
.section-label:first-child {
  margin-top: 0;
}
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}
.form-row.full {
  grid-template-columns: 1fr;
}
.form-group {
  display: flex;
  flex-direction: column;
}
.form-label {
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  margin-bottom: 6px;
}
.form-label .required {
  color: #ef4444;
}
.form-input {
  padding: 9px 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 13px;
  color: #1e293b;
  outline: none;
  transition: border-color 0.2s;
  width: 100%;
}
.form-input:focus {
  border-color: #3b82f6;
}
.form-input::placeholder {
  color: #94a3b8;
}
.form-input.mono {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
}
select.form-input {
  appearance: auto;
  cursor: pointer;
}
.bandwidth-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}
.bandwidth-wrap .form-input {
  flex: 1;
}
.bandwidth-wrap .unit {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
}
.bandwidth-field {
  display: flex;
  align-items: center;
  gap: 8px;
}
.bandwidth-field .el-input {
  flex: 1;
}
.bandwidth-unit {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
}
.ip-count-box {
  padding: 9px 12px;
  background: rgba(59, 130, 246, 0.06);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 10px;
  font-size: 13px;
  color: #3b82f6;
  font-weight: 500;
}
.form-textarea {
  padding: 9px 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 13px;
  color: #1e293b;
  outline: none;
  resize: vertical;
  min-height: 72px;
  font-family: inherit;
  transition: border-color 0.2s;
  width: 100%;
}
.form-textarea:focus {
  border-color: #3b82f6;
}

/* 底部按钮 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.btn-cancel {
  padding: 9px 18px;
  background: transparent;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 13px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-cancel:hover {
  border-color: #94a3b8;
  color: #1e293b;
}
.btn-save {
  padding: 9px 18px;
  background: #3b82f6;
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-save:hover {
  background: #2563eb;
}
.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>

<style>
/* 弹窗全局样式覆盖 */
.line-dialog {
  border-radius: 16px !important;
  overflow: hidden;
}
.line-dialog .el-dialog__header {
  padding: 20px 24px 16px;
  border-bottom: 1px solid #f1f5f9;
}
.line-dialog .el-dialog__title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}
.line-dialog .el-dialog__body {
  padding: 20px 24px;
}
.line-dialog .el-dialog__footer {
  padding: 16px 24px;
  border-top: 1px solid #f1f5f9;
}
.line-dialog .el-select .el-input__inner,
.line-dialog .el-input__inner {
  border-radius: 10px;
  border-color: #e2e8f0;
  font-size: 13px;
  height: 36px;
  line-height: 36px;
}
.line-dialog .el-textarea__inner {
  border-radius: 10px;
  border-color: #e2e8f0;
  font-size: 13px;
}
.line-dialog .el-select .el-input__inner:focus,
.line-dialog .el-input__inner:focus,
.line-dialog .el-textarea__inner:focus {
  border-color: #3b82f6;
}
.line-dialog .el-input--small .el-input__inner {
  height: 36px;
  line-height: 36px;
}
.line-dialog .el-upload--picture-card {
  border-radius: 10px;
  border-color: #e2e8f0;
  width: 80px;
  height: 80px;
  line-height: 80px;
}
.line-dialog .el-upload--picture-card:hover {
  border-color: #3b82f6;
}
.line-dialog .el-upload-list--picture-card .el-upload-list__item {
  border-radius: 10px;
  width: 80px;
  height: 80px;
}
</style>
