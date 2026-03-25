'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

export interface AIModel {
  id: string;
  name: string;
  provider: 'anthropic' | 'openai' | 'google';
  description: string;
  capabilities: string[];
  contextWindow: number;
  costPer1M: number;
  speed: 'fast' | 'medium' | 'slow';
  reasoning: boolean;
}

export const AI_MODELS: AIModel[] = [
  {
    id: 'claude-3-5-sonnet',
    name: 'Claude 3.5 Sonnet',
    provider: 'anthropic',
    description: 'Лучшая модель для кодинга и сложных задач',
    capabilities: ['Кодинг', 'Анализ', 'Рассуждения'],
    contextWindow: 200000,
    costPer1M: 3.0,
    speed: 'fast',
    reasoning: false,
  },
  {
    id: 'claude-4-6-sonnet',
    name: 'Claude 4.6 Sonnet',
    provider: 'anthropic',
    description: 'Новейшая модель с улучшенным пониманием',
    capabilities: ['Кодинг', 'Анализ', 'Креатив', 'Рассуждения'],
    contextWindow: 200000,
    costPer1M: 5.0,
    speed: 'medium',
    reasoning: true,
  },
  {
    id: 'gpt-4o',
    name: 'GPT-4o',
    provider: 'openai',
    description: 'Быстрая мультимодальная модель от OpenAI',
    capabilities: ['Кодинг', 'Анализ', 'Изображения'],
    contextWindow: 128000,
    costPer1M: 2.5,
    speed: 'fast',
    reasoning: false,
  },
  {
    id: 'gpt-4o-mini',
    name: 'GPT-4o Mini',
    provider: 'openai',
    description: 'Экономичная версия для простых задач',
    capabilities: ['Кодинг', 'Анализ'],
    contextWindow: 128000,
    costPer1M: 0.15,
    speed: 'fast',
    reasoning: false,
  },
  {
    id: 'gemini-2-flash',
    name: 'Gemini 2.0 Flash',
    provider: 'google',
    description: 'Сверхбыстрая модель от Google',
    capabilities: ['Кодинг', 'Анализ', 'Мультимодальность'],
    contextWindow: 1000000,
    costPer1M: 0.1,
    speed: 'fast',
    reasoning: false,
  },
];

interface ModelSelectorProps {
  selectedModel: string;
  onSelectModel: (modelId: string) => void;
}

export function ModelSelector({ selectedModel, onSelectModel }: ModelSelectorProps) {
  const [showDetails, setShowDetails] = useState(false);

  const currentModel = AI_MODELS.find((m) => m.id === selectedModel) || AI_MODELS[0];

  const getProviderColor = (provider: string) => {
    switch (provider) {
      case 'anthropic':
        return 'bg-orange-500/20 text-orange-400 border-orange-500/40';
      case 'openai':
        return 'bg-green-500/20 text-green-400 border-green-500/40';
      case 'google':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/40';
      default:
        return 'bg-zinc-500/20 text-zinc-400 border-zinc-500/40';
    }
  };

  const getSpeedColor = (speed: string) => {
    switch (speed) {
      case 'fast':
        return 'text-green-400';
      case 'medium':
        return 'text-yellow-400';
      case 'slow':
        return 'text-red-400';
      default:
        return 'text-zinc-400';
    }
  };

  return (
    <div className="space-y-4">
      {/* Current model display */}
      <Card className="glass-strong border-white/10">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm text-zinc-400">🧠 Активная модель</CardTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowDetails(!showDetails)}
              className="text-xs text-zinc-400 hover:text-white"
            >
              {showDetails ? 'Скрыть' : 'Выбрать'}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Badge className={getProviderColor(currentModel.provider)}>
                {currentModel.provider.toUpperCase()}
              </Badge>
              <div>
                <p className="text-white font-semibold">{currentModel.name}</p>
                <p className="text-xs text-zinc-500">{currentModel.description}</p>
              </div>
            </div>
            {currentModel.reasoning && (
              <Badge className="bg-purple-500/20 text-purple-400 border-purple-500/40">
                💭 Reasoning
              </Badge>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Model selection grid */}
      {showDetails && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          className="grid grid-cols-1 md:grid-cols-2 gap-4"
        >
          {AI_MODELS.map((model) => (
            <motion.div
              key={model.id}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
            >
              <Card
                className={`cursor-pointer transition-all ${
                  selectedModel === model.id
                    ? 'glass-strong border-indigo-500/50 shadow-glow-sm'
                    : 'glass border-white/10 hover:border-white/20'
                }`}
                onClick={() => onSelectModel(model.id)}
              >
                <CardContent className="p-4">
                  <div className="space-y-3">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-white font-semibold">{model.name}</p>
                        <p className="text-xs text-zinc-500 mt-1">{model.description}</p>
                      </div>
                      {selectedModel === model.id && (
                        <Badge className="bg-green-500/20 text-green-400 border-green-500/40">
                          ✓
                        </Badge>
                      )}
                    </div>

                    <div className="flex flex-wrap gap-1">
                      {model.capabilities.map((cap, i) => (
                        <Badge
                          key={i}
                          variant="outline"
                          className="text-xs border-white/10 text-zinc-400"
                        >
                          {cap}
                        </Badge>
                      ))}
                    </div>

                    <div className="grid grid-cols-3 gap-2 text-xs">
                      <div>
                        <p className="text-zinc-500">Скорость</p>
                        <p className={`font-semibold ${getSpeedColor(model.speed)}`}>
                          {model.speed === 'fast' ? '⚡ Быстро' : model.speed === 'medium' ? '⏱️ Средне' : '🐌 Медленно'}
                        </p>
                      </div>
                      <div>
                        <p className="text-zinc-500">Контекст</p>
                        <p className="text-white font-semibold">
                          {(model.contextWindow / 1000).toFixed(0)}K
                        </p>
                      </div>
                      <div>
                        <p className="text-zinc-500">Цена/1M</p>
                        <p className="text-white font-semibold">${model.costPer1M}</p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </motion.div>
      )}
    </div>
  );
}
