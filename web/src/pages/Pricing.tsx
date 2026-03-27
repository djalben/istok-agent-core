import { useState } from "react";
import { motion } from "framer-motion";
import { Check, Zap, Crown, ArrowLeft, Coins } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useLanguage } from "@/hooks/useLanguage";
import NeuralBackground from "@/components/NeuralBackground";
import SBPModal from "@/components/SBPModal";

const Pricing = () => {
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
    <div className="min-h-screen bg-background relative overflow-hidden">
      <NeuralBackground />

      <div className="relative z-10">
        <header className="h-14 sticky top-0 glass border-b border-border/50 flex items-center px-4 md:px-6 z-50">
          <button
            onClick={() => navigate("/")}
            className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors text-sm"
          >
            <ArrowLeft size={16} />
            {t("back")}
          </button>
          <div className="flex-1 text-center font-bold text-lg tracking-tight text-foreground">
            {t("brand")}
          </div>
          <div className="w-16" />
        </header>

        <div className="max-w-4xl mx-auto px-4 py-16 md:py-24">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center mb-12 md:mb-16"
          >
            <h1 className="text-3xl md:text-5xl font-extrabold text-foreground tracking-tight mb-4 text-glow">
              {t("pricingTitle")}
            </h1>
            <p className="text-muted-foreground text-sm md:text-lg max-w-lg mx-auto">
              {t("pricingSubtitle")}
            </p>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 md:gap-8">
            {/* Старт */}
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="glass-subtle rounded-2xl p-8 md:p-10 border border-border/30 flex flex-col"
            >
              <div className="flex items-center gap-3 mb-8">
                <div className="w-12 h-12 rounded-xl bg-secondary flex items-center justify-center">
                  <Zap size={24} className="text-muted-foreground" />
                </div>
                <h2 className="text-2xl font-bold text-foreground">{t("pricingFree")}</h2>
              </div>

              <div className="mb-8">
                <span className="text-5xl font-extrabold text-foreground">{t("pricingFreePrice")}</span>
                <span className="text-muted-foreground text-sm ml-2">{t("pricingFreePeriod")}</span>
              </div>

              <ul className="space-y-4 mb-10 flex-1">
                {[t("pricingFreeFeature1"), t("pricingFreeFeature2"), t("pricingFreeFeature3"), t("pricingFreeFeature4")].map((f, i) => (
                  <li key={i} className="flex items-center gap-3 text-sm text-muted-foreground">
                    <Check size={18} className="text-primary shrink-0" />
                    {f}
                  </li>
                ))}
              </ul>

              <button
                onClick={() => navigate("/auth")}
                className="w-full py-3.5 rounded-xl text-sm font-semibold bg-secondary text-foreground hover:bg-secondary/80 transition-colors"
              >
                {t("pricingFreeCta")}
              </button>
            </motion.div>

            {/* Бизнес */}
            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.3 }}
              className="relative glass-subtle rounded-2xl p-8 md:p-10 border border-primary/30 flex flex-col shadow-[0_0_60px_hsla(243,76%,58%,0.1)]"
            >
              <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                <span className="px-4 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-primary text-primary-foreground">
                  {t("pricingPopular")}
                </span>
              </div>

              <div className="flex items-center gap-3 mb-8">
                <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
                  <Crown size={24} className="text-primary" />
                </div>
                <h2 className="text-2xl font-bold text-foreground">{t("pricingBusiness")}</h2>
              </div>

              <div className="mb-8">
                <span className="text-5xl font-extrabold text-foreground">{t("pricingBusinessPrice")}</span>
                <span className="text-muted-foreground text-sm ml-2">{t("pricingBusinessPeriod")}</span>
              </div>

              <ul className="space-y-4 mb-10 flex-1">
                {[t("pricingBusinessFeature1"), t("pricingBusinessFeature2"), t("pricingBusinessFeature3"), t("pricingBusinessFeature4"), t("pricingBusinessFeature5")].map((f, i) => (
                  <li key={i} className="flex items-center gap-3 text-sm text-muted-foreground">
                    <Check size={18} className="text-primary shrink-0" />
                    {f}
                  </li>
                ))}
              </ul>

              <button
                onClick={() => {
                  setSelectedPackage(null);
                  setSbpOpen(true);
                }}
                className="w-full py-3.5 rounded-xl text-sm font-semibold btn-gradient text-primary-foreground"
              >
                {t("pricingBusinessCta")}
              </button>
            </motion.div>
          </div>

          {/* Credit Packages */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
            className="mt-16"
          >
            <div className="text-center mb-8">
              <div className="flex items-center justify-center gap-2 mb-2">
                <Coins size={22} className="text-primary" />
                <h2 className="text-2xl font-bold text-foreground">{t("creditPackagesTitle")}</h2>
              </div>
              <p className="text-muted-foreground text-sm">{t("creditPackagesSubtitle")}</p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
              {creditPackages.map((pkg, i) => (
                <motion.button
                  key={i}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3, delay: 0.6 + i * 0.1 }}
                  onClick={() => handleBuyCredits(pkg)}
                  className="glass-subtle rounded-xl p-6 border border-border/30 hover:border-primary/40 transition-all text-left group"
                >
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-base font-semibold text-foreground">{pkg.name}</span>
                    <span className="text-xl font-extrabold text-primary">{pkg.price}</span>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {t("creditTokens", pkg.credits)}
                  </p>
                </motion.button>
              ))}
            </div>
          </motion.div>

          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5, delay: 0.8 }}
            className="text-center mt-10 text-sm text-muted-foreground/50"
          >
            {t("pricingContact")}
          </motion.p>
        </div>
      </div>

      <SBPModal
        open={sbpOpen}
        onClose={() => setSbpOpen(false)}
        packageInfo={selectedPackage}
      />
    </div>
  );
};

export default Pricing;
