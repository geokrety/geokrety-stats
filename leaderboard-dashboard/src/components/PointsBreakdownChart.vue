<script setup>
import { ref, onMounted, watch } from 'vue'

const props = defineProps({
  data: {
    type: Array,
    required: true
  },
  height: {
    type: Number,
    default: 300
  }
})

const container = ref(null)

onMounted(() => {
  if (props.data && props.data.length > 0) {
    drawChart()
  }
})

watch(() => props.data, () => {
  if (container.value && window.d3) {
    const d3 = window.d3
    d3.select(container.value).selectAll('*').remove()
    drawChart()
  }
})

function drawChart() {
  if (!props.data || props.data.length === 0 || !container.value || !window.d3) return

  const d3 = window.d3
  const margin = { top: 20, right: 30, bottom: 100, left: 60 }
  const width = container.value.offsetWidth - margin.left - margin.right
  const height = props.height - margin.top - margin.bottom

  // Color scheme for different bonus types
  const colorScale = d3.scaleOrdinal()
    .domain(['base_move', 'relay_mover', 'rescuer', 'chain', 'country_crossing', 'diversity', 'handover', 'reach'])
    .range(['#0d6efd', '#198754', '#fd7e14', '#6f42c1', '#20c997', '#ffc107', '#dc3545', '#0dcaf0'])

  const svg = d3.select(container.value)
    .append('svg')
    .attr('width', width + margin.left + margin.right)
    .attr('height', height + margin.top + margin.bottom)
    .append('g')
    .attr('transform', `translate(${margin.left},${margin.top})`)

  const x = d3.scaleBand()
    .domain(props.data.map(d => d.source || d.label))
    .range([0, width])
    .padding(0.2)

  const y = d3.scaleLinear()
    .domain([0, d3.max(props.data, d => d.points)])
    .range([height, 0])

  // Bars
  svg.selectAll('.bar')
    .data(props.data)
    .enter()
    .append('rect')
    .attr('class', 'bar')
    .attr('x', d => x(d.source || d.label))
    .attr('y', d => y(d.points))
    .attr('width', x.bandwidth())
    .attr('height', d => height - y(d.points))
    .attr('fill', d => colorScale(d.label || d.source))
    .attr('opacity', 0.8)
    .on('mouseover', function () {
      d3.select(this).attr('opacity', 1)
    })
    .on('mouseout', function () {
      d3.select(this).attr('opacity', 0.8)
    })

  // Value labels on bars
  svg.selectAll('.value-label')
    .data(props.data)
    .enter()
    .append('text')
    .attr('class', 'value-label')
    .attr('x', d => x(d.source || d.label) + x.bandwidth() / 2)
    .attr('y', d => y(d.points) - 5)
    .attr('text-anchor', 'middle')
    .attr('font-size', '12px')
    .attr('font-weight', 'bold')
    .text(d => d.points ? d.points.toFixed(1) : '0')

  // X Axis
  svg.append('g')
    .attr('transform', `translate(0,${height})`)
    .call(d3.axisBottom(x))
    .append('text')
    .attr('x', width / 2)
    .attr('y', 40)
    .attr('fill', 'black')
    .attr('text-anchor', 'middle')
    .text('Bonus Type')

  svg.selectAll('.tick text')
    .style('font-size', '12px')
    .attr('transform', 'rotate(-45)')
    .style('text-anchor', 'end')

  // Y Axis
  svg.append('g')
    .call(d3.axisLeft(y))
    .append('text')
    .attr('transform', 'rotate(-90)')
    .attr('y', 0 - margin.left)
    .attr('x', 0 - (height / 2))
    .attr('dy', '1em')
    .attr('fill', 'black')
    .attr('text-anchor', 'middle')
    .text('Points')
}
</script>

<template>
  <div ref="container" style="width:100%;"></div>
</template>

<style scoped>
:deep(svg) {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}
:deep(.tick) {
  font-size: 12px;
}
:deep(.axis-label) {
  font-size: 14px;
  font-weight: 500;
}
</style>
