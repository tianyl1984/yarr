<script>
import { icons } from '@/icons/index.js'

export default {
  name: 'Icon',
  props: ['name'],
  computed: {
    svg() {
      return icons[this.name] || ''
    },
  },
}
</script>

<template>
  <span class="icon" v-html="svg"></span>
</template>
