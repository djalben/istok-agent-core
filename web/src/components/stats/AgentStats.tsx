'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useAgentStats } from '@/lib/hooks/useAgentStats';

export function AgentStats() {
  const { stats, loading, error } = useAgentStats();

  if (loading) {
    return (
      <Card className="bg-black/40 backdrop-blur-xl border-white/20">
        <CardContent className="pt-6">
          <p className="text-zinc-400">Загрузка статистики...</p>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="bg-black/40 backdrop-blur-xl border-white/20">
        <CardContent className="pt-6">
          <p className="text-red-400">Ошибка: {error}</p>
        </CardContent>
      </Card>
    );
  }

  if (!stats) return null;

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'idle':
        return 'bg-green-500/20 text-green-400 border-green-500/30';
      case 'analyzing':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
      case 'coding':
        return 'bg-purple-500/20 text-purple-400 border-purple-500/30';
      case 'deploying':
        return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
      case 'error':
        return 'bg-red-500/20 text-red-400 border-red-500/30';
      default:
        return 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30';
    }
  };

  const getStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      idle: 'Готов',
      analyzing: 'Анализ',
      coding: 'Генерация',
      deploying: 'Деплой',
      error: 'Ошибка',
    };
    return labels[status] || status;
  };

  return (
    <Card className="bg-black/40 backdrop-blur-xl border-white/20 shadow-2xl">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-xl font-bold text-white">
            {stats.name}
          </CardTitle>
          <Badge className={getStatusColor(stats.status)}>
            {getStatusLabel(stats.status)}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1">
            <p className="text-xs text-zinc-400">Баланс токенов</p>
            <p className="text-2xl font-bold text-white">
              {stats.token_balance.toLocaleString()}
            </p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-zinc-400">Успешность</p>
            <p className="text-2xl font-bold text-white">
              {(stats.success_rate * 100).toFixed(0)}%
            </p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-zinc-400">Всего задач</p>
            <p className="text-lg font-semibold text-white">{stats.total_tasks}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-zinc-400">Узлов знаний</p>
            <p className="text-lg font-semibold text-white">{stats.knowledge_nodes}</p>
          </div>
        </div>
        
        {stats.learning_confidence > 0 && (
          <div className="pt-2 border-t border-white/10">
            <p className="text-xs text-zinc-400 mb-1">Уверенность обучения</p>
            <div className="w-full bg-white/10 rounded-full h-2">
              <div
                className="bg-gradient-to-r from-blue-500 to-purple-500 h-2 rounded-full transition-all"
                style={{ width: `${stats.learning_confidence * 100}%` }}
              />
            </div>
            <p className="text-xs text-zinc-400 mt-1 text-right">
              {(stats.learning_confidence * 100).toFixed(1)}%
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
