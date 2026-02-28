<script setup>
import { ref, onMounted, watch } from 'vue'

const props = defineProps({
  data:   { type: Array,  required: true },
  xKey:   { type: String, default: 'x' },
  yKey:   { type: String, default: 'y' },
  color:  { type: String, default: '#0d6efd' },
  height: { type: Number, default: 180 },
})

const container = ref(null)

function draw() {
  if (!container.value || !window.d3 || !props.data.length) return
  const d3 = window.d3

  const el = container.value
  el.innerHTML = ''

  const margin = { top: 10, right: 20, bottom: 30, left: 50 }
  const width  = el.clientWidth - margin.left - margin.right
  const height = props.height - margin.top - margin.bottom

  const svg = d3.select(el)
    .append('svg')
    .attr('width', '100%')
    .attr('height', props.height)
    .append('g')
    .attr('transform', `translate(${margin.left},${margin.top})`)

  const parseDate = d3.timeParse('%Y-%m-%d')

  const x = d3.scaleTime()
    .domain(d3.extent(props.data, d => parseDate(d[props.xKey]) || new Date(d[props.xKey])))
    .range([0, width])

  const y = d3.scaleLinear()
    .domain([0, d3.max(props.data, d => +d[props.yKey]) * 1.1])
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
    .x(d => x(parseDate(d[props.xKey]) || new Date(d[props.xKey])))
    .y0(height)
    .y1(d => y(+d[props.yKey]))
    .curve(d3.curveMonotoneX)

  svg.append('path')
    .datum(props.data)
    .attr('fill', props.color)
    .attr('fill-opacity', 0.15)
    .attr('d', area)

  // Line
  const line = d3.line()
    .x(d => x(parseDate(d[props.xKey]) || new Date(d[props.xKey])))
    .y(d => y(+d[props.yKey]))
    .curve(d3.curveMonotoneX)

  svg.append('path')
    .datum(props.data)
    .attr('fill', 'none')
    .attr('stroke', props.color)
    .attr('stroke-width', 2)
    .attr('d', line)

  // Axes
  svg.append('g')
    .attr('transform', `translate(0,${height})`)
    .call(d3.axisBottom(x).ticks(6).tickFormat(d3.timeFormat('%b %d')))
    .selectAll('text').attr('font-size', '11px')

  svg.append('g')
    .call(d3.axisLeft(y).ticks(5).tickFormat(d3.format('~s')))
    .selectAll('text').attr('font-size', '11px')
}

onMounted(draw)
watch(() => props.data, draw, { deep: true })
</script>

<template>
  <div ref="container" style="width:100%;" :style="{ height: height + 'px' }"></div>
</template>
