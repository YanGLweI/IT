<template>
  <el-dialog
    :title="isEdit ? '编辑VPN隧道' : '新增VPN隧道'"
    :visible.sync="visible"
    width="720px"
    :close-on-click-modal="false"
    :append-to-body="true"
    custom-class="vpn-dialog"
    @close="handleClose"
  >
    <div class="dialog-body">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 网络 -->
        <el-tab-pane label="网络" name="network">
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">隧道名 <span class="required">*</span></label>
              <el-input v-model="form.tunnel_name" placeholder="请输入隧道名称" size="small" />
            </div>
            <div class="form-group">
              <label class="form-label">负责人 <span class="required">*</span></label>
              <el-input v-model="form.owner" placeholder="请输入负责人" size="small" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">对端IP <span class="required">*</span></label>
              <el-input v-model="form.remote_ip" placeholder="如 203.0.113.1" size="small" class="mono-input" />
            </div>
            <div class="form-group">
              <label class="form-label">本端IP <span class="required">*</span></label>
              <el-input v-model="form.local_ip" placeholder="如 10.0.0.1" size="small" class="mono-input" />
            </div>
          </div>
          <div class="form-row full">
            <div class="form-group" @mousedown="focusedUploadTarget = 'network'">
              <label class="form-label">配置截图 <span class="required">*</span> <span class="paste-hint">支持 Ctrl+V 粘贴</span></label>
              <el-upload
                ref="networkUpload"
                action=""
                :http-request="noopUpload"
                list-type="picture-card"
                :auto-upload="false"
                :limit="1"
                :on-change="(f) => handleImageChange('network', f)"
                :on-remove="() => handleImageRemove('network')"
                :file-list="networkFileList"
                accept=".jpg,.jpeg,.png,.gif,.webp"
              >
                <i class="el-icon-plus"></i>
              </el-upload>
            </div>
          </div>
        </el-tab-pane>

        <!-- 认证 -->
        <el-tab-pane label="认证" name="auth">
          <div class="form-row full">
            <div class="form-group">
              <label class="form-label">预共享密钥 (PSK) <span class="required">*</span></label>
              <el-input v-model="form.psk" :type="showPSK ? 'text' : 'password'" placeholder="请输入预共享密钥" size="small">
                <i slot="suffix" :class="showPSK ? 'el-icon-view' : 'el-icon-view'" class="psk-toggle" @click="showPSK = !showPSK"></i>
              </el-input>
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">IKE版本 <span class="required">*</span></label>
              <el-radio-group v-model="form.ike_version" size="small">
                <el-radio :label="1">IKEv1</el-radio>
                <el-radio :label="2">IKEv2</el-radio>
              </el-radio-group>
            </div>
            <div class="form-group" v-if="form.ike_version === 1">
              <label class="form-label">模式 <span class="required">*</span></label>
              <el-radio-group v-model="form.mode" size="small">
                <el-radio label="野蛮模式">野蛮模式</el-radio>
                <el-radio label="主模式">主模式</el-radio>
              </el-radio-group>
            </div>
          </div>
        </el-tab-pane>

        <!-- 阶段一 -->
        <el-tab-pane label="阶段一" name="phase1">
          <div class="form-row full">
            <div class="form-group" @mousedown="focusedUploadTarget = 'phase1'">
              <label class="form-label">配置截图 <span class="required">*</span> <span class="paste-hint">支持 Ctrl+V 粘贴</span></label>
              <el-upload
                ref="phase1Upload"
                action=""
                :http-request="noopUpload"
                list-type="picture-card"
                :auto-upload="false"
                :limit="1"
                :on-change="(f) => handleImageChange('phase1', f)"
                :on-remove="() => handleImageRemove('phase1')"
                :file-list="phase1FileList"
                accept=".jpg,.jpeg,.png,.gif,.webp"
              >
                <i class="el-icon-plus"></i>
              </el-upload>
            </div>
          </div>
        </el-tab-pane>

        <!-- 阶段二 -->
        <el-tab-pane label="阶段二" name="phase2">
          <el-tabs v-model="activePhase2" type="card" editable @edit="handlePhase2TabEdit">
            <el-tab-pane
              v-for="(p2, idx) in phase2List"
              :key="p2.key"
              :label="'条目 ' + (idx + 1)"
              :name="p2.key"
              :closable="phase2List.length > 1"
            >
              <div class="form-row">
                <div class="form-group">
                  <label class="form-label">本端地址 <span class="required">*</span></label>
                  <el-input v-model="p2.local_addr" placeholder="如 192.168.1.0/24" size="small" class="mono-input" />
                </div>
                <div class="form-group">
                  <label class="form-label">对端地址 <span class="required">*</span></label>
                  <el-input v-model="p2.remote_addr" placeholder="如 172.16.1.0/24" size="small" class="mono-input" />
                </div>
              </div>
              <div class="form-row full">
                <div class="form-group" @mousedown="focusedUploadTarget = 'phase2_' + idx">
                  <label class="form-label">配置截图 <span class="required">*</span> <span class="paste-hint">支持 Ctrl+V 粘贴</span></label>
                  <el-upload
                    :ref="'phase2Upload_' + idx"
                    action=""
                    :http-request="noopUpload"
                    list-type="picture-card"
                    :auto-upload="false"
                    :limit="1"
                    :on-change="(f) => handleImageChange('phase2_' + idx, f)"
                    :on-remove="() => handleImageRemove('phase2_' + idx)"
                    :file-list="getPhase2FileList(idx)"
                    accept=".jpg,.jpeg,.png,.gif,.webp"
                  >
                    <i class="el-icon-plus"></i>
                  </el-upload>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-tab-pane>
      </el-tabs>
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
import { createIPsecVpn, updateIPsecVpn } from '@/api/ipsec_vpn'
import DualControlDialog from '@/components/DualControlDialog.vue'

let phase2KeySeq = 0

export default {
  name: 'IPsecVpnDialog',
  components: { DualControlDialog },
  data() {
    return {
      visible: false,
      isEdit: false,
      editId: null,
      saving: false,
      showPSK: false,
      activeTab: 'network',
      activePhase2: 'p2_0',
      focusedUploadTarget: null,
      form: {
        tunnel_name: '',
        owner: '',
        remote_ip: '',
        local_ip: '',
        psk: '',
        ike_version: 2,
        mode: ''
      },
      // 图片状态: { file, url, isExisting, path }
      networkImage: null,
      phase1Image: null,
      phase2List: [{ key: 'p2_0', local_addr: '', remote_addr: '', image: null }]
    }
  },
  computed: {
    networkFileList() {
      if (!this.networkImage) return []
      return [{ name: this.networkImage.name, url: this.networkImage.url }]
    },
    phase1FileList() {
      if (!this.phase1Image) return []
      return [{ name: this.phase1Image.name, url: this.phase1Image.url }]
    }
  },
  mounted() {
    this._pasteHandler = (e) => this.onGlobalPaste(e)
    document.addEventListener('paste', this._pasteHandler)
  },
  beforeDestroy() {
    document.removeEventListener('paste', this._pasteHandler)
  },
  methods: {
    noopUpload() {},
    open(item) {
      this.visible = true
      this.activeTab = 'network'
      this.focusedUploadTarget = null
      if (item) {
        this.isEdit = true
        this.editId = item.id
        this.form = {
          tunnel_name: item.tunnel_name,
          owner: item.owner,
          remote_ip: item.remote_ip,
          local_ip: item.local_ip,
          psk: item.psk,
          ike_version: item.ike_version,
          mode: item.mode || ''
        }
        // 已有图片
        this.networkImage = item.network_image ? { path: item.network_image, url: '/' + item.network_image, name: item.network_image.split('/').pop(), isExisting: true } : null
        this.phase1Image = item.phase1_image ? { path: item.phase1_image, url: '/' + item.phase1_image, name: item.phase1_image.split('/').pop(), isExisting: true } : null
        // 阶段二
        let entries = []
        if (item.phase2_entries && item.phase2_entries !== '[]') {
          try { entries = JSON.parse(item.phase2_entries) } catch { entries = [] }
        }
        this.phase2List = entries.length ? entries.map((e, i) => ({
          key: 'p2_' + (phase2KeySeq++),
          local_addr: e.local_addr || '',
          remote_addr: e.remote_addr || '',
          image: e.image ? { path: e.image, url: '/' + e.image, name: e.image.split('/').pop(), isExisting: true } : null
        })) : [{ key: 'p2_0', local_addr: '', remote_addr: '', image: null }]
        this.activePhase2 = this.phase2List[0].key
      } else {
        this.isEdit = false
        this.editId = null
        this.form = { tunnel_name: '', owner: '', remote_ip: '', local_ip: '', psk: '', ike_version: 2, mode: '' }
        this.networkImage = null
        this.phase1Image = null
        this.phase2List = [{ key: 'p2_0', local_addr: '', remote_addr: '', image: null }]
        this.activePhase2 = 'p2_0'
      }
    },
    handleClose() {
      this.revokeAllUrls()
      this.form = { tunnel_name: '', owner: '', remote_ip: '', local_ip: '', psk: '', ike_version: 2, mode: '' }
      this.networkImage = null
      this.phase1Image = null
      this.phase2List = [{ key: 'p2_0', local_addr: '', remote_addr: '', image: null }]
      this.showPSK = false
    },
    revokeAllUrls() {
      if (this.networkImage && !this.networkImage.isExisting && this.networkImage.url) URL.revokeObjectURL(this.networkImage.url)
      if (this.phase1Image && !this.phase1Image.isExisting && this.phase1Image.url) URL.revokeObjectURL(this.phase1Image.url)
      this.phase2List.forEach(p2 => {
        if (p2.image && !p2.image.isExisting && p2.image.url) URL.revokeObjectURL(p2.image.url)
      })
    },
    getPhase2FileList(idx) {
      const p2 = this.phase2List[idx]
      if (!p2 || !p2.image) return []
      return [{ name: p2.image.name, url: p2.image.url }]
    },
    // 图片选择
    handleImageChange(target, file) {
      if (!file.raw) return
      const imgObj = { file: file.raw, url: URL.createObjectURL(file.raw), name: file.raw.name, isExisting: false }
      if (target === 'network') {
        if (this.networkImage && !this.networkImage.isExisting && this.networkImage.url) URL.revokeObjectURL(this.networkImage.url)
        this.networkImage = imgObj
      } else if (target === 'phase1') {
        if (this.phase1Image && !this.phase1Image.isExisting && this.phase1Image.url) URL.revokeObjectURL(this.phase1Image.url)
        this.phase1Image = imgObj
      } else if (target.startsWith('phase2_')) {
        const idx = parseInt(target.split('_')[1])
        if (this.phase2List[idx]) {
          if (this.phase2List[idx].image && !this.phase2List[idx].image.isExisting && this.phase2List[idx].image.url) URL.revokeObjectURL(this.phase2List[idx].image.url)
          this.$set(this.phase2List[idx], 'image', imgObj)
        }
      }
    },
    handleImageRemove(target) {
      if (target === 'network') {
        if (this.networkImage && !this.networkImage.isExisting && this.networkImage.url) URL.revokeObjectURL(this.networkImage.url)
        this.networkImage = null
      } else if (target === 'phase1') {
        if (this.phase1Image && !this.phase1Image.isExisting && this.phase1Image.url) URL.revokeObjectURL(this.phase1Image.url)
        this.phase1Image = null
      } else if (target.startsWith('phase2_')) {
        const idx = parseInt(target.split('_')[1])
        if (this.phase2List[idx]) {
          if (this.phase2List[idx].image && !this.phase2List[idx].image.isExisting && this.phase2List[idx].image.url) URL.revokeObjectURL(this.phase2List[idx].image.url)
          this.$set(this.phase2List[idx], 'image', null)
        }
      }
    },
    // 粘贴上传
    onGlobalPaste(e) {
      if (!this.visible || !this.focusedUploadTarget) return
      const items = (e.clipboardData || window.clipboardData).items
      if (!items) return
      for (const item of items) {
        if (item.type.startsWith('image/')) {
          const file = item.getAsFile()
          if (!file) continue
          if (file.size > 10 * 1024 * 1024) { this.$message.error('图片大小不能超过 10MB'); return }
          const wrappedFile = { raw: file, name: file.name || `pasted-${Date.now()}.png` }
          this.handleImageChange(this.focusedUploadTarget, wrappedFile)
          this.$message.success('图片已粘贴')
          e.preventDefault()
          break
        }
      }
    },
    // 阶段二动态标签页
    handlePhase2TabEdit(targetName, action) {
      if (action === 'add') {
        const newKey = 'p2_' + (phase2KeySeq++)
        this.phase2List.push({ key: newKey, local_addr: '', remote_addr: '', image: null })
        this.activePhase2 = newKey
      } else if (action === 'remove') {
        const idx = this.phase2List.findIndex(p => p.key === targetName)
        if (idx === -1) return
        const removed = this.phase2List[idx]
        if (removed.image && !removed.image.isExisting && removed.image.url) URL.revokeObjectURL(removed.image.url)
        this.phase2List.splice(idx, 1)
        if (this.activePhase2 === targetName) {
          this.activePhase2 = this.phase2List[Math.min(idx, this.phase2List.length - 1)].key
        }
      }
    },
    // 验证 + 保存
    async handleSave() {
      // 验证网络
      if (!this.form.tunnel_name.trim()) { this.activeTab = 'network'; this.$message.warning('请输入隧道名'); return }
      if (!this.form.owner.trim()) { this.activeTab = 'network'; this.$message.warning('请输入负责人'); return }
      if (!this.form.remote_ip.trim()) { this.activeTab = 'network'; this.$message.warning('请输入对端IP'); return }
      if (!this.form.local_ip.trim()) { this.activeTab = 'network'; this.$message.warning('请输入本端IP'); return }
      if (!this.networkImage) { this.activeTab = 'network'; this.$message.warning('请上传网络配置截图'); return }
      // 验证认证
      if (!this.form.psk.trim()) { this.activeTab = 'auth'; this.$message.warning('请输入预共享密钥'); return }
      if (this.form.ike_version === 1 && !this.form.mode) { this.activeTab = 'auth'; this.$message.warning('IKEv1时必须选择模式'); return }
      // 验证阶段一
      if (!this.phase1Image) { this.activeTab = 'phase1'; this.$message.warning('请上传阶段一配置截图'); return }
      // 验证阶段二
      for (let i = 0; i < this.phase2List.length; i++) {
        const p2 = this.phase2List[i]
        if (!p2.local_addr.trim()) { this.activeTab = 'phase2'; this.activePhase2 = p2.key; this.$message.warning(`阶段二条目${i + 1}：请输入本端地址`); return }
        if (!p2.remote_addr.trim()) { this.activeTab = 'phase2'; this.activePhase2 = p2.key; this.$message.warning(`阶段二条目${i + 1}：请输入对端地址`); return }
        if (!p2.image) { this.activeTab = 'phase2'; this.activePhase2 = p2.key; this.$message.warning(`阶段二条目${i + 1}：请上传配置截图`); return }
      }

      this.saving = true
      try {
        const dualToken = await this.$refs.dualControl.open()
        const formData = new FormData()
        formData.append('tunnel_name', this.form.tunnel_name)
        formData.append('owner', this.form.owner)
        formData.append('remote_ip', this.form.remote_ip)
        formData.append('local_ip', this.form.local_ip)
        formData.append('psk', this.form.psk)
        formData.append('ike_version', String(this.form.ike_version))
        formData.append('mode', this.form.mode)

        // 网络截图
        if (this.networkImage && !this.networkImage.isExisting) {
          formData.append('network_image', this.networkImage.file)
        } else if (this.networkImage && this.networkImage.isExisting) {
          formData.append('existing_network_image', this.networkImage.path)
        }
        // 阶段一截图
        if (this.phase1Image && !this.phase1Image.isExisting) {
          formData.append('phase1_image', this.phase1Image.file)
        } else if (this.phase1Image && this.phase1Image.isExisting) {
          formData.append('existing_phase1_image', this.phase1Image.path)
        }
        // 阶段二
        const phase2Data = this.phase2List.map(p2 => ({
          local_addr: p2.local_addr,
          remote_addr: p2.remote_addr,
          image: (p2.image && p2.image.isExisting) ? p2.image.path : ''
        }))
        formData.append('phase2_entries', JSON.stringify(phase2Data))
        this.phase2List.forEach((p2, i) => {
          if (p2.image && !p2.image.isExisting) {
            formData.append(`phase2_image_${i}`, p2.image.file)
          }
        })

        if (this.isEdit) {
          await updateIPsecVpn(this.editId, formData, dualToken)
          this.$message.success('更新成功')
        } else {
          await createIPsecVpn(formData, dualToken)
          this.$message.success('创建成功')
        }
        this.visible = false
        this.$emit('saved')
      } catch (e) {
        if (e.message !== 'canceled' && e !== 'canceled') console.error(e)
      } finally {
        this.saving = false
      }
    }
  }
}
</script>

<style scoped>
.dialog-body {
  max-height: 60vh;
  overflow-y: auto;
}
.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}
.form-row.full {
  display: block;
}
.form-group {
  flex: 1;
}
.form-label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: #475569;
  margin-bottom: 6px;
}
.required {
  color: #ef4444;
}
.paste-hint {
  font-size: 11px;
  color: #94a3b8;
  font-weight: 400;
  margin-left: 6px;
}
.mono-input >>> .el-input__inner {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
}
.psk-toggle {
  cursor: pointer;
  color: #94a3b8;
  padding: 0 8px;
  line-height: 32px;
}
.psk-toggle:hover {
  color: #3b82f6;
}
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.btn-cancel {
  padding: 9px 20px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #fff;
  color: #475569;
  font-size: 13px;
  cursor: pointer;
}
.btn-cancel:hover {
  border-color: #94a3b8;
}
.btn-save {
  padding: 9px 24px;
  border: none;
  border-radius: 8px;
  background: #3b82f6;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
.btn-save:hover {
  background: #2563eb;
}
.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
