import { useState } from 'react';
import { apiClient, GenerateRequest, GenerateResponse } from '@/lib/api/client';

export function useAgentGenerate() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<GenerateResponse | null>(null);

  const generate = async (request: GenerateRequest) => {
    setLoading(true);
    setError(null);
    setData(null);

    try {
      const response = await apiClient.generateProject(request);
      setData(response);
      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Ошибка генерации';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return {
    generate,
    loading,
    error,
    data,
  };
}
