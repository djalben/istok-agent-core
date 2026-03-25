'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { PROJECT_TEMPLATES, ProjectTemplate } from '@/lib/templates';

interface TemplateSelectorProps {
  onSelectTemplate: (template: ProjectTemplate) => void;
}

export function TemplateSelector({ onSelectTemplate }: TemplateSelectorProps) {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  const categories = [
    { id: 'all', name: 'Все шаблоны', icon: '📦' },
    { id: 'crm', name: 'CRM', icon: '📊' },
    { id: 'messenger', name: 'Мессенджеры', icon: '💬' },
    { id: 'bot', name: 'Боты', icon: '🤖' },
    { id: 'landing', name: 'Landing', icon: '🚀' },
    { id: 'dashboard', name: 'Дашборды', icon: '⚙️' },
    { id: 'ecommerce', name: 'E-commerce', icon: '🛒' },
  ];

  const filteredTemplates =
    selectedCategory === 'all'
      ? PROJECT_TEMPLATES
      : PROJECT_TEMPLATES.filter((t) => t.category === selectedCategory);

  return (
    <div className="space-y-6">
      {/* Category filters */}
      <div className="flex flex-wrap gap-2">
        {categories.map((category) => (
          <motion.div key={category.id} whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
            <Button
              variant={selectedCategory === category.id ? 'default' : 'outline'}
              onClick={() => setSelectedCategory(category.id)}
              className={
                selectedCategory === category.id
                  ? 'gradient-indigo-violet text-white'
                  : 'glass border-white/20 text-white hover:bg-white/10'
              }
            >
              <span className="mr-2">{category.icon}</span>
              {category.name}
            </Button>
          </motion.div>
        ))}
      </div>

      {/* Templates grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredTemplates.map((template, index) => (
          <motion.div
            key={template.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <Card className="glass-strong border-white/10 hover:border-indigo-500/50 transition-all h-full flex flex-col">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="text-4xl mb-2">{template.icon}</div>
                  <Badge className="glass border-white/20 text-xs">
                    {template.category.toUpperCase()}
                  </Badge>
                </div>
                <CardTitle className="text-xl text-white">{template.name}</CardTitle>
                <CardDescription className="text-zinc-400">
                  {template.description}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1 flex flex-col justify-between">
                <div className="space-y-4 mb-4">
                  <div>
                    <p className="text-xs text-zinc-400 mb-2 font-medium">Возможности:</p>
                    <div className="flex flex-wrap gap-1">
                      {template.features.slice(0, 3).map((feature, i) => (
                        <Badge
                          key={i}
                          variant="outline"
                          className="text-xs border-white/10 text-zinc-300"
                        >
                          {feature}
                        </Badge>
                      ))}
                      {template.features.length > 3 && (
                        <Badge
                          variant="outline"
                          className="text-xs border-white/10 text-zinc-400"
                        >
                          +{template.features.length - 3}
                        </Badge>
                      )}
                    </div>
                  </div>
                  <div>
                    <p className="text-xs text-zinc-400 mb-2 font-medium">Технологии:</p>
                    <div className="flex flex-wrap gap-1">
                      {template.techStack.slice(0, 4).map((tech, i) => (
                        <span key={i} className="text-xs text-zinc-500">
                          {tech}
                          {i < Math.min(3, template.techStack.length - 1) && ' •'}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>
                <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
                  <Button
                    onClick={() => onSelectTemplate(template)}
                    className="w-full gradient-indigo-violet text-white font-semibold"
                  >
                    ✨ Использовать шаблон
                  </Button>
                </motion.div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
