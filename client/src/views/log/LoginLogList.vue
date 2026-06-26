<template>
  <div class="login-log-list">
    <el-card>
      <div slot="header">
        <span>登录日志</span>
      </div>
      
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="yyyy-MM-dd"
            @change="handleDateChange"
          />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="filters.username" placeholder="请输入用户名" clearable @clear="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="logs" border stripe v-loading="loading">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="created_at" label="时间" width="180" align="center">
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

      <!-- 分页 -->
      <el-pagination
        style="margin-top: 20px; text-align: right"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        :current-page="pagination.page"
        :page-sizes="[10, 20, 50, 100]"
        :page-size="pagination.pageSize"
        layout="total, sizes, prev, pager, next, jumper"
        :total="pagination.total"
      />
    </el-card>
  </div>
</template>

<script>
import { getLoginLogs } from '@/api/audit_log'

export default {
  name: 'LoginLogList',
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
.filter-form {
  margin-bottom: 16px;
}
</style>
