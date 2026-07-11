<script>
// `class` is always a fallthrough attribute in Vue3 (it cannot be a declared
// prop), so we disable auto-inheritance and bind `$attrs.class` onto the root
// `.dropdown` explicitly, matching the Vue2 original.
export default {
  name: 'Dropdown',
  inheritAttrs: false,
  props: ['toggleClass', 'drop', 'title'],
  data() {
    return { open: false }
  },
  computed: {
    btnToggleClass() {
      let c = this.toggleClass || ''
      c += ' dropdown-toggle dropdown-toggle-no-caret'
      c += this.open ? ' show' : ''
      return c.trim()
    },
  },
  methods: {
    toggle() {
      this.open ? this.hide() : this.show()
    },
    show() {
      this.open = true
      this.$refs.menu.style.top = this.$refs.btn.offsetHeight + 'px'
      const drop = this.drop

      if (drop === 'right') {
        this.$refs.menu.style.left = 'auto'
        this.$refs.menu.style.right = '0'
      } else if (drop === 'center') {
        this.$nextTick(
          function () {
            const btnWidth = this.$refs.btn.getBoundingClientRect().width
            const menuWidth = this.$refs.menu.getBoundingClientRect().width
            this.$refs.menu.style.left = '-' + (menuWidth - btnWidth) / 2 + 'px'
          }.bind(this),
        )
      }

      document.addEventListener('click', this.clickHandler)
    },
    hide() {
      this.open = false
      document.removeEventListener('click', this.clickHandler)
    },
    clickHandler(e) {
      const dropdown = e.target.closest('.dropdown')
      if (dropdown == null || dropdown != this.$el) return this.hide()
      if (e.target.closest('.dropdown-item') != null) return this.hide()
    },
  },
}
</script>

<template>
  <div class="dropdown" :class="$attrs.class">
    <button ref="btn" @click="toggle" :class="btnToggleClass" :title="title">
      <slot name="button"></slot>
    </button>
    <div ref="menu" class="dropdown-menu" :class="{ show: open }">
      <slot v-if="open"></slot>
    </div>
  </div>
</template>
