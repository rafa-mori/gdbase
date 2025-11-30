-- ============================================================================
-- ETAPA 5: Business Entities
-- ============================================================================
-- Cria tabelas de domínio de negócio (pipelines, partners, leads, commissions)
-- ============================================================================

\echo '🚀 ETAPA 5: Criando entidades de negócio...'

-- Pipelines
CREATE TABLE IF NOT EXISTS pipeline (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela pipeline criada'

-- Pipeline Stages
CREATE TABLE IF NOT EXISTS pipeline_stage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES pipeline(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    order_index INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela pipeline_stage criada'

-- Partners
CREATE TABLE IF NOT EXISTS partner (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,

    -- Dados da empresa parceira
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    cnpj VARCHAR(14),

    -- Relacionamento
    tier TEXT,
    status TEXT,

    -- Usuário de contato principal
    primary_contact_user_id UUID REFERENCES "user"(id) ON DELETE SET NULL,

    -- Auditoria
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela partner criada'

-- Partner Updates (histórico)
CREATE TABLE IF NOT EXISTS partner_update (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partner(id) ON DELETE CASCADE,

    updated_by UUID NOT NULL REFERENCES "user"(id),
    changes JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

\echo '  ✅ Tabela partner_update criada'

-- Leads
CREATE TABLE IF NOT EXISTS lead (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    pipeline_id UUID NOT NULL REFERENCES pipeline(id),
    stage_id UUID NOT NULL REFERENCES pipeline_stage(id),

    -- Dados do lead
    company_name TEXT NOT NULL,
    contact_name TEXT,
    contact_email TEXT,
    contact_phone TEXT,

    -- Valor e status
    value NUMERIC(12, 2),
    status TEXT,

    -- Atribuição
    assigned_to UUID REFERENCES "user"(id) ON DELETE SET NULL,
    partner_id UUID REFERENCES partner(id) ON DELETE SET NULL,

    -- Auditoria
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela lead criada'

-- Commissions
CREATE TABLE IF NOT EXISTS commission (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,

    deal_id UUID NOT NULL REFERENCES lead(id),
    partner_id UUID NOT NULL REFERENCES partner(id),

    amount NUMERIC(12, 2) NOT NULL,
    percentage NUMERIC(5, 2),

    status TEXT,

    approved_by UUID REFERENCES "user"(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela commission criada'

-- Clawbacks
CREATE TABLE IF NOT EXISTS clawback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    commission_id UUID NOT NULL REFERENCES commission(id) ON DELETE CASCADE,

    amount NUMERIC(12, 2) NOT NULL,
    reason TEXT,

    status TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela clawback criada'
\echo '✨ ETAPA 5 concluída com sucesso!'
