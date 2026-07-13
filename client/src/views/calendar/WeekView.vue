<template>
  <div class="week-view">
    <div class="week-grid" ref="weekBody">
      <!-- 表头行 -->
      <div class="time-gutter-header"></div>
      <div
        v-for="(day, index) in weekDays"
        :key="'header-' + index"
        class="day-header"
        :class="{ 'is-today': isToday(day) }"
      >
        <div class="day-name">{{ getDayName(index) }}</div>
        <div class="day-number" :class="{ 'today-number': isToday(day) }">{{ day.getDate() }}</div>
      </div>

      <!-- 全天日程行 -->
      <div class="allday-gutter" v-if="hasAnyAllDayEvents"></div>
      <template v-if="hasAnyAllDayEvents">
        <div
          v-for="(day, dayIndex) in weekDays"
          :key="'allday-col-' + dayIndex"
          class="allday-column"
        >
          <div
            v-for="event in allDayEventsByDay[day.toISOString().split('T')[0]] || []"
            :key="'allday-' + event.id"
            class="allday-card"
            :style="{ backgroundColor: getEventColor(event) }"
            @click.stop="$emit('event-click', event)"
            @dblclick.stop="$emit('event-dblclick', event)"
          >
            <div class="event-title">{{ event.title }}</div>
            <div class="event-all-day-tag">全天</div>
          </div>
        </div>
      </template>

      <!-- 时间标签列 -->
      <div class="time-gutter">
        <div v-for="hour in 23" :key="'time-' + hour" class="time-label" :style="{ top: hour * HOUR_HEIGHT + 'px' }">
          {{ String(hour).padStart(2, '0') }}:00
        </div>
      </div>

      <!-- 日程列 -->
      <div
        v-for="(day, dayIndex) in weekDays"
        :key="'column-' + dayIndex"
        class="day-column"
        :class="{ 'is-today': isToday(day) }"
      >
        <div v-for="hour in 24" :key="'slot-' + hour" class="hour-slot" @click="handleSlotClick(day, hour - 1)"></div>

        <div
          v-for="item in layoutEventsByDay[day.toISOString().split('T')[0]] || []"
          :key="item.event.id + '-' + item.groupIndex"
          class="event-card"
          :class="{ 'overlap-active': item.groupSize <= 1 || getActiveIdx(item.groupKey, item.groupSize) === item.groupIndex, 'overlap-hidden': item.groupSize > 1 && getActiveIdx(item.groupKey, item.groupSize) !== item.groupIndex }"
          :style="item.style"
          @click.stop="$emit('event-click', item.event)"
          @dblclick.stop="$emit('event-dblclick', item.event)"
        >
          <div class="overlap-nav" v-if="item.groupSize > 1" @click.stop>
            <button class="overlap-btn" @click="navigateOverlap(item.groupKey, -1)">
              <i class="el-icon-arrow-left"></i>
            </button>
            <span class="overlap-count">{{ getActiveIdx(item.groupKey, item.groupSize) + 1 }}/{{ item.groupSize }}</span>
            <button class="overlap-btn" @click="navigateOverlap(item.groupKey, 1)">
              <i class="el-icon-arrow-right"></i>
            </button>
          </div>
          <div class="event-title">{{ item.event.title }}</div>
          <div class="event-all-day-tag" v-if="item.event.is_all_day">全天</div>
          <div class="event-time" v-if="!item.event.is_all_day">
            {{ formatTime(item.event.start_time) }} - {{ formatTime(item.event.end_time) }}
          </div>
          <div class="event-participants" v-if="item.event.participants && item.event.participants.length > 0">
            <span v-for="(p, i) in item.event.participants.slice(0, 3)" :key="i" class="participant-avatar">
              {{ p.display_name ? p.display_name[0] : '?' }}
            </span>
            <span v-if="item.event.participants.length > 3" class="participant-more">+{{ item.event.participants.length - 3 }}</span>
          </div>
        </div>

        <div class="current-time-line" v-if="isToday(day)" :style="currentTimeStyle"></div>
      </div>
    </div>
  </div>
</template>

<script>
const HOUR_HEIGHT = 56

export default {
  name: 'WeekView',
  props: {
    currentDate: { type: Date, required: true },
    events: { type: Array, default: () => [] }
  },
  data() {
    return {
      currentTime: new Date(),
      timeTimer: null,
      HOUR_HEIGHT,
      overlapActive: {}
    }
  },
  computed: {
    weekDays() {
      const start = this.getWeekStart(this.currentDate)
      const days = []
      for (let i = 0; i < 7; i++) {
        const d = new Date(start)
        d.setDate(d.getDate() + i)
        days.push(d)
      }
      return days
    },
    currentTimeStyle() {
      const now = this.currentTime
      const minutes = now.getHours() * 60 + now.getMinutes()
      const top = (minutes / 60) * HOUR_HEIGHT
      return { top: `${top}px` }
    },
    eventsByDay() {
      const map = {}
      for (const day of this.weekDays) {
        const key = day.toISOString().split('T')[0]
        const dayStart = new Date(day)
        dayStart.setHours(0, 0, 0, 0)
        const dayEnd = new Date(day)
        dayEnd.setHours(23, 59, 59, 999)
        map[key] = this.events.filter(event => {
          const start = new Date(event.start_time)
          const end = new Date(event.end_time)
          return start <= dayEnd && end >= dayStart
        })
      }
      return map
    },
    layoutEventsByDay() {
      const result = {}
      for (const day of this.weekDays) {
        const key = day.toISOString().split('T')[0]
        const events = this.eventsByDay[key] || []
        result[key] = this.computeDayLayout(events, day)
      }
      return result
    },
    allDayEventsByDay() {
      const map = {}
      for (const day of this.weekDays) {
        const key = day.toISOString().split('T')[0]
        const events = this.eventsByDay[key] || []
        map[key] = events.filter(e => e.is_all_day)
      }
      return map
    },
    hasAnyAllDayEvents() {
      return Object.values(this.allDayEventsByDay).some(arr => arr.length > 0)
    }
  },
  mounted() {
    this.timeTimer = setInterval(() => {
      this.currentTime = new Date()
    }, 60000)
    this.scrollToCurrentTime()
  },
  beforeDestroy() {
    if (this.timeTimer) clearInterval(this.timeTimer)
  },
  methods: {
    computeDayLayout(events, day) {
      if (!events.length) return []
      const dayStart = new Date(day)
      dayStart.setHours(0, 0, 0, 0)

      // Separate all-day and timed events
      const allDay = events.filter(e => e.is_all_day)
      const timed = events.filter(e => !e.is_all_day)

      // Process timed events for overlap detection
      const processed = timed.map(event => {
        const start = new Date(event.start_time)
        const end = new Date(event.end_time)
        const startMin = Math.max(0, (start - dayStart) / 60000)
        const endMin = Math.min(1440, (end - dayStart) / 60000)
        return { event, startMin, endMin: Math.max(startMin + 30, endMin) }
      }).sort((a, b) => a.startMin - b.startMin || a.endMin - b.endMin)

      // Find overlap groups
      const groups = []
      const used = new Set()
      for (let i = 0; i < processed.length; i++) {
        if (used.has(i)) continue
        const group = [i]
        used.add(i)
        let groupEnd = processed[i].endMin
        for (let j = i + 1; j < processed.length; j++) {
          if (used.has(j)) continue
          if (processed[j].startMin < groupEnd) {
            group.push(j)
            used.add(j)
            groupEnd = Math.max(groupEnd, processed[j].endMin)
          }
        }
        groups.push(group)
      }

      // Build layout items for timed events - all full width, stacked
      const result = []
      for (const group of groups) {
        const groupKey = day.toISOString().split('T')[0] + '_' + group.map(i => processed[i].event.id).sort().join('_')
        group.forEach((idx, col) => {
          const p = processed[idx]
          const duration = Math.max(30, p.endMin - p.startMin)
          const top = (p.startMin / 60) * HOUR_HEIGHT
          const height = (duration / 60) * HOUR_HEIGHT
          result.push({
            event: p.event,
            groupKey,
            groupSize: group.length,
            groupIndex: col,
            style: {
              top: `${top}px`,
              height: `${height}px`,
              backgroundColor: this.getEventColor(p.event),
              left: '4px',
              right: '4px',
              position: 'absolute'
            }
          })
        })
      }

      // All-day events are now rendered in the separate allday section, not here
      return result
    },
    navigateOverlap(groupKey, dir) {
      // Find group size from any layout item
      for (const dayKey in this.layoutEventsByDay) {
        const items = this.layoutEventsByDay[dayKey]
        const match = items.find(i => i.groupKey === groupKey)
        if (match) {
          const groupSize = match.groupSize
          const current = this.overlapActive[groupKey] || 0
          const next = (current + dir + groupSize) % groupSize
          this.$set(this.overlapActive, groupKey, next)
          return
        }
      }
    },
    getActiveIdx(groupKey, groupSize) {
      const idx = this.overlapActive[groupKey] || 0
      return idx < groupSize ? idx : 0
    },
    getWeekStart(date) {
      const d = new Date(date)
      const day = d.getDay()
      const diff = d.getDate() - day + (day === 0 ? -6 : 1)
      d.setDate(diff)
      d.setHours(0, 0, 0, 0)
      return d
    },
    getDayName(index) {
      const names = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
      return names[index]
    },
    isToday(date) {
      const today = new Date()
      return date.getFullYear() === today.getFullYear() &&
        date.getMonth() === today.getMonth() &&
        date.getDate() === today.getDate()
    },
    getEventColor(event) {
      const colors = ['#93c5fd', '#6ee7b7', '#fcd34d', '#fca5a5', '#c4b5fd', '#f9a8d4']
      return colors[parseInt(event.id || 0) % colors.length]
    },
    formatTime(timeStr) {
      const d = new Date(timeStr)
      return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    handleSlotClick(day, hour) {
      // Could open create dialog with pre-filled time
    },
    scrollToCurrentTime() {
      this.$nextTick(() => {
        if (this.$refs.weekBody) {
          const now = new Date()
          const scrollTo = Math.max(0, (now.getHours() - 2) * HOUR_HEIGHT)
          this.$refs.weekBody.scrollTop = scrollTo
        }
      })
    }
  }
}
</script>

<style scoped>
.week-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.week-allday-section {
  display: contents;
}

.allday-gutter {
  grid-column: 1;
  grid-row: 2;
  border-right: 1px solid #f1f5f9;
  border-bottom: 1px solid #f1f5f9;
  background: #ffffff;
  position: sticky;
  top: 55px;
  z-index: 24;
}

.allday-column {
  grid-row: 2;
  padding: 4px 4px;
  display: flex;
  flex-direction: column;
  gap: 3px;
  border-right: 1px solid #f1f5f9;
  border-bottom: 1px solid #f1f5f9;
  background: #ffffff;
  position: sticky;
  top: 55px;
  z-index: 24;
}

.allday-column:last-child {
  border-right: none;
}

.allday-card {
  position: relative;
  border-radius: 6px;
  padding: 4px 8px;
  color: #1e293b;
  font-size: 12px;
  min-height: 45px;
  cursor: pointer;
  opacity: 0.85;
  box-sizing: border-box;
}

.allday-card:hover {
  opacity: 1;
}

.week-grid {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: 56px repeat(7, 1fr);
  grid-template-rows: 55px 53px repeat(24, 56px);
  position: relative;
}

.time-gutter-header {
  grid-column: 1;
  grid-row: 1;
  border-right: 1px solid #f1f5f9;
  border-bottom: 1px solid #f1f5f9;
  box-sizing: border-box;
  background: #ffffff;
  position: sticky;
  top: 0;
  z-index: 25;
}

.day-header {
  grid-row: 1;
  text-align: center;
  padding: 10px 0;
  border-right: 1px solid #f1f5f9;
  border-bottom: 1px solid #f1f5f9;
  box-sizing: border-box;
  background: #ffffff;
  position: sticky;
  top: 0;
  z-index: 25;
}

.day-header:last-child {
  border-right: none;
}

.day-name {
  font-size: 11px;
  color: #94a3b8;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.day-number {
  font-size: 20px;
  font-weight: 600;
  color: #475569;
  margin-top: 2px;
  letter-spacing: -0.01em;
}

.today-number {
  color: #3b82f6;
}

.is-today.day-header {
  background: #eff6ff;
}

.is-today.day-column {
  background: rgba(239, 246, 255, 0.45);
}

.time-gutter {
  grid-column: 1;
  grid-row: 3 / span 24;
  width: 56px;
  border-right: 1px solid #f1f5f9;
  box-sizing: border-box;
  background: #ffffff;
  z-index: 5;
  position: relative;
}

.time-label {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: #94a3b8;
  font-weight: 500;
  line-height: 1;
  transform: translateY(-50%);
}

.day-column {
  grid-row: 3 / span 24;
  position: relative;
  border-right: 1px solid #f1f5f9;
  box-sizing: border-box;
}

.day-column:last-child {
  border-right: none;
}

.hour-slot {
  height: 56px;
  border-bottom: 1px solid #f1f5f9;
  box-sizing: border-box;
  cursor: pointer;
  transition: background 0.15s ease;
}

.hour-slot:hover {
  background: #fafbfc;
}

.event-card {
  position: absolute;
  left: 4px;
  right: 4px;
  border-radius: 6px;
  padding: 4px 8px;
  color: #1e293b;
  font-size: 12px;
  cursor: pointer;
  overflow: hidden;
  box-sizing: border-box;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  transition: opacity 0.25s ease, transform 0.2s ease, box-shadow 0.2s ease, z-index 0s;
}

.event-card.overlap-active {
  z-index: 15;
  opacity: 1;
}

.event-card.overlap-hidden {
  z-index: 5;
  opacity: 0;
  pointer-events: none;
}

.event-card.overlap-active:hover {
  transform: scale(1.02);
  z-index: 20;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.15);
}

.overlap-nav {
  position: absolute;
  top: 3px;
  right: 3px;
  display: flex;
  align-items: center;
  gap: 1px;
  z-index: 5;
}

.overlap-btn {
  width: 16px;
  height: 16px;
  border: none;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.7);
  color: #475569;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  padding: 0;
  transition: all 0.15s ease;
}

.overlap-btn:hover {
  background: rgba(255, 255, 255, 0.95);
  color: #1e293b;
}

.overlap-count {
  font-size: 8px;
  font-weight: 600;
  color: #475569;
  min-width: 16px;
  text-align: center;
}

.event-title {
  font-weight: 600;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.3;
}

.event-all-day-tag {
  font-size: 10px;
  color: #475569;
  margin-top: 1px;
}

.event-time {
  font-size: 10px;
  color: #475569;
  margin-top: 1px;
}

.event-participants {
  display: flex;
  gap: 3px;
  margin-top: 3px;
}

.participant-avatar {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 600;
  color: #475569;
}

.participant-more {
  font-size: 10px;
  color: #64748b;
  line-height: 18px;
}

.current-time-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 2px;
  background: #ef4444;
  z-index: 15;
  pointer-events: none;
}

.current-time-line::before {
  content: '';
  position: absolute;
  left: -5px;
  top: -4px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #ef4444;
}

/* 滚动条美化 */
.week-grid::-webkit-scrollbar {
  width: 6px;
}

.week-grid::-webkit-scrollbar-track {
  background: transparent;
}

.week-grid::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 3px;
}

.week-grid::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
</style>
