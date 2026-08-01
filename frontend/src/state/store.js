import { reactive, computed, watch, nextTick } from 'vue'
import { api } from '@/api/api.js'

export const TITLE = 'yarr!'

export function debounce(callback, wait) {
  let timeout
  return function () {
    const ctx = this
    const args = arguments
    clearTimeout(timeout)
    timeout = setTimeout(function () {
      callback.apply(ctx, args)
    }, wait)
  }
}

export function scrollto(target, scroll) {
  const padding = 10
  const targetRect = target.getBoundingClientRect()
  const scrollRect = scroll.getBoundingClientRect()

  const relativeOffset = targetRect.y - scrollRect.y
  const absoluteOffset = relativeOffset + scroll.scrollTop

  if (
    padding <= relativeOffset &&
    relativeOffset + targetRect.height <= scrollRect.height - padding
  )
    return

  let newPos = scroll.scrollTop
  if (relativeOffset < padding) {
    newPos = absoluteOffset - padding
  } else {
    newPos = absoluteOffset - scrollRect.height + targetRect.height + padding
  }
  scroll.scrollTop = Math.round(newPos)
}

function dateRepr(d) {
  let sec = (new Date().getTime() - d.getTime()) / 1000
  const neg = sec < 0
  let out = ''

  sec = Math.abs(sec)
  if (sec < 2700)
    // less than 45 minutes
    out = Math.round(sec / 60) + 'm'
  else if (sec < 86400)
    // less than 24 hours
    out = Math.round(sec / 3600) + 'h'
  else if (sec < 604800)
    // less than a week
    out = Math.round(sec / 86400) + 'd'
  else
    out = d.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })

  if (neg) return '-' + out
  return out
}
export { dateRepr }

// Refs to scroll containers, registered by components (replaces vm.$refs).
export const refs = {
  itemlist: null,
  content: null,
}

export const refreshRateOptions = [
  { title: '0', value: 0 },
  { title: '10m', value: 10 },
  { title: '30m', value: 30 },
  { title: '1h', value: 60 },
  { title: '2h', value: 120 },
  { title: '4h', value: 240 },
  { title: '12h', value: 720 },
  { title: '24h', value: 1440 },
]

export const state = reactive({
  filterSelected: '',
  folders: [],
  feeds: [],
  feedSelected: null,
  feedListWidth: 300,
  feedNewChoice: [],
  feedNewChoiceSelected: '',
  items: [],
  itemsHasMore: true,
  itemSelected: null,
  itemSelectedDetails: null,
  itemSelectedReadability: '',
  itemSearch: '',
  itemListWidth: 300,

  settings: '',
  loading: {
    newfeed: false,
    items: false,
    readability: false,
  },
  fonts: ['', 'serif', 'monospace'],
  feedStats: {},
  // Article body typography (the colour theme feature was removed).
  theme: {
    font: '',
    size: 1,
  },
  refreshRate: 0,
  authenticated: false,
  feed_errors: {},

  opmlCompareResult: [],
})

// Hydrate reactive state from persisted settings + status (replaces Go template
// injection of window.app.settings / window.app.authenticated).
export function hydrate(s, status) {
  // 'starred' is a leftover from the removed feature; fall back to no filter.
  state.filterSelected = s.filter === 'unread' ? 'unread' : ''
  state.feedSelected = s.feed
  state.feedListWidth = s.feed_list_width || 300
  state.itemListWidth = s.item_list_width || 300
  state.theme.font = s.theme_font
  state.theme.size = s.theme_size
  state.refreshRate = s.refresh_rate
  state.authenticated = status.authenticated
}

// ---- computed ----

export const foldersWithFeeds = computed(function () {
  const feedsByFolders = state.feeds.reduce(function (folders, feed) {
    if (!folders[feed.folder_id]) folders[feed.folder_id] = [feed]
    else folders[feed.folder_id].push(feed)
    return folders
  }, {})
  const folders = state.folders.map(function (folder) {
    const feeds = (feedsByFolders[folder.id] || []).slice().sort((a, b) => {
      const a1 = state.feedStats[a.id] || { unread: 0 }
      const b1 = state.feedStats[b.id] || { unread: 0 }
      return b1.unread - a1.unread
    })
    return {
      id: folder.id,
      title: folder.title,
      is_expanded: folder.is_expanded,
      feeds,
    }
  })
  folders.push({ id: null, feeds: feedsByFolders[null] || [] })
  return folders
})

export const feedsById = computed(function () {
  return state.feeds.reduce(function (acc, f) {
    acc[f.id] = f
    return acc
  }, {})
})

export const foldersById = computed(function () {
  return state.folders.reduce(function (acc, f) {
    acc[f.id] = f
    return acc
  }, {})
})

// Per-feed / per-folder / total counters for the active filter, shown in the
// feed list. Computed rather than assigned so it recomputes whenever any of its
// three inputs lands — feeds and stats come from separate requests and either
// can arrive last on a slow connection.
export const filteredStats = computed(function () {
  const filter = state.filterSelected
  if (!filter) return { feeds: {}, folders: {}, total: null }

  const feeds = {},
    folders = {}
  let total = 0

  for (let i = 0; i < state.feeds.length; i++) {
    const feed = state.feeds[i]
    if (!state.feedStats[feed.id]) continue

    const n = state.feedStats[feed.id][filter] || 0

    if (!folders[feed.folder_id]) folders[feed.folder_id] = 0

    feeds[feed.id] = n
    folders[feed.folder_id] += n
    total += n
  }

  return { feeds, folders, total }
})

export const current = computed(function () {
  const parts = (state.feedSelected || '').split(':', 2)
  const type = parts[0]
  const guid = parts[1]

  let folder = {},
    feed = {}

  if (type == 'feed') feed = feedsById.value[guid] || {}
  if (type == 'folder') folder = foldersById.value[guid] || {}

  return { type: type, feed: feed, folder: folder }
})

export const itemSelectedContent = computed(function () {
  if (!state.itemSelected) return ''
  if (state.itemSelectedReadability) return state.itemSelectedReadability
  return state.itemSelectedDetails.content || ''
})

export const contentImages = computed(function () {
  if (!state.itemSelectedDetails) return []
  return (state.itemSelectedDetails.media_links || []).filter(
    (l) => l.type === 'image',
  )
})

export const contentAudios = computed(function () {
  if (!state.itemSelectedDetails) return []
  return (state.itemSelectedDetails.media_links || []).filter(
    (l) => l.type === 'audio',
  )
})

export const contentVideos = computed(function () {
  if (!state.itemSelectedDetails) return []
  return (state.itemSelectedDetails.media_links || []).filter(
    (l) => l.type === 'video',
  )
})

export const refreshRateTitle = computed(function () {
  const entry = refreshRateOptions.find((o) => o.value === state.refreshRate)
  return entry ? entry.title : '0'
})

// ---- methods ----

export function refreshStats() {
  return api.status().then(function (data) {
    state.feedStats = data.stats.reduce(function (acc, stat) {
      acc[stat.feed_id] = stat
      return acc
    }, {})

    api.feeds.list_errors().then(function (errors) {
      state.feed_errors = errors
    })
  })
}

export function getItemsQuery() {
  const query = {}
  if (state.feedSelected) {
    const parts = state.feedSelected.split(':', 2)
    const type = parts[0]
    const guid = parts[1]
    if (type == 'feed') {
      query.feed_id = guid
    } else if (type == 'folder') {
      query.folder_id = guid
    }
  }
  if (state.filterSelected) {
    query.status = state.filterSelected
  }
  if (state.itemSearch) {
    query.search = state.itemSearch
  }
  return query
}

export function refreshFeeds() {
  return Promise.all([api.folders.list(), api.feeds.list()]).then(function (
    values,
  ) {
    state.folders = values[0]
    state.feeds = values[1]
  })
}

export function refreshItems(loadMore = false) {
  if (state.feedSelected === null) {
    state.items = []
    state.itemsHasMore = false
    return
  }

  const query = getItemsQuery()
  if (loadMore) {
    query.after = state.items[state.items.length - 1].id
  }

  state.loading.items = true
  return api.items.list(query).then(function (data) {
    if (loadMore) {
      state.items = state.items.concat(data.list)
    } else {
      state.items = data.list
    }
    state.itemsHasMore = data.has_more
    state.loading.items = false

    // load more if there's some space left at the bottom of the item list.
    nextTick(function () {
      if (state.itemsHasMore && !state.loading.items && itemListCloseToBottom()) {
        refreshItems(true)
      }
    })
  })
}

export function itemListCloseToBottom() {
  // approx. vertical space at the bottom of the list (loading el & paddings) when 1rem = 16px
  const bottomSpace = 70
  const scale =
    (parseFloat(getComputedStyle(document.documentElement).fontSize) || 16) / 16

  const el = refs.itemlist
  if (!el) return false

  if (el.scrollHeight === 0) return false // element is invisible (responsive design)

  const closeToBottom =
    el.scrollHeight - el.scrollTop - el.offsetHeight < bottomSpace * scale
  return closeToBottom
}

export function loadMoreItems() {
  if (!state.itemsHasMore) return
  if (state.loading.items) return
  if (itemListCloseToBottom()) return refreshItems(true)
  if (
    state.itemSelected &&
    state.itemSelected === state.items[state.items.length - 1].id
  )
    return refreshItems(true)
}

export function markItemsRead() {
  const query = getItemsQuery()
  api.items.mark_read(query).then(function () {
    state.items = []
    state.itemSelected = null
    state.itemsHasMore = false
    refreshStats()
  })
}

export function toggleFolderExpanded(folder) {
  const real = state.folders.find((f) => f.id === folder.id)
  if (!real) return
  real.is_expanded = !real.is_expanded
  api.folders.update(real.id, { is_expanded: real.is_expanded })
}

export function formatDate(datestr) {
  const options = {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }
  return new Date(datestr).toLocaleDateString(undefined, options)
}

export function moveFeed(feed, folder) {
  const folder_id = folder ? folder.id : null
  api.feeds.update(feed.id, { folder_id: folder_id }).then(function () {
    feed.folder_id = folder_id
    refreshStats()
  })
}

export function moveFeedToNewFolder(feed) {
  const title = prompt('Enter folder name:')
  if (!title) return
  api.folders.create({ title: title }).then(function (folder) {
    api.feeds.update(feed.id, { folder_id: folder.id }).then(function () {
      refreshFeeds().then(function () {
        refreshStats()
      })
    })
  })
}

export function createNewFeedFolder() {
  const title = prompt('Enter folder name:')
  if (!title) return
  return api.folders.create({ title: title }).then(function (result) {
    return refreshFeeds().then(function () {
      return result
    })
  })
}

export function renameFolder(folder) {
  const newTitle = prompt('Enter new title', folder.title)
  if (newTitle) {
    api.folders.update(folder.id, { title: newTitle }).then(function () {
      folder.title = newTitle
      state.folders.sort(function (a, b) {
        return a.title.localeCompare(b.title)
      })
    })
  }
}

export function deleteFolder(folder) {
  if (confirm('Are you sure you want to delete ' + folder.title + '?')) {
    api.folders.delete(folder.id).then(function () {
      state.feedSelected = null
      refreshStats()
      refreshFeeds()
    })
  }
}

export function updateFeedLink(feed) {
  const newLink = prompt('Enter feed link', feed.feed_link)
  if (newLink) {
    api.feeds.update(feed.id, { feed_link: newLink }).then(function () {
      feed.feed_link = newLink
    })
  }
}

export function renameFeed(feed) {
  const newTitle = prompt('Enter new title', feed.title)
  if (newTitle) {
    api.feeds.update(feed.id, { title: newTitle }).then(function () {
      feed.title = newTitle
    })
  }
}

export function deleteFeed(feed) {
  if (confirm('Are you sure you want to delete ' + feed.title + '?')) {
    api.feeds.delete(feed.id).then(function () {
      state.feedSelected = null
      refreshStats()
      refreshFeeds()
    })
  }
}

export function createFeed(event) {
  const form = event.target
  const data = {
    url: form.querySelector('input[name=url]').value,
    folder_id:
      parseInt(form.querySelector('select[name=folder_id]').value) || null,
    use_proxy: form.querySelector('select[name=use_proxy]').value === 'true',
  }
  if (state.feedNewChoiceSelected) {
    data.url = state.feedNewChoiceSelected
  }
  state.loading.newfeed = true
  api.feeds.create(data).then(function (result) {
    if (result.status === 'success') {
      refreshFeeds()
      refreshStats()
      state.settings = ''
      state.feedSelected = 'feed:' + result.feed.id
    } else if (result.status === 'multiple') {
      state.feedNewChoice = result.choice
      state.feedNewChoiceSelected = result.choice[0].url
    } else {
      alert('No feeds found at the given url.')
    }
    state.loading.newfeed = false
  })
}

export function toggleItemStatus(item, targetstatus, fallbackstatus) {
  const oldstatus = item.status
  const newstatus =
    item.status !== targetstatus ? targetstatus : fallbackstatus

  const updateStats = function (status, incr) {
    if (status == 'unread') {
      state.feedStats[item.feed_id][status] += incr
    }
  }

  api.items.update(item.id, { status: newstatus }).then(function () {
    updateStats(oldstatus, -1)
    updateStats(newstatus, +1)

    const itemInList = state.items.find(function (i) {
      return i.id == item.id
    })
    if (itemInList) itemInList.status = newstatus
    item.status = newstatus
  })
}

export function toggleItemRead(item) {
  toggleItemStatus(item, 'unread', 'read')
}

export function logout() {
  api.logout().then(function () {
    document.location.reload()
  })
}

export function toggleReadability() {
  if (state.itemSelectedReadability) {
    state.itemSelectedReadability = null
    return
  }
  const item = state.itemSelectedDetails
  if (!item) return
  if (item.link) {
    state.loading.readability = true
    api.crawl(item.link).then(function (data) {
      state.itemSelectedReadability = data && data.content
      state.loading.readability = false
    })
  }
}

export function showSettings(settings) {
  state.settings = settings
  if (settings === 'create') {
    state.feedNewChoice = []
    state.feedNewChoiceSelected = ''
  }
}

export function resizeFeedList(width) {
  state.feedListWidth = Math.min(Math.max(200, width), 700)
}

export function resizeItemList(width) {
  state.itemListWidth = Math.min(Math.max(200, width), 700)
}

export function resetFeedChoice() {
  state.feedNewChoice = []
  state.feedNewChoiceSelected = ''
}

export function incrFont(x) {
  state.theme.size = +(state.theme.size + 0.1 * x).toFixed(1)
}

export function fetchAllFeeds() {
  // Concurrent refreshes are rejected server-side (worker.RefreshFeeds).
  api.feeds.refresh().then(function () {
    refreshStats()
  })
}

export function mustHideFolder() {
  return false
}

export function mustHideFeed() {
  return false
}

// navigation helper, navigate relative to selected item
export function navigateToItem(relativePosition) {
  if (state.itemSelected == null) {
    if (state.items.length !== 0) state.itemSelected = state.items[0].id
    return
  }

  const itemPosition = state.items.findIndex(function (x) {
    return x.id === state.itemSelected
  })
  if (itemPosition === -1) {
    if (state.items.length !== 0) state.itemSelected = state.items[0].id
    return
  }

  const newPosition = itemPosition + relativePosition
  if (newPosition < 0 || newPosition >= state.items.length) return

  state.itemSelected = state.items[newPosition].id

  nextTick(function () {
    const scroll = document.querySelector('#item-list-scroll')
    const handle = scroll.querySelector('input[type=radio]:checked')
    const target = handle && handle.parentElement
    if (target && scroll) scrollto(target, scroll)
    loadMoreItems()
  })
}

// navigation helper, navigate relative to selected feed
export function navigateToFeed(relativePosition) {
  const navigationList = foldersWithFeeds.value
    .filter((folder) => !folder.id || !mustHideFolder(folder))
    .map((folder) => {
      if (mustHideFolder(folder)) return []
      const folds = folder.id ? [`folder:${folder.id}`] : []
      const feeds =
        folder.is_expanded || !folder.id
          ? folder.feeds
              .filter((f) => !mustHideFeed(f))
              .map((f) => `feed:${f.id}`)
          : []
      return folds.concat(feeds)
    })
    .flat()
  navigationList.unshift('')

  const currentFeedPosition = navigationList.indexOf(state.feedSelected)

  if (currentFeedPosition == -1) {
    state.feedSelected = ''
    return
  }

  const newPosition = currentFeedPosition + relativePosition
  if (newPosition < 0 || newPosition >= navigationList.length) return

  state.feedSelected = navigationList[newPosition]

  nextTick(function () {
    const scroll = document.querySelector('#feed-list-scroll')
    const handle = scroll.querySelector('input[type=radio]:checked')
    const target = handle && handle.parentElement
    if (target && scroll) scrollto(target, scroll)
  })
}

export function changeRefreshRate(offset) {
  const curIdx = refreshRateOptions.findIndex(
    (o) => o.value === state.refreshRate,
  )
  if (curIdx <= 0 && offset < 0) return
  if (curIdx >= refreshRateOptions.length - 1 && offset > 0) return
  state.refreshRate = refreshRateOptions[curIdx + offset].value
}

// ---- watchers ----
// Registered once at startup (after hydrate) so initial-setup guards behave
// like the Vue2 `oldVal === undefined` checks.

export function registerWatchers() {
  watch(
    () => state.theme,
    function (theme) {
      api.settings.update({
        theme_font: theme.font,
        theme_size: theme.size,
      })
    },
    { deep: true },
  )

  watch(
    () => state.feedStats,
    debounce(function () {
      let title = TITLE
      const unreadCount = Object.values(state.feedStats).reduce(function (
        acc,
        stat,
      ) {
        return acc + stat.unread
      }, 0)
      if (unreadCount) {
        title += ' (' + unreadCount + ')'
      }
      document.title = title
    }, 500),
    { deep: true },
  )

  watch(
    () => state.filterSelected,
    function (newVal) {
      api.settings
        .update({ filter: newVal })
        .then(() => refreshItems(false))
      state.itemSelected = null
    },
  )

  watch(
    () => state.feedSelected,
    function (newVal) {
      api.settings
        .update({ feed: newVal })
        .then(() => refreshItems(false))
      state.itemSelected = null
      if (refs.itemlist) refs.itemlist.scrollTop = 0
    },
  )

  watch(
    () => state.itemSelected,
    function (newVal) {
      state.itemSelectedReadability = ''
      if (newVal === null) {
        state.itemSelectedDetails = null
        return
      }
      if (refs.content) refs.content.scrollTop = 0

      api.items.get(newVal).then(function (item) {
        state.itemSelectedDetails = item
        if (state.itemSelectedDetails.status == 'unread') {
          api.items
            .update(state.itemSelectedDetails.id, { status: 'read' })
            .then(function () {
              state.feedStats[state.itemSelectedDetails.feed_id].unread -= 1
              const itemInList = state.items.find(function (i) {
                return i.id == item.id
              })
              if (itemInList) itemInList.status = 'read'
              state.itemSelectedDetails.status = 'read'
            })
        }
      })
    },
  )

  watch(
    () => state.itemSearch,
    debounce(function () {
      refreshItems()
    }, 500),
  )

  watch(
    () => state.feedListWidth,
    debounce(function (newVal) {
      api.settings.update({ feed_list_width: newVal })
    }, 1000),
  )

  watch(
    () => state.itemListWidth,
    debounce(function (newVal) {
      api.settings.update({ item_list_width: newVal })
    }, 1000),
  )

  watch(
    () => state.refreshRate,
    function (newVal) {
      api.settings.update({ refresh_rate: newVal })
    },
  )
}
