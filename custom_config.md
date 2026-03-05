## 自定义抓取逻辑

### 添加抓取配置

feed_config 表新增配置数据

url: 区分不同站点，添加时判断是否走自定义抓取

config: 配置rss解析逻辑

```json
{
  "title": "xxx",
  "link": "https://xxx.xxx.xxx/",
  "items": {
    "root": "", // 每个文章的root，根据本配置解析dom数组
    "title": "", // 解析标题
    "link": "", // 解析文章链接
    "linkHost": "", // link是相对地址时，linkHost为base url
    "pubDate": "", // 发布日期
    "pubDateFmt": "", // 发布日期格式，支持  "September 11, 2025":"January 2, 2006":"  "2026-02-23":"2006-1-2"
    "description": "" // 描述
  }
}
```

解析表达式

- 标签: a、p、div 之类
- css筛选: .classA、.classB
- 第几个元素: :eq(0)、:eq(1)

### 测试抓取接口

```bash
curl http://127.0.0.1:8080/htmlFeed\?url\=https://xxx.xxx.xxx/
```
