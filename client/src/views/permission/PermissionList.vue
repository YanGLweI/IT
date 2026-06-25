<template>
  <div class="permission-page">
    <el-card>
      <div slot="header" class="page-header">
        <span>岗位权限设置规则</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-setting" @click="showManagement = true">管理配置</el-button>
          <el-button type="primary" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>

      <div class="table-wrapper" ref="tableWrapper">
        <el-table
          :data="rules"
          border
          stripe
          style="width: 100%"
          v-loading="loading"
          :max-height="tableMaxHeight"
        >
          <el-table-column label="岗位" width="200" fixed>
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

    <!-- 角色授权/取消授权确认弹窗 -->
    <el-dialog title="保存确认" :visible.sync="dialogVisible" width="400px">
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

    <!-- ==================== 管理配置弹窗 ==================== -->
    <el-dialog title="管理配置" :visible.sync="showManagement" width="750px" top="5vh">
      <el-tabs v-model="activeTab">
        <!-- Tab 1: 岗位管理 -->
        <el-tab-pane label="岗位管理" name="position">
          <div class="mgmt-add-row">
            <el-input v-model="posInput" placeholder="输入岗位名称" size="small" class="mgmt-input" @keyup.enter="addPosition" />
            <el-button type="primary" size="small" @click="addPosition">添加</el-button>
          </div>
          <el-table :data="rules" stripe size="small" max-height="380">
            <el-table-column label="#" width="64">
              <template slot-scope="{ $index }">
                <div class="sort-btns">
                  <el-button size="mini" type="text" icon="el-icon-arrow-up"
                    :disabled="$index === 0" @click="movePosition(rules[$index], 'up')" />
                  <el-button size="mini" type="text" icon="el-icon-arrow-down"
                    :disabled="$index === rules.length - 1" @click="movePosition(rules[$index], 'down')" />
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="position_name" label="岗位名称" min-width="200" />
            <el-table-column label="操作" width="160">
              <template slot-scope="{ row }">
                <el-button size="mini" type="text" @click="startRenamePosition(row)">重命名</el-button>
                <el-button size="mini" type="text" style="color:#F56C6C" @click="confirmDeletePosition(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 2: 系统管理 -->
        <el-tab-pane label="系统管理" name="system">
          <div class="mgmt-add-row">
            <el-input v-model="sysInput" placeholder="输入系统名称" size="small" class="mgmt-input" @keyup.enter="addSystem" />
            <el-button type="primary" size="small" @click="addSystem">添加</el-button>
          </div>
          <el-table :data="systemList" stripe size="small" max-height="380">
            <el-table-column label="#" width="64">
              <template slot-scope="{ $index }">
                <div class="sort-btns">
                  <el-button size="mini" type="text" icon="el-icon-arrow-up"
                    :disabled="$index === 0" @click="moveSystem(systems[$index], 'up')" />
                  <el-button size="mini" type="text" icon="el-icon-arrow-down"
                    :disabled="$index === systems.length - 1" @click="moveSystem(systems[$index], 'down')" />
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="系统名称" min-width="200" />
            <el-table-column label="操作" width="160">
              <template slot-scope="{ row }">
                <el-button size="mini" type="text" @click="startRenameSystem(row)">重命名</el-button>
                <el-button size="mini" type="text" style="color:#F56C6C" @click="confirmDeleteSystem(row.name)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 3: 角色管理 -->
        <el-tab-pane label="角色管理" name="role">
          <div class="mgmt-field">
            <label class="mgmt-label">选择系统：</label>
            <el-select v-model="selectedSystemForRole" placeholder="请选择系统" size="small" @change="loadRolesForSystem">
              <el-option v-for="sys in systems" :key="sys" :label="sys" :value="sys" />
            </el-select>
          </div>
          <div class="mgmt-add-row" style="margin-top:12px">
            <el-input v-model="roleInput" placeholder="输入角色名称" size="small" class="mgmt-input" @keyup.enter="addRole" />
            <el-button type="primary" size="small" @click="addRole" :disabled="!selectedSystemForRole">添加</el-button>
          </div>
          <el-table v-if="selectedSystemForRole && currentRoles.length > 0" :data="currentRoles" stripe size="small" max-height="250">
            <el-table-column prop="name" label="角色名称" min-width="200" />
            <el-table-column label="操作" width="160">
              <template slot-scope="{ row }">
                <el-button size="mini" type="text" @click="startRenameRole(row)">重命名</el-button>
                <el-button size="mini" type="text" style="color:#F56C6C" @click="confirmDeleteRole(row.name)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="selectedSystemForRole && currentRoles.length === 0" description="暂无角色" />
          <div v-if="!selectedSystemForRole" style="color:#909399;text-align:center;padding:40px 0">请先选择系统</div>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <!-- 重命名弹窗（复用：岗位/系统/角色） -->
    <el-dialog :title="renameTitle" :visible.sync="renameDialogVisible" width="400px">
      <el-form>
        <el-form-item :label="renameLabel" required>
          <el-input v-model="renameValue" @keyup.enter="confirmRename" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="renameDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmRename">确定</el-button>
      </span>
    </el-dialog>
    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getPermissionRules, createPermissionRule, addSystemToPermissions, updatePermissionRule, deletePermissionRule, removeSystemFromPermissions, renameSystemInPermissions, manageRolesInSystem, reorderPermissionRule, reorderSystemInPermissions } from '@/api/permission'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  components: { DualControlDialog },
  name: 'PermissionList',
  data() {
    return {
      loading: false,
      saving: false,
      rules: [],
      // 所有系统名称（按顺序从数据中提取）
      systems: [],
      // 角色授权相关
      dialogVisible: false,
      editingPosition: '',
      editingSystem: '',
      editingRoleName: '',
      editingNewStatus: false,
      editingRuleId: null,
      // 管理配置弹窗
      showManagement: false,
      activeTab: 'position',
      posInput: '',
      sysInput: '',
      selectedSystemForRole: '',
      roleInput: '',
      currentRoles: [],
      // 重命名弹窗
      renameDialogVisible: false,
      renameTarget: null, // { type: 'position'|'system'|'role', id?, oldName?, systemName? }
      renameValue: '',
      // 表格最大高度（动态按页面剩余空间计算，使横向滚动条始终可见）
      tableMaxHeight: null
    }
  },
  computed: {
    systemList() {
      return this.systems.map(name => ({ name }))
    },
    renameTitle() {
      if (!this.renameTarget) return '重命名'
      const map = { position: '重命名岗位', system: '重命名系统', role: '重命名角色' }
      return map[this.renameTarget.type] || '重命名'
    },
    renameLabel() {
      if (!this.renameTarget) return ''
      const map = { position: '岗位名称', system: '系统名称', role: '角色名称' }
      return `新${map[this.renameTarget.type]}`
    }
  },
  mounted() {
    this.fetchData()
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
        this.$nextTick(() => this.calcTableHeight())
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

        // 双控验证
        const dualToken = await this.$refs.dualControl.open()

        // 保存到后端
        await updatePermissionRule(rule.id, {
          rules_json: JSON.stringify(rule._rules)
        }, dualToken)

        this.$message.success('保存成功')
        this.dialogVisible = false
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error('保存失败')
          console.error(e)
        }
      } finally {
        this.saving = false
      }
    },

    // ---- 岗位管理 ----
    async addPosition() {
      const name = this.posInput.trim()
      if (!name) {
        this.$message.warning('请输入岗位名称')
        return
      }
      try {
        const dualToken = await this.$refs.dualControl.open()
        await createPermissionRule({ position_name: name }, dualToken)
        this.$message.success('岗位添加成功')
        this.posInput = ''
        await this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error('添加失败')
          console.error(e)
        }
      }
    },
    async confirmDeletePosition(row) {
      try {
        await this.$confirm(`确定要删除岗位「${row.position_name}」吗？`, '删除确认', {
          confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deletePermissionRule(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    startRenamePosition(row) {
      this.renameTarget = { type: 'position', id: row.id, oldName: row.position_name }
      this.renameValue = row.position_name
      this.renameDialogVisible = true
    },

    // ---- 系统管理 ----
    async addSystem() {
      const name = this.sysInput.trim()
      if (!name) {
        this.$message.warning('请输入系统名称')
        return
      }
      try {
        const dualToken = await this.$refs.dualControl.open()
        await addSystemToPermissions({ system_name: name, roles: [] }, dualToken)
        this.$message.success('系统添加成功（可到角色管理添加角色）')
        this.sysInput = ''
        await this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error('添加失败')
          console.error(e)
        }
      }
    },
    async confirmDeleteSystem(systemName) {
      try {
        await this.$confirm(`确定要删除系统「${systemName}」吗？将从所有岗位移除该系统及其角色。`, '删除确认', {
          confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await removeSystemFromPermissions({ system_name: systemName }, dualToken)
        this.$message.success('删除成功')
        if (this.selectedSystemForRole === systemName) {
          this.selectedSystemForRole = ''
          this.currentRoles = []
        }
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    startRenameSystem(row) {
      this.renameTarget = { type: 'system', oldName: row.name }
      this.renameValue = row.name
      this.renameDialogVisible = true
    },

    // ---- 排序移动（不需要双控）----
    async movePosition(row, direction) {
      try {
        await reorderPermissionRule({ id: row.id, direction })
        await this.fetchData()
      } catch (e) {
        this.$message.error(e.response?.data?.message || '移动失败')
      }
    },
    async moveSystem(systemName, direction) {
      try {
        await reorderSystemInPermissions({ system_name: systemName, direction })
        await this.fetchData()
      } catch (e) {
        this.$message.error(e.response?.data?.message || '移动失败')
      }
    },

    // ---- 角色管理 ----
    loadRolesForSystem() {
      if (!this.selectedSystemForRole || this.rules.length === 0) {
        this.currentRoles = []
        return
      }
      const sysRule = this.rules[0]._rules.find(sr => sr.system === this.selectedSystemForRole)
      this.currentRoles = sysRule ? sysRule.roles.map(r => ({ ...r })) : []
    },
    async addRole() {
      const name = this.roleInput.trim()
      if (!name) {
        this.$message.warning('请输入角色名称')
        return
      }
      // 检查是否已存在
      if (this.currentRoles.some(r => r.name === name)) {
        this.$message.warning('该角色已存在')
        return
      }
      try {
        const dualToken = await this.$refs.dualControl.open()
        await manageRolesInSystem({ system_name: this.selectedSystemForRole, action: 'add', new_name: name }, dualToken)
        this.$message.success('角色添加成功')
        this.roleInput = ''
        await this.fetchData()
        this.loadRolesForSystem()
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error('添加失败')
          console.error(e)
        }
      }
    },
    async confirmDeleteRole(roleName) {
      try {
        await this.$confirm(`确定要删除角色「${roleName}」吗？将从所有岗位中移除此角色。`, '删除确认', {
          confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await manageRolesInSystem({ system_name: this.selectedSystemForRole, action: 'delete', old_name: roleName }, dualToken)
        this.$message.success('删除成功')
        await this.fetchData()
        this.loadRolesForSystem()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    startRenameRole(row) {
      this.renameTarget = { type: 'role', systemName: this.selectedSystemForRole, oldName: row.name }
      this.renameValue = row.name
      this.renameDialogVisible = true
    },

    // ---- 通用重命名 ----
    async confirmRename() {
      const newName = this.renameValue.trim()
      if (!newName) {
        this.$message.warning('名称不能为空')
        return
      }
      const target = this.renameTarget
      if (!target) return
      if (newName === target.oldName) {
        this.renameDialogVisible = false
        return
      }
      try {
        const dualToken = await this.$refs.dualControl.open()
        switch (target.type) {
          case 'position':
            await updatePermissionRule(target.id, { position_name: newName }, dualToken)
            break
          case 'system':
            await renameSystemInPermissions({ old_name: target.oldName, new_name: newName }, dualToken)
            break
          case 'role':
            await manageRolesInSystem({
              system_name: target.systemName,
              action: 'rename',
              old_name: target.oldName,
              new_name: newName
            }, dualToken)
            break
        }
        this.$message.success('重命名成功')
        this.renameDialogVisible = false
        await this.fetchData()
        if (target.type === 'role') this.loadRolesForSystem()
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error('重命名失败')
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

.sort-btns {
  display: flex;
  gap: 2px;
  justify-content: center;
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

.mgmt-add-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.mgmt-input {
  flex: 1;
}

.mgmt-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mgmt-label {
  white-space: nowrap;
  color: #606266;
  font-size: 14px;
}
</style>
