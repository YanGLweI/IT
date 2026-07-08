<template>
  <div class="operation-log-list">
    <el-card>
      <template #header><div>
        <span>操作日志</span>
      </div>
      </template>
      
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
            <el-option label="上传" value="上传" />
            <el-option label="替换" value="替换" />
            <el-option label="管理" value="管理" />
            <el-option label="排序" value="排序" />
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
          <template v-slot="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" width="120" align="center" />
        <el-table-column prop="display_name" label="姓名" width="120" align="center" />
        <el-table-column prop="action" label="操作" width="200" align="center" show-overflow-tooltip/>
        <el-table-column prop="resource_type" label="资源类型" width="120" align="center">
          <template v-slot="scope">
            {{ resourceTypeLabels[scope.row.resource_type] || scope.row.resource_type }}
          </template>
        </el-table-column>
        <el-table-column prop="resource_name" label="资源名称" show-overflow-tooltip />
        <el-table-column prop="approver" label="审批人" width="120" align="center" />
        <el-table-column label="操作" width="100" align="center">
          <template v-slot="scope">
            <el-button size="small" @click="handleViewDetail(scope.row)">详情</el-button>
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
    <el-dialog title="操作日志详情" v-model="detailDialogVisible" width="700px">
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
              <template v-slot="scope">
                <pre v-if="isJson(scope.row.old_value)" class="json-display">{{ formatFieldValue(scope.row.field_name, scope.row.old_value) }}</pre>
                <span v-else>{{ formatFieldValue(scope.row.field_name, scope.row.old_value) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="new_value" label="新值" show-overflow-tooltip>
              <template v-slot="scope">
                <pre v-if="isJson(scope.row.new_value)" class="json-display">{{ formatFieldValue(scope.row.field_name, scope.row.new_value) }}</pre>
                <span v-else>{{ formatFieldValue(scope.row.field_name, scope.row.new_value) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getOperationLogs, getOperationLogDetails } from '@/api/audit_log'
import { getRegions } from '@/api/region'
import { getVulnerabilityScans } from '@/api/vulnerability_scan'
import { getApprovedSoftware } from '@/api/approved_software'

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
      changeTypesMap: {}, // 变更类型ID到名称的映射
      regionsMap: {}, // 区域ID到名称的映射
      vulnScanMap: {}, // 漏洞扫描报告ID到文件名的映射
      softwareMap: {}, // 核准软件ID到名称的映射
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
        'asset_software': '资产软件关联',
        'change_record': '变更记录',
        'change_record_template': '变更记录模板',
        'change_type': '变更类型',
        'monthly_check_history': '月度检查记录',
        'quarterly_check_history': '季度检查记录',
        'user_change_history': '用户变更记录',
        'vulnerability_scan': '漏洞扫描',
        'system_hardening_history': '系统加固记录',
        'penetration_test': '渗透测试',
        'firewall_check': '防火墙检查',
        'patch_update': '补丁更新',
        'backup': '备份记录',
        'backup_recovery': '恢复还原记录',
        'backup_template': '备份模板'
      }
    }
  },
  mounted() {
    this.fetchChangeTypes()
    this.fetchRegions()
    this.fetchVulnerabilityScans()
    this.fetchApprovedSoftware()
    this.fetchData()
  },
  methods: {
    async fetchChangeTypes() {
      try {
        const { getChangeTypes } = await import('@/api/change_record')
        const res = await getChangeTypes()
        const types = res.data || []
        this.changeTypesMap = {}
        types.forEach(t => {
          this.changeTypesMap[t.id] = t.name
        })
      } catch (e) {
        console.error('获取变更类型失败', e)
      }
    },
    async fetchRegions() {
      try {
        const res = await getRegions()
        const regions = res.data || []
        this.regionsMap = {}
        regions.forEach(r => {
          this.regionsMap[r.id] = r.name
        })
      } catch (e) {
        console.error('获取区域列表失败', e)
      }
    },
    async fetchVulnerabilityScans() {
      try {
        this.vulnScanMap = {}
        let page = 1
        const pageSize = 100 // 后端允许的最大值
        let hasMore = true
        
        while (hasMore) {
          const res = await getVulnerabilityScans({ page, page_size: pageSize })
          const scans = Array.isArray(res.data) ? res.data : []
          scans.forEach(s => {
            this.vulnScanMap[s.id] = s.file_name
          })
          
          // 检查是否还有更多数据
          const total = res.total || 0
          hasMore = scans.length === pageSize && (page * pageSize) < total
          page++
        }
        
        console.log('Loaded vuln scan map:', Object.keys(this.vulnScanMap).length, 'items')
      } catch (e) {
        console.error('获取漏洞扫描报告列表失败', e)
      }
    },
    async fetchApprovedSoftware() {
      try {
        this.softwareMap = {}
        const res = await getApprovedSoftware()
        const softwareList = Array.isArray(res.data) ? res.data : []
        softwareList.forEach(s => {
          this.softwareMap[s.id] = s.name
        })
      } catch (e) {
        console.error('获取核准软件列表失败', e)
      }
    },
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
      if (!value || value === '<nil>') return '-'

      // 特殊处理 Compliance：合规性中文翻译
      if (fieldName === 'Compliance') {
        if (value === 'compliant') return '合规'
        if (value === 'non_compliant') return '不合规'
        return value
      }
      
      // 特殊处理 ChangeTypes：将ID数组转换为名称列表
      if (fieldName === 'ChangeTypes') {
        try {
          const parsed = JSON.parse(value)
          if (Array.isArray(parsed)) {
            if (parsed.length === 0) return '[]'
            // 尝试将ID转换为名称
            const names = parsed.map(id => this.changeTypesMap[id] || id).filter(n => n)
            return names.join('、')
          }
        } catch (e) {
          // 解析失败，返回原值
        }
      }
      
      // 特殊处理 Regions：将ID数组转换为名称列表
      if (fieldName === 'Regions') {
        try {
          const parsed = JSON.parse(value)
          if (Array.isArray(parsed)) {
            if (parsed.length === 0) return '[]'
            // 尝试将ID转换为名称
            const names = parsed.map(id => this.regionsMap[id] || id).filter(n => n)
            return names.join('、')
          }
        } catch (e) {
          // 解析失败，返回原值
        }
      }
      
      // 特殊处理 VulnerabilityScans：将ID数组转换为文件名列表
      if (fieldName === 'VulnerabilityScans') {
        try {
          const parsed = JSON.parse(value)
          if (Array.isArray(parsed)) {
            if (parsed.length === 0) return '[]'
            // 尝试将ID转换为文件名
            const names = parsed.map(id => {
              // 确保ID是数字类型进行匹配
              const numId = typeof id === 'string' ? parseInt(id, 10) : id
              return this.vulnScanMap[numId] || this.vulnScanMap[id] || id
            }).filter(n => n)
            return names.join('、')
          }
        } catch (e) {
          // 解析失败，返回原值
        }
      }
      
      // 特殊处理 SoftwareIDs / SoftwareAdded / SoftwareRemoved：将ID数组转换为软件名称列表
      if (fieldName === 'SoftwareIDs' || fieldName === 'SoftwareAdded' || fieldName === 'SoftwareRemoved') {
        try {
          const parsed = JSON.parse(value)
          if (Array.isArray(parsed)) {
            if (parsed.length === 0) return '[]'
            // 尝试将ID转换为软件名称
            const names = parsed.map(id => {
              const numId = typeof id === 'string' ? parseInt(id, 10) : id
              return this.softwareMap[numId] || this.softwareMap[id] || id
            }).filter(n => n)
            return names.join('、')
          }
        } catch (e) {
          // 解析失败，返回原值
        }
      }
      
      // 尝试解析JSON格式的值
      try {
        const parsed = JSON.parse(value)
        // 如果是对象或数组，格式化显示
        if (typeof parsed === 'object' && parsed !== null) {
          // 特殊处理已知的JSON字段
          if (fieldName === 'SystemRolesJSON') {
            if (Array.isArray(parsed) && parsed.length > 0) {
              return parsed.map(r => r.system + ': ' + (Array.isArray(r.roles) ? r.roles.join(', ') : r.roles)).join('; ')
            }
          } else if (fieldName === 'PermissionsJSON') {
            if (Array.isArray(parsed) && parsed.length > 0) {
              const labels = parsed.map(p => p === 'read' ? '读' : p === 'write' ? '写' : p)
              return labels.join('、')
            }
          } else if (fieldName === 'WhitelistJSON') {
            if (Array.isArray(parsed) && parsed.length > 0) {
              return parsed.join(', ')
            }
          }
          
          // 通用JSON格式化：美化显示
          return JSON.stringify(parsed, null, 2)
        }
        // 如果是简单类型（数字、布尔等），直接返回
        return String(parsed)
      } catch (e) {
        // 不是JSON格式，直接返回原值
        return value
      }
    },
    isJson(value) {
      if (!value || value === '<nil>') return false
      try {
        const parsed = JSON.parse(value)
        return typeof parsed === 'object' && parsed !== null
      } catch (e) {
        return false
      }
    }
  }
}
</script>

<style scoped>
.filter-form {
  margin-bottom: 16px;
}
.json-display {
  margin: 0;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  max-width: 400px;
  overflow-x: auto;
}
</style>
