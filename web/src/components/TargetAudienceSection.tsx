import { motion } from "framer-motion";
import { ShoppingBag, Store, Palette } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

const item = {
  hidden: { opacity: 0, y: 24 },
  show: { opacity: 1, y: 0, transition: { duration: 0.45, ease: "easeOut" as const } },
};

const TargetAudienceSection = () => {
  const { t } = useLanguage();

  const audiences = [
    {
      icon: ShoppingBag,
      title: t("targetSellers"),
      description: t("targetSellersDesc"),
      gradient: "from-orange-500/20 to-amber-500/10",
      screenshot: "CRM",
    },
    {
      icon: Store,
      title: t("targetSmallBiz"),
      description: t("targetSmallBizDesc"),
      gradient: "from-blue-500/20 to-cyan-500/10",
      screenshot: "Dashboard",
    },
    {
      icon: Palette,
      title: t("targetFreelance"),
      description: t("targetFreelanceDesc"),
      gradient: "from-violet-500/20 to-purple-500/10",
      screenshot: "Portfolio",
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
          {t("targetTitle")}
        </h2>
        <p className="text-muted-foreground text-sm md:text-base max-w-lg mx-auto">
          {t("targetSubtitle")}
        </p>
      </motion.div>

      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.1 }}
        transition={{ staggerChildren: 0.15 }}
        className="max-w-5xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-6"
      >
        {audiences.map((a, i) => (
          <motion.div
            key={i}
            variants={item}
            className="group relative glass-subtle rounded-2xl overflow-hidden border border-border/30 hover:border-primary/20 transition-all duration-300 hover:-translate-y-1"
          >
            {/* Screenshot placeholder */}
            <div className={`h-40 bg-gradient-to-br ${a.gradient} flex items-center justify-center relative overflow-hidden`}>
              <div className="absolute inset-0 bg-background/30" />
              <div className="relative z-10 w-[85%] h-[75%] rounded-lg bg-background/60 backdrop-blur-sm border border-border/20 flex flex-col p-3 gap-2">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-destructive/60" />
                  <div className="w-2 h-2 rounded-full bg-yellow-500/60" />
                  <div className="w-2 h-2 rounded-full bg-green-500/60" />
                  <div className="flex-1 h-3 rounded bg-muted/30 ml-2" />
                </div>
                <div className="flex-1 flex gap-2">
                  <div className="w-1/4 rounded bg-muted/20 space-y-1.5 p-1.5">
                    <div className="h-1.5 w-full rounded-full bg-primary/20" />
                    <div className="h-1.5 w-3/4 rounded-full bg-muted/30" />
                    <div className="h-1.5 w-full rounded-full bg-muted/30" />
                  </div>
                  <div className="flex-1 rounded bg-muted/10 p-1.5 space-y-1.5">
                    <div className="h-2 w-1/2 rounded bg-primary/15" />
                    <div className="h-1.5 w-full rounded-full bg-muted/20" />
                    <div className="h-1.5 w-4/5 rounded-full bg-muted/15" />
                    <div className="h-6 w-1/3 rounded bg-primary/10 mt-2" />
                  </div>
                </div>
              </div>
            </div>
            {/* Content */}
            <div className="p-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
                  <a.icon size={20} className="text-primary" />
                </div>
                <h3 className="text-base font-semibold text-foreground">{a.title}</h3>
              </div>
              <p className="text-sm text-muted-foreground leading-relaxed">{a.description}</p>
            </div>
          </motion.div>
        ))}
      </motion.div>
    </section>
  );
};

export default TargetAudienceSection;
