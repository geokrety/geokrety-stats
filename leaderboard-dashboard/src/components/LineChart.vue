<script setup>
import { ref, onMounted, watch, computed } from 'vue'

const props = defineProps({
  data:      { type: Array,  required: true },
  xKey:      { type: String, default: 'x' },
  yKey:      { type: String, default: 'y' },
  datasets:  { type: Array,  default: null }, // [{ key, label, color }]
  color:     { type: String, default: '#0d6efd' },
  height:    { type: Number, default: 200 },
  type:      { type: String, default: 'line' }, // 'line' or 'area'
  stacked:   { type: Boolean, default: false },
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
    .attr('viewBox', `0 0 ${totalWidth} ${props.height}`)
    .attr('preserveAspectRatio', 'xMidYMid meet')
    .append('g')
    .attr('transform', `translate(${margin.left},${margin.top})`)

  const x = d3.scaleTime()
    .domain([domainStart, domainEnd])
    .range([0, width])

  const datasets = props.datasets || [{ key: props.yKey, label: '', color: props.color }]

  let yDomainMax = 0
  if (props.stacked) {
    yDomainMax = d3.max(displayData, d => d3.sum(datasets, ds => +d[ds.key] || 0))
  } else {
    yDomainMax = d3.max(displayData, d => d3.max(datasets, ds => +d[ds.key] || 0))
  }

  const y = d3.scaleLinear()
    .domain([0, yDomainMax * 1.1 || 1])
    .nice()
    .range([height, 0])

  // Grid lines
  svg.append('g')
    .attr('class', 'grid')
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(''))
    .select('.domain').remove()
  svg.selectAll('.grid line').attr('stroke', '#e9ecef').attr('stroke-dasharray', '3,3')

  if (props.stacked) {
    const stack = d3.stack().keys(datasets.map(ds => ds.key))
    const stackedData = stack(displayData)

    const area = d3.area()
      .x(d => x(parseDate(d.data[props.xKey])))
      .y0(d => y(d[0]))
      .y1(d => y(d[1]))
      .curve(d3.curveMonotoneX)

    svg.selectAll('.layer')
      .data(stackedData)
      .join('path')
      .attr('class', 'layer')
      .attr('fill', d => datasets.find(ds => ds.key === d.key).color)
      .attr('fill-opacity', 0.6)
      .attr('d', area)

    // Optional lines on top of stacked area
    svg.selectAll('.layer-line')
      .data(stackedData)
      .join('path')
      .attr('fill', 'none')
      .attr('stroke', d => datasets.find(ds => ds.key === d.key).color)
      .attr('stroke-width', 1)
      .attr('d', d3.line()
        .x(d => x(parseDate(d.data[props.xKey])))
        .y(d => y(d[1]))
        .curve(d3.curveMonotoneX)
      )
  } else {
    datasets.forEach(ds => {
      // Area
      const area = d3.area()
        .defined(d => !isNaN(+d[ds.key]))
        .x(d => x(parseDate(d[props.xKey])))
        .y0(height)
        .y1(d => y(+d[ds.key]))
        .curve(d3.curveMonotoneX)

      svg.append('path')
        .datum(displayData)
        .attr('fill', ds.color)
        .attr('fill-opacity', 0.1)
        .attr('d', area)

      // Line
      const line = d3.line()
        .defined(d => !isNaN(+d[ds.key]))
        .x(d => x(parseDate(d[props.xKey])))
        .y(d => y(+d[ds.key]))
        .curve(d3.curveMonotoneX)

      svg.append('path')
        .datum(displayData)
        .attr('fill', 'none')
        .attr('stroke', ds.color)
        .attr('stroke-width', 2)
        .attr('d', line)
    })
  }

  // Legend
  if (datasets.length > 0) {
    const legend = svg.append('g')
      .attr('font-family', 'sans-serif')
      .attr('font-size', 10)
      .selectAll('g')
      .data(datasets)
      .join('g')
      .attr('transform', (d, i) => `translate(${i * 65}, -5)`)

    legend.append('rect')
      .attr('x', 0)
      .attr('width', 10)
      .attr('height', 10)
      .attr('fill', d => d.color)

    legend.append('text')
      .attr('x', 14)
      .attr('y', 5)
      .attr('dy', '0.35em')
      .text(d => d.label || d.key)
    }

    const dayMs = 24 * 60 * 60 * 1000
    const rangeMs = domainEnd - domainStart
    let xFormat = d3.timeFormat('%Y'); let xTicks = 8

    if (rangeMs < 40 * dayMs) {
      xFormat = d3.timeFormat('%b %d'); xTicks = 7
    } else if (rangeMs < 200 * dayMs) {
      xFormat = d3.timeFormat('%b %d'); xTicks = 6
    } else if (rangeMs < 800 * dayMs) {
      xFormat = d3.timeFormat('%b %Y'); xTicks = 8
    }

    svg.append('g')
      .attr('transform', `translate(0,${height})`)
      .call(d3.axisBottom(x).ticks(xTicks).tickFormat(xFormat))

  // Tooltip on hover
  const tooltip = d3.select(el).append('div')
    .attr('class', 'position-absolute p-2 bg-dark text-white rounded small')
    .style('pointer-events', 'none')
    .style('display', 'none')
    .style('z-index', '1000')
    .style('white-space', 'nowrap')
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
      const dateStr = new Date(parseDate(nearest[props.xKey])).toLocaleDateString()

      const lines = datasets.map(ds => {
        const val = +nearest[ds.key]
        return `<div class="d-flex justify-content-between gap-3">
          <span style="color:${ds.color}">●</span> <span>${ds.label || ds.key}:</span>
          <span class="fw-bold">${val.toLocaleString()}</span>
        </div>`
      }).join('')

      tooltip
        .style('display', 'block')
        .html(`<strong>${dateStr}</strong><hr class="my-1 border-secondary" />${lines}`)

      // Position tooltip near cursor but ensure it doesn't go off-screen
      const tooltipRect = tooltip.node().getBoundingClientRect()
      const tooltipWidth = tooltipRect.width || 140
      const tooltipHeight = tooltipRect.height || 50

      const mousePos = d3.pointer(event, el)
      let left = mousePos[0] - tooltipWidth / 2
      let top = mousePos[1] - tooltipHeight - 15

      // Keep tooltip within bounds
      if (left < 0) left = 0
      if (left + tooltipWidth > totalWidth + margin.left + margin.right) left = (totalWidth + margin.left + margin.right) - tooltipWidth
      if (top < 0) top = mousePos[1] + 15

      tooltip
        .style('left', left + 'px')
        .style('top', top + 'px')
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
    <div ref="container" style="width:100%; position: relative;" :style="{ height: height + 'px' }"></div>
  </div>
</template>
