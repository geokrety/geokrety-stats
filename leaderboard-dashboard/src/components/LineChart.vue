<script setup>
import { ref, onMounted, watch, computed } from 'vue'

const props = defineProps({
  data:      { type: Array,  required: true },
  xKey:      { type: String, default: 'x' },
  yKey:      { type: String, default: 'y' },
  color:     { type: String, default: '#0d6efd' },
  height:    { type: Number, default: 200 },
  startDate: { type: String, default: null },  // ISO date string, e.g. '2020-03-15'
  endDate:   { type: String, default: null },  // ISO date string, defaults to today
  showRangeButtons: { type: Boolean, default: false },
})

const container = ref(null)
const activeRange = ref('all')

const RANGES = [
  { label: 'All',  key: 'all',  months: null },
  { label: '5Y',   key: '5y',   months: 60 },
  { label: '1Y',   key: '1y',   months: 12 },
  { label: '6M',   key: '6m',   months: 6 },
  { label: '3M',   key: '3m',   months: 3 },
  { label: '1M',   key: '1m',   months: 1 },
]

function parseDate(d) {
  if (!d) return null
  const p = window.d3?.timeParse('%Y-%m-%d')
  return p ? (p(d) || new Date(d)) : new Date(d)
}

function draw() {
  if (!container.value || !window.d3 || !props.data.length) return
  const d3 = window.d3

  const el = container.value
  el.innerHTML = ''

  const margin = { top: 10, right: 20, bottom: 30, left: 52 }
  const totalWidth = el.clientWidth || 600
  const width  = totalWidth - margin.left - margin.right
  const height = props.height - margin.top - margin.bottom

  // Determine the x domain
  const today = new Date()
  const dataEnd = props.endDate ? parseDate(props.endDate) : today
  let domainStart, domainEnd

  if (activeRange.value !== 'all') {
    const rangeObj = RANGES.find(r => r.key === activeRange.value)
    const cutoff = new Date(dataEnd)
    cutoff.setMonth(cutoff.getMonth() - rangeObj.months)
    domainStart = cutoff
    domainEnd = dataEnd
  } else {
    domainStart = props.startDate ? parseDate(props.startDate) : d3.min(props.data, d => parseDate(d[props.xKey]))
    domainEnd = dataEnd
  }

  // Filter data to visible domain
  const visibleData = props.data.filter(d => {
    const t = parseDate(d[props.xKey])
    return t && t >= domainStart && t <= domainEnd
  })

  if (!visibleData.length && activeRange.value !== 'all') {
    // fallback: show all data if filter too narrow
  }

  const displayData = visibleData.length ? visibleData : props.data

  const svg = d3.select(el)
    .append('svg')
    .attr('width', '100%')
    .attr('height', props.height)
    .append('g')
    .attr('transform', `translate(${margin.left},${margin.top})`)

  const x = d3.scaleTime()
    .domain([domainStart, domainEnd])
    .range([0, width])

  const y = d3.scaleLinear()
    .domain([0, d3.max(displayData, d => +d[props.yKey]) * 1.1 || 1])
    .nice()
    .range([height, 0])

  // Grid lines
  svg.append('g')
    .attr('class', 'grid')
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(''))
    .select('.domain').remove()
  svg.selectAll('.grid line').attr('stroke', '#e9ecef').attr('stroke-dasharray', '3,3')

  // Area
  const area = d3.area()
    .defined(d => !isNaN(+d[props.yKey]))
    .x(d => x(parseDate(d[props.xKey])))
    .y0(height)
    .y1(d => y(+d[props.yKey]))
    .curve(d3.curveMonotoneX)

  svg.append('path')
    .datum(displayData)
    .attr('fill', props.color)
    .attr('fill-opacity', 0.15)
    .attr('d', area)

  // Line
  const line = d3.line()
    .defined(d => !isNaN(+d[props.yKey]))
    .x(d => x(parseDate(d[props.xKey])))
    .y(d => y(+d[props.yKey]))
    .curve(d3.curveMonotoneX)

  svg.append('path')
    .datum(displayData)
    .attr('fill', 'none')
    .attr('stroke', props.color)
    .attr('stroke-width', 2)
    .attr('d', line)

  // Smart x tick format based on range
  const rangeMs = domainEnd - domainStart
  const dayMs = 86400000
  let xFormat, xTicks
  if (rangeMs < 40 * dayMs) {
    xFormat = d3.timeFormat('%b %d'); xTicks = 7
  } else if (rangeMs < 200 * dayMs) {
    xFormat = d3.timeFormat('%b %d'); xTicks = 6
  } else if (rangeMs < 800 * dayMs) {
    xFormat = d3.timeFormat('%b %Y'); xTicks = 8
  } else {
    xFormat = d3.timeFormat('%Y'); xTicks = 8
  }

  svg.append('g')
    .attr('transform', `translate(0,${height})`)
    .call(d3.axisBottom(x).ticks(xTicks).tickFormat(xFormat))
    .selectAll('text').attr('font-size', '11px')

  svg.append('g')
    .call(d3.axisLeft(y).ticks(5).tickFormat(d3.format('~s')))
    .selectAll('text').attr('font-size', '11px')

  // Tooltip on hover
  const tooltip = d3.select(el).append('div')
    .attr('class', 'position-absolute p-2 bg-dark text-white rounded small')
    .style('pointer-events', 'none')
    .style('display', 'none')
    .attr('role', 'tooltip')

  const mouseMoveHandler = (event) => {
    const [mouseX] = d3.pointer(event, svg.node())
    const date = x.invert(mouseX)
    
    // Find the nearest data point
    let nearest = null
    let minDist = Infinity
    displayData.forEach(d => {
      const dDate = parseDate(d[props.xKey])
      const dist = Math.abs(dDate - date)
      if (dist < minDist) {
        minDist = dist
        nearest = d
      }
    })

    if (nearest) {
      const xPos = x(parseDate(nearest[props.xKey]))
      const yPos = y(+nearest[props.yKey])
      const dateStr = new Date(parseDate(nearest[props.xKey])).toLocaleDateString()
      const value = +nearest[props.yKey]

      tooltip
        .style('display', 'block')
        .style('left', (margin.left + xPos - 40) + 'px')
        .style('top', (margin.top + yPos - 30) + 'px')
        .html(`<strong>${dateStr}</strong><br/>${value.toLocaleString()}`)
    }
  }

  const mouseOutHandler = () => {
    tooltip.style('display', 'none')
  }

  svg.append('rect')
    .attr('width', width)
    .attr('height', height)
    .attr('fill', 'none')
    .attr('pointer-events', 'all')
    .on('mousemove', mouseMoveHandler)
    .on('mouseout', mouseOutHandler)
}

onMounted(() => { setTimeout(draw, 50) })
watch(() => props.data, draw, { deep: true })
watch(() => props.startDate, draw)
watch(activeRange, draw)
</script>

<template>
  <div>
    <!-- Range buttons -->
    <div v-if="showRangeButtons" class="d-flex justify-content-end mb-1 gap-1">
      <button
        v-for="r in RANGES" :key="r.key"
        class="btn btn-xs py-0 px-2"
        style="font-size:0.75rem"
        :class="activeRange === r.key ? 'btn-primary' : 'btn-outline-secondary'"
        @click="activeRange = r.key"
      >{{ r.label }}</button>
    </div>
    <div ref="container" style="width:100%;" :style="{ height: height + 'px' }"></div>
  </div>
</template>
