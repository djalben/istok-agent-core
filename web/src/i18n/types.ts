// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — i18n shared types
//  Изолированы в отдельном файле, чтобы избежать circular import
//  между useLanguage.tsx и dictionary-файлами ru.ts / en.ts.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

/** Allowed argument types for translation functions (`(arg) => string`). */
export type TranslationArg = string | number | boolean | null | undefined;

/** Translation function signature. */
export type TranslationFn = (arg: TranslationArg) => string;

/** Dictionary shape — keys map to either a literal string or a function template. */
export type Dict = Record<string, string | TranslationFn>;
