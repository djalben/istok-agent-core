import { useState } from "react";
import { motion } from "framer-motion";
import { Check, Zap, Crown, Coins } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useLanguage } from "@/hooks/useLanguage";
import SBPModal from "@/components/SBPModal";

const PricingSection = () => {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const [sbpOpen, setSbpOpen] = useState(false);
  const [selectedPackage, setSelectedPackage] = useState<{ name: string; amount: number; credits: number } | null>(null);

  const creditPackages = [
    { name: t("creditStarter"), credits: 100000, price: t("creditStarterPrice"), amount: 990 },
    { name: t("creditPro"), credits: 500000, price: t("creditProPrice"), amount: 2990 },
  ];

  const handleBuyCredits = (pkg: typeof creditPackages[0]) => {
    setSelectedPackage({ name: pkg.name, amount: pkg.amount, credits: pkg.credits });
    setSbpOpen(true);
  };

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
          {t("pricingTitle")}
        </h2>
        <p className="text-muted-foreground text-sm md:text-base max-w-lg mx-auto">
          {t("pricingSubtitle")}
        </p>
      </motion.div>

      <div className="max-w-3xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Free Plan */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="glass-subtle rounded-2xl p-6 md:p-8 border border-border/30 flex flex-col"
        >
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-xl bg-secondary flex items-center justify-center">
              <Zap size={20} className="text-muted-foreground" />
            </div>
            <h3 className="text-lg font-bold text-foreground">{t("pricingFree")}</h3>
          </div>

          <div className="mb-6">
            <span className="text-4xl font-extrabold text-foreground">{t("pricingFreePrice")}</span>
            <span className="text-muted-foreground text-sm ml-2">{t("pricingFreePeriod")}</span>
          </div>

          <ul className="space-y-3 mb-8 flex-1">
            {[t("pricingFreeFeature1"), t("pricingFreeFeature2"), t("pricingFreeFeature3"), t("pricingFreeFeature4")].map((f, i) => (
              <li key={i} className="flex items-center gap-2.5 text-sm text-muted-foreground">
                <Check size={16} className="text-primary shrink-0" />
                {f}
              </li>
            ))}
          </ul>

          <button
            onClick={() => navigate("/auth")}
            className="w-full py-3 rounded-xl text-sm font-semibold bg-secondary text-foreground hover:bg-secondary/80 transition-colors"
          >
            {t("pricingFreeCta")}
          </button>
        </motion.div>

        {/* Business Plan */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.2 }}
          className="relative glass-subtle rounded-2xl p-6 md:p-8 border border-primary/30 flex flex-col shadow-[0_0_40px_hsla(243,76%,58%,0.08)]"
        >
          <div className="absolute -top-3 left-1/2 -translate-x-1/2">
            <span className="px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-primary text-primary-foreground">
              {t("pricingPopular")}
            </span>
          </div>

          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
              <Crown size={20} className="text-primary" />
            </div>
            <h3 className="text-lg font-bold text-foreground">{t("pricingBusiness")}</h3>
          </div>

          <div className="mb-6">
            <span className="text-4xl font-extrabold text-foreground">{t("pricingBusinessPrice")}</span>
            <span className="text-muted-foreground text-sm ml-2">{t("pricingBusinessPeriod")}</span>
          </div>

          <ul className="space-y-3 mb-8 flex-1">
            {[t("pricingBusinessFeature1"), t("pricingBusinessFeature2"), t("pricingBusinessFeature3"), t("pricingBusinessFeature4"), t("pricingBusinessFeature5")].map((f, i) => (
              <li key={i} className="flex items-center gap-2.5 text-sm text-muted-foreground">
                <Check size={16} className="text-primary shrink-0" />
                {f}
              </li>
            ))}
          </ul>

          <button
            onClick={() => {
              setSelectedPackage(null);
              setSbpOpen(true);
            }}
            className="w-full py-3 rounded-xl text-sm font-semibold btn-gradient text-primary-foreground"
          >
            {t("pricingBusinessCta")}
          </button>
        </motion.div>
      </div>

      {/* Credit Packages */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5, delay: 0.3 }}
        className="max-w-3xl mx-auto mt-12"
      >
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-2 mb-2">
            <Coins size={20} className="text-primary" />
            <h3 className="text-xl font-bold text-foreground">{t("creditPackagesTitle")}</h3>
          </div>
          <p className="text-muted-foreground text-sm">{t("creditPackagesSubtitle")}</p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {creditPackages.map((pkg, i) => (
            <motion.button
              key={i}
              initial={{ opacity: 0, y: 10 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.3, delay: 0.4 + i * 0.1 }}
              onClick={() => handleBuyCredits(pkg)}
              className="glass-subtle rounded-xl p-5 border border-border/30 hover:border-primary/40 transition-all text-left group"
            >
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-semibold text-foreground">{pkg.name}</span>
                <span className="text-lg font-extrabold text-primary">{pkg.price}</span>
              </div>
              <p className="text-xs text-muted-foreground">
                {t("creditTokens", pkg.credits)}
              </p>
            </motion.button>
          ))}
        </div>
      </motion.div>

      <motion.p
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5, delay: 0.5 }}
        className="text-center mt-8 text-xs text-muted-foreground/50"
      >
        {t("pricingContact")}
      </motion.p>

      <SBPModal
        open={sbpOpen}
        onClose={() => setSbpOpen(false)}
        packageInfo={selectedPackage}
      />
    </section>
  );
};

export default PricingSection;
