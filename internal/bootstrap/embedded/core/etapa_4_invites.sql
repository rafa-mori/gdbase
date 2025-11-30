-- ============================================================================
-- ETAPA 4: Invites
-- ============================================================================
-- Cria tabelas de convites (partner e internal)
-- ============================================================================

\echo '🚀 ETAPA 4: Criando sistema de convites...'

-- Partner Invitation
CREATE TABLE IF NOT EXISTS partner_invitation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Dados do convidado
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(30) NOT NULL,

    -- Token único
    token VARCHAR(255) NOT NULL UNIQUE,

    -- Contexto tenant-based
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    team_id UUID REFERENCES team(id) ON DELETE SET NULL,
    invited_by UUID NOT NULL REFERENCES "user"(id),

    -- Estado
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '7 days'),
    accepted_at TIMESTAMPTZ,

    -- Auditoria
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

\echo '  ✅ Tabela partner_invitation criada'

-- Internal Invitation
CREATE TABLE IF NOT EXISTS internal_invitation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Dados do convidado
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(30) NOT NULL,

    -- Token único
    token VARCHAR(255) NOT NULL UNIQUE,

    -- Contexto tenant-based
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    team_id UUID REFERENCES team(id) ON DELETE SET NULL,
    invited_by UUID NOT NULL REFERENCES "user"(id),

    -- Estado
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '7 days'),
    accepted_at TIMESTAMPTZ,

    -- Auditoria
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

\echo '  ✅ Tabela internal_invitation criada'
\echo '✨ ETAPA 4 concluída com sucesso!'
