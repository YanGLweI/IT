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

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="search-wrap">
        <i class="el-icon-search search-icon"></i>
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索厂区、运营商、IP..."
          @input="handleSearch"
        />
        <i v-if="keyword" class="el-icon-close clear-icon" @click="keyword = ''; handleSearch()"></i>
      </div>
      <select v-model="filterCarrier" class="filter-select" @change="fetchList">
        <option value="">全部运营商</option>
        <option value="电信">电信</option>
        <option value="联通">联通</option>
        <option value="移动">移动</option>
        <option value="广电">广电</option>
      </select>
      <select v-model="filterFactory" class="filter-select" @change="fetchList">
        <option value="">全部厂区</option>
        <option v-for="f in factoryOptions" :key="f" :value="f">{{ f }}</option>
      </select>
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
            <td class="td-actions">
              <button class="action-btn" @click="openDialog(item)">编辑</button>
              <button class="action-btn danger" @click="handleDelete(item)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增/编辑弹窗 -->
    <DedicatedLineDialog ref="lineDialog" @saved="fetchList" />

    <!-- 双控验证 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getDedicatedLines, deleteDedicatedLine } from '@/api/dedicated_line'
import DedicatedLineDialog from './DedicatedLineDialog.vue'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'DedicatedLineList',
  components: { DedicatedLineDialog, DualControlDialog },
  data() {
    return {
      list: [],
      loading: false,
      keyword: '',
      filterCarrier: '',
      filterFactory: '',
      searchTimer: null
    }
  },
  computed: {
    factoryOptions() {
      const set = new Set(this.list.map(i => i.factory))
      return [...set].sort()
    }
  },
  created() {
    this.fetchList()
  },
  methods: {
    async fetchList() {
      this.loading = true
      try {
        const params = {}
        if (this.keyword) params.keyword = this.keyword
        if (this.filterCarrier) params.carrier = this.filterCarrier
        if (this.filterFactory) params.factory = this.filterFactory
        const res = await getDedicatedLines(params)
        this.list = res.data || []
      } catch (e) {
        // handled by interceptor
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(() => this.fetchList(), 300)
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
    }
  }
}
</script>

<style scoped>
.dedicated-line-page {
  padding: 24px;
  height: 100%;
  overflow-y: auto;
  background: #f8fafc;
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

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.search-wrap {
  position: relative;
  width: 240px;
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
  padding: 9px 32px 9px 34px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 13px;
  color: #1e293b;
  outline: none;
  transition: border-color 0.2s;
}
.search-input:focus {
  border-color: #3b82f6;
}
.search-input::placeholder {
  color: #94a3b8;
}
.clear-icon {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
  cursor: pointer;
  font-size: 13px;
}
.clear-icon:hover {
  color: #64748b;
}
.filter-select {
  padding: 9px 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 13px;
  color: #1e293b;
  outline: none;
  cursor: pointer;
  transition: border-color 0.2s;
}
.filter-select:focus {
  border-color: #3b82f6;
}

/* 表格卡片 */
.table-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  overflow: hidden;
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
  border-collapse: collapse;
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
tr:hover td {
  background: #f8fafc;
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
  display: flex;
  gap: 6px;
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
</style>
