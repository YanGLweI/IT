<template>
  <div class="ipsec-vpn-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">IPsec VPN</h2>
        <p class="page-subtitle">记录VPN隧道协商配置信息</p>
      </div>
      <button class="btn-primary" @click="openDialog(null)">
        <i class="el-icon-plus"></i>
        <span>新增隧道</span>
      </button>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="search-wrap">
        <i class="el-icon-search search-icon"></i>
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索隧道名、IP、负责人..."
          @input="handleSearch"
        />
        <i v-if="keyword" class="el-icon-close clear-icon" @click="keyword = ''; handleSearch()"></i>
      </div>
    </div>

    <!-- 表格 -->
    <div class="table-card">
      <div v-if="loading" class="loading-wrap">
        <i class="el-icon-loading"></i> 加载中...
      </div>
      <div v-else-if="!list.length" class="empty-wrap">
        <i class="el-icon-document"></i>
        <p>暂无VPN隧道信息</p>
        <button class="btn-primary btn-sm" @click="openDialog(null)">新增第一条</button>
      </div>
      <table v-else>
        <thead>
          <tr>
            <th>隧道名</th>
            <th>对端IP</th>
            <th>本端IP</th>
            <th>负责人</th>
            <th>阶段一</th>
            <th>阶段二</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td class="td-name">{{ item.tunnel_name }}</td>
            <td class="td-mono">{{ item.remote_ip }}</td>
            <td class="td-mono">{{ item.local_ip }}</td>
            <td>{{ item.owner }}</td>
            <td class="td-images">
              <template v-if="item.phase1_image">
                <div class="img-thumb-wrap">
                  <img :src="'/' + item.phase1_image" class="img-thumb" @click="previewImage(item.phase1_image)" />
                </div>
              </template>
              <span v-else class="td-empty">-</span>
            </td>
            <td class="td-images">
              <template v-if="getFirstPhase2Image(item)">
                <div class="img-thumb-wrap">
                  <img :src="'/' + getFirstPhase2Image(item)" class="img-thumb" @click="previewPhase2Images(item)" />
                  <span v-if="parsePhase2(item).length > 1" class="img-count">+{{ parsePhase2(item).length - 1 }}</span>
                </div>
              </template>
              <span v-else class="td-empty">-</span>
            </td>
            <td class="td-actions">
              <button class="action-btn" @click="showDetail(item)">详情</button>
              <button class="action-btn" @click="openDialog(item)">编辑</button>
              <button class="action-btn danger" @click="handleDelete(item)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <el-pagination
      class="page-pagination"
      background
      layout="total, sizes, prev, pager, next, jumper"
      :total="total"
      :page-size.sync="pageSize"
      :current-page.sync="page"
      :page-sizes="[10, 20, 50]"
      @size-change="handleSizeChange"
      @current-change="fetchList"
    />

    <!-- 新增/编辑弹窗 -->
    <IPsecVpnDialog ref="vpnDialog" @saved="fetchList" />

    <!-- 图片预览弹窗 -->
    <el-dialog title="图片预览" :visible.sync="previewVisible" width="700px" append-to-body :close-on-click-modal="true">
      <div class="preview-gallery">
        <img v-if="previewUrl" :src="'/' + previewUrl" class="preview-img" />
        <div v-if="previewImages.length > 1" class="preview-nav">
          <button class="nav-arrow left" :disabled="previewIndex === 0" @click="prevPreview">
            <i class="el-icon-arrow-left"></i>
          </button>
          <span class="nav-indicator">{{ previewIndex + 1 }} / {{ previewImages.length }}</span>
          <button class="nav-arrow right" :disabled="previewIndex === previewImages.length - 1" @click="nextPreview">
            <i class="el-icon-arrow-right"></i>
          </button>
        </div>
      </div>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog title="VPN隧道详情" :visible.sync="detailVisible" width="700px" append-to-body :close-on-click-modal="true">
      <div v-if="detailItem" class="detail-content">
        <div class="detail-section">
          <div class="detail-section-title">网络信息</div>
          <div class="detail-grid">
            <div class="detail-field">
              <span class="detail-label">隧道名</span>
              <span class="detail-value highlight">{{ detailItem.tunnel_name }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">负责人</span>
              <span class="detail-value">{{ detailItem.owner }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">对端IP</span>
              <span class="detail-value mono">{{ detailItem.remote_ip }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">本端IP</span>
              <span class="detail-value mono">{{ detailItem.local_ip }}</span>
            </div>
          </div>
          <div v-if="detailItem.network_image" class="detail-image-row">
            <img :src="'/' + detailItem.network_image" class="detail-img" @click="previewImage(detailItem.network_image)" />
          </div>
        </div>

        <div class="detail-section">
          <div class="detail-section-title">认证信息</div>
          <div class="detail-grid">
            <div class="detail-field">
              <span class="detail-label">预共享密钥</span>
              <span class="detail-value mono psk-row">{{ maskPSK(detailItem.psk) }}
                <i class="el-icon-document-copy psk-copy" title="复制密钥" @click="copyPSK(detailItem.psk)"></i>
              </span>
            </div>
            <div class="detail-field">
              <span class="detail-label">IKE版本</span>
              <span class="detail-value">IKEv{{ detailItem.ike_version }}</span>
            </div>
            <div v-if="detailItem.ike_version === 1" class="detail-field">
              <span class="detail-label">模式</span>
              <span class="detail-value">{{ detailItem.mode }}</span>
            </div>
          </div>
        </div>

        <div class="detail-section">
          <div class="detail-section-title">阶段一</div>
          <div v-if="detailItem.phase1_image" class="detail-image-row">
            <img :src="'/' + detailItem.phase1_image" class="detail-img" @click="previewImage(detailItem.phase1_image)" />
          </div>
          <span v-else class="td-empty">无截图</span>
        </div>

        <div class="detail-section">
          <div class="detail-section-title">阶段二（{{ parsePhase2(detailItem).length }} 条）</div>
          <div v-for="(p2, idx) in parsePhase2(detailItem)" :key="idx" class="phase2-item">
            <div class="phase2-header">条目 {{ idx + 1 }}</div>
            <div class="detail-grid">
              <div class="detail-field">
                <span class="detail-label">本端地址</span>
                <span class="detail-value mono">{{ p2.local_addr }}</span>
              </div>
              <div class="detail-field">
                <span class="detail-label">对端地址</span>
                <span class="detail-value mono">{{ p2.remote_addr }}</span>
              </div>
            </div>
            <div v-if="p2.image" class="detail-image-row">
              <img :src="'/' + p2.image" class="detail-img" @click="previewPhase2FromDetail(detailItem, idx)" />
            </div>
          </div>
        </div>

        <div class="detail-section">
          <div class="detail-section-title">操作记录</div>
          <div class="detail-grid">
            <div class="detail-field">
              <span class="detail-label">创建人</span>
              <span class="detail-value">{{ detailItem.created_by || '-' }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">创建时间</span>
              <span class="detail-value">{{ formatTime(detailItem.created_at) }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">更新人</span>
              <span class="detail-value">{{ detailItem.updated_by || '-' }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">更新时间</span>
              <span class="detail-value">{{ formatTime(detailItem.updated_at) }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 双控验证 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getIPsecVpns, deleteIPsecVpn } from '@/api/ipsec_vpn'
import IPsecVpnDialog from './IPsecVpnDialog.vue'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'IPsecVpnList',
  components: { IPsecVpnDialog, DualControlDialog },
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      page: 1,
      pageSize: 10,
      keyword: '',
      searchTimer: null,
      previewVisible: false,
      previewUrl: '',
      previewImages: [],
      previewIndex: 0,
      detailVisible: false,
      detailItem: null
    }
  },
  created() {
    this.fetchList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const params = { page: this.page, page_size: this.pageSize }
        if (this.keyword) params.keyword = this.keyword
        const res = await getIPsecVpns(params)
        this.list = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(() => {
        this.page = 1
        this.fetchList()
      }, 300)
    },
    handleSizeChange(val) {
      this.pageSize = val
      this.page = 1
      this.fetchList()
    },
    openDialog(item) {
      this.$refs.vpnDialog.open(item)
    },
    parsePhase2(item) {
      if (!item.phase2_entries || item.phase2_entries === '[]') return []
      try { return JSON.parse(item.phase2_entries) } catch { return [] }
    },
    getFirstPhase2Image(item) {
      const entries = this.parsePhase2(item)
      for (const e of entries) {
        if (e.image) return e.image
      }
      return ''
    },
    previewImage(path) {
      this.previewImages = [path]
      this.previewIndex = 0
      this.previewUrl = path
      this.previewVisible = true
    },
    previewPhase2Images(item) {
      const entries = this.parsePhase2(item)
      const images = entries.filter(e => e.image).map(e => e.image)
      if (!images.length) return
      this.previewImages = images
      this.previewIndex = 0
      this.previewUrl = images[0]
      this.previewVisible = true
    },
    previewPhase2FromDetail(item, entryIdx) {
      const entries = this.parsePhase2(item)
      const images = entries.filter(e => e.image).map(e => e.image)
      if (!images.length) return
      const clickedImage = entries[entryIdx] && entries[entryIdx].image
      const startIdx = clickedImage ? images.indexOf(clickedImage) : 0
      this.previewImages = images
      this.previewIndex = startIdx >= 0 ? startIdx : 0
      this.previewUrl = images[this.previewIndex]
      this.previewVisible = true
    },
    prevPreview() {
      if (this.previewIndex > 0) {
        this.previewIndex--
        this.previewUrl = this.previewImages[this.previewIndex]
      }
    },
    nextPreview() {
      if (this.previewIndex < this.previewImages.length - 1) {
        this.previewIndex++
        this.previewUrl = this.previewImages[this.previewIndex]
      }
    },
    showDetail(item) {
      this.detailItem = item
      this.detailVisible = true
    },
    maskPSK(psk) {
      if (!psk) return '-'
      if (psk.length <= 6) return '***'
      return psk.substring(0, 3) + '***' + psk.substring(psk.length - 3)
    },
    copyPSK(psk) {
      if (!psk) return
      const textarea = document.createElement('textarea')
      textarea.value = psk
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      try {
        document.execCommand('copy')
        this.$message.success('密钥已复制到剪贴板')
      } catch (e) {
        this.$message.error('复制失败')
      }
      document.body.removeChild(textarea)
    },
    formatTime(t) {
      if (!t) return '-'
      const d = new Date(t)
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    },
    async handleDelete(item) {
      try {
        await this.$confirm(`确定删除隧道「${item.tunnel_name}」？此操作不可恢复。`, '删除确认', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteIPsecVpn(item.id, dualToken)
        this.$message.success('删除成功')
        this.fetchList()
      } catch (e) {
        if (e.message !== 'canceled' && e !== 'canceled') console.error(e)
      }
    }
  }
}
</script>

<style scoped>
.ipsec-vpn-page {
  padding: 24px;
  height: calc(100% - 85px);
  overflow-y: auto;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  margin: 20px;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.page-title {
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}
.page-subtitle {
  font-size: 13px;
  color: #64748b;
  margin: 4px 0 0;
}
.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 18px;
  background: #3b82f6;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-primary:hover {
  background: #2563eb;
}
.btn-sm {
  padding: 6px 14px;
  font-size: 12px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.search-wrap {
  position: relative;
  width: 280px;
}
.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
  font-size: 14px;
}
.search-input {
  width: 100%;
  padding: 8px 32px 8px 34px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  background: #fff;
  outline: none;
  transition: border-color 0.2s;
}
.search-input:focus {
  border-color: #3b82f6;
}
.clear-icon {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
  cursor: pointer;
}
.clear-icon:hover {
  color: #64748b;
}
.table-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  overflow-x: auto;
}
.page-pagination {
  margin-top: 16px;
  text-align: right;
}
.loading-wrap, .empty-wrap {
  text-align: center;
  padding: 60px 20px;
  color: #64748b;
}
.empty-wrap p {
  margin: 12px 0;
}
table {
  width: 100%;
  min-width: 900px;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 13px;
}
th {
  padding: 12px 14px;
  text-align: left;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}
td {
  padding: 12px 14px;
  border-bottom: 1px solid #f1f5f9;
  color: #1e293b;
  white-space: nowrap;
}
tbody tr:nth-child(odd) td {
  background: #fff;
}
tbody tr:nth-child(even) td {
  background: #f8fafc;
}
tbody tr:hover td {
  background: #f1f5f9;
}
tbody tr:hover td:first-child {
  box-shadow: inset 3px 0 0 #3b82f6;
}
tr:last-child td {
  border-bottom: none;
}
.td-name {
  font-weight: 600;
  color: #1e293b;
}
.td-mono {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  color: #3b82f6;
}
.td-images {
  min-width: 56px;
}
.img-thumb-wrap {
  position: relative;
  display: inline-block;
}
.img-thumb {
  width: 36px;
  height: 36px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: transform 0.2s;
}
.img-thumb:hover {
  transform: scale(1.05);
}
.img-count {
  position: absolute;
  bottom: -4px;
  right: -4px;
  background: #3b82f6;
  color: #fff;
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 8px;
  line-height: 1.2;
}
.td-empty {
  color: #94a3b8;
}
.td-actions {
  vertical-align: middle;
  position: sticky;
  right: 0;
  background: #fff;
}
.td-actions .action-btn {
  margin-right: 6px;
}
.td-actions .action-btn:last-child {
  margin-right: 0;
}
th:last-child {
  position: sticky;
  right: 0;
  background: #f8fafc;
}
tbody tr:nth-child(even) .td-actions {
  background: #f8fafc;
}
tbody tr:hover .td-actions {
  background: #f1f5f9;
}
tr:hover .td-actions {
  background: #f1f5f9;
}
.action-btn {
  padding: 4px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: #fff;
  color: #475569;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.action-btn:hover {
  border-color: #3b82f6;
  color: #3b82f6;
}
.action-btn.danger:hover {
  border-color: #ef4444;
  color: #ef4444;
}
.preview-gallery {
  text-align: center;
}
.preview-img {
  max-width: 100%;
  max-height: 400px;
  border-radius: 8px;
  object-fit: contain;
}
.preview-nav {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 12px;
}
.nav-arrow {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid #e2e8f0;
  background: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #334155;
  transition: all 0.2s;
}
.nav-arrow:hover:not(:disabled) {
  background: #f1f5f9;
  border-color: #3b82f6;
  color: #3b82f6;
}
.nav-arrow:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.nav-indicator {
  font-size: 13px;
  color: #64748b;
}
/* 详情弹窗 */
.detail-content {
  max-height: 65vh;
  overflow-y: auto;
}
.detail-section {
  margin-bottom: 20px;
}
.detail-section:last-child {
  margin-bottom: 0;
}
.detail-section-title {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid #f1f5f9;
}
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 20px;
}
.detail-field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.detail-label {
  font-size: 11px;
  color: #94a3b8;
}
.detail-value {
  font-size: 13px;
  color: #1e293b;
  font-weight: 500;
}
.detail-value.mono {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  color: #3b82f6;
}
.detail-value.highlight {
  color: #3b82f6;
  font-weight: 600;
}
.psk-row {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.psk-copy {
  cursor: pointer;
  color: #94a3b8;
  font-size: 13px;
  transition: color 0.2s;
}
.psk-copy:hover {
  color: #3b82f6;
}
.detail-image-row {
  margin-top: 10px;
}
.detail-img {
  width: 100px;
  height: 70px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: transform 0.2s;
}
.detail-img:hover {
  transform: scale(1.05);
}
.phase2-item {
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  margin-bottom: 10px;
}
.phase2-header {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 8px;
}
</style>
