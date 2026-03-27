import { useLanguage } from "@/hooks/useLanguage";
import { Globe } from "lucide-react";

const FooterSection = () => {
  const { lang, setLang, t } = useLanguage();

  return (
    <footer className="border-t border-border/50 py-12 px-4 md:px-6">
      <div className="max-w-5xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="text-sm text-muted-foreground/60">
          {t("footerCopyright")}
        </div>
        <div className="flex items-center gap-6 text-sm text-muted-foreground/40">
          <span className="hover:text-foreground transition-colors cursor-pointer">{t("footerDocs")}</span>
          <span className="hover:text-foreground transition-colors cursor-pointer">{t("footerBlog")}</span>
          <span className="hover:text-foreground transition-colors cursor-pointer">{t("footerSupport")}</span>
          <div className="flex items-center gap-1.5 border-l border-border/30 pl-6">
            <Globe size={14} className="text-muted-foreground/40" />
            <button
              onClick={() => setLang("ru")}
              className={`text-xs px-1.5 py-0.5 rounded transition-colors ${lang === "ru" ? "text-primary font-medium" : "hover:text-foreground"}`}
            >
              RU
            </button>
            <span className="text-muted-foreground/20">/</span>
            <button
              onClick={() => setLang("en")}
              className={`text-xs px-1.5 py-0.5 rounded transition-colors ${lang === "en" ? "text-primary font-medium" : "hover:text-foreground"}`}
            >
              EN
            </button>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default FooterSection;
