-- ============================================================================
-- ETAPA 1: Extensions + Multi-Tenancy
-- ============================================================================
-- Cria extensions necessárias e tabelas de multi-tenancy (org, tenant, team)
-- ============================================================================

\echo '🚀 ETAPA 1: Criando extensions e estrutura de multi-tenancy...'

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext"; -- Case-insensitive text

\echo '  ✅ Extensions criadas'

-- Organização (nível mais alto)
CREATE TABLE IF NOT EXISTS org (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

\echo '  ✅ Tabela org criada'

-- Tenant (empresa/cliente individual)
CREATE TABLE IF NOT EXISTS tenant (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES org(id) ON DELETE CASCADE,

    -- Identidade
    name TEXT NOT NULL,
    slug TEXT UNIQUE,
    domain TEXT UNIQUE,
    logo_url TEXT,

    -- Contato
    phone TEXT,
    address TEXT,

    -- Tiers/Plans (SaaS)
    plan TEXT,
    is_active BOOLEAN DEFAULT true,
    is_trial BOOLEAN DEFAULT false,
    trial_ends_at TIMESTAMPTZ,

    -- Auditoria
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela tenant criada'

-- Teams (times dentro de um tenant)
CREATE TABLE IF NOT EXISTS team (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,

    name VARCHAR(255) NOT NULL,
    description TEXT,

    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,

    created_by UUID, -- FK adicionada na etapa 3
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

\echo '  ✅ Tabela team criada'
\echo '✨ ETAPA 1 concluída com sucesso!'
