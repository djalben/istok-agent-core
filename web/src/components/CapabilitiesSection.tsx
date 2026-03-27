import { motion } from "framer-motion";
import { Brain, Download, Globe, MessageSquare, Eye, Database } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

const item = {
  hidden: { opacity: 0, y: 24 },
  show: { opacity: 1, y: 0, transition: { duration: 0.45, ease: "easeOut" as const } },
};

const CapabilitiesSection = () => {
  const { t } = useLanguage();

  const capabilities = [
    {
      icon: Brain,
      title: t("capCodeGen"),
      description: t("capCodeGenDesc"),
      gradient: "from-violet-500/20 to-purple-600/10",
    },
    {
      icon: Download,
      title: t("capExport"),
      description: t("capExportDesc"),
      gradient: "from-blue-500/20 to-cyan-500/10",
    },
    {
      icon: Globe,
      title: t("capPublish"),
      description: t("capPublishDesc"),
      gradient: "from-emerald-500/20 to-green-500/10",
    },
    {
      icon: MessageSquare,
      title: t("capAI"),
      description: t("capAIDesc"),
      gradient: "from-orange-500/20 to-amber-500/10",
    },
    {
      icon: Eye,
      title: t("capPreview"),
      description: t("capPreviewDesc"),
      gradient: "from-pink-500/20 to-rose-500/10",
    },
    {
      icon: Database,
      title: t("capDB"),
      description: t("capDBDesc"),
      gradient: "from-indigo-500/20 to-blue-600/10",
    },
  ];

  return (
    <section className="py-20 md:py-32 px-4 md:px-6 relative">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.5 }}
        className="text-center mb-12 md:mb-16"
      >
        <h2 className="text-2xl md:text-4xl font-bold text-foreground tracking-tight mb-3">
          {t("capabilitiesTitle")}
        </h2>
        <p className="text-muted-foreground text-sm md:text-base max-w-lg mx-auto">
          {t("capabilitiesSubtitle")}
        </p>
      </motion.div>

      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.1 }}
        transition={{ staggerChildren: 0.1 }}
        className="max-w-5xl mx-auto grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5"
      >
        {capabilities.map((cap, i) => (
          <motion.div
            key={i}
            variants={item}
            className="group relative glass-subtle rounded-2xl p-6 hover:-translate-y-1 transition-all duration-300 border border-border/30 hover:border-primary/20"
          >
            <div className={`absolute inset-0 rounded-2xl bg-gradient-to-br ${cap.gradient} opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none`} />
            <div className="relative z-10">
              <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center mb-4 group-hover:bg-primary/20 transition-colors duration-300">
                <cap.icon size={24} className="text-primary" />
              </div>
              <h3 className="text-base font-semibold text-foreground mb-2">{cap.title}</h3>
              <p className="text-sm text-muted-foreground leading-relaxed">{cap.description}</p>
            </div>
          </motion.div>
        ))}
      </motion.div>
    </section>
  );
};

export default CapabilitiesSection;
