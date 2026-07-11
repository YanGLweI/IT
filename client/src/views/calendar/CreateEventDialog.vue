<template>
  <el-dialog
    :title="mode === 'create' ? '新建日程' : '编辑日程'"
    :visible.sync="dialogVisible"
    width="520px"
    :close-on-click-modal="false"
    custom-class="create-event-dialog"
    @close="handleClose"
  >
    <el-form ref="form" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" placeholder="请输入日程标题" maxlength="200" show-word-limit />
      </el-form-item>

      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="3" placeholder="日程描述（可选）" />
      </el-form-item>

      <el-form-item label="参与人">
        <el-select
          v-model="form.participants"
          multiple
          filterable
          placeholder="选择参与者"
          style="width: 100%"
          :loading="ldapLoading"
        >
          <el-option
            v-for="user in ldapUsers"
            :key="user.sAMAccountName"
            :label="user.display_name"
            :value="user.sAMAccountName"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="全天">
        <el-switch v-model="form.is_all_day" />
      </el-form-item>

      <el-form-item label="开始时间" v-if="!form.is_all_day">
        <el-date-picker
          v-model="form.start_time"
          type="datetime"
          placeholder="选择开始时间"
          value-format="yyyy-MM-dd HH:mm:ss"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="结束时间" v-if="!form.is_all_day">
        <el-date-picker
          v-model="form.end_time"
          type="datetime"
          placeholder="选择结束时间"
          value-format="yyyy-MM-dd HH:mm:ss"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="开始日期" v-if="form.is_all_day">
        <el-date-picker
          v-model="form.start_date"
          type="date"
          placeholder="选择开始日期"
          value-format="yyyy-MM-dd"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="结束日期" v-if="form.is_all_day">
        <el-date-picker
          v-model="form.end_date"
          type="date"
          placeholder="选择结束日期"
          value-format="yyyy-MM-dd"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="重复">
        <el-select v-model="selectedRepeatOption" placeholder="选择重复规则" style="width: 100%" @change="handleRepeatChange">
          <el-option
            v-for="opt in repeatOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <div slot="footer" class="dialog-footer">
      <el-button @click="handleClose">取消</el-button>
      <el-button v-if="mode === 'edit'" type="danger" @click="handleDelete">删除</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ mode === 'create' ? '确定' : '保存' }}
      </el-button>
    </div>

    <repeat-rule-dialog
      :visible="customDialogVisible"
      :rule="customRule"
      @close="customDialogVisible = false"
      @confirm="handleCustomRuleConfirm"
    />
  </el-dialog>
</template>

<script>
import { getLDAPUsers, createCalendar, updateCalendar, deleteCalendar, checkConflict } from '@/api/calendar'
import RepeatRuleDialog from './RepeatRuleDialog.vue'

export default {
  name: 'CreateEventDialog',
  components: { RepeatRuleDialog },
  props: {
    visible: { type: Boolean, default: false },
    mode: { type: String, default: 'create' },
    event: { type: Object, default: null }
  },
  data() {
    return {
      form: {
        title: '',
        description: '',
        participants: [],
        is_all_day: false,
        start_time: '',
        end_time: '',
        start_date: '',
        end_date: ''
      },
      rules: {
        title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
      },
      ldapUsers: [],
      ldapLoading: false,
      submitting: false,
      selectedRepeatOption: 'none',
      customRule: null,
      customDialogVisible: false,
      repeatOptions: []
    }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(val) { this.$emit('close') }
    }
  },
  watch: {
    visible(val) {
      if (val) {
        this.loadLDAPUsers()
        this.initForm()
      }
    },
    'form.start_time'(val) {
      if (val) this.generateRepeatOptions()
    },
    'form.start_date'(val) {
      if (val) this.generateRepeatOptions()
    }
  },
  methods: {
    initForm() {
      if (this.mode === 'edit' && this.event) {
        this.form.title = this.event.title
        this.form.description = this.event.description || ''
        this.form.is_all_day = this.event.is_all_day
        this.form.start_time = this.formatDateTime(this.event.start_time)
        this.form.end_time = this.formatDateTime(this.event.end_time)
        this.form.participants = (this.event.participants || []).map(p => p.user_dn)
        this.customRule = this.event.repeat_rule_json ? JSON.parse(this.event.repeat_rule_json) : null
        this.selectedRepeatOption = this.customRule ? this.customRule.type : 'none'
      } else {
        this.form = {
          title: '',
          description: '',
          participants: [],
          is_all_day: false,
          start_time: '',
          end_time: '',
          start_date: '',
          end_date: ''
        }
        this.selectedRepeatOption = 'none'
        this.customRule = null
      }
      this.generateRepeatOptions()
    },
    formatDateTime(dateStr) {
      if (!dateStr) return ''
      const d = new Date(dateStr)
      const y = d.getFullYear()
      const m = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      const h = String(d.getHours()).padStart(2, '0')
      const min = String(d.getMinutes()).padStart(2, '0')
      const s = String(d.getSeconds()).padStart(2, '0')
      return `${y}-${m}-${day} ${h}:${min}:${s}`
    },
    async loadLDAPUsers() {
      this.ldapLoading = true
      try {
        const res = await getLDAPUsers()
        if (res && res.code === 200) {
          this.ldapUsers = res.data || []
        }
      } catch (err) {
        console.error('获取LDAP用户失败:', err)
      } finally {
        this.ldapLoading = false
      }
    },
    generateRepeatOptions() {
      const dateStr = this.form.is_all_day ? this.form.start_date : this.form.start_time
      if (!dateStr) {
        this.repeatOptions = [{ label: '不重复', value: 'none' }]
        return
      }

      const startDate = new Date(dateStr)
      if (isNaN(startDate.getTime())) {
        this.repeatOptions = [{ label: '不重复', value: 'none' }]
        return
      }

      const day = startDate.getDate()
      const weekday = startDate.getDay()
      const month = startDate.getMonth() + 1
      const weekOfMonth = this.getWeekOfMonth(startDate)
      const weekdayNames = ['日', '一', '二', '三', '四', '五', '六']
      const ordinals = ['一', '二', '三', '四', '五']

      this.repeatOptions = [
        { label: '不重复', value: 'none', data: null },
        { label: '每天', value: 'daily', data: { type: 'daily', interval: 1 } },
        { label: `每周${weekdayNames[weekday]}`, value: 'weekly', data: { type: 'weekly', interval: 1, weekday } },
        { label: `每月第${ordinals[weekOfMonth - 1]}个周${weekdayNames[weekday]}`, value: 'monthly_week', data: { type: 'monthly_week', interval: 1, weekOfMonth, weekday } },
        { label: `每月${day}日`, value: 'monthly_day', data: { type: 'monthly_day', interval: 1, monthDay: day } },
        { label: `每年${month}月${day}日`, value: 'yearly', data: { type: 'yearly', interval: 1, monthOfYear: month, monthDay: day } },
        { label: '每个工作日（周一至周五）', value: 'workday', data: { type: 'workday', interval: 1 } },
        { label: '自定义...', value: 'custom', data: null }
      ]
    },
    getWeekOfMonth(date) {
      const firstDay = new Date(date.getFullYear(), date.getMonth(), 1)
      const firstDayOfWeek = firstDay.getDay()
      const offset = firstDayOfWeek === 0 ? 6 : firstDayOfWeek - 1
      return Math.ceil((date.getDate() + offset) / 7)
    },
    handleRepeatChange(val) {
      if (val === 'custom') {
        this.customDialogVisible = true
      } else {
        const opt = this.repeatOptions.find(o => o.value === val)
        this.customRule = opt ? opt.data : null
      }
    },
    handleCustomRuleConfirm(rule) {
      this.customRule = rule
      this.customDialogVisible = false
    },
    buildRepeatRuleJSON() {
      if (this.selectedRepeatOption === 'none' || this.selectedRepeatOption === 'custom' && !this.customRule) {
        return ''
      }
      if (this.selectedRepeatOption === 'custom') {
        return JSON.stringify(this.customRule)
      }
      const opt = this.repeatOptions.find(o => o.value === this.selectedRepeatOption)
      return opt && opt.data ? JSON.stringify(opt.data) : ''
    },
    async handleSubmit() {
      try {
        await this.$refs.form.validate()
      } catch {
        return
      }

      let startTime, endTime
      if (this.form.is_all_day) {
        if (!this.form.start_date || !this.form.end_date) {
          this.$message.warning('请选择日期')
          return
        }
        startTime = `${this.form.start_date}T00:00:00+08:00`
        endTime = `${this.form.end_date}T23:59:59+08:00`
      } else {
        if (!this.form.start_time || !this.form.end_time) {
          this.$message.warning('请选择时间')
          return
        }
        startTime = this.form.start_time.replace(' ', 'T') + '+08:00'
        endTime = this.form.end_time.replace(' ', 'T') + '+08:00'
      }

      if (new Date(endTime) <= new Date(startTime)) {
        this.$message.warning('结束时间必须大于开始时间')
        return
      }

      // 冲突检测
      const participants = this.form.participants.map(username => {
        const user = this.ldapUsers.find(u => u.sAMAccountName === username)
        return {
          user_dn: username,
          display_name: user ? user.display_name : username
        }
      })

      try {
        const conflictRes = await checkConflict({
          start_time: startTime,
          end_time: endTime,
          repeat_rule_json: this.buildRepeatRuleJSON(),
          participants,
          exclude_id: this.mode === 'edit' ? this.event.id : null
        })

        if (conflictRes && conflictRes.data && conflictRes.data.length > 0) {
          const conflictNames = [...new Set(conflictRes.data.map(c => c.user_name))].join(', ')
          try {
            await this.$confirm(`以下参与者在该时间段已有日程: ${conflictNames}，是否继续？`, '时间冲突提示', {
              confirmButtonText: '继续提交',
              cancelButtonText: '返回修改',
              type: 'warning'
            })
          } catch {
            return
          }
        }
      } catch (err) {
        console.error('冲突检测失败:', err)
      }

      this.submitting = true
      try {
        const data = {
          title: this.form.title,
          description: this.form.description,
          start_time: startTime,
          end_time: endTime,
          is_all_day: this.form.is_all_day,
          repeat_rule_json: this.buildRepeatRuleJSON(),
          participants
        }

        let res
        if (this.mode === 'edit') {
          res = await updateCalendar(this.event.id, data)
        } else {
          res = await createCalendar(data)
        }

        if (res && res.code === 200) {
          this.$message.success(this.mode === 'create' ? '创建成功' : '更新成功')
          this.$emit('saved')
        } else {
          this.$message.error(res ? res.message : '操作失败')
        }
      } catch (err) {
        this.$message.error('操作失败: ' + (err.response ? err.response.data.message : err.message))
      } finally {
        this.submitting = false
      }
    },
    async handleDelete() {
      try {
        await this.$confirm('确定要删除此日程吗？', '确认删除', {
          type: 'warning'
        })
      } catch {
        return
      }

      try {
        const res = await deleteCalendar(this.event.id)
        if (res && res.code === 200) {
          this.$message.success('删除成功')
          this.$emit('saved')
        } else {
          this.$message.error(res ? res.message : '删除失败')
        }
      } catch (err) {
        this.$message.error('删除失败')
      }
    },
    handleClose() {
      this.$emit('close')
    }
  }
}
</script>

<style>
.create-event-dialog.el-dialog {
  border-radius: 16px;
  overflow: hidden;
}

.create-event-dialog .el-dialog__header {
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
  margin: 0;
}

.create-event-dialog .el-dialog__title {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
}

.create-event-dialog .el-dialog__body {
  padding: 20px;
}

.create-event-dialog .el-dialog__footer {
  padding: 12px 20px;
  border-top: 1px solid #f1f5f9;
}

.create-event-dialog .el-form-item__label {
  font-size: 13px;
  color: #475569;
  font-weight: 500;
}

.create-event-dialog .el-input__inner,
.create-event-dialog .el-textarea__inner {
  border-radius: 8px;
  border-color: #e2e8f0;
  transition: all 0.2s ease;
}

.create-event-dialog .el-input__inner:focus,
.create-event-dialog .el-textarea__inner:focus {
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.08);
}

.create-event-dialog .el-button--primary {
  border-radius: 8px;
  background: #3b82f6;
  border-color: #3b82f6;
}

.create-event-dialog .el-button--primary:hover {
  background: #2563eb;
  border-color: #2563eb;
}

.create-event-dialog .el-button--danger {
  border-radius: 8px;
}
</style>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
