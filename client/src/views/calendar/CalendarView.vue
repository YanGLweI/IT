<template>
  <div class="calendar-page">
    <div class="calendar-card">
      <!-- 工具栏 -->
      <div class="calendar-toolbar">
        <div class="toolbar-left">
          <button class="nav-btn today-btn" @click="goToday">今天</button>
          <div class="nav-arrows">
            <button class="nav-btn arrow-btn" @click="navigate(-1)">
              <i class="el-icon-arrow-left"></i>
            </button>
            <button class="nav-btn arrow-btn" @click="navigate(1)">
              <i class="el-icon-arrow-right"></i>
            </button>
          </div>
          <span class="current-date-label">{{ dateLabel }}</span>
          <div class="view-switch">
            <button
              v-for="v in views"
              :key="v.key"
              class="view-btn"
              :class="{ active: currentView === v.key }"
              @click="currentView = v.key"
            >{{ v.label }}</button>
          </div>
        </div>
        <div class="toolbar-right">
          <div class="search-box">
            <i class="el-icon-search search-icon"></i>
            <input
              v-model="searchKeyword"
              class="search-input"
              placeholder="搜索日程..."
              @input="handleSearch"
            />
            <i v-if="searchKeyword" class="el-icon-close clear-icon" @click="searchKeyword = ''"></i>
          </div>
          <button class="nav-btn create-btn" @click="openCreateDialog">
            <i class="el-icon-plus"></i>
            <span>新建日程</span>
          </button>
        </div>
      </div>

      <!-- 日历内容 -->
      <div class="calendar-content">
        <week-view
          v-if="currentView === 'week'"
          :current-date="currentDate"
          :events="filteredEvents"
          @event-click="handleEventClick"
          @event-dblclick="handleEventDblClick"
        />
        <day-view
          v-if="currentView === 'day'"
          :current-date="currentDate"
          :events="filteredEvents"
          @event-click="handleEventClick"
          @event-dblclick="handleEventDblClick"
        />
        <month-view
          v-if="currentView === 'month'"
          :current-date="currentDate"
          :events="filteredEvents"
          @event-click="handleEventClick"
          @event-dblclick="handleEventDblClick"
        />
      </div>
    </div>

    <create-event-dialog
      :visible="createDialogVisible"
      :mode="createDialogMode"
      :event="editingEvent"
      @close="createDialogVisible = false"
      @saved="handleEventSaved"
    />

    <event-detail-dialog
      :visible="detailDialogVisible"
      :event="selectedEvent"
      @close="detailDialogVisible = false"
      @edit="handleEditFromDetail"
      @deleted="handleEventDeleted"
    />
  </div>
</template>

<script>
import { getCalendars } from '@/api/calendar'
import WeekView from './WeekView.vue'
import DayView from './DayView.vue'
import MonthView from './MonthView.vue'
import CreateEventDialog from './CreateEventDialog.vue'
import EventDetailDialog from './EventDetailDialog.vue'

export default {
  name: 'CalendarView',
  components: { WeekView, DayView, MonthView, CreateEventDialog, EventDetailDialog },
  data() {
    return {
      currentView: 'week',
      currentDate: new Date(),
      events: [],
      searchKeyword: '',
      createDialogVisible: false,
      createDialogMode: 'create',
      editingEvent: null,
      detailDialogVisible: false,
      selectedEvent: null,
      searchTimer: null,
      views: [
        { key: 'day', label: '日' },
        { key: 'week', label: '周' },
        { key: 'month', label: '月' }
      ]
    }
  },
  computed: {
    dateLabel() {
      const d = this.currentDate
      const year = d.getFullYear()
      const month = d.getMonth() + 1
      if (this.currentView === 'day') {
        return `${year}年${month}月${d.getDate()}日`
      } else if (this.currentView === 'week') {
        const weekStart = this.getWeekStart(d)
        const weekEnd = new Date(weekStart)
        weekEnd.setDate(weekEnd.getDate() + 6)
        const sm = weekStart.getMonth() + 1
        const em = weekEnd.getMonth() + 1
        if (sm === em) {
          return `${year}年${sm}月${weekStart.getDate()}日 - ${weekEnd.getDate()}日`
        }
        return `${year}年${sm}月${weekStart.getDate()}日 - ${em}月${weekEnd.getDate()}日`
      } else {
        return `${year}年${month}月`
      }
    },
    filteredEvents() {
      if (!this.searchKeyword) return this.events
      const kw = this.searchKeyword.toLowerCase()
      return this.events.filter(e =>
        e.title.toLowerCase().includes(kw) ||
        (e.participants && e.participants.some(p => p.display_name && p.display_name.toLowerCase().includes(kw)))
      )
    }
  },
  watch: {
    currentDate() {
      this.fetchEvents()
    },
    currentView() {
      this.fetchEvents()
    }
  },
  created() {
    this.fetchEvents()
  },
  methods: {
    getWeekStart(date) {
      const d = new Date(date)
      const day = d.getDay()
      const diff = d.getDate() - day + (day === 0 ? -6 : 1)
      d.setDate(diff)
      d.setHours(0, 0, 0, 0)
      return d
    },
    getDateRange() {
      if (this.currentView === 'day') {
        const start = new Date(this.currentDate)
        start.setHours(0, 0, 0, 0)
        const end = new Date(start)
        end.setDate(end.getDate() + 1)
        return { start, end }
      } else if (this.currentView === 'week') {
        const start = this.getWeekStart(this.currentDate)
        const end = new Date(start)
        end.setDate(end.getDate() + 7)
        return { start, end }
      } else {
        const start = new Date(this.currentDate.getFullYear(), this.currentDate.getMonth(), 1)
        const day = start.getDay()
        start.setDate(start.getDate() - (day === 0 ? 6 : day - 1))
        const end = new Date(start)
        end.setDate(end.getDate() + 42)
        return { start, end }
      }
    },
    async fetchEvents() {
      try {
        const { start, end } = this.getDateRange()
        const params = {
          start_date: this.formatDate(start),
          end_date: this.formatDate(end)
        }
        const res = await getCalendars(params)
        if (res && res.code === 200) {
          const rawEvents = res.data || []
          this.events = this.expandRecurringEvents(rawEvents, start, end)
        }
      } catch (err) {
        console.error('获取日程失败:', err)
      }
    },
    expandRecurringEvents(events, rangeStart, rangeEnd) {
      const result = []
      for (const event of events) {
        const rule = this.parseRepeatRule(event.repeat_rule_json)
        if (!rule) {
          result.push(event)
          continue
        }
        const origStart = new Date(event.start_time)
        const origEnd = new Date(event.end_time)
        const duration = origEnd.getTime() - origStart.getTime()
        const instances = this.generateInstances(rule, origStart, rangeStart, rangeEnd)
        for (const inst of instances) {
          result.push({
            ...event,
            start_time: inst.toISOString(),
            end_time: new Date(inst.getTime() + duration).toISOString()
          })
        }
      }
      return result
    },
    parseRepeatRule(json) {
      if (!json || json === '' || json === 'null') return null
      try {
        const rule = JSON.parse(json)
        if (!rule.type || rule.type === 'none') return null
        return rule
      } catch { return null }
    },
    generateInstances(rule, startDate, rangeStart, rangeEnd) {
      const instances = []
      const interval = Math.max(rule.interval || 1, 1)
      const maxOcc = rule.occurrences || 1000
      let count = 0
      const endDate = rule.endDate ? new Date(rule.endDate + 'T00:00:00') : null

      switch (rule.type) {
        case 'daily': {
          let cur = new Date(startDate)
          while (cur <= rangeEnd && count < maxOcc) {
            if (cur >= rangeStart) instances.push(new Date(cur))
            count++
            if (endDate && cur > endDate) break
            cur = new Date(cur.getTime() + interval * 86400000)
          }
          break
        }
        case 'weekly': {
          let cur = new Date(startDate)
          while (cur <= rangeEnd && count < maxOcc) {
            if (cur >= rangeStart && cur.getDay() === rule.weekday) instances.push(new Date(cur))
            count++
            if (endDate && cur > endDate) break
            cur = new Date(cur.getTime() + interval * 7 * 86400000)
          }
          break
        }
        case 'monthly_day': {
          let year = startDate.getFullYear()
          let month = startDate.getMonth() + 1
          while (count < maxOcc) {
            const candidate = new Date(year, month - 1, rule.monthDay, startDate.getHours(), startDate.getMinutes(), startDate.getSeconds())
            if (candidate > rangeEnd) break
            if (candidate >= rangeStart && candidate >= startDate && candidate.getDate() === rule.monthDay) {
              instances.push(candidate)
            }
            count++
            if (endDate && candidate > endDate) break
            month += interval
            while (month > 12) { month -= 12; year++ }
          }
          break
        }
        case 'monthly_week': {
          let year = startDate.getFullYear()
          let month = startDate.getMonth() + 1
          while (count < maxOcc) {
            const candidate = this.getNthWeekdayOfMonth(year, month, rule.weekOfMonth, rule.weekday, startDate)
            // 该月不存在第N个周X，跳过
            if (candidate) {
              if (candidate > rangeEnd) break
              if (candidate >= rangeStart && candidate >= startDate) instances.push(candidate)
            }
            count++
            if (endDate && candidate && candidate > endDate) break
            month += interval
            while (month > 12) { month -= 12; year++ }
          }
          break
        }
        case 'yearly': {
          let year = startDate.getFullYear()
          while (count < maxOcc) {
            const candidate = new Date(year, rule.monthOfYear - 1, rule.monthDay, startDate.getHours(), startDate.getMinutes(), startDate.getSeconds())
            if (candidate > rangeEnd) break
            if (candidate >= rangeStart && candidate >= startDate && candidate.getMonth() === rule.monthOfYear - 1 && candidate.getDate() === rule.monthDay) {
              instances.push(candidate)
            }
            count++
            if (endDate && candidate > endDate) break
            year += interval
          }
          break
        }
        case 'workday': {
          let cur = new Date(startDate)
          while (cur <= rangeEnd && count < maxOcc) {
            const dow = cur.getDay()
            if (cur >= rangeStart && dow !== 0 && dow !== 6) instances.push(new Date(cur))
            count++
            if (endDate && cur > endDate) break
            cur = new Date(cur.getTime() + 86400000)
          }
          break
        }
        case 'custom': {
          if (rule.unit === 'weeks' && rule.weekdays && rule.weekdays.length > 0) {
            // 周频率+多选周几：逐日遍历，对齐到startDate所在周的周一
            const weekdaySet = new Set(rule.weekdays)
            const sdWeekday = startDate.getDay() // 0=周日
            const mondayOffset = sdWeekday === 0 ? 6 : sdWeekday - 1
            let cur = new Date(startDate)
            cur.setDate(cur.getDate() - mondayOffset)
            let weekCounter = 0
            while (cur <= rangeEnd && count < maxOcc) {
              if (weekCounter % interval === 0 && weekdaySet.has(cur.getDay())) {
                const inst = new Date(cur)
                inst.setHours(startDate.getHours(), startDate.getMinutes(), startDate.getSeconds())
                // 使用本地时间的日期字符串进行比较，避免时区问题
                const instDateStr = inst.getFullYear() + '-' + String(inst.getMonth() + 1).padStart(2, '0') + '-' + String(inst.getDate()).padStart(2, '0')
                const startDateStr = startDate.getFullYear() + '-' + String(startDate.getMonth() + 1).padStart(2, '0') + '-' + String(startDate.getDate()).padStart(2, '0')
                const rangeStartStr = rangeStart.getFullYear() + '-' + String(rangeStart.getMonth() + 1).padStart(2, '0') + '-' + String(rangeStart.getDate()).padStart(2, '0')
                
                if (instDateStr >= startDateStr && instDateStr >= rangeStartStr) {
                  instances.push(new Date(inst))
                  count++
                }
              }
              if (endDate && cur > endDate) break
              const next = new Date(cur)
              next.setDate(next.getDate() + 1)
              // next是周一时递增周计数器（说明本周结束）
              if (next.getDay() === 1) {
                weekCounter++
              }
              cur = next
            }
          } else {
            let cur = new Date(startDate)
            while (cur <= rangeEnd && count < maxOcc) {
              if (cur >= rangeStart) instances.push(new Date(cur))
              count++
              if (endDate && cur > endDate) break
              switch (rule.unit) {
                case 'days': cur = new Date(cur.getTime() + interval * 86400000); break
                case 'months': {
                  const d = new Date(cur)
                  const origDay = startDate.getDate()
                  d.setDate(1)
                  d.setMonth(d.getMonth() + interval)
                  d.setDate(Math.min(origDay, new Date(d.getFullYear(), d.getMonth() + 1, 0).getDate()))
                  cur = d
                  break
                }
                case 'years': { const d = new Date(cur); d.setFullYear(d.getFullYear() + interval); cur = d; break }
                default: cur = new Date(cur.getTime() + interval * 86400000)
              }
            }
          }
          break
        }
      }
      return instances
    },
    getNthWeekdayOfMonth(year, month, weekOfMonth, weekday, baseDate) {
      const firstDay = new Date(year, month - 1, 1, baseDate.getHours(), baseDate.getMinutes(), baseDate.getSeconds())
      const firstWeekday = firstDay.getDay()
      let diff = weekday - firstWeekday
      if (diff < 0) diff += 7
      const firstTarget = new Date(firstDay.getTime() + diff * 86400000)
      const target = new Date(firstTarget.getTime() + (weekOfMonth - 1) * 7 * 86400000)
      // 第N个周X超出本月范围，返回null表示该月不存在
      if (target.getMonth() !== month - 1) {
        return null
      }
      return target
    },
    formatDate(date) {
      const y = date.getFullYear()
      const m = String(date.getMonth() + 1).padStart(2, '0')
      const d = String(date.getDate()).padStart(2, '0')
      return `${y}-${m}-${d}`
    },
    goToday() {
      this.currentDate = new Date()
    },
    navigate(direction) {
      const d = new Date(this.currentDate)
      if (this.currentView === 'day') {
        d.setDate(d.getDate() + direction)
      } else if (this.currentView === 'week') {
        d.setDate(d.getDate() + direction * 7)
      } else {
        d.setMonth(d.getMonth() + direction)
      }
      this.currentDate = d
    },
    handleSearch() {
      clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(() => {}, 300)
    },
    openCreateDialog() {
      this.createDialogMode = 'create'
      this.editingEvent = null
      this.createDialogVisible = true
    },
    handleEventClick(event) {
      this.selectedEvent = event
      this.detailDialogVisible = true
    },
    handleEventDblClick(event) {
      // 重复日程编辑时提示用户将修改整个系列
      const rule = this.parseRepeatRule(event.repeat_rule_json)
      if (rule) {
        this.$confirm('编辑重复日程将修改整个系列，是否继续？', '提示', {
          confirmButtonText: '继续编辑',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          this.editingEvent = event
          this.createDialogMode = 'edit'
          this.createDialogVisible = true
        }).catch(() => {})
      } else {
        this.editingEvent = event
        this.createDialogMode = 'edit'
        this.createDialogVisible = true
      }
    },
    handleEditFromDetail(event) {
      const rule = this.parseRepeatRule(event.repeat_rule_json)
      if (rule) {
        this.$confirm('编辑重复日程将修改整个系列，是否继续？', '提示', {
          confirmButtonText: '继续编辑',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          this.detailDialogVisible = false
          this.editingEvent = event
          this.createDialogMode = 'edit'
          this.createDialogVisible = true
        }).catch(() => {})
      } else {
        this.detailDialogVisible = false
        this.editingEvent = event
        this.createDialogMode = 'edit'
        this.createDialogVisible = true
      }
    },
    handleEventSaved() {
      this.createDialogVisible = false
      this.fetchEvents()
    },
    handleEventDeleted() {
      this.detailDialogVisible = false
      this.fetchEvents()
    }
  }
}
</script>

<style scoped>
.calendar-page {
  padding: 20px;
  height: 100%;
  box-sizing: border-box;
}

.calendar-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 4px 16px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

/* 工具栏 */
.calendar-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid #f1f5f9;
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 按钮 */
.nav-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  padding: 0 14px;
  border: none;
  border-radius: 10px;
  background: #f1f5f9;
  color: #475569;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  gap: 4px;
}

.nav-btn:hover {
  background: #e2e8f0;
  color: #334155;
}

.arrow-btn {
  width: 34px;
  padding: 0;
  font-size: 14px;
}

.today-btn {
  background: #eff6ff;
  color: #3b82f6;
  font-weight: 600;
}

.today-btn:hover {
  background: #dbeafe;
}

.create-btn {
  background: #3b82f6;
  color: #ffffff;
}

.create-btn:hover {
  background: #2563eb;
}

.nav-arrows {
  display: flex;
  gap: 4px;
}

/* 日期标签 */
.current-date-label {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  letter-spacing: -0.01em;
  min-width: 180px;
}

/* 视图切换 */
.view-switch {
  display: flex;
  background: #f1f5f9;
  border-radius: 10px;
  padding: 3px;
  margin-left: 6px;
}

.view-btn {
  height: 28px;
  padding: 0 14px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.view-btn:hover {
  color: #334155;
}

.view-btn.active {
  background: #ffffff;
  color: #1e293b;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* 搜索框 */
.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 10px;
  color: #94a3b8;
  font-size: 13px;
  pointer-events: none;
}

.search-input {
  height: 34px;
  width: 180px;
  padding: 0 30px 0 30px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
  color: #334155;
  font-size: 13px;
  outline: none;
  transition: all 0.2s ease;
}

.search-input::placeholder {
  color: #94a3b8;
}

.search-input:focus {
  border-color: #93c5fd;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.08);
}

.clear-icon {
  position: absolute;
  right: 8px;
  color: #94a3b8;
  font-size: 12px;
  cursor: pointer;
}

.clear-icon:hover {
  color: #64748b;
}

/* 内容区 */
.calendar-content {
  flex: 1;
  overflow: hidden;
  background: #fafbfc;
}
</style>
