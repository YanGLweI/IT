<template>
  <el-dialog title="管理配置" :visible.sync="dialogVisible" width="750px" top="5vh" @close="handleClose">
    <el-tabs v-model="activeTab">
      <!-- Tab 1: 部门管理 -->
      <el-tab-pane label="部门管理" name="dept">
        <div class="mgmt-add-row">
          <el-input v-model="deptInput" placeholder="输入部门名称" size="small" class="mgmt-input" @keyup.enter="addDepartment" />
          <el-button type="primary" size="small" @click="addDepartment">添加</el-button>
        </div>
        <el-table :data="departments" stripe size="small" max-height="380">
          <el-table-column label="部门名称" prop="name" min-width="200" />
          <el-table-column label="操作" width="160">
            <template slot-scope="{ row }">
              <el-button size="mini" type="text" @click="startRenameDept(row)">重命名</el-button>
              <el-button size="mini" type="text" style="color:#F56C6C" @click="confirmDeleteDept(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 2: 部门岗位配置 -->
      <el-tab-pane label="部门岗位配置" name="position">
        <div class="mgmt-field">
          <label class="mgmt-label">选择部门：</label>
          <el-select v-model="selectedDeptId" placeholder="请选择部门" size="small" @change="loadDeptPositions" style="width:200px">
            <el-option v-for="dept in departments" :key="dept.id" :label="dept.name" :value="dept.id" />
          </el-select>
        </div>
        <template v-if="selectedDeptId">
          <div class="mgmt-add-row" style="margin-top:12px">
            <el-select v-model="positionInput" placeholder="选择岗位" size="small" filterable style="flex:1" @keyup.enter.native="addPosition">
              <el-option
                v-for="pos in availablePermissionPositions"
                :key="pos"
                :label="pos"
                :value="pos"
                :disabled="currentPositions.includes(pos)"
              />
            </el-select>
            <el-button type="primary" size="small" @click="addPosition" :disabled="!positionInput">添加</el-button>
          </div>
          <el-table v-if="currentPositions.length > 0" :data="currentPositions" stripe size="small" max-height="250">
            <el-table-column label="岗位名称" prop="name" min-width="200" />
            <el-table-column label="操作" width="100">
              <template slot-scope="{ row }">
                <el-button size="mini" type="text" style="color:#F56C6C" @click="removePosition(row)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="currentPositions.length === 0" description="暂无岗位" />
        </template>
        <div v-if="!selectedDeptId" style="color:#909399;text-align:center;padding:40px 0">请先选择部门</div>
      </el-tab-pane>
    </el-tabs>

    <!-- 重命名弹窗 -->
    <el-dialog title="重命名部门" :visible.sync="renameDialogVisible" width="400px" append-to-body>
      <el-form>
        <el-form-item label="部门名称" required>
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
  </el-dialog>
</template>

<script>
import { getDepartments, createDepartment, updateDepartment, deleteDepartment, getDepartmentPositions, addDepartmentPosition, removeDepartmentPosition } from '@/api/department'
import { getPermissionRules } from '@/api/permission'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  components: { DualControlDialog },
  props: {
    visible: { type: Boolean, default: false }
  },
  data() {
    return {
      activeTab: 'dept',
      departments: [],
      deptInput: '',
      // 部门岗位
      selectedDeptId: null,
      deptPositions: {}, // { deptId: [{id, position_name}] }
      positionInput: '',
      allPositions: [], // 权限规则中的所有岗位
      // 重命名
      renameDialogVisible: false,
      renameDeptId: null,
      renameValue: ''
    }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(val) { this.$emit('update:visible', val) }
    },
    currentPositions() {
      if (!this.selectedDeptId) return []
      const positions = this.deptPositions[this.selectedDeptId] || []
      return positions.map(p => ({ id: p.id, name: p.position_name }))
    },
    availablePermissionPositions() {
      return this.allPositions
    }
  },
  watch: {
    visible(val) {
      if (val) {
        this.fetchData()
      }
    }
  },
  methods: {
    async fetchData() {
      try {
        // 获取部门列表
        const deptRes = await getDepartments()
        this.departments = deptRes.data || []

        // 获取权限规则中的所有岗位
        const ruleRes = await getPermissionRules()
        const rules = ruleRes.data || []
        this.allPositions = rules.map(r => r.position_name)

        // 获取每个部门的岗位
        for (const dept of this.departments) {
          try {
            const posRes = await getDepartmentPositions(dept.id)
            this.$set(this.deptPositions, dept.id, posRes.data || [])
          } catch (e) {
            this.$set(this.deptPositions, dept.id, [])
          }
        }
      } catch (e) {
        console.error(e)
      }
    },
    // ---- 部门管理 ----
    async addDepartment() {
      const name = this.deptInput.trim()
      if (!name) {
        this.$message.warning('请输入部门名称')
        return
      }
      try {
        const dualToken = await this.$refs.dualControl.open()
        await createDepartment({ name }, dualToken)
        this.$message.success('部门添加成功')
        this.deptInput = ''
        await this.fetchData()
        this.$emit('updated')
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error(e.response?.data?.message || '添加失败')
        }
      }
    },
    startRenameDept(row) {
      this.renameDeptId = row.id
      this.renameValue = row.name
      this.renameDialogVisible = true
    },
    async confirmRename() {
      const newName = this.renameValue.trim()
      if (!newName) {
        this.$message.warning('名称不能为空')
        return
      }
      try {
        const dualToken = await this.$refs.dualControl.open()
        await updateDepartment(this.renameDeptId, { name: newName }, dualToken)
        this.$message.success('重命名成功')
        this.renameDialogVisible = false
        await this.fetchData()
        this.$emit('updated')
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error(e.response?.data?.message || '重命名失败')
        }
      }
    },
    async confirmDeleteDept(row) {
      try {
        await this.$confirm(`确定要删除部门「${row.name}」吗？`, '删除确认', {
          confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteDepartment(row.id, dualToken)
        this.$message.success('删除成功')
        if (this.selectedDeptId === row.id) {
          this.selectedDeptId = null
        }
        await this.fetchData()
        this.$emit('updated')
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error(e.response?.data?.message || '删除失败')
        }
      }
    },
    // ---- 部门岗位管理 ----
    loadDeptPositions() {
      this.positionInput = ''
    },
    async addPosition() {
      if (!this.selectedDeptId || !this.positionInput) {
        this.$message.warning('请选择岗位')
        return
      }
      try {
        const dualToken = await this.$refs.dualControl.open()
        await addDepartmentPosition(this.selectedDeptId, { position_name: this.positionInput }, dualToken)
        this.$message.success('岗位添加成功')
        this.positionInput = ''
        await this.fetchData()
        this.$emit('updated')
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error(e.response?.data?.message || '添加失败')
        }
      }
    },
    async removePosition(row) {
      try {
        const dualToken = await this.$refs.dualControl.open()
        await removeDepartmentPosition(this.selectedDeptId, row.id, dualToken)
        this.$message.success('岗位移除成功')
        await this.fetchData()
        this.$emit('updated')
      } catch (e) {
        if (e.message !== 'canceled') {
          this.$message.error(e.response?.data?.message || '移除失败')
        }
      }
    },
    handleClose() {
      this.activeTab = 'dept'
      this.selectedDeptId = null
      this.positionInput = ''
    }
  }
}
</script>

<style scoped>
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
