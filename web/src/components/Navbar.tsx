import { useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Zap } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useLanguage } from "@/hooks/useLanguage";

const Navbar = () => {
  const [scrolled, setScrolled] = useState(false);
  const { user, signOut } = useAuth();
  const { t } = useLanguage();
  const navigate = useNavigate();

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <nav
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ${
        scrolled
          ? "bg-[rgba(9,9,11,0.85)] backdrop-blur-xl border-b border-white/5 shadow-[0_1px_40px_rgba(0,0,0,0.4)]"
          : "bg-transparent"
      }`}
    >
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2.5 group">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-violet-600 to-indigo-700 flex items-center justify-center shadow-[0_0_20px_rgba(124,58,237,0.4)] group-hover:shadow-[0_0_30px_rgba(124,58,237,0.6)] transition-all duration-300">
            <Zap size={16} className="text-white" fill="white" />
          </div>
          <span className="font-bold text-lg tracking-tight text-white">ИСТОК</span>
        </Link>

        {/* Links */}
        <div className="hidden md:flex items-center gap-8">
          {[
            { label: t("navFeatures") || "Возможности", href: "#features" },
            { label: t("navPricing") || "Цены", href: "/pricing" },
            { label: t("navDocs") || "Документация", href: "#" },
          ].map((item) => (
            <a
              key={item.label}
              href={item.href}
              className="text-sm text-white/60 hover:text-white transition-colors duration-200 font-medium"
            >
              {item.label}
            </a>
          ))}
        </div>

        {/* CTA */}
        <div className="flex items-center gap-3">
          {user ? (
            <>
              <button
                onClick={() => navigate("/project/new")}
                className="h-9 px-4 rounded-lg bg-white/5 border border-white/10 text-sm text-white hover:bg-white/10 transition-all duration-200 font-medium"
              >
                {t("navDashboard") || "Мои проекты"}
              </button>
              <button
                onClick={signOut}
                className="h-9 px-4 rounded-lg text-sm text-white/50 hover:text-white transition-colors duration-200"
              >
                {t("navLogout") || "Выйти"}
              </button>
            </>
          ) : (
            <>
              <Link
                to="/auth"
                className="h-9 px-4 rounded-lg text-sm text-white/60 hover:text-white transition-colors duration-200 font-medium"
              >
                {t("navLogin") || "Войти"}
              </Link>
              <Link
                to="/auth"
                className="h-9 px-5 rounded-lg bg-gradient-to-r from-violet-600 to-indigo-600 text-sm text-white font-semibold hover:from-violet-500 hover:to-indigo-500 transition-all duration-200 shadow-[0_0_20px_rgba(124,58,237,0.35)] hover:shadow-[0_0_30px_rgba(124,58,237,0.5)] flex items-center"
              >
                {t("navStart") || "Начать бесплатно"}
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
