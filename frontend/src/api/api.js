// Latched once a 401 is seen so the reload happens exactly once. Without this,
// every concurrent request (the initial refresh burst plus the stats poll) would
// each fire its own reload() and keep hammering the API in a loop before the
// navigation to the SSO login page completes.
let sessionExpired = false
const never = new Promise(function () {}) // halts a chain without resolving

const xfetch = function (resource, init) {
  if (sessionExpired) return never
  init = init || {}
  if (['post', 'put', 'delete'].indexOf(init.method) !== -1) {
    init['headers'] = init['headers'] || {}
    init['headers']['x-requested-by'] = 'yarr'
  }
  return fetch(resource, init).then(function (res) {
    // Session expired: send the browser to the SSO login page. The backend
    // gives us the SSO base URL in a header; we build the callback from the
    // browser's own origin (the backend can't see it behind the proxy) so the
    // SSO service redirects back to the address the user actually opened.
    if (res.status === 401) {
      if (!sessionExpired) {
        sessionExpired = true
        const authUrl = res.headers.get('X-Auth-Url')
        if (authUrl) {
          const callback = window.location.origin + '/api/auth/callback'
          document.location.href =
            authUrl.replace(/\/+$/, '') + '/login?callback=' + encodeURIComponent(callback)
        }
      }
      return never // halt the chain while redirecting
    }
    return res
  })
}

const request = function (method, endpoint, data) {
  const headers = { 'Content-Type': 'application/json' }
  return xfetch(endpoint, {
    method: method,
    headers: headers,
    body: JSON.stringify(data),
  })
}

const json = function (res) {
  return res.json()
}

const param = function (query) {
  if (!query) return ''
  return (
    '?' +
    Object.keys(query)
      .map(function (key) {
        return encodeURIComponent(key) + '=' + encodeURIComponent(query[key])
      })
      .join('&')
  )
}

export const api = {
  feeds: {
    list: function () {
      return request('get', './api/feeds').then(json)
    },
    create: function (data) {
      return request('post', './api/feeds', data).then(json)
    },
    update: function (id, data) {
      return request('put', './api/feeds/' + id, data)
    },
    delete: function (id) {
      return request('delete', './api/feeds/' + id)
    },
    refresh: function () {
      return request('post', './api/feeds/refresh')
    },
    list_errors: function () {
      return request('get', './api/feeds/errors').then(json)
    },
  },
  folders: {
    list: function () {
      return request('get', './api/folders').then(json)
    },
    create: function (data) {
      return request('post', './api/folders', data).then(json)
    },
    update: function (id, data) {
      return request('put', './api/folders/' + id, data)
    },
    delete: function (id) {
      return request('delete', './api/folders/' + id)
    },
  },
  items: {
    get: function (id) {
      return request('get', './api/items/' + id).then(json)
    },
    list: function (query) {
      return request('get', './api/items' + param(query)).then(json)
    },
    update: function (id, data) {
      return request('put', './api/items/' + id, data)
    },
    mark_read: function (query) {
      return request('put', './api/items' + param(query))
    },
  },
  settings: {
    get: function () {
      return request('get', './api/settings').then(json)
    },
    update: function (data) {
      return request('put', './api/settings', data)
    },
  },
  status: function () {
    return request('get', './api/status').then(json)
  },
  upload_opml: function (form) {
    return xfetch('./api/opml/import', {
      method: 'post',
      body: new FormData(form),
    })
  },
  compare_opml: function (form) {
    return xfetch('./api/opml/compare', {
      method: 'post',
      body: new FormData(form),
    }).then(json)
  },
  logout: function () {
    return request('post', './api/logout')
  },
  crawl: function (url) {
    return request('get', './api/page?url=' + encodeURIComponent(url)).then(json)
  },
}
