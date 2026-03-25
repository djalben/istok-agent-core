import { motion } from "framer-motion";
import { Cpu, Clock, Globe, FileCode, Activity, TrendingUp } from "lucide-react";
import { MOCK_STATS } from "@lib/api";

const stats = [
  {
    icon: Cpu,
    label: "Модель",
    value: "Claude 4.6",
    sublabel: "thinking",
    color: "text-indigo-400",
    bgColor: "bg-indigo-500/10",
    borderColor: "border-indigo-500/20",
  },
  {
    icon: Clock,
    label: "Латентность",
    value: `${MOCK_STATS.responseTimeMs}ms`,
    sublabel: "среднее",
    color: "text-emerald-400",
    bgColor: "bg-emerald-500/10",
    borderColor: "border-emerald-500/20",
  },
  {
    icon: Globe,
    label: "Узлы",
    value: MOCK_STATS.crawlerNodesFound.toLocaleString("ru-RU"),
    sublabel: "краулер",
    color: "text-amber-400",
    bgColor: "bg-amber-500/10",
    borderColor: "border-amber-500/20",
  },
  {
    icon: FileCode,
    label: "Файлы",
    value: String(MOCK_STATS.generatedFilesCount),
    sublabel: "создано",
    color: "text-violet-400",
    bgColor: "bg-violet-500/10",
    borderColor: "border-violet-500/20",
  },
];

function MiniChart() {
  const bars = [35, 55, 42, 68, 45, 72, 58, 82, 65, 90, 75, 95];
  return (
    <div className="flex items-end gap-[3px] h-10">
      {bars.map((h, i) => (
        <motion.div
          key={i}
          initial={{ height: 0 }}
          animate={{ height: `${h}%` }}
          transition={{ duration: 0.5, delay: 0.6 + i * 0.04, ease: "easeOut" }}
          className="w-[4px] rounded-full bg-gradient-to-t from-indigo-500/30 to-indigo-500/60"
        />
      ))}
    </div>
  );
}

export function ProjectStats() {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.3, ease: [0.25, 0.1, 0.25, 1] }}
      className="space-y-3"
    >
      {/* Stats Header */}
      <div className="flex items-center gap-2 px-1">
        <Activity className="w-3.5 h-3.5 text-indigo-400" />
        <span className="text-[11px] uppercase tracking-[0.15em] text-zinc-500 font-medium">
          Статистика Проекта
        </span>
      </div>

      {/* Stat Cards */}
      <div className="grid grid-cols-2 gap-2.5">
        {stats.map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{
              duration: 0.4,
              delay: 0.35 + i * 0.07,
              ease: [0.25, 0.1, 0.25, 1],
            }}
            whileHover={{ scale: 1.02 }}
            className="glass rounded-xl p-3.5 noise relative overflow-hidden cursor-default"
          >
            <div className="relative z-10">
              <div className="flex items-center justify-between mb-2.5">
                <div
                  className={`w-7 h-7 rounded-lg ${stat.bgColor} border ${stat.borderColor} flex items-center justify-center`}
                >
                  <stat.icon className={`w-3.5 h-3.5 ${stat.color}`} />
                </div>
                <span className="text-[9px] uppercase tracking-[0.1em] text-zinc-600">
                  {stat.sublabel}
                </span>
              </div>
              <div className="text-lg font-bold tracking-tight text-white leading-none mb-0.5">
                {stat.value}
              </div>
              <div className="text-[11px] text-zinc-500 font-medium">
                {stat.label}
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Performance chart */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.4, delay: 0.65 }}
        className="glass rounded-xl p-4 noise relative overflow-hidden"
      >
        <div className="relative z-10">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <TrendingUp className="w-3.5 h-3.5 text-indigo-400" />
              <span className="text-[11px] text-zinc-500 font-medium">
                Производительность
              </span>
            </div>
            <span className="text-[10px] text-green-400 font-medium">+23%</span>
          </div>
          <MiniChart />
        </div>
      </motion.div>
    </motion.div>
  );
}
