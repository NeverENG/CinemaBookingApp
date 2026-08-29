# LTerm Design System

> 前端原型设计视觉规范 | 仅在设计 UI/前端原型时生效

## Brand Identity

**产品定位**: 电影订票系统
**视觉基调**: 用户端胶片灰与影院红，管理端深灰控制台；简约、专业、现代、克制

---

## Color System

```css
:root {
    /* 用户端：胶片灰底 + 白色表面 + 影院红 */
    --user-bg: #f4f5f6;
    --user-surface: #ffffff;
    --user-surface-strong: #ffffff;
    --user-border: #e2e5e8;
    --user-text: #202124;
    --user-text-secondary: #6d7278;
    --user-text-muted: #9ca2a8;
    --user-accent: #df5960;
    --user-accent-hover: #c94850;
    --user-accent-soft: #fff0f1;

    /* 管理端：深灰背景 + 灰色表面 + 克制金 */
    --admin-bg: #0f1216;
    --admin-surface: #151a20;
    --admin-surface-raised: #1b222a;
    --admin-border: rgba(255, 255, 255, 0.09);
    --admin-text: #f3f5f7;
    --admin-text-secondary: #9da8b4;
    --admin-text-muted: #64707d;
    --admin-gold: #d4af65;

    /* 兼容通用组件命名 */
    --primary: var(--user-accent);
    --primary-hover: var(--user-accent-hover);
    --bg-light: var(--user-surface-strong);
    --bg-input: #f1f3f4;
    --text-primary: var(--user-text);
    --text-secondary: var(--user-text-secondary);
    --text-light: #ffffff;
    --text-muted: var(--user-text-muted);
    --border: var(--user-border);
    --border-focus: var(--user-accent);
}
```

## 使用原则

1. 用户端默认使用胶片灰背景和白色内容表面，不使用古典米色大面积铺陈。
2. 影院红只用于主操作、当前选中态、评分强调和关键 CTA；不作为大面积背景。
3. 首页以当前影院热映为主线，使用一个电影大屏 Hero 承担氛围，不继续堆叠无关 Banner 和推荐专区。
4. 管理端可以使用深灰工作台，但信息层级优先于装饰，图表和操作状态仍使用克制的金色与语义色。
