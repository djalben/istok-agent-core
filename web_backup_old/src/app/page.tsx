'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { AgentTerminal } from '@/components/agent/AgentTerminal';
import { SandboxPreview } from '@/components/sandbox/SandboxPreview';
import { AgentStats } from '@/components/stats/AgentStats';

export default function Home() {
  const [generatedCode, setGeneratedCode] = useState('');

  return (
    <div className="min-h-screen bg-black overflow-hidden">
      {/* Animated gradient mesh background */}
      <div className="fixed inset-0 gradient-mesh opacity-50" />
      <div className="fixed inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-indigo-900/20 via-black to-black" />
      
      {/* Floating orbs */}
      <motion.div
        className="fixed top-20 left-20 w-96 h-96 bg-indigo-600/30 rounded-full blur-3xl"
        animate={{
          scale: [1, 1.2, 1],
          opacity: [0.3, 0.5, 0.3],
        }}
        transition={{
          duration: 8,
          repeat: Infinity,
          ease: "easeInOut",
        }}
      />
      <motion.div
        className="fixed bottom-20 right-20 w-96 h-96 bg-violet-600/30 rounded-full blur-3xl"
        animate={{
          scale: [1.2, 1, 1.2],
          opacity: [0.5, 0.3, 0.5],
        }}
        transition={{
          duration: 10,
          repeat: Infinity,
          ease: "easeInOut",
        }}
      />

      <div className="relative z-10">
        {/* Header */}
        <motion.header
          initial={{ y: -100, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.6, ease: "easeOut" }}
          className="border-b border-white/5 glass-strong sticky top-0 z-50"
        >
          <div className="container mx-auto px-6 py-5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <motion.div
                  className="text-4xl"
                  animate={{ rotate: [0, 10, -10, 0] }}
                  transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
                >
                  🤖
                </motion.div>
                <div>
                  <h1 className="text-3xl font-bold text-gradient tracking-tight">
                    Исток Agent
                  </h1>
                  <p className="text-sm text-zinc-400 font-medium mt-0.5">
                    Автономный AI агент нового поколения
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-2 glass px-4 py-2 rounded-full">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse-glow" />
                  <span className="text-xs text-zinc-300 font-medium">Online</span>
                </div>
              </div>
            </div>
          </div>
        </motion.header>

        {/* Main Content */}
        <main className="container mx-auto px-6 py-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="mb-8"
          >
            <AgentStats />
          </motion.div>

          <div className="grid grid-cols-1 xl:grid-cols-2 gap-8 min-h-[calc(100vh-320px)]">
            {/* Terminal */}
            <motion.div
              initial={{ opacity: 0, x: -50 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.6, delay: 0.3 }}
              className="h-full"
            >
              <AgentTerminal onCodeGenerated={setGeneratedCode} />
            </motion.div>

            {/* Preview */}
            <motion.div
              initial={{ opacity: 0, x: 50 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.6, delay: 0.4 }}
              className="h-full"
            >
              <SandboxPreview code={generatedCode} />
            </motion.div>
          </div>
        </main>

        {/* Footer */}
        <motion.footer
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.5 }}
          className="border-t border-white/5 glass mt-12"
        >
          <div className="container mx-auto px-6 py-6">
            <div className="flex items-center justify-between">
              <p className="text-sm text-zinc-500">
                Построено на <span className="text-gradient font-semibold">Clean Architecture</span>
              </p>
              <div className="flex items-center gap-6 text-xs text-zinc-500">
                <span>Go + Next.js 15</span>
                <span>•</span>
                <span>Claude AI</span>
                <span>•</span>
                <span className="text-gradient font-semibold">v1.0.0</span>
              </div>
            </div>
          </div>
        </motion.footer>
      </div>
    </div>
  );
}
