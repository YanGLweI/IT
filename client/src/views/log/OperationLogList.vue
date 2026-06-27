<template>
  <div class="operation-log-list">
    <el-card>
      <div slot="header">
        <span>操作日志</span>
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
        <el-form-item label="操作类型">
          <el-select v-model="filters.action" placeholder="请选择" clearable @change="handleSearch">
            <el-option label="创建" value="创建" />
            <el-option label="更新" value="更新" />
            <el-option label="删除" value="删除" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源类型">
          <el-select v-model="filters.resource_type" placeholder="请选择" clearable @change="handleSearch">
            <el-option v-for="(label, key) in resourceTypeLabels" :key="key" :label="label" :value="key" />
          </el-select>
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
        <el-table-column prop="action" label="操作" width="150" align="center" />
        <el-table-column prop="resource_type" label="资源类型" width="120" align="center">
          <template slot-scope="scope">
            {{ resourceTypeLabels[scope.row.resource_type] || scope.row.resource_type }}
          </template>
        </el-table-column>
        <el-table-column prop="resource_name" label="资源名称" show-overflow-tooltip />
        <el-table-column prop="approver" label="审批人" width="120" align="center" />
        <el-table-column label="操作" width="100" align="center">
          <template slot-scope="scope">
            <el-button size="mini" @click="handleViewDetail(scope.row)">详情</el-button>
          </template>
        </el-table-column>
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

    <!-- 详情弹窗 -->
    <el-dialog title="操作日志详情" :visible.sync="detailDialogVisible" width="700px">
      <div v-if="currentLog">
        <el-descriptions :column="2" border size="medium">
          <el-descriptions-item label="操作时间">{{ formatDate(currentLog.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ currentLog.display_name }} ({{ currentLog.username }})</el-descriptions-item>
          <el-descriptions-item label="操作类型">{{ currentLog.action }}</el-descriptions-item>
          <el-descriptions-item label="资源类型">{{ resourceTypeLabels[currentLog.resource_type] || currentLog.resource_type }}</el-descriptions-item>
          <el-descriptions-item label="资源名称">{{ currentLog.resource_name }}</el-descriptions-item>
          <el-descriptions-item label="审批人">{{ currentLog.approver || '-' }}</el-descriptions-item>
          <el-descriptions-item label="IP地址">{{ currentLog.ip_address }}</el-descriptions-item>
        </el-descriptions>
        
        <div style="margin-top: 20px">
          <h4>变更明细</h4>
          <el-table :data="currentDetails" border stripe size="small">
            <el-table-column prop="field_label" label="字段名" width="150" />
            <el-table-column prop="old_value" label="旧值" show-overflow-tooltip>
              <template slot-scope="scope">{{ formatFieldValue(scope.row.field_name, scope.row.old_value) }}</template>
            </el-table-column>
            <el-table-column prop="new_value" label="新值" show-overflow-tooltip>
              <template slot-scope="scope">{{ formatFieldValue(scope.row.field_name, scope.row.new_value) }}</template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getOperationLogs, getOperationLogDetails } from '@/api/audit_log'

export default {
  name: 'OperationLogList',
  data() {
    return {
      logs: [],
      loading: false,
      dateRange: [],
      filters: {
        username: '',
        action: '',
        resource_type: '',
        start_date: '',
        end_date: ''
      },
      pagination: {
        page: 1,
        pageSize: 10,
        total: 0
      },
      detailDialogVisible: false,
      currentLog: null,
      currentDetails: [],
      resourceTypeLabels: {
        'asset': '资产',
        'region': '区域',
        'policy': '政策',
        'topology': '拓扑图',
        'os_type': '操作系统',
        'department': '部门',
        'department_position': '部门岗位',
        'permission_rule': '岗位权限',
        'user_permission': '用户权限',
        'sftp_server': 'SFTP服务器',
        'sftp_account': 'SFTP账号',
        'approved_software': '核准软件',
        'asset_software': '资产软件关联'
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
          action: this.filters.action,
          resource_type: this.filters.resource_type,
          start_date: this.filters.start_date,
          end_date: this.filters.end_date
        }
        const res = await getOperationLogs(params)
        this.logs = res.data || []
        this.pagination.total = res.total || 0
      } catch (e) {
        console.error(e)
        this.$message.error('获取操作日志失败')
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
    async handleViewDetail(row) {
      try {
        const res = await getOperationLogDetails(row.id)
        this.currentLog = res.data.log
        this.currentDetails = res.data.details || []
        this.detailDialogVisible = true
      } catch (e) {
        console.error(e)
        this.$message.error('获取日志详情失败')
      }
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
    formatFieldValue(fieldName, value) {
      if (!value) return '-'
      // 系统角色JSON解析为可读格式
      if (fieldName === 'SystemRolesJSON') {
        try {
          const roles = JSON.parse(value)
          if (Array.isArray(roles) && roles.length > 0) {
            return roles.map(r => r.system + ': ' + (Array.isArray(r.roles) ? r.roles.join(', ') : r.roles)).join('; ')
          }
          return '-'
        } catch (e) {
          return value
        }
      }
      // 权限JSON解析为可读格式
      if (fieldName === 'PermissionsJSON') {
        try {
          const perms = JSON.parse(value)
          if (Array.isArray(perms) && perms.length > 0) {
            const labels = perms.map(p => p === 'read' ? '读' : p === 'write' ? '写' : p)
            return labels.join('、')
          }
          return '-'
        } catch (e) {
          return value
        }
      }
      // 白名单JSON解析为可读格式
      if (fieldName === 'WhitelistJSON') {
        try {
          const ips = JSON.parse(value)
          if (Array.isArray(ips) && ips.length > 0) {
            return ips.join(', ')
          }
          return '-'
        } catch (e) {
          return value
        }
      }
      return value
    }
  }
}
</script>

<style scoped>
.filter-form {
  margin-bottom: 16px;
}
</style>
