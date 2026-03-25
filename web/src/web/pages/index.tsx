import { motion } from "framer-motion";
import { Sidebar } from "../components/sidebar";
import { IntelligenceBar } from "../components/intelligence-bar";
import { AgentTerminal } from "../components/agent-terminal";
import { PreviewPanel } from "../components/preview-panel";
import { ProjectStats } from "../components/project-stats";
import {
  Cpu,
  Wifi,
  Shield,
} from "lucide-react";

function StatusPill({ icon: Icon, label, status }: { icon: React.ElementType; label: string; status: "active" | "idle" }) {
  return (
    <div className="flex items-center gap-2 px-3 py-1.5 glass rounded-lg">
      <Icon className="w-3 h-3 text-zinc-500" />
      <span className="text-[10px] text-zinc-500 font-medium">{label}</span>
      <div
        className={`w-1.5 h-1.5 rounded-full ${
          status === "active" ? "bg-green-500 green-pulse" : "bg-zinc-600"
        }`}
      />
    </div>
  );
}

function Dashboard() {
  return (
    <div className="min-h-screen bg-[#09090b] text-white relative overflow-hidden">
      {/* Ambient background gradients */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-[-20%] left-[-10%] w-[60%] h-[60%] bg-indigo-500/[0.03] rounded-full blur-[120px]" />
        <div className="absolute bottom-[-20%] right-[-10%] w-[50%] h-[50%] bg-violet-500/[0.03] rounded-full blur-[120px]" />
      </div>

      <Sidebar />

      {/* Main Content — offset by sidebar */}
      <main className="ml-16 min-h-screen relative z-10">
        <div className="max-w-[1600px] mx-auto px-5 py-4">
          {/* Top Bar: Status Indicators */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.4 }}
            className="flex items-center justify-between mb-4"
          >
            <div className="flex items-center gap-2">
              <StatusPill icon={Cpu} label="GPU A100" status="active" />
              <StatusPill icon={Wifi} label="API Gateway" status="active" />
              <StatusPill icon={Shield} label="Auth" status="active" />
            </div>
            <div className="flex items-center gap-3">
              <span className="text-[10px] text-zinc-600 font-mono">
                v2.4.0-rc.1
              </span>
              <div className="w-7 h-7 rounded-full bg-gradient-to-br from-indigo-500 to-violet-500 flex items-center justify-center text-[11px] font-semibold">
                А
              </div>
            </div>
          </motion.div>

          {/* Intelligence Bar */}
          <div className="mb-5">
            <IntelligenceBar />
          </div>

          {/* Bento Grid Workspace */}
          <div className="grid grid-cols-12 gap-3 h-[calc(100vh-200px)] min-h-[500px]">
            {/* Panel A — Agent Terminal (left) */}
            <div className="col-span-12 lg:col-span-4 xl:col-span-3">
              <AgentTerminal />
            </div>

            {/* Panel B — Preview (center) */}
            <div className="col-span-12 lg:col-span-5 xl:col-span-6">
              <PreviewPanel />
            </div>

            {/* Panel C — Stats (right) */}
            <div className="col-span-12 lg:col-span-3 xl:col-span-3">
              <ProjectStats />
            </div>
          </div>

          {/* Bottom Status Bar */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.8 }}
            className="flex items-center justify-between mt-3 px-1"
          >
            <div className="flex items-center gap-4">
              <span className="text-[10px] text-zinc-700 font-mono">
                pid: 84729 • mem: 2.4GB • cpu: 12%
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] text-zinc-700">
                © 2026 ИСТОК АГЕНТ
              </span>
              <span className="text-[10px] text-zinc-800">•</span>
              <span className="text-[10px] text-zinc-700 font-mono">
                Российская Федерация
              </span>
            </div>
          </motion.div>
        </div>
      </main>
    </div>
  );
}

export default Dashboard;
