'use client';

import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useAgentStats } from '@/lib/hooks/useAgentStats';

export function AgentStats() {
  const { stats, loading, error } = useAgentStats();

  if (loading) {
    return (
      <Card className="glass-strong border-white/10">
        <CardContent className="pt-6">
          <div className="flex items-center gap-3">
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
              className="text-2xl"
            >
              ⚡
            </motion.div>
            <p className="text-zinc-400 font-medium">Подключение к агенту...</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="glass border-red-500/20">
        <CardContent className="pt-6">
          <div className="flex items-center gap-3">
            <span className="text-2xl">⚠️</span>
            <div>
              <p className="text-red-400 font-medium">Backend недоступен</p>
              <p className="text-xs text-zinc-500 mt-1">Запустите сервер на порту 8080</p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!stats) return null;

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'idle':
        return 'bg-green-500/20 text-green-400 border-green-500/40';
      case 'analyzing':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/40';
      case 'coding':
        return 'bg-purple-500/20 text-purple-400 border-purple-500/40';
      case 'deploying':
        return 'bg-orange-500/20 text-orange-400 border-orange-500/40';
      case 'error':
        return 'bg-red-500/20 text-red-400 border-red-500/40';
      default:
        return 'bg-zinc-500/20 text-zinc-400 border-zinc-500/40';
    }
  };

  const getStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      idle: '✓ Готов к работе',
      analyzing: '🔍 Анализирует',
      coding: '⚡ Генерирует',
      deploying: '🚀 Деплоит',
      error: '❌ Ошибка',
    };
    return labels[status] || status;
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'idle':
        return '🟢';
      case 'analyzing':
        return '🔵';
      case 'coding':
        return '🟣';
      case 'deploying':
        return '🟠';
      case 'error':
        return '🔴';
      default:
        return '⚪';
    }
  };

  return (
    <Card className="glass-strong border-white/10 shadow-glow-sm overflow-hidden">
      <CardHeader className="border-b border-white/10 bg-gradient-to-r from-indigo-500/10 to-violet-500/10">
        <div className="flex items-center justify-between">
          <CardTitle className="text-2xl font-bold text-gradient flex items-center gap-3">
            <motion.span
              animate={{ scale: [1, 1.1, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
            >
              {getStatusIcon(stats.status)}
            </motion.span>
            {stats.name}
          </CardTitle>
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ type: "spring", stiffness: 200 }}
          >
            <Badge className={`${getStatusColor(stats.status)} px-4 py-1.5 font-semibold`}>
              {getStatusLabel(stats.status)}
            </Badge>
          </motion.div>
        </div>
      </CardHeader>
      <CardContent className="p-6">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="glass border-white/10 rounded-xl p-4"
          >
            <p className="text-xs text-zinc-400 mb-2 font-medium">💰 Баланс токенов</p>
            <p className="text-3xl font-bold text-gradient">
              {stats.token_balance.toLocaleString()}
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="glass border-white/10 rounded-xl p-4"
          >
            <p className="text-xs text-zinc-400 mb-2 font-medium">📊 Успешность</p>
            <p className="text-3xl font-bold text-gradient">
              {(stats.success_rate * 100).toFixed(0)}%
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="glass border-white/10 rounded-xl p-4"
          >
            <p className="text-xs text-zinc-400 mb-2 font-medium">✅ Всего задач</p>
            <p className="text-3xl font-bold text-white">{stats.total_tasks}</p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
            className="glass border-white/10 rounded-xl p-4"
          >
            <p className="text-xs text-zinc-400 mb-2 font-medium">🧠 Узлов знаний</p>
            <p className="text-3xl font-bold text-white">{stats.knowledge_nodes}</p>
          </motion.div>
        </div>
        
        {stats.learning_confidence > 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
            className="mt-6 glass border-white/10 rounded-xl p-4"
          >
            <div className="flex items-center justify-between mb-3">
              <p className="text-sm text-zinc-300 font-medium">🎯 Уверенность обучения</p>
              <p className="text-sm font-bold text-gradient">
                {(stats.learning_confidence * 100).toFixed(1)}%
              </p>
            </div>
            <div className="relative w-full bg-white/5 rounded-full h-3 overflow-hidden">
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: `${stats.learning_confidence * 100}%` }}
                transition={{ duration: 1, ease: "easeOut", delay: 0.6 }}
                className="gradient-indigo-violet h-3 rounded-full shadow-glow-sm"
              />
            </div>
          </motion.div>
        )}
      </CardContent>
    </Card>
  );
}
