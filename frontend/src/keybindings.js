import * as store from '@/state/store.js'

const helperFunctions = {
  scrollContent: function (direction) {
    const padding = 40
    const scroll = document.querySelector('.content')
    if (!scroll) return

    const height = scroll.getBoundingClientRect().height
    const newpos = scroll.scrollTop + (height - padding) * direction

    if (typeof scroll.scrollTo == 'function') {
      scroll.scrollTo({ top: newpos, left: 0, behavior: 'smooth' })
    } else {
      scroll.scrollTop = newpos
    }
  },
}

const shortcutFunctions = {
  openItemLink: function () {
    if (store.state.itemSelectedDetails && store.state.itemSelectedDetails.link) {
      window.open(
        store.state.itemSelectedDetails.link,
        '_blank',
        'noopener,noreferrer',
      )
    }
  },
  toggleReadability: function () {
    store.toggleReadability()
  },
  toggleItemRead: function () {
    if (store.state.itemSelected != null) {
      store.toggleItemRead(store.state.itemSelectedDetails)
    }
  },
  markAllRead: function () {
    // same condition as 'Mark all read button'
    if (store.state.filterSelected == 'unread') {
      store.markItemsRead()
    }
  },
  focusSearch: function () {
    document.getElementById('searchbar').focus()
  },
  nextItem() {
    store.navigateToItem(+1)
  },
  previousItem() {
    store.navigateToItem(-1)
  },
  nextFeed() {
    store.navigateToFeed(+1)
  },
  previousFeed() {
    store.navigateToFeed(-1)
  },
  scrollForward: function () {
    helperFunctions.scrollContent(+1)
  },
  scrollBackward: function () {
    helperFunctions.scrollContent(-1)
  },
  closeItem: function () {
    store.state.itemSelected = null
  },
  showAll() {
    store.state.filterSelected = ''
  },
  showUnread() {
    store.state.filterSelected = 'unread'
  },
  closeModal() {
    store.state.settings = null
  },
}

// If you edit, make sure you update the help modal
const keybindings = {
  v: shortcutFunctions.openItemLink,
  o: shortcutFunctions.openItemLink,
  i: shortcutFunctions.toggleReadability,
  r: shortcutFunctions.toggleItemRead,
  R: shortcutFunctions.markAllRead,
  '/': shortcutFunctions.focusSearch,
  j: shortcutFunctions.nextItem,
  k: shortcutFunctions.previousItem,
  l: shortcutFunctions.nextFeed,
  h: shortcutFunctions.previousFeed,
  f: shortcutFunctions.scrollForward,
  b: shortcutFunctions.scrollBackward,
  q: shortcutFunctions.closeItem,
  1: shortcutFunctions.showUnread,
  2: shortcutFunctions.showAll,
  3: shortcutFunctions.showAll, // legacy binding, kept for muscle memory
}

const codebindings = {
  KeyV: shortcutFunctions.openItemLink,
  KeyO: shortcutFunctions.openItemLink,
  KeyI: shortcutFunctions.toggleReadability,
  //"r": shortcutFunctions.toggleItemRead,
  //"KeyR": shortcutFunctions.markAllRead,
  Slash: shortcutFunctions.focusSearch,
  KeyJ: shortcutFunctions.nextItem,
  KeyK: shortcutFunctions.previousItem,
  KeyL: shortcutFunctions.nextFeed,
  KeyH: shortcutFunctions.previousFeed,
  KeyF: shortcutFunctions.scrollForward,
  KeyB: shortcutFunctions.scrollBackward,
  KeyQ: shortcutFunctions.closeItem,
  Digit1: shortcutFunctions.showUnread,
  Digit2: shortcutFunctions.showAll,
  Digit3: shortcutFunctions.showAll, // legacy binding, kept for muscle memory
  Escape: shortcutFunctions.closeModal,
}

function isTextBox(element) {
  const tagName = element.tagName.toLowerCase()
  // Input elements that aren't text
  const inputBlocklist = [
    'button',
    'checkbox',
    'color',
    'file',
    'hidden',
    'image',
    'radio',
    'range',
    'reset',
    'search',
    'submit',
  ]

  return (
    tagName === 'textarea' ||
    (tagName === 'input' &&
      inputBlocklist.indexOf(element.getAttribute('type').toLowerCase()) == -1)
  )
}

export function registerKeybindings() {
  document.addEventListener('keydown', function (event) {
    // Ignore while focused on text or
    // when using modifier keys (to not clash with browser behaviour)
    if (
      isTextBox(event.target) ||
      event.metaKey ||
      event.ctrlKey ||
      event.altKey
    ) {
      return
    }
    const keybindFunction =
      keybindings[event.key] || codebindings[event.code]
    if (keybindFunction) {
      event.preventDefault()
      keybindFunction()
    }
  })
}
