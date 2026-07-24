<template>
  <div class="dedicated-line-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">专线信息</h2>
        <p class="page-subtitle">记录各厂区运营商专线配置信息</p>
      </div>
      <button class="btn-primary" @click="openDialog(null)">
        <i class="el-icon-plus"></i>
        <span>新增专线</span>
      </button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-input v-model="keyword" placeholder="搜索厂区、运营商、IP..." prefix-icon="el-icon-search" clearable style="width: 240px;" @input="handleSearch" />
      <el-select v-model="filterCarrier" placeholder="全部运营商" clearable class="filter-select" @change="handleFilterChange">
        <el-option label="电信" value="电信" />
        <el-option label="联通" value="联通" />
        <el-option label="移动" value="移动" />
        <el-option label="广电" value="广电" />
      </el-select>
      <el-select v-model="filterFactory" placeholder="全部厂区" clearable class="filter-select" @change="handleFilterChange">
        <el-option v-for="f in factoryOptions" :key="f" :label="f" :value="f" />
      </el-select>
    </div>

    <!-- 表格 -->
    <div class="table-card">
      <div v-if="loading" class="loading-wrap">
        <i class="el-icon-loading"></i> 加载中...
      </div>
      <div v-else-if="!list.length" class="empty-wrap">
        <i class="el-icon-document"></i>
        <p>暂无专线信息</p>
        <button class="btn-primary btn-sm" @click="openDialog(null)">新增第一条</button>
      </div>
      <table v-else>
        <thead>
          <tr>
            <th>厂区</th>
            <th>运营商</th>
            <th>带宽 (上行/下行)</th>
            <th>IP范围</th>
            <th>掩码</th>
            <th>网关</th>
            <th>DNS</th>
            <th>IP数</th>
            <th>图片</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td class="td-factory">{{ item.factory }}</td>
            <td>
              <span class="carrier-badge" :class="'carrier-' + carrierClass(item.carrier)">{{ item.carrier }}</span>
            </td>
            <td class="td-bandwidth">{{ item.bandwidth_up }}M / {{ item.bandwidth_down }}M</td>
            <td class="td-mono">{{ item.ip_start }} - {{ item.ip_end }}</td>
            <td class="td-mono">{{ item.subnet_mask }}</td>
            <td class="td-mono">{{ item.gateway }}</td>
            <td class="td-mono">{{ item.dns || '-' }}</td>
            <td class="td-count">{{ item.ip_count }}</td>
            <td class="td-images">
              <template v-if="parseImages(item.images).length">
                <div class="img-thumb-wrap">
                  <img :src="'/' + parseImages(item.images)[0]" class="img-thumb" @click="previewImages(item)" />
                  <i class="el-icon-close img-close" @click.stop="deleteImage(item, parseImages(item.images)[0])"></i>
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
    <div class="pagination-wrap">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size.sync="pageSize"
        :current-page.sync="page"
        :page-sizes="[10, 20, 50]"
        @size-change="handleSizeChange"
        @current-change="fetchList"
      />
    </div>

    <!-- 新增/编辑弹窗 -->
    <DedicatedLineDialog ref="lineDialog" @saved="fetchList" />

    <!-- 双控验证 -->
    <DualControlDialog ref="dualControl" />

    <!-- 图片预览弹窗 -->
    <el-dialog class="preview-dialog" title="图片预览" :visible.sync="previewVisible" width="700px" append-to-body :close-on-click-modal="true">
      <div class="preview-gallery">
        <img v-for="(img, idx) in previewList" :key="idx" :src="'/' + img" class="preview-img" />
      </div>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog title="专线详情" :visible.sync="detailVisible" width="600px" append-to-body :close-on-click-modal="true" custom-class="detail-dialog">
      <div v-if="detailItem" class="detail-content">
        <div class="detail-section">
          <div class="detail-section-title">基本信息</div>
          <div class="detail-grid">
            <div class="detail-field">
              <span class="detail-label">厂区</span>
              <span class="detail-value">{{ detailItem.factory }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">运营商</span>
              <span class="detail-value">{{ detailItem.carrier }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">上行带宽</span>
              <span class="detail-value">{{ detailItem.bandwidth_up }} Mbps</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">下行带宽</span>
              <span class="detail-value">{{ detailItem.bandwidth_down }} Mbps</span>
            </div>
          </div>
        </div>
        <div class="detail-section">
          <div class="detail-section-title">网络配置</div>
          <div class="detail-grid">
            <div class="detail-field">
              <span class="detail-label">IP范围</span>
              <span class="detail-value mono">{{ detailItem.ip_start }} - {{ detailItem.ip_end }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">子网掩码</span>
              <span class="detail-value mono">{{ detailItem.subnet_mask }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">网关</span>
              <span class="detail-value mono">{{ detailItem.gateway }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">DNS</span>
              <span class="detail-value mono">{{ detailItem.dns || '-' }}</span>
            </div>
            <div class="detail-field">
              <span class="detail-label">可用IP数</span>
              <span class="detail-value highlight">{{ detailItem.ip_count }} 个</span>
            </div>
          </div>
        </div>
        <div v-if="detailItem.notes" class="detail-section">
          <div class="detail-section-title">备注</div>
          <p class="detail-notes">{{ detailItem.notes }}</p>
        </div>
        <div v-if="parseImages(detailItem.images).length" class="detail-section">
          <div class="detail-section-title">图片附件</div>
          <div class="detail-images">
            <img v-for="(img, idx) in parseImages(detailItem.images)" :key="idx" :src="'/' + img" class="detail-img" @click="previewList = [img]; previewVisible = true" />
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
  </div>
</template>

<script>
import { getDedicatedLines, deleteDedicatedLine, deleteDedicatedLineImage } from '@/api/dedicated_line'
import DedicatedLineDialog from './DedicatedLineDialog.vue'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'DedicatedLineList',
  components: { DedicatedLineDialog, DualControlDialog },
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      page: 1,
      pageSize: 10,
      factoryOptions: [],
      keyword: '',
      filterCarrier: '',
      filterFactory: '',
      searchTimer: null,
      previewVisible: false,
      previewList: [],
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
        if (this.filterCarrier) params.carrier = this.filterCarrier
        if (this.filterFactory) params.factory = this.filterFactory
        const res = await getDedicatedLines(params)
        this.list = res.data || []
        this.total = res.total || 0
        if (res.factories) this.factoryOptions = res.factories
      } catch (e) {
        // handled by interceptor
      } finally {
        this.loading = false
      }
    },
    handleSizeChange(val) {
      this.pageSize = val
      this.page = 1
      this.fetchList()
    },
    handleFilterChange() {
      this.page = 1
      this.fetchList()
    },
    handleSearch() {
      clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(() => {
        this.page = 1
        this.fetchList()
      }, 300)
    },
    carrierClass(carrier) {
      const map = { '电信': 'telecom', '联通': 'unicom', '移动': 'mobile', '广电': 'broadcast' }
      return map[carrier] || 'telecom'
    },
    openDialog(item) {
      this.$refs.lineDialog.open(item)
    },
    async handleDelete(item) {
      try {
        await this.$confirm(`确定删除「${item.factory} - ${item.carrier}」的专线信息？`, '删除确认', {
          type: 'warning',
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          confirmButtonClass: 'el-button--danger'
        })
      } catch {
        return
      }
      try {
        const token = await this.$refs.dualControl.open()
        await deleteDedicatedLine(item.id, token)
        this.$message.success('删除成功')
        this.fetchList()
      } catch (e) {
        if (e.message !== 'canceled') {
          // error handled by interceptor
        }
      }
    },
    parseImages(images) {
      if (!images || images === '[]') return []
      try {
        return JSON.parse(images)
      } catch {
        return []
      }
    },
    previewImages(item) {
      this.previewList = this.parseImages(item.images)
      this.previewVisible = true
    },
    async deleteImage(item, imagePath) {
      try {
        await this.$confirm('确定删除这张图片？', '删除确认', {
          type: 'warning',
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          confirmButtonClass: 'el-button--danger'
        })
      } catch {
        return
      }
      try {
        const token = await this.$refs.dualControl.open()
        await deleteDedicatedLineImage(imagePath, item.id, token)
        this.$message.success('图片已删除')
        this.fetchList()
      } catch (e) {
        if (e.message !== 'canceled') {
          // handled
        }
      }
    },
    showDetail(item) {
      this.detailItem = item
      this.detailVisible = true
    },
    formatTime(t) {
      if (!t) return '-'
      const d = new Date(t)
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    }
  }
}
</script>

<style scoped>
.dedicated-line-page {
  padding: 24px;
  height: calc(100% - 85px);
  overflow-y: auto;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  margin: 20px;
}

/* 页面头部 */
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
.btn-primary {
  display: flex;
  align-items: center;
  gap: 6px;
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
.btn-primary:hover {
  background: #2563eb;
}
.btn-primary.btn-sm {
  padding: 7px 14px;
  font-size: 12px;
}

/* 筛选栏 */
.filter-select {
  width: 130px;
  flex-shrink: 0;
}

/* 表格卡片 */
.table-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  overflow-x: auto;
}

.loading-wrap, .empty-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #94a3b8;
  font-size: 14px;
  gap: 12px;
}
.empty-wrap i {
  font-size: 40px;
}
table {
  width: 100%;
  min-width: 900px;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 13px;
}
thead {
  background: #f8fafc;
}
th {
  padding: 12px 14px;
  text-align: left;
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}
td {
  padding: 12px 14px;
  border-bottom: 1px solid #f1f5f9;
  color: #1e293b;
  white-space: nowrap;
}
tr:last-child td {
  border-bottom: none;
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
.td-factory {
  font-weight: 500;
}
.td-mono {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  color: #3b82f6;
}
.td-bandwidth {
  font-weight: 500;
}
.td-count {
  font-weight: 600;
  color: #3b82f6;
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

/* 运营商徽章 */
.carrier-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 500;
}
.carrier-telecom {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}
.carrier-unicom {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}
.carrier-mobile {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}
.carrier-broadcast {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

/* 操作按钮 */
.action-btn {
  padding: 5px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: transparent;
  color: #64748b;
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

/* 图片列 */
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
.img-close {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 16px;
  height: 16px;
  line-height: 16px;
  text-align: center;
  font-size: 10px;
  color: #fff;
  background: #ef4444;
  border-radius: 50%;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}
.img-thumb-wrap:hover .img-close {
  opacity: 1;
}
.img-close:hover {
  background: #dc2626;
}
.td-empty {
  color: #94a3b8;
}

/* 图片预览弹窗 */
.preview-gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
}
.preview-img {
  max-width: 100%;
  max-height: 400px;
  border-radius: 8px;
  object-fit: contain;
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
.detail-notes {
  font-size: 13px;
  color: #475569;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
}
.detail-images {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.detail-img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: transform 0.2s;
}
.detail-img:hover {
  transform: scale(1.05);
}
</style>
