'use client';

import { useState } from 'react';
import { AgentTerminal } from '@/components/agent/AgentTerminal';
import { SandboxPreview } from '@/components/sandbox/SandboxPreview';
import { AgentStats } from '@/components/stats/AgentStats';

export default function Home() {
  const [generatedCode, setGeneratedCode] = useState('');

  return (
    <div className="min-h-screen bg-gradient-to-br from-zinc-950 via-black to-zinc-900">
      {/* Header */}
      <header className="border-b border-white/10 bg-black/40 backdrop-blur-xl">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="text-3xl">🤖</div>
              <div>
                <h1 className="text-2xl font-bold text-white">Исток Agent</h1>
                <p className="text-sm text-zinc-400">Автономный AI агент для генерации проектов</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
          <div className="lg:col-span-3">
            <AgentStats />
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 h-[calc(100vh-280px)]">
          {/* Terminal */}
          <div className="h-full">
            <AgentTerminal onCodeGenerated={setGeneratedCode} />
          </div>

          {/* Preview */}
          <div className="h-full">
            <SandboxPreview code={generatedCode} />
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-white/10 bg-black/40 backdrop-blur-xl mt-6">
        <div className="container mx-auto px-4 py-4">
          <p className="text-center text-sm text-zinc-400">
            Построено на Clean Architecture | Go + Next.js + Claude AI
          </p>
        </div>
      </footer>
    </div>
  );
}
