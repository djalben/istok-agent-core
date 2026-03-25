'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useAgentGenerate } from '@/lib/hooks/useAgentGenerate';

interface Message {
  type: 'user' | 'agent' | 'error';
  content: string;
  timestamp: Date;
}

interface AgentTerminalProps {
  onCodeGenerated?: (code: string) => void;
}

export function AgentTerminal({ onCodeGenerated }: AgentTerminalProps) {
  const [messages, setMessages] = useState<Message[]>([
    {
      type: 'agent',
      content: '👋 Привет! Я Исток - автономный AI агент. Опишите проект, который хотите создать.',
      timestamp: new Date(),
    },
  ]);
  const [specification, setSpecification] = useState('');
  const [analyzeUrl, setAnalyzeUrl] = useState('');
  const { generate, loading } = useAgentGenerate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!specification.trim()) return;

    // Добавляем сообщение пользователя
    const userMessage: Message = {
      type: 'user',
      content: specification,
      timestamp: new Date(),
    };
    setMessages((prev) => [...prev, userMessage]);

    // Добавляем информацию об анализе URL если указан
    if (analyzeUrl.trim()) {
      setMessages((prev) => [
        ...prev,
        {
          type: 'agent',
          content: `🕷️ Анализирую сайт: ${analyzeUrl}...`,
          timestamp: new Date(),
        },
      ]);
    }

    try {
      const response = await generate({
        specification,
        language: 'JavaScript',
        framework: 'React',
        analyze_url: analyzeUrl.trim() || undefined,
      });

      // Добавляем ответ агента
      setMessages((prev) => [
        ...prev,
        {
          type: 'agent',
          content: `✅ Проект сгенерирован!\n\n${response.explanation}\n\nИспользовано токенов: ${response.tokens_used}\nМодель: ${response.model}`,
          timestamp: new Date(),
        },
      ]);

      // Передаем код в preview
      if (onCodeGenerated) {
        onCodeGenerated(response.code);
      }

      // Очищаем поля
      setSpecification('');
      setAnalyzeUrl('');
    } catch (error) {
      setMessages((prev) => [
        ...prev,
        {
          type: 'error',
          content: `❌ Ошибка: ${error instanceof Error ? error.message : 'Неизвестная ошибка'}`,
          timestamp: new Date(),
        },
      ]);
    }
  };

  return (
    <Card className="h-full bg-black/40 backdrop-blur-xl border-white/20 shadow-2xl flex flex-col">
      <CardHeader>
        <CardTitle className="text-xl font-bold text-white flex items-center gap-2">
          <span className="text-2xl">🤖</span>
          Терминал Агента
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col gap-4 p-4">
        <ScrollArea className="flex-1 pr-4">
          <div className="space-y-4">
            {messages.map((message, index) => (
              <div
                key={index}
                className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div
                  className={`max-w-[80%] rounded-lg p-3 ${
                    message.type === 'user'
                      ? 'bg-blue-600/30 border border-blue-500/30 text-white'
                      : message.type === 'error'
                      ? 'bg-red-600/30 border border-red-500/30 text-red-200'
                      : 'bg-zinc-800/50 border border-zinc-700/30 text-zinc-100'
                  }`}
                >
                  <p className="text-sm whitespace-pre-wrap">{message.content}</p>
                  <p className="text-xs opacity-50 mt-1">
                    {message.timestamp.toLocaleTimeString()}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </ScrollArea>

        <form onSubmit={handleSubmit} className="space-y-3 border-t border-white/10 pt-4">
          <div>
            <label className="text-xs text-zinc-400 mb-1 block">
              URL для анализа (опционально)
            </label>
            <Input
              type="url"
              placeholder="https://example.com"
              value={analyzeUrl}
              onChange={(e) => setAnalyzeUrl(e.target.value)}
              disabled={loading}
              className="bg-black/30 border-white/20 text-white placeholder:text-zinc-500"
            />
          </div>
          <div>
            <label className="text-xs text-zinc-400 mb-1 block">
              Опишите проект
            </label>
            <Input
              placeholder="Создай landing page с формой подписки..."
              value={specification}
              onChange={(e) => setSpecification(e.target.value)}
              disabled={loading}
              className="bg-black/30 border-white/20 text-white placeholder:text-zinc-500"
            />
          </div>
          <Button
            type="submit"
            disabled={loading || !specification.trim()}
            className="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-semibold"
          >
            {loading ? '⏳ Генерация...' : '🚀 Сгенерировать'}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
