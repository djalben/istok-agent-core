import { Moon, Sun, Globe, Coins } from "lucide-react";
import { useNavigate } from "react-router-dom";
import UserMenu from "@/components/UserMenu";
import { useTheme } from "@/hooks/useTheme";
import { useLanguage } from "@/hooks/useLanguage";
import { useAuth } from "@/hooks/useAuth";
import { useCredits } from "@/hooks/useCredits";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
} from "@/components/ui/tooltip";

const HeaderBar = () => {
  const { theme, toggleTheme } = useTheme();
  const { lang, setLang, t } = useLanguage();
  const { user } = useAuth();
  const { credits } = useCredits();
  const navigate = useNavigate();

  const formatCredits = (n: number) => {
    if (n >= 1000) return `${(n / 1000).toFixed(n >= 10000 ? 0 : 1)}k`;
    return n.toString();
  };

  return (
    <header className="h-14 sticky top-0 glass border-b border-border/50 flex items-center justify-between px-4 md:px-6 z-50">
      <div className="font-bold text-lg tracking-tight text-foreground">
        {t("brand")}
      </div>

      <div className="flex items-center gap-1.5">
        {/* Pricing link */}
        <button
          onClick={() => navigate("/pricing")}
          className="h-8 px-3 rounded-lg flex items-center text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-all duration-200 text-xs font-medium"
        >
          {t("pricing")}
        </button>

        {/* Language switcher */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="h-8 px-2 rounded-lg flex items-center gap-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-all duration-200 text-xs font-medium">
              <Globe size={14} />
              <span className="hidden sm:inline">{lang.toUpperCase()}</span>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-32 bg-card border-border/50">
            <DropdownMenuItem onClick={() => setLang("ru")} className={`text-sm cursor-pointer ${lang === "ru" ? "text-primary" : ""}`}>
              🇷🇺 Русский
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setLang("en")} className={`text-sm cursor-pointer ${lang === "en" ? "text-primary" : ""}`}>
              🇺🇸 English
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Theme toggle */}
        <button
          onClick={toggleTheme}
          className="w-8 h-8 rounded-lg flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-all duration-200"
        >
          {theme === "dark" ? <Moon size={15} /> : <Sun size={15} />}
        </button>

        {user && credits !== null && (
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                onClick={() => navigate("/pricing")}
                className="h-8 px-2.5 rounded-lg flex items-center gap-1.5 hover:bg-secondary/50 transition-all duration-200 text-xs font-medium"
              >
                <Coins size={14} className="text-primary" />
                <span className={credits < 5000 ? "text-destructive font-semibold" : "text-foreground"}>
                  {formatCredits(credits)} {t("creditsTokens")}
                </span>
              </button>
            </TooltipTrigger>
            <TooltipContent side="bottom">
              <p>{t("creditsTooltip")}</p>
            </TooltipContent>
          </Tooltip>
        )}

        <UserMenu />
      </div>
    </header>
  );
};

export default HeaderBar;
