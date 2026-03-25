export interface ProjectTemplate {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: 'crm' | 'messenger' | 'bot' | 'landing' | 'dashboard' | 'ecommerce';
  prompt: string;
  features: string[];
  techStack: string[];
}

export const PROJECT_TEMPLATES: ProjectTemplate[] = [
  {
    id: 'crm-modern',
    name: 'Современная CRM',
    description: 'Полнофункциональная CRM-система с управлением клиентами, сделками и аналитикой',
    icon: '📊',
    category: 'crm',
    prompt: `Создай современную CRM-систему на React с TypeScript. Включи:
- Дашборд с графиками и метриками (использовать Recharts)
- Управление клиентами (таблица с поиском, фильтрами, пагинацией)
- Управление сделками (Kanban-доска с drag-and-drop)
- Календарь встреч и задач
- Аналитику продаж (графики, воронка продаж)
- Темную и светлую тему
- Адаптивный дизайн
- Использовать Tailwind CSS, shadcn/ui, Lucide icons
- Добавить анимации с Framer Motion`,
    features: [
      'Управление клиентами',
      'Kanban-доска сделок',
      'Аналитика и отчеты',
      'Календарь задач',
      'Темная тема',
    ],
    techStack: ['React', 'TypeScript', 'Tailwind CSS', 'shadcn/ui', 'Recharts', 'Framer Motion'],
  },
  {
    id: 'messenger-realtime',
    name: 'Мессенджер в реальном времени',
    description: 'Современный чат с поддержкой групп, файлов и эмодзи',
    icon: '💬',
    category: 'messenger',
    prompt: `Создай современный мессенджер на React с TypeScript. Включи:
- Список чатов с поиском и фильтрацией
- Окно переписки с пузырьками сообщений
- Поддержка текста, эмодзи, файлов
- Индикатор "печатает..."
- Групповые чаты
- Профили пользователей
- Темная тема с glassmorphism
- Адаптивный дизайн для мобильных
- Использовать Tailwind CSS, shadcn/ui, Lucide icons
- Анимации сообщений с Framer Motion`,
    features: [
      'Личные и групповые чаты',
      'Отправка файлов',
      'Эмодзи и стикеры',
      'Typing indicator',
      'Glassmorphism дизайн',
    ],
    techStack: ['React', 'TypeScript', 'Tailwind CSS', 'shadcn/ui', 'Framer Motion'],
  },
  {
    id: 'telegram-bot-dashboard',
    name: 'Панель управления Telegram-ботом',
    description: 'Дашборд для настройки и мониторинга Telegram-бота',
    icon: '🤖',
    category: 'bot',
    prompt: `Создай панель управления Telegram-ботом на React с TypeScript. Включи:
- Дашборд со статистикой (пользователи, сообщения, активность)
- Конструктор команд бота (визуальный редактор)
- Настройка автоответов и сценариев
- Управление рассылками
- Аналитика взаимодействий (графики)
- Логи сообщений в реальном времени
- Настройки бота (токен, webhook)
- Темная тема
- Использовать Tailwind CSS, shadcn/ui, Lucide icons
- Добавить анимации переходов`,
    features: [
      'Статистика бота',
      'Конструктор команд',
      'Автоответы',
      'Рассылки',
      'Аналитика',
    ],
    techStack: ['React', 'TypeScript', 'Tailwind CSS', 'shadcn/ui', 'Recharts'],
  },
  {
    id: 'landing-saas',
    name: 'Landing Page для SaaS',
    description: 'Современный лендинг с анимациями и формой подписки',
    icon: '🚀',
    category: 'landing',
    prompt: `Создай премиум landing page для SaaS-продукта на React с TypeScript. Включи:
- Hero-секция с градиентами и анимациями
- Блок с преимуществами (features grid)
- Секция "Как это работает" (steps)
- Pricing таблица (3 тарифа)
- Отзывы клиентов (карусель)
- FAQ (аккордеон)
- Форма подписки с валидацией
- Футер с ссылками
- Glassmorphism эффекты
- Плавные анимации при скролле
- Адаптивный дизайн
- Использовать Tailwind CSS, shadcn/ui, Lucide icons, Framer Motion`,
    features: [
      'Hero с анимациями',
      'Features grid',
      'Pricing таблица',
      'Отзывы',
      'FAQ',
      'Форма подписки',
    ],
    techStack: ['React', 'TypeScript', 'Tailwind CSS', 'shadcn/ui', 'Framer Motion'],
  },
  {
    id: 'admin-dashboard',
    name: 'Админ-панель',
    description: 'Универсальная админ-панель с таблицами и графиками',
    icon: '⚙️',
    category: 'dashboard',
    prompt: `Создай универсальную админ-панель на React с TypeScript. Включи:
- Sidebar с навигацией (сворачиваемый)
- Дашборд с KPI-метриками (карточки)
- Таблица данных (сортировка, фильтры, пагинация)
- Графики и диаграммы (линейные, круговые, столбчатые)
- Формы создания/редактирования записей
- Модальные окна
- Уведомления (toast)
- Профиль пользователя
- Темная и светлая тема
- Адаптивный дизайн
- Использовать Tailwind CSS, shadcn/ui, Lucide icons, Recharts`,
    features: [
      'Сворачиваемый Sidebar',
      'KPI метрики',
      'Таблицы с фильтрами',
      'Графики',
      'Формы',
      'Темная тема',
    ],
    techStack: ['React', 'TypeScript', 'Tailwind CSS', 'shadcn/ui', 'Recharts'],
  },
  {
    id: 'ecommerce-store',
    name: 'Интернет-магазин',
    description: 'Современный e-commerce с корзиной и каталогом',
    icon: '🛒',
    category: 'ecommerce',
    prompt: `Создай интернет-магазин на React с TypeScript. Включи:
- Каталог товаров (grid с карточками)
- Фильтры и сортировка товаров
- Страница товара (галерея, описание, характеристики)
- Корзина покупок (добавление, удаление, изменение количества)
- Форма оформления заказа
- Поиск товаров
- Избранное
- Адаптивный дизайн
- Темная тема
- Анимации добавления в корзину
- Использовать Tailwind CSS, shadcn/ui, Lucide icons, Framer Motion`,
    features: [
      'Каталог товаров',
      'Фильтры и поиск',
      'Корзина',
      'Оформление заказа',
      'Избранное',
      'Анимации',
    ],
    techStack: ['React', 'TypeScript', 'Tailwind CSS', 'shadcn/ui', 'Framer Motion'],
  },
];

export function getTemplateById(id: string): ProjectTemplate | undefined {
  return PROJECT_TEMPLATES.find((template) => template.id === id);
}

export function getTemplatesByCategory(category: ProjectTemplate['category']): ProjectTemplate[] {
  return PROJECT_TEMPLATES.filter((template) => template.category === category);
}
