<template>
  <div class="permission-page">
    <el-card>
      <div slot="header" class="page-header">
        <span>岗位权限设置规则</span>
        <el-button type="primary" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
      </div>

      <div class="table-wrapper">
        <el-table
          :data="rules"
          border
          stripe
          max-height="calc(100vh - 200px)"
          style="width: 100%"
          v-loading="loading"
        >
          <el-table-column label="系统\角色\岗位" width="160" fixed>
            <template slot-scope="{ row }">
              <strong>{{ row.position_name }}</strong>
            </template>
          </el-table-column>

          <el-table-column
            v-for="sys in systems"
            :key="sys"
            :label="sys"
            min-width="180"
          >
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
import { getPermissionRules, updatePermissionRule } from '@/api/permission'

export default {
  name: 'PermissionList',
  data() {
    return {
      loading: false,
      saving: false,
      rules: [],
      // 所有系统名称（按顺序从数据中提取）
      systems: [],
      // 编辑相关
      dialogVisible: false,
      editingPosition: '',
      editingSystem: '',
      editingRoleName: '',
      editingNewStatus: false,
      editingRuleId: null,
      // 存放待保存的变更队列
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
    toggleRole(row, systemName, role) {
      this.editingRuleId = row.id
      this.editingPosition = row.position_name
      this.editingSystem = systemName
      this.editingRoleName = role.name
      this.editingNewStatus = !role.enabled
      this.dialogVisible = true
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
</style>
