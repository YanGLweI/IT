<template>
  <div class="user-permission-page">
    <el-card>
      <div slot="header" class="page-header">
        <span>用户权限一览</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-setting" @click="showDeptManage = true">管理配置</el-button>
          <el-button type="primary" size="small" icon="el-icon-plus" @click="openUserForm">新增用户</el-button>
          <el-button type="primary" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>

      <!-- 部门 Tabs -->
      <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane
          v-for="dept in departments"
          :key="dept.id"
          :label="dept.name"
          :name="String(dept.id)"
        />
      </el-tabs>

      <!-- 用户表格 -->
      <div class="table-wrapper" ref="tableWrapper">
        <el-table
          :data="filteredUsers"
          border
          stripe
          style="width: 100%"
          v-loading="loading"
          :max-height="tableMaxHeight"
        >
          <el-table-column prop="name" label="姓名" width="150" />
          <el-table-column prop="position_name" label="岗位" width="150" />
          <el-table-column label="系统角色" min-width="400">
            <template slot-scope="{ row }">
              <div class="cell-roles">
                <template v-if="parseSystemRoles(row.system_roles_json).length > 0">
                  <el-tag
                    v-for="sr in parseSystemRoles(row.system_roles_json)"
                    :key="sr.system"
                    size="small"
                    class="role-tag"
                    type="success"
                  >
                    {{ sr.system }}: {{ sr.roles.join(', ') }}
                  </el-tag>
                </template>
                <span v-else class="empty-role">-</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template slot-scope="{ row }">
              <el-button size="mini" type="text" @click="openUserForm(row)">编辑</el-button>
              <el-button size="mini" type="text" style="color:#F56C6C" @click="confirmDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 新增/编辑用户弹窗 -->
    <el-dialog :title="isEdit ? '编辑用户' : '新增用户'" :visible.sync="userFormVisible" width="650px" top="5vh" @close="resetUserForm">
      <el-form :model="userForm" :rules="userRules" ref="userForm" label-width="100px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="userForm.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="部门" prop="department_id">
          <el-select v-model="userForm.department_id" placeholder="请选择部门" @change="handleDeptChange" style="width:100%">
            <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="岗位" prop="position_name">
          <el-select v-model="userForm.position_name" placeholder="请先选择部门" :disabled="!userForm.department_id" style="width:100%">
            <el-option v-for="pos in availablePositions" :key="pos" :label="pos" :value="pos" />
          </el-select>
        </el-form-item>
        <el-form-item label="系统角色">
          <div class="role-select-area">
            <!-- 已选系统角色展示 -->
            <div v-if="userForm.systemRoles.length > 0" class="selected-roles">
              <el-tag
                v-for="(sr, idx) in userForm.systemRoles"
                :key="sr.system"
                closable
                @close="userForm.systemRoles.splice(idx, 1)"
                size="small"
                class="role-tag"
                type="success"
              >
                {{ sr.system }}: {{ sr.roles.join(', ') }}
              </el-tag>
            </div>
            <!-- 添加系统角色 -->
            <div class="add-role-row">
              <el-select v-model="selectedSystem" placeholder="选择系统" size="small" @change="handleSystemChange" style="width:200px">
                <el-option v-for="sys in allSystems" :key="sys" :label="sys" :value="sys" />
              </el-select>
              <el-checkbox-group
                v-if="selectedSystem && getSystemRoles(selectedSystem).length > 0"
                v-model="tempSelectedRoles"
                size="small"
                class="role-checkboxes"
              >
                <el-checkbox v-for="role in getSystemRoles(selectedSystem)" :key="role" :label="role">{{ role }}</el-checkbox>
              </el-checkbox-group>
              <el-button v-if="selectedSystem && tempSelectedRoles.length > 0" type="primary" size="small" @click="addSystemRoles">添加</el-button>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="userFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">确定</el-button>
      </span>
    </el-dialog>

    <!-- 管理配置弹窗 -->
    <DepartmentManage ref="deptManage" :visible.sync="showDeptManage" @updated="fetchData" />

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getDepartments, getDepartmentPositions } from '@/api/department'
import { getUserPermissions, createUserPermission, updateUserPermission, deleteUserPermission } from '@/api/userPermission'
import { getPermissionRules } from '@/api/permission'
import DualControlDialog from '@/components/DualControlDialog.vue'
import DepartmentManage from './DepartmentManage.vue'

export default {
  components: { DualControlDialog, DepartmentManage },
  name: 'UserPermissionList',
  data() {
    return {
      loading: false,
      saving: false,
      departments: [],
      users: [],
      activeTab: 'all',
      // 用户表单
      userFormVisible: false,
      isEdit: false,
      editingUserId: null,
      userForm: {
        name: '',
        department_id: null,
        position_name: '',
        systemRoles: [] // [{system: '防火墙', roles: ['admin', 'viewer']}]
      },
      userRules: {
        name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
        department_id: [{ required: true, message: '请选择部门', trigger: 'change' }],
        position_name: [{ required: true, message: '请选择岗位', trigger: 'change' }]
      },
      // 系统角色选择
      allSystems: [],
      allSystemRoles: {}, // { '防火墙': ['admin', 'viewer'], ... }
      selectedSystem: '',
      tempSelectedRoles: [],
      // 部门岗位
      deptPositions: {}, // { deptId: ['岗位1', '岗位2'] }
      // 管理配置
      showDeptManage: false,
      // 表格高度
      tableMaxHeight: null
    }
  },
  computed: {
    filteredUsers() {
      if (this.activeTab === 'all') return this.users
      return this.users.filter(u => String(u.department_id) === this.activeTab)
    },
    availablePositions() {
      if (!this.userForm.department_id) return []
      return this.deptPositions[this.userForm.department_id] || []
    }
  },
  mounted() {
    this.fetchData()
    this.fetchPermissionRules()
    this.$nextTick(() => this.calcTableHeight())
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.handleResize)
  },
  methods: {
    handleResize() {
      this.calcTableHeight()
    },
    calcTableHeight() {
      this.$nextTick(() => {
        if (this.$refs.tableWrapper) {
          const top = this.$refs.tableWrapper.getBoundingClientRect().top
          this.tableMaxHeight = window.innerHeight - top - 20
        }
      })
    },
    async fetchData() {
      this.loading = true
      try {
        // 获取部门列表
        const deptRes = await getDepartments()
        this.departments = deptRes.data || []

        // 获取所有用户
        const userRes = await getUserPermissions()
        this.users = userRes.data || []

        // 获取每个部门的岗位
        for (const dept of this.departments) {
          try {
            const posRes = await getDepartmentPositions(dept.id)
            this.$set(this.deptPositions, dept.id, (posRes.data || []).map(p => p.position_name))
          } catch (e) {
            this.$set(this.deptPositions, dept.id, [])
          }
        }
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
        this.$nextTick(() => this.calcTableHeight())
      }
    },
    async fetchPermissionRules() {
      try {
        const res = await getPermissionRules()
        const rules = (res.data || []).map(r => ({
          ...r,
          _rules: JSON.parse(r.rules_json || '[]')
        }))
        if (rules.length > 0) {
          // 提取所有系统和角色
          const sysMap = {}
          for (const sr of rules[0]._rules) {
            sysMap[sr.system] = (Array.isArray(sr.roles) ? sr.roles : []).map(r => r.name)
          }
          this.allSystems = Object.keys(sysMap)
          this.allSystemRoles = sysMap
        }
      } catch (e) {
        console.error(e)
      }
    },
    parseSystemRoles(json) {
      if (!json) return []
      try {
        const arr = JSON.parse(json)
        return Array.isArray(arr) ? arr.filter(sr => sr.roles && sr.roles.length > 0) : []
      } catch (e) {
        return []
      }
    },
    getSystemRoles(systemName) {
      return this.allSystemRoles[systemName] || []
    },
    handleSystemChange() {
      this.tempSelectedRoles = []
    },
    addSystemRoles() {
      if (!this.selectedSystem || this.tempSelectedRoles.length === 0) return
      // 检查是否已有该系统的角色
      const existing = this.userForm.systemRoles.find(sr => sr.system === this.selectedSystem)
      if (existing) {
        // 合并角色（去重）
        const merged = [...new Set([...existing.roles, ...this.tempSelectedRoles])]
        existing.roles = merged
      } else {
        this.userForm.systemRoles.push({
          system: this.selectedSystem,
          roles: [...this.tempSelectedRoles]
        })
      }
      this.selectedSystem = ''
      this.tempSelectedRoles = []
    },
    handleDeptChange() {
      this.userForm.position_name = ''
    },
    handleTabClick() {
      // tab 切换后自动过滤
    },
    resetUserForm() {
      // 重置表单状态
      this.isEdit = false
      this.editingUserId = null
      this.userForm = {
        name: '',
        department_id: null,
        position_name: '',
        systemRoles: []
      }
      this.selectedSystem = ''
      this.tempSelectedRoles = []
      if (this.$refs.userForm) {
        this.$refs.userForm.clearValidate()
      }
    },
    openUserForm(row) {
      // 先重置状态，确保干净
      this.resetUserForm()
      
      if (row && typeof row === 'object' && row.id !== undefined && row.id !== null) {
        // 编辑
        this.isEdit = true
        this.editingUserId = row.id
        console.log('编辑模式: isEdit=', this.isEdit, ', editingUserId=', this.editingUserId)
        this.userForm = {
          name: row.name,
          department_id: row.department_id,
          position_name: row.position_name,
          systemRoles: this.parseSystemRoles(row.system_roles_json).map(sr => ({
            system: sr.system,
            roles: [...sr.roles]
          }))
        }
      } else {
        // 新增
        this.isEdit = false
        this.editingUserId = null
        console.log('新增模式: isEdit=', this.isEdit, ', editingUserId=', this.editingUserId)
        this.userForm = {
          name: '',
          department_id: null,
          position_name: '',
          systemRoles: []
        }
      }
      this.selectedSystem = ''
      this.tempSelectedRoles = []
      this.userFormVisible = true
      this.$nextTick(() => {
        if (this.$refs.userForm) this.$refs.userForm.clearValidate()
      })
    },
    async handleSave() {
      this.$refs.userForm.validate(async (valid) => {
        if (!valid) return

        try {
          const dualToken = await this.$refs.dualControl.open()

          const submitData = {
            name: this.userForm.name,
            department_id: this.userForm.department_id,
            position_name: this.userForm.position_name,
            system_roles_json: JSON.stringify(this.userForm.systemRoles)
          }

          if (this.isEdit) {
            await updateUserPermission(this.editingUserId, submitData, dualToken)
            this.$message.success('更新成功')
          } else {
            await createUserPermission(submitData, dualToken)
            this.$message.success('创建成功')
          }
          this.userFormVisible = false
          await this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') {
            this.$message.error(e.response?.data?.message || '操作失败')
            console.error(e)
          }
        }
      })
    },
    async confirmDelete(row) {
      try {
        await this.$confirm(`确定要删除用户「${row.name}」吗？`, '删除确认', {
          confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteUserPermission(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') {
          console.error(e)
        }
      }
    }
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.page-header-right {
  display: flex;
  gap: 8px;
}
.table-wrapper {
  overflow-x: auto;
  padding-bottom: 8px;
}
.cell-roles {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 4px 0;
}
.role-tag {
  margin: 2px;
}
.empty-role {
  color: #ccc;
  font-size: 12px;
}
.role-select-area {
  width: 100%;
}
.selected-roles {
  margin-bottom: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.add-role-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.role-checkboxes {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>
