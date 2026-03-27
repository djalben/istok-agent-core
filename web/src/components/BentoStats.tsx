import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Activity, Cpu, Clock, Globe, TrendingUp } from "lucide-react";
import { Card } from "@/components/ui/card";
import { api, type AgentStats } from "@/lib/api";

const BentoStats = () => {
  const [stats, setStats] = useState<AgentStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await api.getStats();
        setStats(data);
      } catch (error) {
        console.error("Failed to fetch stats:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
    // Обновляем статистику каждые 30 секунд
    const interval = setInterval(fetchStats, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 p-6">
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} className="glass-subtle p-6 animate-pulse">
            <div className="h-20 bg-muted/20 rounded" />
          </Card>
        ))}
      </div>
    );
  }

  if (!stats) {
    return null;
  }

  const statCards = [
    {
      icon: Cpu,
      label: "Модель",
      value: stats.model,
      sublabel: stats.modelVersion,
      color: "text-indigo-400",
      bgColor: "bg-indigo-500/10",
      borderColor: "border-indigo-500/20",
    },
    {
      icon: Clock,
      label: "Латентность",
      value: `${stats.responseTimeMs}ms`,
      sublabel: "среднее",
      color: "text-emerald-400",
      bgColor: "bg-emerald-500/10",
      borderColor: "border-emerald-500/20",
    },
    {
      icon: Globe,
      label: "Узлы",
      value: stats.crawlerNodesFound.toLocaleString("ru-RU"),
      sublabel: "краулер",
      color: "text-amber-400",
      bgColor: "bg-amber-500/10",
      borderColor: "border-amber-500/20",
    },
    {
      icon: Activity,
      label: "Файлы",
      value: String(stats.generatedFilesCount),
      sublabel: "создано",
      color: "text-violet-400",
      bgColor: "bg-violet-500/10",
      borderColor: "border-violet-500/20",
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 p-6">
      {statCards.map((stat, i) => (
        <motion.div
          key={stat.label}
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{
            duration: 0.4,
            delay: i * 0.1,
            ease: [0.25, 0.1, 0.25, 1],
          }}
          whileHover={{ scale: 1.02 }}
        >
          <Card className="glass-subtle p-6 relative overflow-hidden cursor-default">
            <div className="relative z-10">
              <div className="flex items-center justify-between mb-3">
                <div
                  className={`w-10 h-10 rounded-lg ${stat.bgColor} border ${stat.borderColor} flex items-center justify-center`}
                >
                  <stat.icon className={`w-5 h-5 ${stat.color}`} />
                </div>
                <span className="text-xs uppercase tracking-wider text-muted-foreground">
                  {stat.sublabel}
                </span>
              </div>
              <div className="text-2xl font-bold tracking-tight text-foreground leading-none mb-1">
                {stat.value}
              </div>
              <div className="text-sm text-muted-foreground font-medium">
                {stat.label}
              </div>
            </div>
            {/* Gradient overlay */}
            <div className="absolute inset-0 bg-gradient-to-br from-transparent via-transparent to-primary/5 pointer-events-none" />
          </Card>
        </motion.div>
      ))}

      {/* Performance chart card */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.4, delay: 0.4 }}
        className="md:col-span-2 lg:col-span-4"
      >
        <Card className="glass-subtle p-6 relative overflow-hidden">
          <div className="relative z-10">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-indigo-400" />
                <span className="text-sm text-muted-foreground font-medium">
                  Производительность
                </span>
              </div>
              <div className="flex items-center gap-4 text-xs text-muted-foreground">
                <span>Токены: {stats.tokensUsed.toLocaleString("ru-RU")}</span>
                <span>Стоимость: {stats.costRub} ₽</span>
                <span className="text-green-400 font-medium">Статус: {stats.status}</span>
              </div>
            </div>
            <div className="h-2 bg-muted/20 rounded-full overflow-hidden">
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: "85%" }}
                transition={{ duration: 1, delay: 0.5 }}
                className="h-full bg-gradient-to-r from-indigo-500 to-violet-500"
              />
            </div>
          </div>
        </Card>
      </motion.div>
    </div>
  );
};

export default BentoStats;
