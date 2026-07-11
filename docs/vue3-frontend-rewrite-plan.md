# Vue3 前端重构（构建模式，暂不嵌入 Go 二进制）

## Context

`yarr` 的前端目前完全活在 `backend/src/assets/` 里：一个 Go `html/template` 渲染的
`index.html`（自定义分隔符 `{% %}`，避免与 Vue 的 `{{ }}` 冲突）+ 一个无构建步骤的
全局 Vue 2.6.11 实例（`javascripts/app.js`，无 SFC，纯 `Vue.component`/inline
template）+ `api.js`（fetch 封装）+ `key.js`（键盘快捷键）+ 两个全局 CSS 文件
（`app.css` 614 行 + 精简版 `bootstrap.min.css`）+ 28 个通过 Go 模板函数
`{% inline "x.svg" %}` 服务端内联的 SVG 图标。这套无构建管线导致前端代码风格老旧、
难以组件化维护，且与 Go 后端强耦合（依赖服务端模板注入初始数据）。

本次目标：把前端整体迁移到 Vue3 + Vite 构建模式，**保留现有的全部功能和行为**（不
是重新设计 UI/UX），并且按用户要求**这一阶段不考虑把构建产物重新嵌入 Go 二进制**——
新前端是 `frontend/` 目录下一个独立的、可 `npm run dev`/`npm run build` 的项目，
开发时通过 Vite 代理访问现有 Go 后端 API。嵌入构建产物到 Go 二进制是未来阶段的工作。

已与用户确认的关键决策（不再重新讨论）：
1. **纯 JavaScript**，不用 TypeScript。
2. **不引入 vue-router**——保持现状：没有真实 URL 路由，选中的 feed/item/filter
   等状态纯粹活在响应式 JS 状态里，刷新页面靠服务端持久化的 settings 恢复。
3. **允许一个很小的后端改动**：给 `GET /api/status`（`handleStatus`）的 JSON 响应
   增加一个 `authenticated` 字段，这样新前端能在启动时用 AJAX 获取，而不再依赖
   Go 模板注入的 `window.app.authenticated`。
4. **CSS 原样迁移**：`app.css`/`bootstrap.min.css` 原封不动复制为全局样式表，不做
   scoped/组件化拆分，不引入预处理器，保证视觉和响应式行为（`theme-light`/
   `theme-sepia`/`theme-night` 三主题、767.98px/991.98px 断点）不变。

## 现状核对（已读源码确认，非猜测）

- `backend/src/server/routes.go` `handleStatus`（约第 102 行）当前返回
  `{"running": ..., "stats": ...}`，改动是纯增量加一个 key，不影响任何现有消费者。
- `backend/src/server/routes.go` `handleSettings` GET 直接返回
  `s.db.GetSettings()`，字段名（`filter`/`feed`/`feed_list_width`/
  `item_list_width`/`sort_newest_first`/`theme_name`/`theme_font`/`theme_size`/
  `refresh_rate`）与 `app.js` 里 `data()` 读取 `window.app.settings` 的字段一一对应
  （已在 `app.js` 第 235-279 行核实）。
- `app.js` 里所有这些 settings 字段的 watcher（theme/filterSelected/feedSelected/
  itemSortNewestFirst/feedListWidth/itemListWidth/refreshRate）都调用
  `api.settings.update({...})` 做部分字段 PUT 到同一个 `/api/settings` 端点——
  这个机制必须原样保留，**不是死代码**。
- 确认的真正死代码：`mustHideFolder`/`mustHideFeed`（`app.js` 约 921-934 行）恒
  返回 `false`，真实逻辑被注释掉了；`api.feeds.list_items`/`api.folders.list_items`
  在 `routes.go` 路由表里没有对应路由，是死代码，迁移时直接去掉。
- **`index.html` 第 249-251 行有一个 `v-for`+`v-if` 同时用在同一个 `<button>`
  元素上**（"移动到文件夹"下拉菜单项），Vue2 允许（`v-if` 逐项生效），但 **Vue3
  默认禁止/警告 `v-for`+`v-if` 同元素**，这是一个必须处理的、非可选的模板改造点
  （详见下方组件章节）。
- 静态资源当前通过 `/static/...` 前缀服务，`index.html`/`api.js` 全部用**相对
  URL**（`./static/...`、`./api/...`），这是当前 `YARR_BASE` 子路径部署无需前端
  感知 base path 的原因；新 Vite 构建要用 `base: './'` 延续这一约定。
- 图标 SVG 都是独立的 `<svg>...</svg>` 标记（已核实 `anchor.svg`），Go 的 `inline`
  模板函数只是原样读取注入，没有做任何服务端变换——所以 Vite 的 `?raw` 字符串导入
  与现状等价。

## 实施方案

### 1. 项目脚手架

新项目落在 `frontend/`（替换掉现有占位 `README.md`）。手写文件而非依赖网络脚手架
命令：

```
frontend/
├── package.json          # vite + @vitejs/plugin-vue + vue3
├── vite.config.js         # base: './', dev server proxy 到 127.0.0.1:7070
├── index.html             # Vite 根 HTML，替代 Go 模板渲染的 index.html
├── .gitignore              # node_modules, dist
└── src/
    ├── main.js
    ├── App.vue
    ├── state/store.js       # 见 §3
    ├── api/api.js           # 见 §4
    ├── directives/{scroll,focus}.js
    ├── keybindings.js        # 移植 key.js
    ├── icons/index.js         # 28 个 svg 的 ?raw 导入 barrel
    ├── components/
    │   ├── common/{Drag,Dropdown,Modal,RelativeTime,Icon}.vue
    │   ├── FeedList/{FeedList,SettingsDropdown}.vue
    │   ├── ItemList/{ItemList,FeedSettingsDropdown,FolderSettingsDropdown}.vue
    │   ├── ItemPane/ItemPane.vue
    │   └── modals/{NewFeedModal,ShortcutsModal,CompareOpmlModal}.vue
    ├── styles/{app.css,bootstrap.min.css}   # 原样拷贝
    └── assets/icons/*.svg                     # 原样拷贝（favicon 除外，继续走后端 /static）
```

`vite.config.js` 关键内容：

```js
export default defineConfig({
  base: './',
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://127.0.0.1:7070', changeOrigin: true },
      '/opml': { target: 'http://127.0.0.1:7070', changeOrigin: true },
      '/logout': { target: 'http://127.0.0.1:7070', changeOrigin: true },
      '/page': { target: 'http://127.0.0.1:7070', changeOrigin: true },
      '/manifest.json': { target: 'http://127.0.0.1:7070', changeOrigin: true },
    }
  }
})
```
（`/api` 前缀已覆盖 `/api/feeds/:id/icon` 图标请求，无需单独规则。）

### 2. 组件拆分（中等粒度，优先复用现有结构，不过度拆分）

- `App.vue`：根组件，持有启动时的异步加载态（见 §5），渲染三栏布局容器
  （`feed-selected`/`item-selected` class 逻辑与现状一致），组合下面各组件，同一
  时刻只显示一个 modal（`settings` 状态值 `''`/`'create'`/`'shortcuts'`/
  `'compare-opml'`，与现状完全一致）。
- `FeedList.vue`：左栏整体（Drag 手柄、过滤按钮、kebab 设置菜单内容委托给
  `SettingsDropdown.vue`、文件夹/feed 树）。文件夹/feed 树先内联在
  `FeedList.vue` 里，如果实现时发现太臃肿再拆 `FeedTree.vue`——不预先固定。
- `ItemList.vue`：搜索栏、全部已读、`FeedSettingsDropdown.vue`/
  `FolderSettingsDropdown.vue`（依 `current.type` 条件渲染）、无限滚动列表
  （`v-scroll`）。
- `ItemPane.vue`：工具栏（星标/已读/外观/可读性/外链/上下篇导航/关闭）+ 内容渲染
  （含图片/音频/视频提取）。
- `modals/*`：`NewFeedModal.vue`（含发现多个 feed 时的消歧 UI）、
  `ShortcutsModal.vue`、`CompareOpmlModal.vue`，均包一层共享 `Modal.vue`。
- `common/{Drag,Dropdown,Modal,RelativeTime}.vue`：Vue2 全局组件的直接 Options API
  移植（生命周期钩子改名，见 §11 第 7 步）。
- `common/Icon.vue`：`<span class="icon" v-html="icons[name]"></span>`，替代
  `{% inline "x.svg" %}`。

**必须处理的模板改造**：`index.html` 第 249-251 行的移动到文件夹菜单项
（`v-if="folder.id != current.feed.folder_id"` 与 `v-for="folder in folders"`
同元素）在 Vue3 中不被允许，需要改为 `<template v-for="folder in folders">` 包裹
内层 `v-if` 的元素，或者用一个 computed 提前过滤列表。实现时二选一即可，功能行为
不变。

**Options API vs Composition API**：`common/` 下的 4 个全局组件用 Options API
直接移植（与 Vue2 源码结构几乎 1:1，翻译风险最低）；页面级组件
（`App.vue`/`FeedList.vue`/`ItemList.vue`/`ItemPane.vue`）用 `<script setup>` +
从 `state/store.js` 导入共享状态，风险集中在一个可逐行核对的文件里。

### 3. 状态管理：单一 `reactive()` 模块，不引入 Pinia/Vuex

现状本身就是一个扁平的单一 Vue 实例（`data()` ~40 个属性，无 Vuex），引入 Pinia
需要设计源码里不存在的 store 模块边界，属于重新设计而非保留迁移，增加风险却没有
功能收益。`frontend/src/state/store.js` 用 Vue3 的 `reactive()`/`computed()`/
`watch()`（可在组件外直接从 `'vue'` 导入使用）原样对应现有 `data`/`computed`/
`watch`/`methods`：

```js
export const state = reactive({ filterSelected: '', folders: [], ... })
export const foldersWithFeeds = computed(() => { ... })
export function refreshItems(loadMore) { ... }
watch(() => state.theme, (theme) => { ... }, { deep: true })
```

组件里 `import { state, current, refreshItems, ... } from '@/state/store'` 直接用，
Vue3 响应式系统跨组件边界透明追踪，行为等价于现状的单一共享 VM。所有防抖时长
（搜索 500ms、feedStats/title 500ms、宽度持久化 1000ms、滚动 200ms）原样保留。

### 4. API 层

`backend/src/assets/javascripts/api.js` 近乎逐字移植到 `frontend/src/api/api.js`
为 ES module（`export const api = {...}`，保持 `api.feeds.list()` 这种嵌套调用形式
不变，减少调用点改动）。`xfetch` 的 401→`document.location.reload()` 行为不变。
移除确认为死代码的 `api.feeds.list_items`/`api.folders.list_items`。保留
`api.settings.get`/`api.settings.update`（确认在用）。所有相对 URL
（`./api/...`、`./opml/...`、`./page?url=...`、`./logout`）原样保留。

### 5. 启动引导序列

替代现状 `index.html` 里 Go 模板注入的 `window.app.settings`/
`window.app.authenticated`：`main.js`/`App.vue` 在挂载真实 UI 前先并发
`await Promise.all([api.settings.get(), api.status()])`，用返回结果 hydrate
`state`（`state.filterSelected = s.filter` 等，与现状 `data()` 里的赋值一一对应），
`authenticated` 从（改造后的）`GET /api/status` 里取。hydrate 完成前展示一个轻量
loading 态（对应现状的 `v-cloak` 隐藏效果，这里再叠加一次异步等待）。主题 class
应用到 `document.body`、`<meta name="theme-color">` 更新、初始
`refreshStats().then(refreshFeeds).then(() => refreshItems(false))` +
`api.feeds.list_errors()` 均在 hydrate 完成后立即执行，顺序与现状一致。

### 6. 图标

把 `backend/src/assets/graphicarts/*.svg`（除 favicon.png/svg，继续走后端
`/static/graphicarts/` 服务）原样拷贝到 `frontend/src/assets/icons/`，用 Vite 的
`?raw` 后缀导入原始字符串，不引入额外依赖：

```js
import anchor from '../assets/icons/anchor.svg?raw'
export const icons = { anchor, /* ...全部 28 个 */ }
```

`Icon.vue` 用 `v-html` 渲染，DOM 结构与现状字节级一致。

### 7. CSS

`app.css`/`bootstrap.min.css` 原样拷贝到 `frontend/src/styles/`，在 `main.js` 里
按现状 `<link>` 顺序全局 import（先 bootstrap 再 app.css）。不做 scoped/CSS
modules/预处理器改造，所有 class 名称在移植后的模板里保持不变，确保选择器
（含三主题、两个响应式断点）继续生效。

### 8. 开发工作流

两个进程并行：
- 后端：`cd backend && go run ./cmd/yarr -db <dsn>`（默认 `127.0.0.1:7070`）。
- 前端：`cd frontend && npm run dev`（Vite dev server，默认 5173，代理规则见 §1）。

浏览器只访问 `localhost:5173`，Vite 代理转发 API 请求，同源避免 CORS，cookie/
session 行为不受影响。

### 9. 构建产物

`npm run build` 产出 `frontend/dist/`（`base: './'`，相对路径资源引用）。**本阶段
到此为止**——不设计任何把 `dist/` 嵌入/接入 Go 二进制的工作（不碰
`backend/src/assets/assetsfs.go` 的 embed 指令、`handleStatic`、路由表），除了下面
唯一允许的后端改动。

### 10. 唯一的后端改动

**文件**：`backend/src/server/routes.go`，函数 `handleStatus`（约第 102-107 行）

```go
func (s *Server) handleStatus(c *router.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"running":       s.worker.FeedsPending(),
		"stats":         s.db.FeedStats(),
		"authenticated": s.AuthURL != "",
	})
}
```

纯增量改动，`s.AuthURL != ""` 与 `handleIndex` 里计算 `authenticated` 的表达式
完全一致，无破坏性。除此之外不改动任何其他后端文件。

### 11. 实施顺序（每步可独立验证）

1. 脚手架：`package.json`/`vite.config.js`/`index.html`/`main.js` + 一个占位
   `App.vue`，验证 `npm run dev` 能跑起来且能通过代理访问 `/api/status`。
2. 后端改动（§10）——独立、可单独提交。
3. API 层移植（`api.js` → ES module），在浏览器控制台里逐个函数核对。
4. CSS + 图标：拷贝文件，搭好 `Icon.vue`/`icons/index.js`，`main.js` 里 import
   CSS，渲染几个图标与旧前端做视觉比对。
5. Store：逐个属性/计算属性/watcher/方法把 `app.js` 搬进 `state/store.js`，
   对照本次探索得到的完整清单逐项打勾。
6. 启动引导序列接入真实 store，验证 hydrate 后主题/标题正确应用。
7. 公共组件：`Drag`/`Dropdown`/`Modal`/`RelativeTime` + `v-scroll`/`v-focus`
   指令，注意 Vue3 生命周期/指令钩子改名（`bind`/`inserted`/`unbind` →
   `mounted`/`updated`/`unmounted`；`destroyed` → `unmounted`），并核实
   `Dropdown` 组件里把 `class` 声明为具名 prop 这种非常规写法在 Vue3 下的
   `$attrs`/`inheritAttrs` 语义是否仍按预期工作。
8. 主布局组件：`FeedList`/`ItemList`/`ItemPane`，模板从 `index.html` 对应片段
   近乎逐字移植，重点处理第 249-251 行的 `v-for`+`v-if` 改造。
9. Modal + 设置下拉：`NewFeedModal`/`ShortcutsModal`/`CompareOpmlModal`/
   `SettingsDropdown`（OPML 导入/对比表单沿用 `FormData(form)` + template ref）。
10. 键盘快捷键：`key.js` → `keybindings.js`，把 `vm.*` 引用换成 store 导出的
    state/函数，在 `App.vue` 的 `onMounted` 里注册一次。
11. Auth/登出收尾：确认 `authenticated`（现在来自 `/api/status`）正确控制登出菜单
    项显隐，401 触发的整页刷新在代理下仍然生效。
12. 完整验证（见下）。

## 验证计划

当前前端没有自动化测试，验证方式是手动对照。建议：新旧前端同时指向同一个数据库
运行——旧前端用 `go run ./cmd/yarr` 直接跑 `backend/src/assets`（`127.0.0.1:7070`），
新前端 `npm run dev`（`localhost:5173`，代理到同一个后端）——两个浏览器窗口并排
逐项走查：

- 布局：三栏桌面布局，767.98px/991.98px 两个断点的响应式折叠。
- 侧栏：三个过滤器、kebab 菜单全部项（新建 feed、刷新、三套主题、8 档刷新间隔、
  排序切换、OPML 对比/导入/导出、快捷键帮助、登出）、文件夹展开/折叠持久化、
  可拖拽分隔条、feed 错误指示图标。
- 列表栏：防抖搜索、全部已读、无限滚动、已读/未读/星标指示、相对时间自刷新。
- 阅读栏：星标/已读切换、外观下拉（3 种字体+字号）、Read Here 可读性模式、外部
  链接打开、上一篇/下一篇（按钮+快捷键）、关闭、图片/音频/视频提取渲染。
- Feed 设置：重命名、改链接、移动到已有/新建文件夹（含改造后的 v-for+v-if 逻辑）、
  删除确认。
- 文件夹设置：重命名、删除。
- 新建 Feed 弹窗三种响应场景（成功/多选/无结果）、快捷键帮助表内容与 `key.js`
  绑定一致、OPML 对比结果表。
- OPML 导入导出。
- Auth：切换后端 `-auth-url`，确认登出菜单项显隐随 `authenticated` 变化，强制
  401 后确认整页刷新重定向。
- 键盘快捷键：全部绑定，包括输入框聚焦时不劫持按键、修饰键（Cmd/Ctrl/Alt）绕过。

不为此阶段引入新的自动化测试框架，除非用户后续要求——验收标准是手动行为对等，
与现状代码库的验证方式一致。

## 关键文件

- `backend/src/server/routes.go`（`handleStatus` 改动）
- `backend/src/assets/javascripts/app.js`（迁移源）
- `backend/src/assets/javascripts/api.js`（迁移源）
- `backend/src/assets/index.html`（迁移源，含需改造的 v-for+v-if）
- `backend/src/storage/settings.go`（settings 字段名确认）
- 新建：`frontend/package.json`、`frontend/vite.config.js`、
  `frontend/src/**`（见 §1 目录树）
