<script>
export default {
  name: 'Drag',
  props: ['width'],
  emits: ['resize'],
  mounted() {
    const self = this
    let startX
    let initW
    const onMouseMove = function (e) {
      const offset = e.clientX - startX
      const newWidth = initW + offset
      self.$emit('resize', newWidth)
    }
    const onMouseUp = function () {
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
    }
    this.$el.addEventListener('mousedown', function (e) {
      startX = e.clientX
      initW = self.width
      document.addEventListener('mousemove', onMouseMove)
      document.addEventListener('mouseup', onMouseUp)
    })
  },
}
</script>

<template>
  <div class="drag"></div>
</template>
