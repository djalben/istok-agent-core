import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import { Coffee, ShoppingBag, UserPlus, LayoutDashboard, ShoppingCart, FileText, Bot, BarChart3, CalendarDays, MessageCircle } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

const item = {
  hidden: { opacity: 0, y: 24 },
  show: { opacity: 1, y: 0, transition: { duration: 0.45, ease: "easeOut" as const } },
};

const TemplatesSection = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();

  const popularTemplates = [
    { icon: Coffee, title: t("tplCafeLanding"), description: t("tplCafeLandingDesc"), prompt: "Современный одностраничный сайт для кофейни в Москве с меню и картой", tag: t("tagLanding") },
    { icon: ShoppingBag, title: t("tplProductCard"), description: t("tplProductCardDesc"), prompt: "Стильный блок карточки товара для маркетплейса с кнопкой покупки и выбором размера", tag: t("tagEcommerce") },
    { icon: UserPlus, title: t("tplRegForm"), description: t("tplRegFormDesc"), prompt: "Чистая форма входа и регистрации с валидацией полей на русском языке", tag: t("tagForms") },
    { icon: LayoutDashboard, title: t("tplDashboard"), description: t("tplDashboardDesc"), prompt: "Дашборд пользователя с графиками и списком заказов в темной теме", tag: t("tagDashboard") },
  ];

  const moreTemplates = [
    { icon: ShoppingCart, title: t("tplShop"), description: t("tplShopDesc"), tag: t("tagEcommerce"), prompt: "Интернет-магазин с каталогом товаров, корзиной и формой оплаты" },
    { icon: FileText, title: t("tplLanding"), description: t("tplLandingDesc"), tag: t("tagMarketing"), prompt: "Продающая лендинг-страница с призывом к действию и формой заявки" },
    { icon: Bot, title: t("tplChatbot"), description: t("tplChatbotDesc"), tag: t("tagAI"), prompt: "Интерфейс чат-бота с диалоговым окном и вводом сообщений" },
    { icon: BarChart3, title: t("tplCRM"), description: t("tplCRMDesc"), tag: t("tagBusiness"), prompt: "CRM-система для управления клиентами и сделками с таблицами и фильтрами" },
    { icon: CalendarDays, title: t("tplPlanner"), description: t("tplPlannerDesc"), tag: t("tagProductivity"), prompt: "Планировщик задач с календарем и дедлайнами" },
    { icon: MessageCircle, title: t("tplMessenger"), description: t("tplMessengerDesc"), tag: t("tagSocial"), prompt: "Мессенджер с чатом в реальном времени и списком контактов" },
  ];

  const handleTemplateClick = (prompt: string) => {
    navigate("/project/new", { state: { prompt } });
  };

  return (
    <section className="py-20 md:py-32 px-4 md:px-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.5 }}
        className="text-center mb-12 md:mb-16"
      >
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-primary/20 bg-primary/5 mb-4">
          <span className="text-[10px] font-semibold tracking-wider uppercase text-primary">{t("templatesPopular")}</span>
        </div>
        <h2 className="text-2xl md:text-3xl font-bold text-foreground tracking-tight mb-3">
          {t("templatesTitle")}
        </h2>
        <p className="text-muted-foreground text-sm max-w-md mx-auto">
          {t("templatesSubtitle")}
        </p>
      </motion.div>

      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.1 }}
        transition={{ staggerChildren: 0.1 }}
        className="max-w-5xl mx-auto grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8"
      >
        {popularTemplates.map((tpl, i) => (
          <motion.button
            key={i}
            variants={item}
            onClick={() => handleTemplateClick(tpl.prompt)}
            className="group relative glass-subtle rounded-2xl p-5 md:p-6 text-left hover:-translate-y-1 hover:shadow-[0_0_40px_hsla(243,76%,58%,0.12)] transition-all duration-300 cursor-pointer border border-border/30 hover:border-primary/30"
          >
            <div className="absolute inset-0 rounded-2xl bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none" />
            <div className="relative z-10">
              <div className="flex items-start justify-between mb-4">
                <div className="w-11 h-11 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors duration-300">
                  <tpl.icon size={22} className="text-primary" />
                </div>
                <span className="text-[10px] font-semibold tracking-wider uppercase text-muted-foreground/50 border border-border/50 rounded-full px-2.5 py-0.5">
                  {tpl.tag}
                </span>
              </div>
              <h3 className="text-base font-semibold text-foreground mb-1.5">{tpl.title}</h3>
              <p className="text-sm text-muted-foreground mb-3">{tpl.description}</p>
              <span className="text-[11px] text-primary/70 group-hover:text-primary transition-colors">
                {t("templateClick")}
              </span>
            </div>
          </motion.button>
        ))}
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.5 }}
        className="text-center mb-8 mt-12 md:mt-16"
      >
        <h3 className="text-xl font-semibold text-foreground tracking-tight mb-2">
          {t("templatesMore")}
        </h3>
      </motion.div>

      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.1 }}
        transition={{ staggerChildren: 0.08 }}
        className="max-w-5xl mx-auto grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        {moreTemplates.map((tpl, i) => (
          <motion.button
            key={i}
            variants={item}
            onClick={() => handleTemplateClick(tpl.prompt)}
            className="group relative glass-subtle rounded-2xl p-5 md:p-6 text-left hover:-translate-y-1 hover:shadow-[0_0_40px_hsla(243,76%,58%,0.08)] transition-all duration-300 cursor-pointer"
          >
            <div className="flex items-start justify-between mb-4">
              <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors duration-300">
                <tpl.icon size={20} className="text-primary" />
              </div>
              <span className="text-[10px] font-semibold tracking-wider uppercase text-muted-foreground/50 border border-border/50 rounded-full px-2.5 py-0.5">
                {tpl.tag}
              </span>
            </div>
            <h3 className="text-base font-semibold text-foreground mb-1">{tpl.title}</h3>
            <p className="text-sm text-muted-foreground">{tpl.description}</p>
          </motion.button>
        ))}
      </motion.div>
    </section>
  );
};

export default TemplatesSection;
