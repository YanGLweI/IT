<template>
  <div class="sftp-page">
    <el-card>
      <div slot="header" class="page-header">
        <span>SFTP账号一览</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-setting" @click="showServerManage = true">配置管理</el-button>
          <el-button type="primary" size="small" icon="el-icon-plus" @click="openAccountForm" :disabled="!activeServerId">新增账号</el-button>
          <el-button type="primary" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>

      <!-- 服务器 Tabs -->
      <el-tabs v-model="activeServerId" @tab-click="handleTabClick">
        <el-tab-pane
          v-for="server in servers"
          :key="server.id"
          :label="server.name"
          :name="String(server.id)"
        />
      </el-tabs>

      <!-- 导出按钮 -->
      <div v-if="activeServerId" class="export-bar">
        <el-button type="success" size="small" icon="el-icon-download" :loading="exporting" @click="handleExport">导出月度确认表</el-button>
      </div>

      <!-- 账号表格 -->
      <el-table
        :data="accounts"
        border
        stripe
        style="width: 100%"
        v-loading="loading"
      >
        <el-table-column prop="account_name" label="账号名" width="150" />
        <el-table-column prop="created_time" label="创建时间" width="120" />
        <el-table-column prop="validity" label="有效期" width="120" />
        <el-table-column label="权限" width="100">
          <template slot-scope="{ row }">
            <template v-if="parsePermissions(row.permissions_json).length > 0">
              <el-tag
                v-for="perm in parsePermissions(row.permissions_json)"
                :key="perm"
                size="small"
                :type="perm === 'read' ? 'info' : 'warning'"
                style="margin-right: 4px"
              >
                {{ perm === 'read' ? '读' : '写' }}
              </el-tag>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="contact_person" label="所属对接人" width="120" />
        <el-table-column label="所属部门" width="120">
          <template slot-scope="{ row }">
            {{ getDeptName(row.department_id) }}
          </template>
        </el-table-column>
        <el-table-column label="白名单IP" min-width="180">
          <template slot-scope="{ row }">
            <template v-if="parseWhitelist(row.whitelist_json).length > 0">
              <el-tag
                v-for="(ip, idx) in parseWhitelist(row.whitelist_json)"
                :key="idx"
                size="small"
                type="info"
                style="margin-right: 4px; margin-bottom: 4px"
              >
                {{ ip }}
              </el-tag>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template slot-scope="{ row }">
            <el-button size="mini" type="text" @click="openAccountForm(row)">编辑</el-button>
            <el-button size="mini" type="text" style="color:#F56C6C" @click="confirmDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 服务器配置管理弹窗 -->
    <el-dialog title="SFTP服务器配置" :visible.sync="showServerManage" width="500px" @close="fetchData">
      <div class="server-list">
        <div v-for="server in servers" :key="server.id" class="server-item">
          <span>{{ server.name }}</span>
          <div>
            <el-button size="mini" type="text" @click="editServer(server)">编辑</el-button>
            <el-button size="mini" type="text" style="color:#F56C6C" @click="confirmDeleteServer(server)">删除</el-button>
          </div>
        </div>
        <div v-if="servers.length === 0" class="empty-text">暂无服务器，请添加</div>
      </div>
      <div class="server-add">
        <el-input v-model="newServerName" placeholder="输入服务器名称" size="small" style="width: 250px; margin-right: 10px" />
        <el-button type="primary" size="small" @click="addServer" :loading="saving">添加</el-button>
      </div>
    </el-dialog>

    <!-- 账号表单弹窗 -->
    <el-dialog :title="isEdit ? '编辑账号' : '新增账号'" :visible.sync="accountFormVisible" width="600px" @close="resetAccountForm">
      <el-form :model="accountForm" :rules="accountRules" ref="accountForm" label-width="100px">
        <el-form-item label="账号名" prop="account_name">
          <el-input v-model="accountForm.account_name" placeholder="请输入账号名" />
        </el-form-item>
        <el-form-item label="创建时间" prop="created_time">
          <el-input v-model="accountForm.created_time" placeholder="如：2024-01-15" />
        </el-form-item>
        <el-form-item label="有效期" prop="validity">
          <el-input v-model="accountForm.validity" placeholder="如：2025-12-31 或 长期有效" />
        </el-form-item>
        <el-form-item label="权限" prop="permissions">
          <el-checkbox-group v-model="accountForm.permissions">
            <el-checkbox label="read">读</el-checkbox>
            <el-checkbox label="write">写</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="所属对接人" prop="contact_person">
          <el-input v-model="accountForm.contact_person" placeholder="请输入对接人姓名" />
        </el-form-item>
        <el-form-item label="所属部门" prop="department_id">
          <el-select v-model="accountForm.department_id" placeholder="请选择部门" style="width:100%">
            <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="白名单IP" prop="whitelist">
          <el-input v-model="accountForm.whitelist" type="textarea" :rows="3" placeholder="每行一个IP地址，或用逗号分隔" />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="accountFormVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAccountForm" :loading="saving">确定</el-button>
      </div>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getSftpServers, createSftpServer, updateSftpServer, deleteSftpServer, getSftpAccounts, createSftpAccount, updateSftpAccount, deleteSftpAccount, exportSftpConfirmation } from '@/api/sftp'
import { getDepartments } from '@/api/department'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  components: { DualControlDialog },
  name: 'SftpAccountList',
  data() {
    return {
      loading: false,
      saving: false,
      servers: [],
      activeServerId: '',
      accounts: [],
      departments: [],
      // 服务器管理
      showServerManage: false,
      newServerName: '',
      editingServerId: null,
      // 账号表单
      accountFormVisible: false,
      isEdit: false,
      editingAccountId: null,
      accountForm: {
        account_name: '',
        created_time: '',
        validity: '',
        permissions: [],
        contact_person: '',
        department_id: null,
        whitelist: ''
      },
      accountRules: {
        account_name: [{ required: true, message: '请输入账号名', trigger: 'blur' }],
        validity: [{ required: true, message: '请输入有效期', trigger: 'blur' }]
      },
      // 导出
      exporting: false
    }
  },
  mounted() {
    this.fetchData()
    this.fetchDepartments()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const res = await getSftpServers()
        this.servers = res.data || []
        
        // 如果有服务器且当前没有选中，选中第一个
        if (this.servers.length > 0 && !this.activeServerId) {
          this.activeServerId = String(this.servers[0].id)
        }
        
        // 加载当前服务器的账号
        if (this.activeServerId) {
          await this.fetchAccounts()
        }
      } catch (e) {
        console.error('获取服务器列表失败:', e)
      } finally {
        this.loading = false
      }
    },

    async fetchAccounts() {
      if (!this.activeServerId) return
      try {
        const res = await getSftpAccounts({ server_id: this.activeServerId })
        this.accounts = res.data || []
      } catch (e) {
        console.error('获取账号列表失败:', e)
      }
    },

    async fetchDepartments() {
      try {
        const res = await getDepartments()
        this.departments = res.data || []
      } catch (e) {
        console.error('获取部门列表失败:', e)
      }
    },

    handleTabClick() {
      this.fetchAccounts()
    },

    getDeptName(deptId) {
      if (!deptId) return '-'
      const dept = this.departments.find(d => d.id === deptId)
      return dept ? dept.name : '-'
    },

    parsePermissions(json) {
      if (!json || json === '[]') return []
      try {
        return JSON.parse(json)
      } catch {
        return []
      }
    },

    parseWhitelist(json) {
      if (!json || json === '[]') return []
      try {
        return JSON.parse(json)
      } catch {
        return []
      }
    },

    // === 服务器管理 ===
    async addServer() {
      if (!this.newServerName.trim()) {
        this.$message.warning('请输入服务器名称')
        return
      }

      this.saving = true
      try {
        const dualToken = await this.$refs.dualControl.open()
        
        if (this.editingServerId) {
          await updateSftpServer(this.editingServerId, { name: this.newServerName.trim() }, dualToken)
          this.$message.success('更新成功')
        } else {
          await createSftpServer({ name: this.newServerName.trim() }, dualToken)
          this.$message.success('添加成功')
        }
        
        this.newServerName = ''
        this.editingServerId = null
        this.fetchData()
      } catch (e) {
        if (e !== 'cancel') {
          console.error('添加服务器失败:', e)
        }
      } finally {
        this.saving = false
      }
    },

    editServer(server) {
      this.newServerName = server.name
      this.editingServerId = server.id
    },

    async confirmDeleteServer(server) {
      try {
        await this.$confirm(`确定要删除服务器「${server.name}」吗？该服务器下的所有账号也将被删除。`, '删除确认', {
          confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
        })
        
        const dualToken = await this.$refs.dualControl.open()
        await deleteSftpServer(server.id, dualToken)
        this.$message.success('删除成功')
        
        // 如果删除的是当前选中的服务器，重置选中
        if (String(server.id) === this.activeServerId) {
          this.activeServerId = ''
          this.accounts = []
        }
        
        this.fetchData()
      } catch (e) {
        if (e !== 'cancel') {
          console.error('删除服务器失败:', e)
        }
      }
    },

    // === 账号管理 ===
    openAccountForm(row) {
      if (row) {
        this.isEdit = true
        this.editingAccountId = row.id
        this.accountForm = {
          account_name: row.account_name,
          created_time: row.created_time || '',
          validity: row.validity || '',
          permissions: this.parsePermissions(row.permissions_json),
          contact_person: row.contact_person || '',
          department_id: row.department_id,
          whitelist: this.parseWhitelist(row.whitelist_json).join('\n')
        }
      } else {
        this.isEdit = false
        this.editingAccountId = null
      }
      this.accountFormVisible = true
    },

    resetAccountForm() {
      this.accountForm = {
        account_name: '',
        created_time: '',
        validity: '',
        permissions: [],
        contact_person: '',
        department_id: null,
        whitelist: ''
      }
      this.editingAccountId = null
      this.isEdit = false
    },

    async submitAccountForm() {
      this.$refs.accountForm.validate(async valid => {
        if (!valid) return

        this.saving = true
        try {
          const dualToken = await this.$refs.dualControl.open()

          // 处理白名单
          const whitelistArr = this.accountForm.whitelist
            .split(/[\n,]+/)
            .map(ip => ip.trim())
            .filter(ip => ip)

          const submitData = {
            account_name: this.accountForm.account_name,
            created_time: this.accountForm.created_time,
            validity: this.accountForm.validity,
            permissions_json: JSON.stringify(this.accountForm.permissions),
            contact_person: this.accountForm.contact_person,
            department_id: this.accountForm.department_id,
            whitelist_json: JSON.stringify(whitelistArr),
            server_id: Number(this.activeServerId)
          }

          if (this.isEdit) {
            await updateSftpAccount(this.editingAccountId, submitData, dualToken)
            this.$message.success('更新成功')
          } else {
            await createSftpAccount(submitData, dualToken)
            this.$message.success('创建成功')
          }

          this.accountFormVisible = false
          this.fetchAccounts()
        } catch (e) {
          if (e !== 'cancel') {
            console.error('保存账号失败:', e)
          }
        } finally {
          this.saving = false
        }
      })
    },

    async confirmDelete(row) {
      try {
        await this.$confirm(`确定要删除账号「${row.account_name}」吗？`, '删除确认', {
          confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
        })
        
        const dualToken = await this.$refs.dualControl.open()
        await deleteSftpAccount(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchAccounts()
      } catch (e) {
        if (e !== 'cancel') {
          console.error('删除账号失败:', e)
        }
      }
    },

    // === 导出 ===
    async handleExport() {
      if (!this.activeServerId) return

      this.exporting = true
      try {
        const res = await exportSftpConfirmation(this.activeServerId)
        const blob = new Blob([res], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        
        const server = this.servers.find(s => String(s.id) === this.activeServerId)
        const now = new Date()
        const yearMonth = `${now.getFullYear()}年${now.getMonth() + 1}月份`
        link.download = `SFTP账号确认表(${yearMonth})-${server ? server.name : ''}.xlsx`
        
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
        this.$message.success('导出成功')
      } catch (e) {
        console.error('导出失败:', e)
        this.$message.error('导出失败，请重试')
      } finally {
        this.exporting = false
      }
    }
  }
}
</script>

<style scoped>
.sftp-page {
  height: 100%;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.page-header-right {
  display: flex;
  gap: 10px;
}
.export-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 10px;
}
.server-list {
  max-height: 300px;
  overflow-y: auto;
  margin-bottom: 15px;
}
.server-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #eee;
}
.server-item:last-child {
  border-bottom: none;
}
.server-add {
  display: flex;
  align-items: center;
  padding-top: 10px;
  border-top: 1px solid #eee;
}
.empty-text {
  text-align: center;
  color: #999;
  padding: 20px;
}
</style>
