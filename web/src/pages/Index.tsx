import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import MainLayout from "@/components/layout/MainLayout";
import HeroSection from "@/components/HeroSection";

const Index = () => {
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
      >
        <MainLayout withSidebar decorated>
          <HeroSection onGenerate={handleGenerate} />
        </MainLayout>
      </motion.div>
    </AnimatePresence>
  );
};

export default Index;
