# EtherMonitor Pro - UI/UX 设计方案 (v3 最终版)

## 1. 设计哲学："现代金融科技 (Modern Fintech)"
根据您的要求，我们采用了 **"支付网关 (Payment Gateway)"** 风格的极简主义设计。这种风格强调清晰、专业和信任感（类似 Stripe 或 Plaid 的开发者文档风格）。

### 核心视觉特征
- **极简主义 (Minimal Design)**：
    - 大量的留白 (Whitespace)。
    - 干净的排版。
    - 去除多余的装饰，专注于信息本身的层级。
- **专业配色 (Professional Colors)**：
    - **背景**：`#F8FAFC` (Slate 50) - 柔和的灰白背景。
    - **卡片**：`#FFFFFF` (White) - 纯白卡片，带有极淡的边框和阴影。
    - **主色**：`#2563EB` (Royal Blue) - 经典的科技蓝，用于按钮和链接。
    - **文字**：深灰 (`#0F172A`) 为主，浅灰 (`#64748B`) 为辅。
- **开发者友好 (Developer First)**：
    - **代码预览 (Code Preview)**：将实时数据流展示为漂亮的 JSON 代码块（深色主题），致敬开发者文档体验。
    - **安全徽章**：展示 "AES-256" 或 "Audited" 风格的徽章，增加专业度。

## 2. 布局结构 (多页签)

### 导航栏 (Sidebar/TopNav)
- 干净的白色导航栏。
- Logo 使用深蓝色字体，搭配简单的几何图形。
- 菜单项：`Dashboard`, `Liquidations`, `Transfers`, `API Docs`。

### 页面规划

#### 1. Dashboard (概览)
- **欢迎区**：类似 Landing Page 的 Hero Section。
    - "System Operational. Monitoring 185 contracts."
    - 包含两个主要 CTA 按钮或状态指示器。
- **KPI 网格**：
    - 简单的白色卡片，大号数字，带有微小的趋势图（Sparkline）。
- **代码集成预览区**：
    - 展示 "Live API Stream"。
    - 左侧显示说明，右侧是一个黑色背景的代码窗口，实时滚动最新的 JSON 数据。

#### 2. Liquidation Radar (清算)
- **表格风格**：
    - 极简表格，只有横向分隔线。
    - 利润使用绿色胶囊样式 (Badge)。
    - 地址哈希截断并单色显示。

#### 3. Whale Watch (转账)
- **流式布局**：
    - 每一行转账记录像是一条日志。
    - 使用图标区分 "Inbound" (蓝色) 和 "Outbound" (橙色/灰色)。

## 3. 实现细节
- **CSS 变量**:
  ```css
  :root {
    --bg-page: #F8FAFC;
    --bg-card: #FFFFFF;
    --text-primary: #0F172A;
    --text-secondary: #64748B;
    --primary: #2563EB;
    --border: #E2E8F0;
  }
  ```
- **字体**: `Inter` (UI) + `Fira Code` (代码/数据)。

# 确认
我们将基于这个 **"清爽、明亮、专业"** 的风格重写前端代码。
这与之前的 "黑暗/赛博朋克" 风格截然不同，但更适合展示复杂数据和 API 集成场景。
