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
          v-for="event in getEventsForDay(day)"
          :key="event.id"
          class="event-card"
          :style="getEventStyle(event, day)"
          @click.stop="$emit('event-click', event)"
          @dblclick.stop="$emit('event-dblclick', event)"
        >
          <div class="event-title">{{ event.title }}</div>
          <div class="event-time" v-if="!event.is_all_day">
            {{ formatTime(event.start_time) }} - {{ formatTime(event.end_time) }}
          </div>
          <div class="event-participants" v-if="event.participants && event.participants.length > 0">
            <span v-for="(p, i) in event.participants.slice(0, 3)" :key="i" class="participant-avatar">
              {{ p.display_name ? p.display_name[0] : '?' }}
            </span>
            <span v-if="event.participants.length > 3" class="participant-more">+{{ event.participants.length - 3 }}</span>
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
      HOUR_HEIGHT
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
    getEventsForDay(day) {
      return this.events.filter(event => {
        const start = new Date(event.start_time)
        const end = new Date(event.end_time)
        const dayStart = new Date(day)
        dayStart.setHours(0, 0, 0, 0)
        const dayEnd = new Date(day)
        dayEnd.setHours(23, 59, 59, 999)
        return start <= dayEnd && end >= dayStart
      })
    },
    getEventStyle(event, day) {
      if (event.is_all_day) {
        return {
          top: '0px',
          height: `${24 * HOUR_HEIGHT}px`,
          backgroundColor: this.getEventColor(event),
          left: '4px',
          right: '4px',
          opacity: 1
        }
      }
      const start = new Date(event.start_time)
      const end = new Date(event.end_time)
      const dayStart = new Date(day)
      dayStart.setHours(0, 0, 0, 0)

      const startMinutes = Math.max(0, (start - dayStart) / 60000)
      const endMinutes = Math.min(1440, (end - dayStart) / 60000)
      const duration = Math.max(30, endMinutes - startMinutes)

      const top = (startMinutes / 60) * HOUR_HEIGHT
      const height = (duration / 60) * HOUR_HEIGHT

      return {
        top: `${top}px`,
        height: `${height}px`,
        backgroundColor: this.getEventColor(event),
        left: '4px',
        right: '4px'
      }
    },
    getEventColor(event) {
      const colors = ['#93c5fd', '#6ee7b7', '#fcd34d', '#fca5a5', '#c4b5fd', '#f9a8d4']
      return colors[(event.id || 0) % colors.length]
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

.week-grid {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: 56px repeat(7, 1fr);
  grid-template-rows: auto repeat(24, 56px);
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
  grid-row: 2 / span 24;
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
  grid-row: 2 / span 24;
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
  z-index: 10;
  box-sizing: border-box;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.event-card:hover {
  transform: scale(1.02);
  z-index: 20;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.15);
}

.event-title {
  font-weight: 600;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.3;
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
