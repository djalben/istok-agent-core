import { motion } from "framer-motion";
import { Code2, Palette, Globe, Shield, Zap, Database } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

const item = {
  hidden: { opacity: 0, y: 24, scale: 0.97 },
  show: { opacity: 1, y: 0, scale: 1, transition: { duration: 0.45, ease: "easeOut" as const } },
};

const BentoFeatures = () => {
  const { t } = useLanguage();

  const features = [
    {
      icon: Code2,
      title: t("featCodeGen"),
      description: t("featCodeGenDesc"),
      className: "md:col-span-2 md:row-span-1",
      gradient: "from-primary/20 to-transparent",
      iconGradient: "from-[hsla(243,76%,58%,1)] to-[hsla(220,80%,50%,1)]",
    },
    {
      icon: Palette,
      title: t("featDesign"),
      description: t("featDesignDesc"),
      className: "md:col-span-1 md:row-span-2",
      gradient: "from-[hsla(260,70%,40%,0.2)] to-transparent",
      iconGradient: "from-[hsla(260,70%,55%,1)] to-[hsla(300,60%,50%,1)]",
    },
    {
      icon: Globe,
      title: t("featDeploy"),
      description: t("featDeployDesc"),
      className: "md:col-span-1 md:row-span-1",
      gradient: "from-[hsla(220,80%,45%,0.2)] to-transparent",
      iconGradient: "from-[hsla(220,80%,50%,1)] to-[hsla(200,70%,50%,1)]",
    },
    {
      icon: Shield,
      title: t("featAuth"),
      description: t("featAuthDesc"),
      className: "md:col-span-1 md:row-span-1",
      gradient: "from-[hsla(170,60%,35%,0.15)] to-transparent",
      iconGradient: "from-[hsla(170,60%,45%,1)] to-[hsla(200,70%,50%,1)]",
    },
    {
      icon: Database,
      title: t("featDB"),
      description: t("featDBDesc"),
      className: "md:col-span-1 md:row-span-1",
      gradient: "from-primary/15 to-transparent",
      iconGradient: "from-[hsla(243,76%,58%,1)] to-[hsla(260,70%,55%,1)]",
    },
    {
      icon: Zap,
      title: t("featPreview"),
      description: t("featPreviewDesc"),
      className: "md:col-span-1 md:row-span-1",
      gradient: "from-[hsla(45,80%,45%,0.12)] to-transparent",
      iconGradient: "from-[hsla(45,80%,50%,1)] to-[hsla(30,80%,50%,1)]",
    },
  ];

  return (
    <section className="py-20 md:py-32 px-4 md:px-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.5 }}
        className="text-center mb-12 md:mb-16"
      >
        <h2 className="text-2xl md:text-3xl font-bold text-foreground tracking-tight mb-3">
          {t("featuresTitle")}
        </h2>
        <p className="text-muted-foreground text-sm max-w-md mx-auto">
          {t("featuresSubtitle")}
        </p>
      </motion.div>

      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.1 }}
        transition={{ staggerChildren: 0.08 }}
        className="max-w-5xl mx-auto grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4"
      >
        {features.map((f, i) => (
          <motion.div
            key={i}
            variants={item}
            className={`group relative glass-subtle rounded-2xl p-6 md:p-7 overflow-hidden card-border-glow hover:-translate-y-1 transition-all duration-300 ${f.className}`}
          >
            <div className={`absolute inset-0 bg-gradient-to-br ${f.gradient} opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none`} />
            <div className="relative z-10">
              <div className={`w-10 h-10 rounded-lg bg-gradient-to-br ${f.iconGradient} flex items-center justify-center mb-4 shadow-[0_0_16px_hsla(243,76%,58%,0.15)] group-hover:shadow-[0_0_24px_hsla(243,76%,58%,0.25)] transition-shadow duration-300`}>
                <f.icon size={20} className="text-primary-foreground" />
              </div>
              <h3 className="text-base font-semibold text-foreground mb-1.5">{f.title}</h3>
              <p className="text-sm text-muted-foreground leading-relaxed">{f.description}</p>
            </div>
          </motion.div>
        ))}
      </motion.div>
    </section>
  );
};

export default BentoFeatures;
