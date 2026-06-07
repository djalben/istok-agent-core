package application

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Static Component RAG (Knowledge Base)
//  Перфектные референс-шаблоны премиальных UI-примитивов уровня v0/shadcn.
//  Инъектируется в системный промпт Кодера, чтобы он НЕ переизобретал базовые
//  компоненты, а использовал точные структуры и Tailwind-классы.
//  ВАЖНО: бэктики в примерах запрещены — Go raw-string не допускает '`'.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ComponentKnowledgeBase — статическая база знаний компонентов для Кодера.
const ComponentKnowledgeBase = `COMPONENT KNOWLEDGE BASE: You MUST use these exact HTML structures and Tailwind classes when building similar UI elements. Do not reinvent them.

### 1. Premium Card (glassmorphism + hover scale)
import { type ReactNode } from "react";

interface PremiumCardProps {
  children: ReactNode;
  className?: string;
}

export function PremiumCard({ children, className = "" }: PremiumCardProps) {
  return (
    <div
      data-component-name="PremiumCard"
      className={"group rounded-2xl border border-white/10 bg-zinc-900/50 p-6 shadow-2xl shadow-black/40 backdrop-blur-md transition-all duration-300 ease-in-out hover:scale-[1.02] hover:border-white/20 " + className}
    >
      {children}
    </div>
  );
}

### 2. Modern Button (lucide icon slot + ring focus + active transform)
import { type ButtonHTMLAttributes, type ReactNode } from "react";

interface ModernButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: ReactNode;
  children: ReactNode;
}

export function ModernButton({ icon, children, className = "", ...props }: ModernButtonProps) {
  return (
    <button
      data-component-name="ModernButton"
      className={"inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-500 px-4 py-2.5 text-sm font-medium tracking-tight text-white shadow-lg shadow-emerald-500/20 transition-all duration-300 ease-in-out hover:bg-emerald-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-400 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50 " + className}
      {...props}
    >
      {icon ? <span className="h-4 w-4 shrink-0">{icon}</span> : null}
      {children}
    </button>
  );
}

### 3. Dashboard Metric Widget (title + value + trend indicator)
import { ArrowUpRight, ArrowDownRight } from "lucide-react";

interface MetricWidgetProps {
  title: string;
  value: string;
  trend: number; // percentage, e.g. 12.5 or -3.2
}

export function MetricWidget({ title, value, trend }: MetricWidgetProps) {
  const isUp = trend >= 0;
  return (
    <div
      data-component-name="MetricWidget"
      className="rounded-2xl border border-white/10 bg-zinc-900/50 p-5 shadow-2xl shadow-black/40 backdrop-blur-md transition-all duration-300 hover:border-white/20"
    >
      <p className="text-sm font-medium tracking-tight text-zinc-400">{title}</p>
      <div className="mt-2 flex items-end justify-between gap-3">
        <span className="text-3xl font-semibold tracking-tight text-zinc-100">{value}</span>
        <span
          className={
            "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium " +
            (isUp ? "bg-emerald-500/10 text-emerald-400" : "bg-red-500/10 text-red-400")
          }
        >
          {isUp ? <ArrowUpRight className="h-3.5 w-3.5" /> : <ArrowDownRight className="h-3.5 w-3.5" />}
          {isUp ? "+" : ""}{trend}%
        </span>
      </div>
    </div>
  );
}

`
