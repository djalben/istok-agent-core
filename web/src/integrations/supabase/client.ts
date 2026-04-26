// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  SUPABASE CLIENT — ОТКЛЮЧЕН (миграция на Go Auth API)
//  Заглушка сохраняет совместимость с legacy-импортами.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const DISABLED_ERROR = new Error('Supabase отключен — используйте Go API');
const AUTH_DISABLED_ERROR = new Error('Используйте Go Auth API');

interface DisabledQueryBuilder {
  select: () => { data: null; error: Error };
  insert: () => { data: null; error: Error };
  update: () => { data: null; error: Error };
  delete: () => { data: null; error: Error };
  eq: () => DisabledQueryBuilder;
  single: () => DisabledQueryBuilder;
  maybeSingle: () => DisabledQueryBuilder;
  limit: () => DisabledQueryBuilder;
  order: () => DisabledQueryBuilder;
}

interface DisabledSupabase {
  from: (table: string) => DisabledQueryBuilder;
  functions: {
    invoke: () => Promise<{ data: null; error: Error }>;
  };
  auth: {
    signInWithPassword: () => Promise<{ data: null; error: Error }>;
    signUp: () => Promise<{ data: null; error: Error }>;
    signOut: () => Promise<{ error: null }>;
    getSession: () => Promise<{ data: { session: null }; error: null }>;
    onAuthStateChange: () => { data: { subscription: { unsubscribe: () => void } } };
    resetPasswordForEmail: () => Promise<{ error: Error }>;
  };
}

const builder: DisabledQueryBuilder = {
  select: () => ({ data: null, error: DISABLED_ERROR }),
  insert: () => ({ data: null, error: DISABLED_ERROR }),
  update: () => ({ data: null, error: DISABLED_ERROR }),
  delete: () => ({ data: null, error: DISABLED_ERROR }),
  eq: () => builder,
  single: () => builder,
  maybeSingle: () => builder,
  limit: () => builder,
  order: () => builder,
};

export const supabase: DisabledSupabase = {
  from: () => builder,
  functions: {
    invoke: () => Promise.resolve({ data: null, error: DISABLED_ERROR }),
  },
  auth: {
    signInWithPassword: () => Promise.resolve({ data: null, error: AUTH_DISABLED_ERROR }),
    signUp: () => Promise.resolve({ data: null, error: AUTH_DISABLED_ERROR }),
    signOut: () => Promise.resolve({ error: null }),
    getSession: () => Promise.resolve({ data: { session: null }, error: null }),
    onAuthStateChange: () => ({ data: { subscription: { unsubscribe: () => {} } } }),
    resetPasswordForEmail: () => Promise.resolve({ error: AUTH_DISABLED_ERROR }),
  },
};