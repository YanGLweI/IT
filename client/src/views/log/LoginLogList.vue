<template>
  <div class="login-log-list">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">登录日志</h2>
        <p class="page-subtitle">查看用户登录记录与审计信息</p>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="yyyy-MM-dd"
        size="small"
        @change="handleDateChange"
      />
      <el-input v-model="filters.username" placeholder="搜索用户名..." size="small" clearable style="width: 200px" @clear="handleSearch" @keyup.enter.native="handleSearch" />
      <el-button size="small" type="primary" icon="el-icon-search" @click="handleSearch">搜索</el-button>
    </div>

    <!-- 表格 -->
    <div class="table-card" ref="tableCard">
      <div class="table-wrapper">
        <el-table :data="logs" stripe v-loading="loading" :max-height="tableMaxHeight">
          <el-table-column type="index" label="序号" width="75" align="center" />
          <el-table-column prop="created_at" label="时间" width="220" align="center">
            <template slot-scope="scope">
              {{ formatDate(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column prop="username" label="用户名" width="120" align="center" />
          <el-table-column prop="display_name" label="姓名" width="120" align="center" />
          <el-table-column prop="action" label="操作" width="120" align="center">
            <template slot-scope="scope">
              <el-tag :type="getActionTagType(scope.row.action)" size="small">
                {{ getActionLabel(scope.row.action) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="ip_address" label="IP地址" width="150" align="center" />
          <el-table-column prop="detail" label="详情" show-overflow-tooltip />
        </el-table>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="pagination.total"
        :page-size.sync="pagination.pageSize"
        :current-page.sync="pagination.page"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script>
import { getLoginLogs } from '@/api/audit_log'
import tableHeightMixin from '@/mixins/table-height'

export default {
  name: 'LoginLogList',
  mixins: [tableHeightMixin],
  data() {
    return {
      logs: [],
      loading: false,
      dateRange: [],
      filters: {
        username: '',
        start_date: '',
        end_date: ''
      },
      pagination: {
        page: 1,
        pageSize: 10,
        total: 0
      }
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          page_size: this.pagination.pageSize,
          username: this.filters.username,
          start_date: this.filters.start_date,
          end_date: this.filters.end_date
        }
        const res = await getLoginLogs(params)
        this.logs = res.data || []
        this.pagination.total = res.total || 0
      } catch (e) {
        console.error(e)
        this.$message.error('获取登录日志失败')
      } finally {
        this.loading = false
        this.$nextTick(() => this.calcTableHeight())
      }
    },
    handleDateChange(val) {
      if (val) {
        this.filters.start_date = val[0]
        this.filters.end_date = val[1]
      } else {
        this.filters.start_date = ''
        this.filters.end_date = ''
      }
      this.handleSearch()
    },
    handleSearch() {
      this.pagination.page = 1
      this.fetchData()
    },
    handleSizeChange(val) {
      this.pagination.pageSize = val
      this.fetchData()
    },
    handleCurrentChange(val) {
      this.pagination.page = val
      this.fetchData()
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN', { 
        year: 'numeric', 
        month: '2-digit', 
        day: '2-digit', 
        hour: '2-digit', 
        minute: '2-digit', 
        second: '2-digit' 
      })
    },
    getActionTagType(action) {
      const typeMap = {
        'login_success': 'success',
        'login_failure': 'danger',
        'logout': 'info'
      }
      return typeMap[action] || 'info'
    },
    getActionLabel(action) {
      const labelMap = {
        'login_success': '登录成功',
        'login_failure': '登录失败',
        'logout': '退出登录'
      }
      return labelMap[action] || action
    }
  }
}
</script>

<style scoped>
.login-log-list {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.table-card {
}

.table-wrapper {
}

.filter-bar .el-button {
  border-radius: 10px;
}
</style>
