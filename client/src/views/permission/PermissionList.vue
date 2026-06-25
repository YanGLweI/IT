<template>
  <div class="permission-page">
    <el-card>
      <div slot="header" class="page-header">
        <span>岗位权限设置规则</span>
        <div class="page-header-right">
          <el-button type="success" size="small" icon="el-icon-plus" @click="showAddPosition = true">添加岗位</el-button>
          <el-button type="warning" size="small" icon="el-icon-plus" @click="showAddSystem = true">添加系统</el-button>
          <el-button type="primary" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>

      <div class="table-wrapper">
        <el-table
          :data="rules"
          border
          stripe
          style="width: 100%"
          v-loading="loading"
        >
          <el-table-column label="岗位" width="200" fixed>
            <template slot-scope="{ row }">
              <div class="position-cell">
                <strong>{{ row.position_name }}</strong>
                <el-button type="danger" size="mini" icon="el-icon-delete" circle
                  @click="confirmDeletePosition(row)" title="删除岗位"></el-button>
              </div>
            </template>
          </el-table-column>

          <el-table-column
            v-for="sys in systems"
            :key="sys"
            :label="sys"
            min-width="180"
          >
            <template slot="header" slot-scope="{ column }">
              <div class="system-header">
                <span>{{ column.label }}</span>
                <el-button type="danger" size="mini" icon="el-icon-delete" circle
                  @click="confirmDeleteSystem(column.label)" title="删除系统"></el-button>
              </div>
            </template>
            <template slot-scope="{ row }">
              <div class="cell-roles">
                <el-tag
                  v-for="role in getCellRoles(row, sys)"
                  :key="role.name"
                  :type="role.enabled ? 'success' : 'info'"
                  :effect="role.enabled ? 'dark' : 'plain'"
                  size="small"
                  class="role-tag"
                  @click="toggleRole(row, sys, role)"
                >
                  {{ role.name }}
                </el-tag>
                <span v-if="getCellRoles(row, sys).length === 0" class="empty-role">-</span>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 添加岗位弹窗 -->
    <el-dialog title="添加岗位" :visible.sync="showAddPosition" width="400px">
      <el-form :model="addPositionForm">
        <el-form-item label="岗位名称" required>
          <el-input v-model="addPositionForm.name" placeholder="请输入岗位名称" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showAddPosition = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleAddPosition">确定</el-button>
      </span>
    </el-dialog>

    <!-- 添加系统弹窗 -->
    <el-dialog title="添加系统" :visible.sync="showAddSystem" width="500px">
      <el-form :model="addSystemForm">
        <el-form-item label="系统名称" required>
          <el-input v-model="addSystemForm.name" placeholder="请输入系统名称" />
        </el-form-item>
        <el-form-item label="角色列表" required>
          <el-input
            v-model="addSystemForm.roles"
            type="textarea"
            :rows="3"
            placeholder="请输入角色名称，用逗号分隔，例如：管理员,操作员,审计员"
          />
          <span class="form-tip">多个角色请用逗号（,）分隔</span>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showAddSystem = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleAddSystem">确定</el-button>
      </span>
    </el-dialog>

    <!-- 保存确认弹窗 -->
    <el-dialog
      title="保存确认"
      :visible.sync="dialogVisible"
      width="400px"
    >
      <div>
        <p><strong>岗位：</strong>{{ editingPosition }}</p>
        <p><strong>系统：</strong>{{ editingSystem }}</p>
        <p><strong>角色：</strong>{{ editingRoleName }}</p>
        <p><strong>操作：</strong>{{ editingNewStatus ? '授权' : '取消授权' }}</p>
      </div>
      <span slot="footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="confirmSave">确定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getPermissionRules, createPermissionRule, addSystemToPermissions, updatePermissionRule, deletePermissionRule, removeSystemFromPermissions } from '@/api/permission'

export default {
  name: 'PermissionList',
  data() {
    return {
      loading: false,
      saving: false,
      rules: [],
      // 所有系统名称（按顺序从数据中提取）
      systems: [],
      // 添加岗位
      showAddPosition: false,
      addPositionForm: { name: '' },
      // 添加系统
      showAddSystem: false,
      addSystemForm: { name: '', roles: '' },
      // 编辑相关
      dialogVisible: false,
      editingPosition: '',
      editingSystem: '',
      editingRoleName: '',
      editingNewStatus: false,
      editingRuleId: null,
      pendingChanges: []
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const res = await getPermissionRules()
        this.rules = (res.data || []).map(r => ({
          ...r,
          _rules: JSON.parse(r.rules_json || '[]')
        }))
        // 提取所有系统名称
        if (this.rules.length > 0) {
          const sysNames = this.rules[0]._rules.map(sr => sr.system)
          this.systems = sysNames
        }
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    getCellRoles(row, systemName) {
      const sysRule = row._rules.find(sr => sr.system === systemName)
      return sysRule ? sysRule.roles : []
    },
    async handleAddPosition() {
      const name = this.addPositionForm.name.trim()
      if (!name) {
        this.$message.warning('请输入岗位名称')
        return
      }
      this.saving = true
      try {
        await createPermissionRule({ position_name: name })
        this.$message.success('岗位添加成功')
        this.showAddPosition = false
        this.addPositionForm.name = ''
        await this.fetchData()
      } catch (e) {
        this.$message.error('添加失败')
        console.error(e)
      } finally {
        this.saving = false
      }
    },
    async handleAddSystem() {
      const name = this.addSystemForm.name.trim()
      const rolesStr = this.addSystemForm.roles.trim()
      if (!name) {
        this.$message.warning('请输入系统名称')
        return
      }
      if (!rolesStr) {
        this.$message.warning('请输入角色列表')
        return
      }
      const roles = rolesStr.split(',').map(s => s.trim()).filter(s => s)
      if (roles.length === 0) {
        this.$message.warning('请输入有效的角色名称')
        return
      }
      this.saving = true
      try {
        await addSystemToPermissions({ system_name: name, roles })
        this.$message.success('系统添加成功')
        this.showAddSystem = false
        this.addSystemForm = { name: '', roles: '' }
        await this.fetchData()
      } catch (e) {
        this.$message.error('添加失败')
        console.error(e)
      } finally {
        this.saving = false
      }
    },
    toggleRole(row, systemName, role) {
      this.editingRuleId = row.id
      this.editingPosition = row.position_name
      this.editingSystem = systemName
      this.editingRoleName = role.name
      this.editingNewStatus = !role.enabled
      this.dialogVisible = true
    },
    confirmDeletePosition(row) {
      this.$confirm(`确定要删除岗位「${row.position_name}」吗？`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        await deletePermissionRule(row.id)
        this.$message.success('删除成功')
        this.fetchData()
      }).catch(() => {})
    },
    confirmDeleteSystem(systemName) {
      this.$confirm(`确定要删除系统「${systemName}」吗？将从所有岗位移除该系统及其角色。`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        await removeSystemFromPermissions({ system_name: systemName })
        this.$message.success('删除成功')
        this.fetchData()
      }).catch(() => {})
    },
    async confirmSave() {
      this.saving = true
      try {
        const rule = this.rules.find(r => r.id === this.editingRuleId)
        if (!rule) return

        // 修改本地数据
        const sysRule = rule._rules.find(sr => sr.system === this.editingSystem)
        if (sysRule) {
          const targetRole = sysRule.roles.find(r => r.name === this.editingRoleName)
          if (targetRole) {
            targetRole.enabled = this.editingNewStatus
          }
        }

        // 保存到后端
        await updatePermissionRule(rule.id, {
          rules_json: JSON.stringify(rule._rules)
        })

        this.$message.success('保存成功')
        this.dialogVisible = false
      } catch (e) {
        this.$message.error('保存失败')
        console.error(e)
      } finally {
        this.saving = false
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

.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

.table-wrapper {
  overflow-x: auto;
}

.cell-roles {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 4px 0;
}

.role-tag {
  cursor: pointer;
  margin: 2px;
  transition: all 0.2s;
}

.role-tag:hover {
  transform: scale(1.05);
}

.empty-role {
  color: #ccc;
  font-size: 12px;
}

.position-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
}

.system-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  white-space: normal;
  word-break: break-all;
  line-height: 1.3;
}
</style>
