import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  LayoutDashboard,
  Bot,
  FolderKanban,
  Settings,
  Zap,
  ChevronRight,
  Globe,
  FileCode2,
  Database,
  Shield,
} from "lucide-react";
import { api, type TokenBalance } from "@lib/api";

const NAV_ITEMS = [
  { icon: LayoutDashboard, label: "Дашборд", active: true },
  { icon: Bot, label: "Агенты" },
  { icon: FolderKanban, label: "Проекты" },
  { icon: Globe, label: "Деплой" },
  { icon: FileCode2, label: "Шаблоны" },
  { icon: Database, label: "Хранилище" },
  { icon: Shield, label: "Безопасность" },
  { icon: Settings, label: "Настройки" },
];

export function Sidebar() {
  const [expanded, setExpanded] = useState(false);
  const [balance, setBalance] = useState<TokenBalance | null>(null);

  useEffect(() => {
    // Загружаем баланс токенов
    api.getBalance().then(setBalance).catch(console.error);
  }, []);

  const balanceRub = balance?.currentRub ?? 65000;
  const totalRub = balance?.totalRub ?? 100000;
  const percentage = Math.round((balanceRub / totalRub) * 100);

  return (
    <motion.aside
      className="fixed left-0 top-0 h-full z-50 flex flex-col glass border-r border-white/[0.06]"
      initial={false}
      animate={{ width: expanded ? 220 : 64 }}
      transition={{ duration: 0.25, ease: [0.25, 0.1, 0.25, 1] }}
      onMouseEnter={() => setExpanded(true)}
      onMouseLeave={() => setExpanded(false)}
    >
      {/* Logo */}
      <div className="flex items-center h-14 px-4 border-b border-white/[0.06]">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-violet-500 flex items-center justify-center flex-shrink-0">
          <Zap className="w-4 h-4 text-white" />
        </div>
        <AnimatePresence>
          {expanded && (
            <motion.span
              initial={{ opacity: 0, x: -8 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -8 }}
              transition={{ duration: 0.15 }}
              className="ml-3 text-sm font-semibold tracking-tight whitespace-nowrap"
            >
              ИСТОК АГЕНТ
            </motion.span>
          )}
        </AnimatePresence>
      </div>

      {/* Nav Items */}
      <nav className="flex-1 py-3 px-2 space-y-0.5">
        {NAV_ITEMS.map((item) => (
          <motion.button
            key={item.label}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.97 }}
            className={`w-full flex items-center h-10 px-2.5 rounded-lg transition-colors duration-150 group ${
              item.active
                ? "bg-white/[0.06] text-white"
                : "text-zinc-500 hover:text-zinc-300 hover:bg-white/[0.03]"
            }`}
          >
            <item.icon
              className={`w-[18px] h-[18px] flex-shrink-0 ${
                item.active ? "text-indigo-400" : "group-hover:text-zinc-300"
              }`}
            />
            <AnimatePresence>
              {expanded && (
                <motion.span
                  initial={{ opacity: 0, x: -4 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -4 }}
                  transition={{ duration: 0.12 }}
                  className="ml-3 text-[13px] font-medium whitespace-nowrap"
                >
                  {item.label}
                </motion.span>
              )}
            </AnimatePresence>
            <AnimatePresence>
              {expanded && item.active && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="ml-auto"
                >
                  <ChevronRight className="w-3.5 h-3.5 text-indigo-400" />
                </motion.div>
              )}
            </AnimatePresence>
          </motion.button>
        ))}
      </nav>

      {/* Token Balance Widget */}
      <div className="p-2.5 border-t border-white/[0.06]">
        <div className="px-2.5 py-3">
          <AnimatePresence>
            {expanded && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.15 }}
              >
                <div className="flex items-center justify-between mb-2.5">
                  <span className="text-[10px] uppercase tracking-[0.15em] text-zinc-500 font-medium">
                    Текущий Баланс
                  </span>
                  <span className="text-[11px] font-semibold text-indigo-400">
                    {balanceRub.toLocaleString("ru-RU")} ₽
                  </span>
                </div>
                <div className="h-1.5 bg-white/[0.04] rounded-full overflow-hidden">
                  <motion.div
                    className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 pulse-glow"
                    initial={{ width: 0 }}
                    animate={{ width: `${percentage}%` }}
                    transition={{ duration: 1.2, ease: "easeOut", delay: 0.3 }}
                  />
                </div>
              </motion.div>
            )}
          </AnimatePresence>
          {!expanded && (
            <div className="flex justify-center">
              <div className="w-2 h-2 rounded-full bg-indigo-500 pulse-glow" />
            </div>
          )}
        </div>
      </div>
    </motion.aside>
  );
}
