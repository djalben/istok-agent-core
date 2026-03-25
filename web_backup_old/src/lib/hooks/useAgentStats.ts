import { useState, useEffect } from 'react';
import { apiClient, AgentStats } from '@/lib/api/client';

export function useAgentStats(pollInterval: number = 5000) {
  const [stats, setStats] = useState<AgentStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = async () => {
    try {
      const data = await apiClient.getStats();
      setStats(data);
      setError(null);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Ошибка загрузки статистики';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();

    const interval = setInterval(fetchStats, pollInterval);

    return () => clearInterval(interval);
  }, [pollInterval]);

  return {
    stats,
    loading,
    error,
    refetch: fetchStats,
  };
}
