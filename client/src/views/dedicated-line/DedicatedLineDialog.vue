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
          <input v-model="form.factory" class="form-input" placeholder="请输入厂区名称" />
        </div>
        <div class="form-group">
          <label class="form-label">运营商 <span class="required">*</span></label>
          <select v-model="form.carrier" class="form-input">
            <option value="" disabled>请选择运营商</option>
            <option value="电信">电信</option>
            <option value="联通">联通</option>
            <option value="移动">移动</option>
            <option value="广电">广电</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">上行带宽</label>
          <div class="bandwidth-wrap">
            <input v-model.number="form.bandwidth_up" class="form-input" type="number" min="0" placeholder="0" />
            <span class="unit">Mbps</span>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">下行带宽</label>
          <div class="bandwidth-wrap">
            <input v-model.number="form.bandwidth_down" class="form-input" type="number" min="0" placeholder="0" />
            <span class="unit">Mbps</span>
          </div>
        </div>
      </div>

      <!-- 网络配置 -->
      <div class="section-label">网络配置</div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">IP范围（起始） <span class="required">*</span></label>
          <input v-model="form.ip_start" class="form-input mono" placeholder="如 192.168.1.1" @input="calcIP" />
        </div>
        <div class="form-group">
          <label class="form-label">IP范围（结束） <span class="required">*</span></label>
          <input v-model="form.ip_end" class="form-input mono" placeholder="如 192.168.1.254" @input="calcIP" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">子网掩码 <span class="required">*</span></label>
          <input v-model="form.subnet_mask" class="form-input mono" placeholder="如 255.255.255.0" />
        </div>
        <div class="form-group">
          <label class="form-label">网关 <span class="required">*</span></label>
          <input v-model="form.gateway" class="form-input mono" placeholder="如 192.168.1.1" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">DNS</label>
          <input v-model="form.dns" class="form-input mono" placeholder="如 8.8.8.8, 114.114.114.114" />
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
          <div class="upload-area" @click="triggerUpload">
            <input ref="fileInput" type="file" accept=".jpg,.jpeg,.png,.gif,.webp" multiple style="display:none" @change="handleFileChange" />
            <div class="upload-icon">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <rect x="3" y="3" width="18" height="18" rx="2"/>
                <circle cx="8.5" cy="8.5" r="1.5"/>
                <polyline points="21 15 16 10 5 21"/>
              </svg>
            </div>
            <div class="upload-text">点击上传图片</div>
            <div class="upload-hint">支持 JPG、PNG、GIF、WebP 格式</div>
          </div>
          <!-- 已上传图片预览 -->
          <div v-if="images.length" class="image-preview-list">
            <div v-for="(img, idx) in images" :key="idx" class="preview-item">
              <img :src="img.url" :alt="img.name" @click="previewImage(img.url)" />
              <button class="preview-remove" @click="removeImage(idx)">
                <i class="el-icon-close"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="form-row full">
        <div class="form-group">
          <label class="form-label">备注</label>
          <textarea v-model="form.notes" class="form-textarea" placeholder="输入备注信息..." rows="3"></textarea>
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
import { createDedicatedLine, updateDedicatedLine, uploadDedicatedLineImage, deleteDedicatedLineImage } from '@/api/dedicated_line'
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
  methods: {
    open(item) {
      this.visible = true
      this.images = []
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
    // 图片上传
    triggerUpload() {
      this.$refs.fileInput.click()
    },
    async handleFileChange(e) {
      const files = e.target.files
      if (!files || !files.length) return
      try {
        const token = await this.$refs.dualControl.open()
        for (const file of files) {
          const formData = new FormData()
          formData.append('file', file)
          const res = await uploadDedicatedLineImage(formData, token)
          const data = res.data || res
          this.images.push({ path: data.path, url: data.url, name: data.name })
        }
        this.$message.success('图片上传成功')
      } catch (err) {
        if (err.message !== 'canceled') {
          // handled by interceptor
        }
      }
      e.target.value = ''
    },
    async removeImage(idx) {
      const img = this.images[idx]
      try {
        const token = await this.$refs.dualControl.open()
        await deleteDedicatedLineImage(img.path, token)
        this.images.splice(idx, 1)
        this.$message.success('图片已删除')
      } catch (err) {
        if (err.message !== 'canceled') {
          // handled
        }
      }
    },
    previewImage(url) {
      window.open(url, '_blank')
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
        const payload = {
          ...this.form,
          bandwidth_up: this.form.bandwidth_up || 0,
          bandwidth_down: this.form.bandwidth_down || 0,
          images: JSON.stringify(this.images.map(i => i.path))
        }
        if (this.isEdit) {
          await updateDedicatedLine(this.editId, payload, token)
          this.$message.success('更新成功')
        } else {
          await createDedicatedLine(payload, token)
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

/* 上传区域 */
.upload-area {
  border: 2px dashed #e2e8f0;
  border-radius: 12px;
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.2s;
}
.upload-area:hover {
  border-color: #3b82f6;
}
.upload-icon {
  color: #94a3b8;
  margin-bottom: 8px;
}
.upload-text {
  font-size: 13px;
  color: #64748b;
}
.upload-hint {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 4px;
}

/* 图片预览 */
.image-preview-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 12px;
}
.preview-item {
  position: relative;
  width: 80px;
  height: 80px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #e2e8f0;
}
.preview-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  cursor: pointer;
}
.preview-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  opacity: 0;
  transition: opacity 0.2s;
}
.preview-item:hover .preview-remove {
  opacity: 1;
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
</style>
