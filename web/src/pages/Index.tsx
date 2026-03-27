import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import Navbar from "@/components/Navbar";
import Hero from "@/components/Hero";
import ProjectSidebar from "@/components/ProjectSidebar";
import HowItWorks from "@/components/HowItWorks";
import TargetAudienceSection from "@/components/TargetAudienceSection";
import CapabilitiesSection from "@/components/CapabilitiesSection";
import PricingSection from "@/components/PricingSection";
import TemplatesSection from "@/components/TemplatesSection";
import CTASection from "@/components/CTASection";
import FooterSection from "@/components/FooterSection";
import "@/styles/App.css";

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

  const content = (
    <div className="flex-1 flex overflow-hidden relative">
      <ProjectSidebar
        collapsed={sidebarCollapsed}
        onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
      />
      <div className="flex-1 min-w-0 flex flex-col overflow-y-auto bg-[#08080a]">
        <Navbar />
        {!leaving && <Hero onGenerate={handleGenerate} />}
        {leaving && <Hero onGenerate={() => {}} />}
        {!leaving && (
          <>
            <TargetAudienceSection />
            <HowItWorks />
            <CapabilitiesSection />
            <PricingSection />
            <TemplatesSection />
            <CTASection />
            <FooterSection />
          </>
        )}
      </div>
    </div>
  );

  return (
    <AnimatePresence>
      <motion.div
        key={leaving ? "leaving" : "main"}
        initial={{ opacity: leaving ? 1 : 1 }}
        animate={{ opacity: leaving ? 0 : 1, y: leaving ? -20 : 0 }}
        transition={{ duration: 0.45, ease: "easeInOut" }}
        className="h-screen flex flex-col overflow-hidden"
      >
        {content}
      </motion.div>
    </AnimatePresence>
  );
};

export default Index;
