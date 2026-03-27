import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import { ArrowRight } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

const CTASection = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();

  return (
    <section className="py-20 md:py-32 px-4 md:px-6">
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        whileInView={{ opacity: 1, scale: 1 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.5 }}
        className="max-w-3xl mx-auto text-center relative"
      >
        <div className="absolute inset-0 rounded-3xl bg-gradient-to-br from-primary/10 via-transparent to-primary/5 pointer-events-none" />
        <div className="glass-subtle rounded-3xl p-10 md:p-16 border border-border/30 relative">
          <h2 className="text-2xl md:text-4xl font-bold text-foreground tracking-tight mb-4">
            {t("ctaTitle")}
          </h2>
          <p className="text-muted-foreground text-sm md:text-base mb-8 max-w-md mx-auto">
            {t("ctaSubtitle")}
          </p>
          <button
            onClick={() => navigate("/auth")}
            className="inline-flex items-center gap-2 px-8 py-3.5 font-semibold text-sm rounded-xl btn-gradient text-primary-foreground"
          >
            {t("ctaButton")}
            <ArrowRight size={16} />
          </button>
        </div>
      </motion.div>
    </section>
  );
};

export default CTASection;
