import { debounce } from '@/state/store.js'

// Vue3 directive: `bind`/`inserted` -> `mounted`.
export const scroll = {
  mounted(el, binding) {
    el.addEventListener(
      'scroll',
      debounce(function (event) {
        binding.value(event, el)
      }, 200),
    )
  },
}
