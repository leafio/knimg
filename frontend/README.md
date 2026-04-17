# KnImg 前端开发指南

## 📂 项目结构

```
frontend/
├── index.html              # 主HTML文件
├── start.sh                # 开发服务器启动脚本
├── REFACTORING.md          # 重构说明文档
├── css/                    # 样式文件目录
│   ├── base.css           # 基础样式
│   ├── layout.css         # 布局样式
│   ├── components.css     # 组件样式
│   └── responsive.css     # 响应式样式
└── js/                     # JavaScript 模块目录
    ├── app.js             # 主应用模块
    ├── api.js             # API 调用模块
    ├── utils.js           # 工具函数模块
    ├── filters.js         # 筛选功能模块
    ├── file-list.js       # 文件列表模块
    └── compress.js        # 压缩功能模块
```

## 🚀 快速开始

### 方法一: 使用启动脚本(推荐)

```bash
cd frontend
./start.sh
```

然后在浏览器中访问: `http://localhost:8080`

### 方法二: 手动启动 Python 服务器

```bash
cd frontend
python3 -m http.server 8080
```

### 方法三: 使用 npx

```bash
cd frontend
npx http-server -p 8080
```

### 方法四: 使用 Go 后端(完整功能)

```bash
cd ..
go run main.go
```

然后在浏览器中访问: `http://localhost:8080`

## ⚠️ 重要提示

由于使用了 **ES6 模块** (`import/export`),**不能直接双击打开 HTML 文件**,必须通过 HTTP 服务器访问,否则会遇到 CORS 错误。

## 🔧 开发工作流

### 修改 CSS

1. 确定要修改的样式类型:
   - 全局样式 → `css/base.css`
   - 布局结构 → `css/layout.css`
   - UI 组件 → `css/components.css`
   - 响应式 → `css/responsive.css`

2. 修改对应文件
3. 刷新浏览器查看效果

### 修改 JavaScript

1. 确定要修改的功能模块:
   - 状态管理/初始化 → `js/app.js`
   - API 调用 → `js/api.js`
   - 工具函数 → `js/utils.js`
   - 筛选逻辑 → `js/filters.js`
   - 文件列表 → `js/file-list.js`
   - 图片压缩 → `js/compress.js`

2. 修改对应模块
3. 刷新浏览器测试功能

### 添加新功能

1. 创建新的模块文件(如需要)
2. 在 `app.js` 中导入并集成
3. 更新 HTML 中的引用(如需要)

## 📝 代码规范

### JavaScript 模块化

```javascript
// 导出函数
export function myFunction() {
    // ...
}

// 导入其他模块
import { otherFunction } from './other-module.js';

// 默认导出(每个模块只能有一个)
export default class MyClass {
    // ...
}
```

### CSS 组织

- 使用 CSS 变量统一管理颜色和尺寸
- 按功能模块分组
- 添加清晰的注释分隔不同区域

## 🐛 调试技巧

### 浏览器开发者工具

1. **Console**: 查看日志和错误
2. **Network**: 监控 API 请求
3. **Elements**: 检查 DOM 和样式
4. **Sources**: 断点调试 JavaScript

### 常见问题

**问题**: 页面空白,控制台显示 CORS 错误  
**解决**: 确保通过 HTTP 服务器访问,不要直接打开 HTML 文件

**问题**: 样式不生效  
**解决**: 检查 CSS 文件路径是否正确,清除浏览器缓存

**问题**: JavaScript 模块加载失败  
**解决**: 检查 `import` 路径是否正确,确保使用 `.js` 扩展名

## 🔄 回滚到原始版本

如果遇到问题,可以回滚到原始的单文件版本:

```bash
cd frontend
cp index_backup.html index.html
```

## 📊 性能优化建议

### 当前状态
- ✅ 代码已模块化,便于维护
- ✅ 按需加载(浏览器自动处理模块依赖)
- ⚠️ 未压缩,文件体积较大

### 未来优化方向
1. **代码压缩**: 使用构建工具压缩 CSS 和 JS
2. **代码分割**: 将不常用的模块懒加载
3. **缓存策略**: 设置合理的 HTTP 缓存头
4. **CDN 加速**: 将静态资源托管到 CDN

## 🤝 协作开发

### Git 工作流

1. 从 `main` 分支创建功能分支
2. 在功能分支上开发
3. 完成后提交 PR/MR
4. 代码审查后合并

### 避免冲突

- 每个人负责不同的模块文件
- 修改前拉取最新代码
- 频繁提交小改动

## 📚 相关文档

- [REFACTORING.md](REFACTORING.md) - 详细的重构说明
- [../readme.md](../readme.md) - 项目总体说明

---

**最后更新**: 2026-04-17
