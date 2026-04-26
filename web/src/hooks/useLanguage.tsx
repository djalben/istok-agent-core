import { createContext, useContext, useState, type ReactNode } from "react";
import { ru } from "@/i18n/ru";
import { en } from "@/i18n/en";
import type { Dict, TranslationArg } from "@/i18n/types";

export type { TranslationArg } from "@/i18n/types";

type Lang = "ru" | "en";

interface LanguageContextType {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: string, arg?: TranslationArg) => string;
}

const dictionaries: Record<Lang, Dict> = { ru, en };

const LanguageContext = createContext<LanguageContextType>({
  lang: "ru",
  setLang: () => {},
  t: (key: string) => key,
});

export const useLanguage = () => useContext(LanguageContext);

export const LanguageProvider = ({ children }: { children: ReactNode }) => {
  const [lang, setLangState] = useState<Lang>(() => {
    if (typeof window !== "undefined") {
      return (localStorage.getItem("istok-lang") as Lang) || "ru";
    }
    return "ru";
  });

  const setLang = (l: Lang) => {
    setLangState(l);
    localStorage.setItem("istok-lang", l);
  };

  const t = (key: string, arg?: TranslationArg): string => {
    const val = dictionaries[lang][key];
    if (typeof val === "function") return val(arg);
    if (typeof val === "string") return val;
    return key;
  };

  return (
    <LanguageContext.Provider value={{ lang, setLang, t }}>
      {children}
    </LanguageContext.Provider>
  );
};
