'use client';

import { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface SandboxPreviewProps {
  code: string;
}

export function SandboxPreview({ code }: SandboxPreviewProps) {
  const [isFullscreen, setIsFullscreen] = useState(false);
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
    a.download = 'generated-project.html';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <Card
      className={`bg-black/40 backdrop-blur-xl border-white/20 shadow-2xl flex flex-col ${
        isFullscreen ? 'fixed inset-0 z-50 rounded-none' : 'h-full'
      }`}
    >
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-xl font-bold text-white flex items-center gap-2">
            <span className="text-2xl">🖼️</span>
            Предпросмотр
          </CardTitle>
          <div className="flex gap-2">
            {code && (
              <Button
                onClick={downloadCode}
                variant="outline"
                size="sm"
                className="bg-black/30 border-white/20 text-white hover:bg-white/10"
              >
                💾 Скачать
              </Button>
            )}
            <Button
              onClick={toggleFullscreen}
              variant="outline"
              size="sm"
              className="bg-black/30 border-white/20 text-white hover:bg-white/10"
            >
              {isFullscreen ? '🗗 Свернуть' : '🗖 Развернуть'}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 p-4">
        {code ? (
          <iframe
            ref={iframeRef}
            className="w-full h-full bg-white rounded-lg border border-white/20"
            title="Preview"
            sandbox="allow-scripts allow-same-origin"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-zinc-900/50 rounded-lg border border-white/10">
            <div className="text-center space-y-2">
              <p className="text-4xl">🎨</p>
              <p className="text-zinc-400">Здесь появится предпросмотр</p>
              <p className="text-xs text-zinc-500">
                Сгенерируйте проект в терминале
              </p>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
