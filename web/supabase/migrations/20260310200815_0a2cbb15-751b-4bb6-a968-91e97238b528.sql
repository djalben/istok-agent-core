CREATE TABLE IF NOT EXISTS public.admin_finances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  total_deposited_usd numeric NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE public.admin_finances ENABLE ROW LEVEL SECURITY;

-- Only the admin can read/write
CREATE POLICY "Admin only read" ON public.admin_finances FOR SELECT TO authenticated
  USING (auth.uid() = '9cab2dcf-e25d-490e-99f4-f2df9a925a02'::uuid);
CREATE POLICY "Admin only update" ON public.admin_finances FOR UPDATE TO authenticated
  USING (auth.uid() = '9cab2dcf-e25d-490e-99f4-f2df9a925a02'::uuid);
CREATE POLICY "Admin only insert" ON public.admin_finances FOR INSERT TO authenticated
  WITH CHECK (auth.uid() = '9cab2dcf-e25d-490e-99f4-f2df9a925a02'::uuid);

-- Seed initial row
INSERT INTO public.admin_finances (total_deposited_usd) VALUES (0);