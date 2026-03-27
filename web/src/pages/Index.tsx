import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import HeaderBar from "@/components/HeaderBar";
import ProjectSidebar from "@/components/ProjectSidebar";
import HeroSection from "@/components/HeroSection";

const Index = () => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(true);
  const [leaving, setLeaving] = useState(false);
  const navigate = useNavigate();

  const handleGenerate = (prompt: string) => {
    setLeaving(true);
    setTimeout(() => {
      navigate("/project/new", { state: { prompt } });
    }, 500);
  };

  return (
    <AnimatePresence>
      <motion.div
        key={leaving ? "leaving" : "main"}
        initial={{ opacity: 1 }}
        animate={{ opacity: leaving ? 0 : 1, y: leaving ? -20 : 0 }}
        transition={{ duration: 0.45, ease: "easeInOut" }}
        className="h-screen flex flex-col overflow-hidden"
      >
        <div className="flex-1 flex overflow-hidden relative">
          <ProjectSidebar
            collapsed={sidebarCollapsed}
            onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
          />
          <div className="flex-1 min-w-0 flex flex-col">
            <HeaderBar />
            <div className="flex-1 overflow-y-auto mesh-gradient-bg grid-pattern">
              <HeroSection onGenerate={handleGenerate} />
            </div>
          </div>
        </div>
      </motion.div>
    </AnimatePresence>
  );
};

export default Index;
