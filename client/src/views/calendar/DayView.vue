<template>
  <div class="day-view">
    <div class="day-header">
      <div class="date-info" :class="{ 'is-today': isToday }">
        <span class="day-number">{{ currentDate.getDate() }}</span>
        <span class="day-name">{{ dayName }}</span>
      </div>
    </div>

    <div class="all-day-section" v-if="allDayEvents.length > 0">
      <div
        v-for="event in allDayEvents"
        :key="'allday-' + event.id"
        class="all-day-event"
        :style="{ backgroundColor: getEventColor(event) }"
        @click.stop="$emit('event-click', event)"
        @dblclick.stop="$emit('event-dblclick', event)"
      >
        <div class="event-title">{{ event.title }}</div>
        <div class="event-all-day-tag">全天</div>
        <div class="event-description" v-if="event.description">{{ event.description }}</div>
        <div class="event-participants" v-if="event.participants && event.participants.length > 0">
          <span v-for="(p, i) in event.participants.slice(0, 5)" :key="i" class="participant-avatar">
            {{ p.display_name ? p.display_name[0] : '?' }}
          </span>
        </div>
      </div>
    </div>

    <div class="day-body" ref="dayBody">
      <div class="time-gutter">
        <div v-for="hour in 23" :key="hour" class="time-label" :style="{ top: hour * HOUR_HEIGHT + 'px' }">
          {{ String(hour).padStart(2, '0') }}:00
        </div>
      </div>

      <div class="day-content">
        <div v-for="hour in 24" :key="hour" class="hour-slot"></div>

        <div
          v-for="item in timedLayoutEvents"
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
          <div class="event-time" v-if="!item.event.is_all_day">
            {{ formatTime(item.event.start_time) }} - {{ formatTime(item.event.end_time) }}
          </div>
          <div class="event-description" v-if="item.event.description">{{ item.event.description }}</div>
          <div class="event-participants" v-if="item.event.participants && item.event.participants.length > 0">
            <span v-for="(p, i) in item.event.participants.slice(0, 5)" :key="i" class="participant-avatar">
              {{ p.display_name ? p.display_name[0] : '?' }}
            </span>
          </div>
        </div>

        <div class="current-time-line" v-if="isToday" :style="currentTimeStyle"></div>
      </div>
    </div>
  </div>
</template>

<script>
const HOUR_HEIGHT = 56

export default {
  name: 'DayView',
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
    dayName() {
      const names = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
      return names[this.currentDate.getDay()]
    },
    isToday() {
      const today = new Date()
      return this.currentDate.getFullYear() === today.getFullYear() &&
        this.currentDate.getMonth() === today.getMonth() &&
        this.currentDate.getDate() === today.getDate()
    },
    dayEvents() {
      return this.events.filter(event => {
        const start = new Date(event.start_time)
        const end = new Date(event.end_time)
        const dayStart = new Date(this.currentDate)
        dayStart.setHours(0, 0, 0, 0)
        const dayEnd = new Date(this.currentDate)
        dayEnd.setHours(23, 59, 59, 999)
        return start <= dayEnd && end >= dayStart
      })
    },
    allDayEvents() {
      return this.dayEvents.filter(e => e.is_all_day)
    },
    timedEvents() {
      return this.dayEvents.filter(e => !e.is_all_day)
    },
    timedLayoutEvents() {
      const events = this.timedEvents
      if (!events.length) return []
      const dayStart = new Date(this.currentDate)
      dayStart.setHours(0, 0, 0, 0)

      const processed = events.map(event => {
        const start = new Date(event.start_time)
        const end = new Date(event.end_time)
        const startMin = Math.max(0, (start - dayStart) / 60000)
        const endMin = Math.min(1440, (end - dayStart) / 60000)
        return { event, startMin, endMin: Math.max(startMin + 30, endMin) }
      }).sort((a, b) => a.startMin - b.startMin || a.endMin - b.endMin)

      // Find overlap groups using sweep-line
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

      // Build layout items - all full width, stacked
      const result = []
      for (const group of groups) {
        const groupKey = group.map(i => processed[i].event.id).sort().join('_')
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
              left: '6px',
              right: '6px',
              position: 'absolute'
            }
          })
        })
      }
      return result
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
    getEventColor(event) {
      const colors = ['#93c5fd', '#6ee7b7', '#fcd34d', '#fca5a5', '#c4b5fd', '#f9a8d4']
      return colors[parseInt(event.id || 0) % colors.length]
    },
    formatTime(timeStr) {
      const d = new Date(timeStr)
      return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    navigateOverlap(groupKey, dir) {
      const groupSize = this.timedLayoutEvents.find(i => i.groupKey === groupKey)?.groupSize || 1
      const current = this.overlapActive[groupKey] || 0
      const next = (current + dir + groupSize) % groupSize
      this.$set(this.overlapActive, groupKey, next)
    },
    getActiveIdx(groupKey, groupSize) {
      const idx = this.overlapActive[groupKey] || 0
      return idx < groupSize ? idx : 0
    },
    scrollToCurrentTime() {
      this.$nextTick(() => {
        if (this.$refs.dayBody) {
          const now = new Date()
          const scrollTo = Math.max(0, (now.getHours() - 2) * HOUR_HEIGHT)
          this.$refs.dayBody.scrollTop = scrollTo
        }
      })
    }
  }
}
</script>

<style scoped>
.day-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.day-header {
  padding: 12px 20px;
  border-bottom: 1px solid #f1f5f9;
  flex-shrink: 0;
}

.date-info {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.date-info.is-today .day-number {
  color: #3b82f6;
}

.day-number {
  font-size: 26px;
  font-weight: 700;
  color: #1e293b;
  letter-spacing: -0.02em;
}

.day-name {
  font-size: 14px;
  color: #94a3b8;
  font-weight: 500;
}

.all-day-section {
  flex-shrink: 0;
  padding: 6px 12px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: #ffffff;
}

.all-day-event {
  position: relative;
  border-radius: 6px;
  padding: 6px 10px;
  color: #1e293b;
  font-size: 13px;
  cursor: pointer;
  opacity: 0.85;
  box-sizing: border-box;
}

.day-body {
  flex: 1;
  overflow-y: auto;
  display: flex;
  position: relative;
}

.time-gutter {
  width: 56px;
  flex-shrink: 0;
  border-right: 1px solid #f1f5f9;
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

.day-content {
  flex: 1;
  position: relative;
}

.hour-slot {
  height: 56px;
  border-bottom: 1px solid #f1f5f9;
  box-sizing: border-box;
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.hour-slot:hover {
  background: #fafbfc;
}

.event-card {
  position: absolute;
  border-radius: 8px;
  padding: 6px 10px;
  color: #1e293b;
  font-size: 13px;
  cursor: pointer;
  overflow: hidden;
  box-sizing: border-box;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
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
  transform: scale(1.01);
  z-index: 20;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.18);
}

.overlap-nav {
  position: absolute;
  top: 4px;
  right: 4px;
  display: flex;
  align-items: center;
  gap: 2px;
  z-index: 5;
}

.overlap-btn {
  width: 18px;
  height: 18px;
  border: none;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.7);
  color: #475569;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  padding: 0;
  transition: all 0.15s ease;
}

.overlap-btn:hover {
  background: rgba(255, 255, 255, 0.95);
  color: #1e293b;
}

.overlap-count {
  font-size: 9px;
  font-weight: 600;
  color: #475569;
  min-width: 20px;
  text-align: center;
}

.event-title {
  font-weight: 600;
  font-size: 13px;
  line-height: 1.3;
}

.event-all-day-tag {
  font-size: 11px;
  color: #475569;
  margin-top: 2px;
}

.event-time {
  font-size: 11px;
  color: #475569;
  margin-top: 2px;
}

.event-description {
  font-size: 12px;
  color: #64748b;
  margin-top: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.event-participants {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}

.participant-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 600;
  color: #475569;
}

.current-time-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 2px;
  background: #ef4444;
  z-index: 30;
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
.day-body::-webkit-scrollbar {
  width: 6px;
}

.day-body::-webkit-scrollbar-track {
  background: transparent;
}

.day-body::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 3px;
}

.day-body::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
</style>
