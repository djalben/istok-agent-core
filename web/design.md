# ИСТОК АГЕНТ — Design System

## Theme
- **Mode**: Dark mode only
- **Aesthetic**: Hyper-modern, minimal, "expensive" — Apple + Vercel + Linear
- **Layout**: Bento Grid, Glassmorphism panels

## Colors
- **Background**: #09090b (zinc-950)
- **Surface**: rgba(255,255,255,0.03) with backdrop-blur-xl
- **Border**: rgba(255,255,255,0.06) — thin, translucent
- **Primary Gradient**: Indigo (#6366f1) → Violet (#8b5cf6)
- **Text Primary**: #fafafa
- **Text Muted**: #71717a
- **Accent Glow**: #6366f1 with box-shadow glow

## Typography
- **Font**: Geist Sans (loaded from CDN), fallback Inter
- **Kerning**: tracking-wide on labels, tracking-tight on headings
- **Hierarchy**: 
  - Display: 2xl-3xl, font-bold, tracking-tight
  - Labels: xs-sm, uppercase, tracking-widest, text-muted
  - Body: sm-base, normal weight

## Glassmorphism
- `background: rgba(255,255,255,0.03)`
- `backdrop-filter: blur(24px)`
- `border: 1px solid rgba(255,255,255,0.06)`
- `border-radius: 16px`

## Motion (Framer Motion)
- Staggered entry: delay 0.05-0.1s per child
- Panel entry: y:20 → y:0, opacity 0→1
- Buttons: scale 0.97 on press, 1.02 on hover
- Pulse effects: infinite gentle opacity animation on glowing elements

## Spacing
- Generous padding: p-5 to p-6 on panels
- Gap: gap-3 to gap-4 in bento grid
- Sidebar: 64px collapsed, 240px expanded
