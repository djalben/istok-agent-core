import { useState } from "react";
import { motion } from "framer-motion";
import { RefreshCw, ExternalLink, Code2, Eye } from "lucide-react";

export function PreviewPanel() {
  const [activeTab, setActiveTab] = useState<"preview" | "code">("preview");

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.2, ease: [0.25, 0.1, 0.25, 1] }}
      className="glass rounded-2xl flex flex-col h-full overflow-hidden relative noise"
    >
      {/* macOS Window Controls */}
      <div className="flex items-center h-11 px-4 border-b border-white/[0.06] flex-shrink-0">
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-full bg-[#ff5f57] opacity-80 hover:opacity-100 transition-opacity cursor-pointer" />
          <div className="w-3 h-3 rounded-full bg-[#febc2e] opacity-80 hover:opacity-100 transition-opacity cursor-pointer" />
          <div className="w-3 h-3 rounded-full bg-[#28c840] opacity-80 hover:opacity-100 transition-opacity cursor-pointer" />
        </div>

        {/* URL Bar */}
        <div className="flex-1 mx-4">
          <div className="flex items-center gap-2 bg-white/[0.03] border border-white/[0.06] rounded-lg px-3 py-1">
            <div className="w-3 h-3 rounded-full border border-green-500/40 flex items-center justify-center">
              <div className="w-1.5 h-1.5 rounded-full bg-green-500" />
            </div>
            <span className="text-[11px] text-zinc-500 font-mono truncate">
              localhost:3000
            </span>
          </div>
        </div>

        {/* Tab Switcher */}
        <div className="flex items-center gap-0.5 bg-white/[0.03] rounded-lg p-0.5 border border-white/[0.04]">
          <button
            onClick={() => setActiveTab("preview")}
            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[10px] font-medium transition-all ${
              activeTab === "preview"
                ? "bg-white/[0.06] text-white"
                : "text-zinc-500 hover:text-zinc-400"
            }`}
          >
            <Eye className="w-3 h-3" />
            Превью
          </button>
          <button
            onClick={() => setActiveTab("code")}
            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[10px] font-medium transition-all ${
              activeTab === "code"
                ? "bg-white/[0.06] text-white"
                : "text-zinc-500 hover:text-zinc-400"
            }`}
          >
            <Code2 className="w-3 h-3" />
            Код
          </button>
        </div>

        <div className="flex items-center gap-1 ml-2">
          <motion.button
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.9 }}
            className="w-7 h-7 rounded-lg flex items-center justify-center hover:bg-white/[0.04] transition-colors"
          >
            <RefreshCw className="w-3.5 h-3.5 text-zinc-500" />
          </motion.button>
          <motion.button
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.9 }}
            className="w-7 h-7 rounded-lg flex items-center justify-center hover:bg-white/[0.04] transition-colors"
          >
            <ExternalLink className="w-3.5 h-3.5 text-zinc-500" />
          </motion.button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 relative z-10 overflow-hidden">
        {activeTab === "preview" ? (
          <div className="h-full flex flex-col items-center justify-center p-6">
            {/* Mock website preview */}
            <div className="w-full max-w-md space-y-4">
              {/* Mock header */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-6 h-6 rounded-md bg-gradient-to-br from-indigo-500 to-violet-500" />
                  <div className="w-20 h-2.5 bg-white/[0.08] rounded-full" />
                </div>
                <div className="flex gap-3">
                  <div className="w-12 h-2 bg-white/[0.05] rounded-full" />
                  <div className="w-12 h-2 bg-white/[0.05] rounded-full" />
                  <div className="w-12 h-2 bg-white/[0.05] rounded-full" />
                </div>
              </div>

              {/* Mock hero */}
              <div className="mt-8 space-y-3">
                <div className="w-3/4 h-4 bg-white/[0.1] rounded-full" />
                <div className="w-1/2 h-4 bg-white/[0.07] rounded-full" />
                <div className="w-2/3 h-2.5 bg-white/[0.04] rounded-full mt-4" />
                <div className="w-1/2 h-2.5 bg-white/[0.04] rounded-full" />
              </div>

              {/* Mock CTA */}
              <div className="flex gap-2 mt-6">
                <div className="w-28 h-8 rounded-lg bg-gradient-to-r from-indigo-500/20 to-violet-500/20 border border-indigo-500/20" />
                <div className="w-24 h-8 rounded-lg bg-white/[0.04] border border-white/[0.06]" />
              </div>

              {/* Mock cards */}
              <div className="grid grid-cols-3 gap-2 mt-6">
                {[1, 2, 3].map((i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.4 + i * 0.1 }}
                    className="h-20 rounded-lg bg-white/[0.03] border border-white/[0.06] p-2.5"
                  >
                    <div className="w-5 h-5 rounded bg-white/[0.06] mb-2" />
                    <div className="w-full h-1.5 bg-white/[0.04] rounded-full" />
                    <div className="w-2/3 h-1.5 bg-white/[0.03] rounded-full mt-1" />
                  </motion.div>
                ))}
              </div>
            </div>
          </div>
        ) : (
          <div className="h-full p-4 font-mono text-[12px] leading-relaxed overflow-y-auto">
            <div className="space-y-0.5">
              <div>
                <span className="text-violet-400">import</span>
                <span className="text-zinc-300"> {"{ "}</span>
                <span className="text-indigo-300">NextPage</span>
                <span className="text-zinc-300">{" }"} </span>
                <span className="text-violet-400">from</span>
                <span className="text-green-400"> 'next'</span>
                <span className="text-zinc-500">;</span>
              </div>
              <div>
                <span className="text-violet-400">import</span>
                <span className="text-zinc-300"> {"{ "}</span>
                <span className="text-indigo-300">motion</span>
                <span className="text-zinc-300">{" }"} </span>
                <span className="text-violet-400">from</span>
                <span className="text-green-400"> 'framer-motion'</span>
                <span className="text-zinc-500">;</span>
              </div>
              <div className="h-3" />
              <div>
                <span className="text-violet-400">const</span>
                <span className="text-indigo-300"> Home</span>
                <span className="text-zinc-400">: </span>
                <span className="text-yellow-300">NextPage</span>
                <span className="text-zinc-300"> = () </span>
                <span className="text-violet-400">=&gt;</span>
                <span className="text-zinc-300"> {"{"}</span>
              </div>
              <div>
                <span className="text-zinc-500">  </span>
                <span className="text-violet-400">return</span>
                <span className="text-zinc-300"> (</span>
              </div>
              <div>
                <span className="text-zinc-500">    </span>
                <span className="text-zinc-400">&lt;</span>
                <span className="text-indigo-300">motion.main</span>
              </div>
              <div>
                <span className="text-zinc-500">      </span>
                <span className="text-zinc-400">className=</span>
                <span className="text-green-400">"min-h-screen"</span>
              </div>
              <div>
                <span className="text-zinc-500">      </span>
                <span className="text-zinc-400">initial=</span>
                <span className="text-zinc-300">{"{{ opacity: 0 }}"}</span>
              </div>
              <div>
                <span className="text-zinc-500">      </span>
                <span className="text-zinc-400">animate=</span>
                <span className="text-zinc-300">{"{{ opacity: 1 }}"}</span>
              </div>
              <div>
                <span className="text-zinc-500">    </span>
                <span className="text-zinc-400">&gt;</span>
              </div>
              <div>
                <span className="text-zinc-500">      </span>
                <span className="text-zinc-600">{"// Сгенерировано ИСТОК АГЕНТ"}</span>
              </div>
              <div>
                <span className="text-zinc-500">      </span>
                <span className="text-zinc-600">{"// 23 файла • 847 узлов"}</span>
              </div>
              <div>
                <span className="text-zinc-500">    </span>
                <span className="text-zinc-400">&lt;/</span>
                <span className="text-indigo-300">motion.main</span>
                <span className="text-zinc-400">&gt;</span>
              </div>
              <div>
                <span className="text-zinc-500">  </span>
                <span className="text-zinc-300">);</span>
              </div>
              <div>
                <span className="text-zinc-300">{"};"}</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Status Bar */}
      <div className="h-8 px-4 border-t border-white/[0.06] flex items-center justify-between flex-shrink-0 relative z-10">
        <div className="flex items-center gap-2">
          <div className="w-1.5 h-1.5 rounded-full bg-green-500 green-pulse" />
          <span className="text-[10px] text-zinc-500 font-medium">
            Подключено к Go-бэкенду
          </span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-[10px] text-zinc-600 font-mono">ws://localhost:8080</span>
          <span className="text-[10px] text-zinc-600">142ms</span>
        </div>
      </div>
    </motion.div>
  );
}
