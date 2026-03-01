<script setup>
const props = defineProps({
  title: {
    type: String,
    default: 'Move Type Breakdown',
  },
  drops: {
    type: Number,
    default: 0,
  },
  grabs: {
    type: Number,
    default: 0,
  },
  dips: {
    type: Number,
    default: 0,
  },
  seen: {
    type: Number,
    default: 0,
  },
  comments: {
    type: Number,
    default: 0,
  },
  loves: {
    type: Number,
    default: 0,
  },
})

function formatInt(num) {
  if (!num) return '0'
  return Math.round(num).toLocaleString()
}
</script>

<template>
  <div class="card shadow-sm mb-2">
    <div class="card-header bg-light">
      <h5 class="mb-0">{{ props.title }}</h5>
    </div>
    <div class="card-body">
      <div class="row">
        <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
          <div class="fs-2 mb-2">🌳</div>
          <div class="text-muted small" title="GeoKrety placed into a cache">Drops</div>
          <div class="fs-5 fw-bold">{{ formatInt(props.drops) }}</div>
        </div>
        <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
          <div class="fs-2 mb-2">🚀</div>
          <div class="text-muted small" title="GeoKrety taken from a cache or person">Grabs</div>
          <div class="fs-5 fw-bold">{{ formatInt(props.grabs) }}</div>
        </div>
        <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
          <div class="fs-2 mb-2">🥾</div>
          <div class="text-muted small" title="Virtual carry - GeoKrety held digitally without physical cache">DIPs</div>
          <div class="fs-5 fw-bold">{{ formatInt(props.dips) }}</div>
        </div>
        <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
          <div class="fs-2 mb-2">👀</div>
          <div class="text-muted small" title="GeoKrety spotted but not taken or placed">Seen</div>
          <div class="fs-5 fw-bold">{{ formatInt(props.seen) }}</div>
        </div>
        <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
          <div class="fs-2 mb-2">📝</div>
          <div class="text-muted small" title="Log a comment about GeoKrety status">Comments</div>
          <div class="fs-5 fw-bold">{{ formatInt(props.comments) }}</div>
        </div>
        <div class="col-6 col-md-4 col-lg-2 text-center mb-2">
          <div class="fs-2 mb-2">❤️</div>
          <div class="text-muted small" title="Favorite/love marks given to GeoKrety">Loves</div>
          <div class="fs-5 fw-bold text-danger">{{ formatInt(props.loves) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
