'use client';

import { useState, useRef, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useAgentGenerate } from '@/lib/hooks/useAgentGenerate';

interface Message {
  type: 'user' | 'agent' | 'error' | 'system';
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
      content: '👋 Привет! Я Исток - автономный AI агент нового поколения. Опишите проект, который хотите создать, и я воплощу его в жизнь.',
      timestamp: new Date(),
    },
  ]);
  const [specification, setSpecification] = useState('');
  const [analyzeUrl, setAnalyzeUrl] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const { generate, loading, error: apiError } = useAgentGenerate();
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, isTyping]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!specification.trim()) return;

    const userMessage: Message = {
      type: 'user',
      content: specification,
      timestamp: new Date(),
    };
    setMessages((prev) => [...prev, userMessage]);

    if (analyzeUrl.trim()) {
      setMessages((prev) => [
        ...prev,
        {
          type: 'system',
          content: `🕷️ Анализирую конкурента: ${analyzeUrl}`,
          timestamp: new Date(),
        },
      ]);
    }

    setIsTyping(true);

    try {
      const response = await generate({
        specification,
        language: 'JavaScript',
        framework: 'React',
        analyze_url: analyzeUrl.trim() || undefined,
      });

      setIsTyping(false);

      setMessages((prev) => [
        ...prev,
        {
          type: 'agent',
          content: `✨ Проект успешно сгенерирован!\n\n${response.explanation}\n\n📊 Использовано токенов: ${response.tokens_used}\n🤖 Модель: ${response.model}`,
          timestamp: new Date(),
        },
      ]);

      if (onCodeGenerated) {
        onCodeGenerated(response.code);
      }

      setSpecification('');
      setAnalyzeUrl('');
    } catch (error) {
      setIsTyping(false);
      
      const errorMessage = error instanceof Error ? error.message : 'Неизвестная ошибка';
      
      if (errorMessage.includes('fetch') || errorMessage.includes('Failed to fetch')) {
        setMessages((prev) => [
          ...prev,
          {
            type: 'system',
            content: '🔌 Агент подключается к мозгу... Убедитесь, что backend запущен на порту 8080.',
            timestamp: new Date(),
          },
        ]);
      } else {
        setMessages((prev) => [
          ...prev,
          {
            type: 'error',
            content: `⚠️ ${errorMessage}`,
            timestamp: new Date(),
          },
        ]);
      }
    }
  };

  return (
    <Card className="h-full glass-strong shadow-glow-sm border-white/10 flex flex-col overflow-hidden">
      <CardHeader className="border-b border-white/10 bg-gradient-to-r from-indigo-500/10 to-violet-500/10">
        <CardTitle className="text-xl font-bold text-white flex items-center gap-3">
          <motion.span
            className="text-3xl"
            animate={{ scale: [1, 1.1, 1] }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            💬
          </motion.span>
          <span className="text-gradient">Терминал Агента</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col gap-4 p-6">
        <ScrollArea className="flex-1 pr-4" ref={scrollRef}>
          <div className="space-y-4">
            <AnimatePresence>
              {messages.map((message, index) => (
                <motion.div
                  key={index}
                  initial={{ opacity: 0, y: 20, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  transition={{ duration: 0.3 }}
                  className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-[85%] rounded-2xl px-4 py-3 ${
                      message.type === 'user'
                        ? 'bg-gradient-to-br from-indigo-600 to-violet-600 text-white shadow-glow-sm'
                        : message.type === 'error'
                        ? 'glass border-red-500/30 text-red-200'
                        : message.type === 'system'
                        ? 'glass border-blue-500/30 text-blue-200'
                        : 'glass border-white/20 text-zinc-100'
                    }`}
                  >
                    <p className="text-sm leading-relaxed whitespace-pre-wrap font-medium">
                      {message.content}
                    </p>
                    <p className="text-xs opacity-60 mt-2 font-mono">
                      {message.timestamp.toLocaleTimeString('ru-RU')}
                    </p>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>

            {isTyping && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="flex justify-start"
              >
                <div className="glass border-white/20 rounded-2xl px-4 py-3">
                  <div className="flex items-center gap-2">
                    <div className="flex gap-1">
                      <motion.div
                        className="w-2 h-2 bg-indigo-400 rounded-full"
                        animate={{ scale: [1, 1.5, 1], opacity: [0.5, 1, 0.5] }}
                        transition={{ duration: 1, repeat: Infinity, delay: 0 }}
                      />
                      <motion.div
                        className="w-2 h-2 bg-violet-400 rounded-full"
                        animate={{ scale: [1, 1.5, 1], opacity: [0.5, 1, 0.5] }}
                        transition={{ duration: 1, repeat: Infinity, delay: 0.2 }}
                      />
                      <motion.div
                        className="w-2 h-2 bg-purple-400 rounded-full"
                        animate={{ scale: [1, 1.5, 1], opacity: [0.5, 1, 0.5] }}
                        transition={{ duration: 1, repeat: Infinity, delay: 0.4 }}
                      />
                    </div>
                    <span className="text-xs text-zinc-400 font-medium">Агент думает...</span>
                  </div>
                </div>
              </motion.div>
            )}
          </div>
        </ScrollArea>

        <form onSubmit={handleSubmit} className="space-y-4 border-t border-white/10 pt-4">
          <div>
            <label className="text-xs text-zinc-400 mb-2 block font-medium">
              🔗 URL конкурента (опционально)
            </label>
            <Input
              type="url"
              placeholder="https://example.com"
              value={analyzeUrl}
              onChange={(e) => setAnalyzeUrl(e.target.value)}
              disabled={loading}
              className="glass border-white/20 text-white placeholder:text-zinc-500 focus:border-indigo-500/50 transition-all"
            />
          </div>
          <div>
            <label className="text-xs text-zinc-400 mb-2 block font-medium">
              ✨ Опишите проект
            </label>
            <Input
              placeholder="Создай современный landing page с формой подписки и анимациями..."
              value={specification}
              onChange={(e) => setSpecification(e.target.value)}
              disabled={loading}
              className="glass border-white/20 text-white placeholder:text-zinc-500 focus:border-indigo-500/50 transition-all"
            />
          </div>
          <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
            <Button
              type="submit"
              disabled={loading || !specification.trim()}
              className="w-full gradient-indigo-violet hover:shadow-glow text-white font-semibold py-6 rounded-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <motion.span
                    animate={{ rotate: 360 }}
                    transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                  >
                    ⚡
                  </motion.span>
                  Генерация...
                </span>
              ) : (
                '🚀 Сгенерировать проект'
              )}
            </Button>
          </motion.div>
        </form>
      </CardContent>
    </Card>
  );
}
