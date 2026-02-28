<script setup>
import { ref, onMounted, watch } from 'vue'

const props = defineProps({
  data:   { type: Array,  required: true },
  xKey:   { type: String, default: 'x' },
  yKey:   { type: String, default: 'y' },
  color:  { type: String, default: '#0d6efd' },
  height: { type: Number, default: 220 },
})

const container = ref(null)

function draw() {
  if (!container.value || !window.d3 || !props.data.length) return
  const d3 = window.d3

  const el = container.value
  el.innerHTML = ''

  const margin = { top: 10, right: 20, bottom: 60, left: 60 }
  const width  = el.clientWidth - margin.left - margin.right
  const height = props.height - margin.top - margin.bottom

  const svg = d3.select(el)
    .append('svg')
    .attr('width', '100%')
    .attr('height', props.height)
    .attr('viewBox', `0 0 ${el.clientWidth} ${props.height}`)
    .attr('preserveAspectRatio', 'xMidYMid meet')
    .append('g')
    .attr('transform', `translate(${margin.left},${margin.top})`)

  const x = d3.scaleBand()
    .domain(props.data.map(d => d[props.xKey]))
    .range([0, width])
    .padding(0.25)

  const y = d3.scaleLinear()
    .domain([0, d3.max(props.data, d => +d[props.yKey]) * 1.1])
    .nice()
    .range([height, 0])

  // Grid
  svg.append('g')
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(''))
    .select('.domain').remove()
  svg.selectAll('.tick line').attr('stroke', '#e9ecef').attr('stroke-dasharray', '3,3')

  // Bars with tooltip
  const tooltip = d3.select(el).append('div')
    .attr('class', 'position-absolute p-2 bg-dark text-white rounded small')
    .style('pointer-events', 'none')
    .style('display', 'none')
    .style('z-index', '1000')
    .style('white-space', 'nowrap')
    .attr('role', 'tooltip')

  svg.selectAll('rect')
    .data(props.data)
    .enter().append('rect')
    .attr('x', d => x(d[props.xKey]))
    .attr('y', d => y(+d[props.yKey]))
    .attr('width', x.bandwidth())
    .attr('height', d => height - y(+d[props.yKey]))
    .attr('fill', props.color)
    .attr('rx', 2)
    .on('mouseover', (event, d) => {
      const xPos = x(d[props.xKey]) + x.bandwidth() / 2
      const yPos = y(+d[props.yKey])
      
      tooltip
        .html(`<strong>${d[props.xKey]}</strong><br/>${(+d[props.yKey]).toLocaleString()}`)
        .style('display', 'block')

      // Position tooltip above bar but ensure it doesn't go off-screen
      const tooltipWidth = tooltip.node().offsetWidth || 120
      const tooltipHeight = tooltip.node().offsetHeight || 50
      let left = margin.left + xPos - tooltipWidth / 2
      let top = margin.top + yPos - tooltipHeight - 8
      
      // Keep tooltip within bounds
      if (left < 0) left = 0
      if (left + tooltipWidth > el.clientWidth) left = el.clientWidth - tooltipWidth
      if (top < 0) top = margin.top + yPos + 8
      
      tooltip
        .style('left', left + 'px')
        .style('top', top + 'px')
    })
    .on('mouseout', () => {
      tooltip.style('display', 'none')
    })

  // X axis
  svg.append('g')
    .attr('transform', `translate(0,${height})`)
    .call(d3.axisBottom(x))
    .selectAll('text')
    .attr('font-size', '11px')
    .attr('transform', 'rotate(-35)')
    .attr('text-anchor', 'end')

  // Y axis
  svg.append('g')
    .call(d3.axisLeft(y).ticks(5).tickFormat(d3.format('~s')))
    .selectAll('text').attr('font-size', '11px')
}

onMounted(draw)
watch(() => props.data, draw, { deep: true })
</script>

<template>
  <div ref="container" style="width:100%; position: relative;" :style="{ height: height + 'px' }"></div>
</template>
