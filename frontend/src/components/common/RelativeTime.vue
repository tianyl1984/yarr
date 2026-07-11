<script>
import { dateRepr } from '@/state/store.js'

export default {
  name: 'RelativeTime',
  props: ['val'],
  data() {
    const d = new Date(this.val)
    return {
      date: d,
      formatted: dateRepr(d),
      interval: null,
    }
  },
  mounted() {
    this.interval = setInterval(
      function () {
        this.formatted = dateRepr(this.date)
      }.bind(this),
      600000,
    ) // every 10 minutes
  },
  unmounted() {
    clearInterval(this.interval)
  },
}
</script>

<template>
  <time :datetime="val">{{ formatted }}</time>
</template>
