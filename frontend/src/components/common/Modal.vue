<script>
export default {
  name: 'Modal',
  props: ['open'],
  emits: ['hide'],
  data() {
    return { opening: false }
  },
  watch: {
    open(newVal) {
      if (newVal) {
        this.opening = true
        // document.addEventListener("click", this.handleClick);
      } else {
        // document.removeEventListener("click", this.handleClick);
      }
    },
  },
  methods: {
    handleClick(e) {
      if (this.opening) {
        this.opening = false
        return
      }
      if (e.target.closest('.modal-content') == null) this.$emit('hide')
    },
  },
}
</script>

<template>
  <div class="modal custom-modal" tabindex="-1" v-if="open">
    <div class="modal-dialog">
      <div class="modal-content" ref="content">
        <div class="modal-body">
          <slot v-if="open"></slot>
        </div>
      </div>
    </div>
  </div>
</template>
