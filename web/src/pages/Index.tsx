import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import HeaderBar from "@/components/HeaderBar";
import ProjectSidebar from "@/components/ProjectSidebar";
import HeroSection from "@/components/HeroSection";
import HowItWorks from "@/components/HowItWorks";
import TargetAudienceSection from "@/components/TargetAudienceSection";
import CapabilitiesSection from "@/components/CapabilitiesSection";
import PricingSection from "@/components/PricingSection";
import TemplatesSection from "@/components/TemplatesSection";
import CTASection from "@/components/CTASection";
import FooterSection from "@/components/FooterSection";

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
      <div className="flex-1 min-w-0 flex flex-col">
        <HeaderBar />
        <div className="flex-1 overflow-y-auto mesh-gradient-bg grid-pattern">
          {!leaving && <HeroSection onGenerate={handleGenerate} />}
          {leaving && <HeroSection onGenerate={() => {}} />}
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
