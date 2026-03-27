// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  SUPABASE CLIENT - ОТКЛЮЧЕН
//  Переход на Go Auth API
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

import type { Database } from './types';

// Заглушка для обратной совместимости
// Все функции возвращают пустые результаты
export const supabase = {
  from: (table: string) => ({
    select: () => ({ data: null, error: new Error('Supabase отключен - используйте Go API') }),
    insert: () => ({ data: null, error: new Error('Supabase отключен - используйте Go API') }),
    update: () => ({ data: null, error: new Error('Supabase отключен - используйте Go API') }),
    delete: () => ({ data: null, error: new Error('Supabase отключен - используйте Go API') }),
    eq: function() { return this; },
    single: function() { return this; },
    maybeSingle: function() { return this; },
    limit: function() { return this; },
    order: function() { return this; },
  }),
  functions: {
    invoke: () => Promise.resolve({ data: null, error: new Error('Supabase отключен - используйте Go API') }),
  },
  auth: {
    signInWithPassword: () => Promise.resolve({ data: null, error: new Error('Используйте Go Auth API') }),
    signUp: () => Promise.resolve({ data: null, error: new Error('Используйте Go Auth API') }),
    signOut: () => Promise.resolve({ error: null }),
    getSession: () => Promise.resolve({ data: { session: null }, error: null }),
    onAuthStateChange: () => ({ data: { subscription: { unsubscribe: () => {} } } }),
    resetPasswordForEmail: () => Promise.resolve({ error: new Error('Используйте Go Auth API') }),
  },
} as any;