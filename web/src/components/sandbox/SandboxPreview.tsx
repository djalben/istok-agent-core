'use client';

import { useState, useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface SandboxPreviewProps {
  code: string;
}

export function SandboxPreview({ code }: SandboxPreviewProps) {
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const iframeRef = useRef<HTMLIFrameElement>(null);

  useEffect(() => {
    if (code && iframeRef.current) {
      const iframe = iframeRef.current;
      const doc = iframe.contentDocument || iframe.contentWindow?.document;

      if (doc) {
        doc.open();
        doc.write(code);
        doc.close();
      }
    }
  }, [code]);

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  const downloadCode = () => {
    const blob = new Blob([code], { type: 'text/html' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'istok-generated-project.html';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const refreshPreview = () => {
    setIsRefreshing(true);
    if (code && iframeRef.current) {
      const iframe = iframeRef.current;
      const doc = iframe.contentDocument || iframe.contentWindow?.document;
      if (doc) {
        doc.open();
        doc.write(code);
        doc.close();
      }
    }
    setTimeout(() => setIsRefreshing(false), 500);
  };

  return (
    <Card
      className={`glass-strong shadow-glow-sm border-white/10 flex flex-col overflow-hidden ${
        isFullscreen ? 'fixed inset-0 z-50 rounded-none' : 'h-full'
      }`}
    >
      {/* Browser-like header */}
      <CardHeader className="p-0 border-b border-white/10">
        {/* Window controls */}
        <div className="flex items-center justify-between px-4 py-3 bg-gradient-to-r from-indigo-500/10 to-violet-500/10">
          <div className="flex items-center gap-2">
            <div className="flex gap-1.5">
              <div className="w-3 h-3 rounded-full bg-red-500/80" />
              <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
              <div className="w-3 h-3 rounded-full bg-green-500/80" />
            </div>
            <span className="text-xs text-zinc-400 ml-2 font-medium">Предпросмотр</span>
          </div>
          <div className="flex gap-2">
            {code && (
              <>
                <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
                  <Button
                    onClick={refreshPreview}
                    variant="ghost"
                    size="sm"
                    className="glass border-white/10 text-white hover:bg-white/10 h-8 px-3"
                  >
                    <motion.span
                      animate={isRefreshing ? { rotate: 360 } : {}}
                      transition={{ duration: 0.5 }}
                    >
                      🔄
                    </motion.span>
                  </Button>
                </motion.div>
                <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
                  <Button
                    onClick={downloadCode}
                    variant="ghost"
                    size="sm"
                    className="glass border-white/10 text-white hover:bg-white/10 h-8 px-3"
                  >
                    💾
                  </Button>
                </motion.div>
              </>
            )}
            <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
              <Button
                onClick={toggleFullscreen}
                variant="ghost"
                size="sm"
                className="glass border-white/10 text-white hover:bg-white/10 h-8 px-3"
              >
                {isFullscreen ? '🗗' : '🗖'}
              </Button>
            </motion.div>
          </div>
        </div>

        {/* Address bar */}
        <div className="px-4 py-2 bg-black/20">
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-zinc-500 hover:text-white hover:bg-white/10"
                disabled
              >
                ←
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-zinc-500 hover:text-white hover:bg-white/10"
                disabled
              >
                →
              </Button>
            </div>
            <div className="flex-1 glass border-white/10 rounded-lg px-3 py-1.5 flex items-center gap-2">
              <span className="text-xs text-green-400">🔒</span>
              <span className="text-xs text-zinc-400 font-mono">
                {code ? 'istok://generated-project' : 'istok://preview'}
              </span>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex-1 p-0 bg-zinc-900/50">
        {code ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.3 }}
            className="w-full h-full"
          >
            <iframe
              ref={iframeRef}
              className="w-full h-full bg-white"
              title="Preview"
              sandbox="allow-scripts allow-same-origin"
            />
          </motion.div>
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-center space-y-4"
            >
              <motion.div
                animate={{ scale: [1, 1.1, 1] }}
                transition={{ duration: 2, repeat: Infinity }}
                className="text-6xl"
              >
                🎨
              </motion.div>
              <div>
                <p className="text-lg text-zinc-300 font-medium">Готов к предпросмотру</p>
                <p className="text-sm text-zinc-500 mt-2">
                  Сгенерируйте проект в терминале агента
                </p>
              </div>
              <div className="flex items-center gap-2 justify-center mt-6">
                <div className="w-2 h-2 bg-indigo-500 rounded-full animate-pulse-glow" />
                <div className="w-2 h-2 bg-violet-500 rounded-full animate-pulse-glow" style={{ animationDelay: '0.2s' }} />
                <div className="w-2 h-2 bg-purple-500 rounded-full animate-pulse-glow" style={{ animationDelay: '0.4s' }} />
              </div>
            </motion.div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
