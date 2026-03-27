import { motion } from "framer-motion";
import { MessageSquareText, Cpu, Rocket } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

const container = {
  hidden: {},
  show: { transition: { staggerChildren: 0.15 } },
};

const item = {
  hidden: { opacity: 0, y: 32 },
  show: { opacity: 1, y: 0, transition: { duration: 0.5, ease: "easeOut" as const } },
};

const HowItWorks = () => {
  const { t } = useLanguage();

  const steps = [
    {
      icon: MessageSquareText,
      title: t("step1Title"),
      description: t("step1Desc"),
      iconGradient: "from-[hsla(243,76%,58%,1)] to-[hsla(260,70%,55%,1)]",
    },
    {
      icon: Cpu,
      title: t("step2Title"),
      description: t("step2Desc"),
      iconGradient: "from-[hsla(260,70%,55%,1)] to-[hsla(220,80%,50%,1)]",
    },
    {
      icon: Rocket,
      title: t("step3Title"),
      description: t("step3Desc"),
      iconGradient: "from-[hsla(220,80%,50%,1)] to-[hsla(243,76%,58%,1)]",
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
        <h2 className="text-2xl md:text-3xl font-bold text-foreground tracking-tight mb-3">
          {t("howTitle")}
        </h2>
        <p className="text-muted-foreground text-sm max-w-md mx-auto">
          {t("howSubtitle")}
        </p>
      </motion.div>

      <motion.div
        variants={container}
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.2 }}
        className="max-w-4xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-6 md:gap-8"
      >
        {steps.map((step, i) => (
          <motion.div
            key={i}
            variants={item}
            className="group relative glass-subtle rounded-2xl p-6 md:p-8 text-center card-border-glow hover:-translate-y-1 transition-all duration-300"
          >
            <div className={`w-14 h-14 mx-auto mb-5 rounded-xl bg-gradient-to-br ${step.iconGradient} bg-opacity-20 flex items-center justify-center shadow-[0_0_20px_hsla(243,76%,58%,0.15)] group-hover:shadow-[0_0_30px_hsla(243,76%,58%,0.25)] transition-shadow duration-300`}>
              <step.icon size={24} className="text-primary-foreground" />
            </div>
            <div className="text-xs text-muted-foreground/50 font-semibold tracking-widest uppercase mb-2">
              {t("step")} {i + 1}
            </div>
            <h3 className="text-lg font-semibold text-foreground mb-2">{step.title}</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">{step.description}</p>
          </motion.div>
        ))}
      </motion.div>
    </section>
  );
};

export default HowItWorks;
