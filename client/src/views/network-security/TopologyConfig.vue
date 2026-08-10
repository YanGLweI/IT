<template>
  <div class="topology-config">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">拓扑关系配置</h2>
        <p class="page-subtitle">配置防火墙树结构与区域挂靠关系，用于生成网络拓扑图</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-plus" @click="openCreate">新增防火墙节点</el-button>
        <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchAll" :loading="loading">刷新</el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 防火墙树配置 -->
      <el-tab-pane label="防火墙树配置" name="tree">
        <el-table :data="flatNodes" v-loading="loading" border size="small" style="width: 100%">
          <el-table-column label="防火墙" min-width="220">
            <template slot-scope="{ row }">
              <span :style="{ paddingLeft: row.depth * 22 + 'px' }" class="node-name-cell">
                <i v-if="row.depth > 0" class="el-icon-share tree-indent"></i>
                <svg-icon name="firewall" :size="15" class="fw-icon" />
                <template v-if="!row.invalid">
                  <span class="fw-name">{{ row.asset.computer_name }}</span>
                  <span class="fw-ip" v-if="row.asset.ip_address">{{ row.asset.ip_address }}</span>
                </template>
                <el-tag v-else type="danger" size="mini">无效节点（设备已删除）</el-tag>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="上级防火墙" min-width="150">
            <template slot-scope="{ row }">
              <span v-if="row.parent_id === 0" class="text-muted">—（顶级）</span>
              <span v-else>{{ nodeNameMap[row.parent_id] || '未知' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="挂靠区域" min-width="200">
            <template slot-scope="{ row }">
              <el-tag v-for="r in row.regions" :key="r.id" size="mini" style="margin: 2px">{{ r.name }}</el-tag>
              <span v-if="!row.regions.length" class="text-muted">暂无</span>
            </template>
          </el-table-column>
          <el-table-column prop="sort_order" label="排序" width="70" align="center" />
          <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
          <el-table-column label="操作" width="140" align="center">
            <template slot-scope="{ row }">
              <div class="op-btns">
                <el-button size="mini" @click="openEdit(row)">编辑</el-button>
                <el-button size="mini" type="danger" @click="handleDelete(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="!loading && nodes.length === 0" description="暂无防火墙节点，请点击右上角新增" :image-size="80" />
      </el-tab-pane>

      <!-- 区域分配 -->
      <el-tab-pane label="区域分配" name="region">
        <div class="region-toolbar">
          <span class="text-muted">为每个区域选择其所属防火墙（一个区域只能挂靠一个防火墙）</span>
          <el-button type="primary" size="small" :disabled="!regionDirty" :loading="savingRegions" @click="handleSaveRegions">保存分配</el-button>
        </div>
        <el-table :data="regionRows" v-loading="loading" border size="small" style="width: 100%">
          <el-table-column prop="name" label="区域名称" min-width="180" />
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          <el-table-column label="所属防火墙" min-width="260">
            <template slot-scope="{ row }">
              <el-select v-model="regionAssign[row.id]" placeholder="未分配" clearable size="small" style="width: 100%" @change="regionDirty = true">
                <el-option-group v-for="node in validNodes" :key="node.id" :label="nodeLabel(node)">
                  <el-option :label="nodeOptionLabel(node)" :value="node.id" />
                </el-option-group>
              </el-select>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 新增/编辑节点弹窗 -->
    <el-dialog class="vault-dialog" :title="isEdit ? '编辑防火墙节点' : '新增防火墙节点'" :visible.sync="showForm" width="520px" :close-on-click-modal="false">
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item v-if="!isEdit" label="防火墙设备" prop="asset_id">
          <el-select v-model="form.asset_id" placeholder="选择防火墙设备" filterable style="width: 100%">
            <el-option v-for="a in availableAssets" :key="a.id" :label="`${a.computer_name} (${a.ip_address || '无IP'})`" :value="a.id" />
          </el-select>
          <div class="form-tip" v-if="!firewallRegionFound">未找到"Firewall"区域，请先在区域管理中创建</div>
          <div class="form-tip" v-else-if="availableAssets.length === 0">"Firewall"区域下没有可添加的设备</div>
        </el-form-item>
        <el-form-item v-else label="防火墙设备">
          <el-input :value="editingName" disabled />
        </el-form-item>
        <el-form-item label="上级防火墙">
          <el-select v-model="form.parent_id" placeholder="无（顶级防火墙）" clearable filterable style="width: 100%">
            <el-option v-for="opt in parentOptions" :key="opt.id" :label="opt.label" :value="opt.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="选填" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showForm = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getFirewallTopologyTree, createFirewallNode, updateFirewallNode, deleteFirewallNode, saveRegionFirewall } from '@/api/firewall_topology'
import { getRegions } from '@/api/region'
import { getAssets } from '@/api/asset'
import DualControlDialog from '@/components/DualControlDialog.vue'
import SvgIcon from '@/components/SvgIcon.vue'

export default {
  name: 'TopologyConfig',
  components: { DualControlDialog, SvgIcon },
  data() {
    return {
      loading: false,
      activeTab: 'tree',
      nodes: [],
      regions: [],
      firewallAssets: [],
      firewallRegionFound: true,
      // 区域分配
      regionAssign: {},
      savedAssign: {},
      regionDirty: false,
      savingRegions: false,
      // 弹窗
      showForm: false,
      isEdit: false,
      editingId: null,
      submitting: false,
      form: { asset_id: null, parent_id: null, sort_order: 0, remark: '' },
      formRules: {
        asset_id: [{ required: true, message: '请选择防火墙设备', trigger: 'change' }]
      }
    }
  },
  computed: {
    validNodes() {
      return this.nodes.filter(n => !n.invalid)
    },
    nodeNameMap() {
      const map = {}
      this.nodes.forEach(n => {
        map[n.id] = n.invalid ? '无效节点#' + n.id : n.asset.computer_name
      })
      return map
    },
    // 按树结构展平（含层级深度）
    flatNodes() {
      const roots = this.nodes.filter(n => n.parent_id === 0).sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
      const result = []
      const walk = (node, depth) => {
        result.push({ ...node, depth })
        this.nodes
          .filter(n => n.parent_id === node.id)
          .sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
          .forEach(child => walk(child, depth + 1))
      }
      roots.forEach(r => walk(r, 0))
      // 孤儿节点（上级已被删除）兜底展示
      const ids = new Set(result.map(n => n.id))
      this.nodes.filter(n => !ids.has(n.id)).forEach(n => result.push({ ...n, depth: 0 }))
      return result
    },
    // 新增弹窗中可选的防火墙资产（未建节点的）
    availableAssets() {
      const used = new Set(this.nodes.map(n => n.asset_id))
      return this.firewallAssets.filter(a => !used.has(a.id))
    },
    // 上级防火墙候选（编辑时排除自身及其后代，防止成环）
    parentOptions() {
      const excluded = new Set()
      if (this.isEdit && this.editingId) {
        excluded.add(this.editingId)
        let changed = true
        while (changed) {
          changed = false
          this.nodes.forEach(n => {
            if (excluded.has(n.parent_id) && !excluded.has(n.id)) {
              excluded.add(n.id)
              changed = true
            }
          })
        }
      }
      return this.flatNodes
        .filter(n => !excluded.has(n.id) && !n.invalid)
        .map(n => ({ id: n.id, label: '　'.repeat(n.depth) + this.nodeLabel(n) }))
    },
    editingName() {
      const node = this.nodes.find(n => n.id === this.editingId)
      if (!node || node.invalid) return '无效节点'
      return `${node.asset.computer_name} (${node.asset.ip_address || '无IP'})`
    },
    regionRows() {
      return this.regions
    }
  },
  mounted() {
    this.fetchAll()
  },
  methods: {
    nodeLabel(node) {
      return node.invalid ? `无效节点#${node.id}` : node.asset.computer_name
    },
    nodeOptionLabel(node) {
      return node.asset.ip_address ? `${node.asset.computer_name} (${node.asset.ip_address})` : node.asset.computer_name
    },
    async fetchAll() {
      this.loading = true
      try {
        await Promise.all([this.fetchTree(), this.fetchRegions(), this.fetchFirewallAssets()])
        this.buildRegionAssign()
      } finally {
        this.loading = false
      }
    },
    async fetchTree() {
      const res = await getFirewallTopologyTree()
      this.nodes = res.data || []
    },
    async fetchRegions() {
      const res = await getRegions()
      this.regions = res.data || []
    },
    async fetchFirewallAssets() {
      try {
        const regionsRes = await getRegions()
        const fwRegion = (regionsRes.data || []).find(r => r.name === 'Firewall')
        if (!fwRegion) {
          this.firewallRegionFound = false
          this.firewallAssets = []
          return
        }
        this.firewallRegionFound = true
        const res = await getAssets({ page: 1, page_size: 500, region_id: fwRegion.id })
        this.firewallAssets = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    buildRegionAssign() {
      const assign = {}
      this.regions.forEach(r => { assign[r.id] = null })
      this.nodes.forEach(n => {
        (n.regions || []).forEach(r => { assign[r.id] = n.id })
      })
      this.regionAssign = assign
      this.savedAssign = { ...assign }
      this.regionDirty = false
    },
    // 新增
    openCreate() {
      this.isEdit = false
      this.editingId = null
      this.form = { asset_id: null, parent_id: null, sort_order: 0, remark: '' }
      this.showForm = true
      this.$nextTick(() => this.$refs.formRef && this.$refs.formRef.clearValidate())
    },
    // 编辑
    openEdit(row) {
      this.isEdit = true
      this.editingId = row.id
      this.form = {
        asset_id: row.asset_id,
        parent_id: row.parent_id || null,
        sort_order: row.sort_order,
        remark: row.remark || ''
      }
      this.showForm = true
    },
    async handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        this.submitting = true
        try {
          const dualToken = await this.$refs.dualControl.open()
          const payload = {
            parent_id: this.form.parent_id || 0,
            sort_order: this.form.sort_order,
            remark: this.form.remark
          }
          if (this.isEdit) {
            await updateFirewallNode(this.editingId, payload, dualToken)
            this.$message.success('更新成功')
          } else {
            await createFirewallNode({ ...payload, asset_id: this.form.asset_id }, dualToken)
            this.$message.success('创建成功')
          }
          this.showForm = false
          this.fetchAll()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        } finally {
          this.submitting = false
        }
      })
    },
    async handleDelete(row) {
      try {
        await this.$confirm(`确定要删除防火墙节点"${this.nodeLabel(row)}"吗？其挂靠的区域将变为未分配。`, '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteFirewallNode(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchAll()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    async handleSaveRegions() {
      this.savingRegions = true
      try {
        const dualToken = await this.$refs.dualControl.open()
        const items = Object.keys(this.regionAssign).map(k => ({
          region_id: Number(k),
          firewall_node_id: this.regionAssign[k] || 0
        }))
        await saveRegionFirewall(items, dualToken)
        this.$message.success('保存成功')
        this.fetchAll()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      } finally {
        this.savingRegions = false
      }
    }
  }
}
</script>

<style scoped>
.topology-config {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow: auto;
  display: flex;
  flex-direction: column;
}

/* --- 页面头部 --- */
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
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.header-actions .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 500;
}
.header-actions .el-button--primary:hover {
  background: #2563eb;
}
.header-actions .el-button--default {
  border-radius: 10px;
  border-color: #e2e8f0;
  color: #64748b;
}
.header-actions .el-button--default:hover {
  border-color: #94a3b8;
  color: #1e293b;
}

.node-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.tree-indent {
  color: #cbd5e1;
}
.fw-icon {
  color: #3b82f6;
  flex-shrink: 0;
}
.fw-name {
  font-weight: 500;
  color: #1e293b;
}
.fw-ip {
  color: #64748b;
  font-size: 12px;
  font-family: 'SF Mono', Menlo, monospace;
}
.text-muted {
  color: #94a3b8;
}

.region-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 13px;
}

.op-btns {
  display: flex;
  gap: 6px;
  justify-content: center;
}

.form-tip {
  font-size: 12px;
  color: #f56c6c;
  line-height: 1.6;
  margin-top: 4px;
}
</style>
