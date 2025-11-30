-- ============================================================================
-- ETAPA 6: Índices de Performance
-- ============================================================================
-- Cria todos os índices para otimização de queries
-- ============================================================================

\echo '🚀 ETAPA 6: Criando índices de performance...'

-- Tenant
CREATE INDEX IF NOT EXISTS idx_tenant_org_id ON tenant(org_id);
CREATE INDEX IF NOT EXISTS idx_tenant_slug ON tenant(slug);
CREATE INDEX IF NOT EXISTS idx_tenant_domain ON tenant(domain);
CREATE INDEX IF NOT EXISTS idx_tenant_plan ON tenant(plan);
CREATE INDEX IF NOT EXISTS idx_tenant_is_active ON tenant(is_active) WHERE is_active = true;

\echo '  ✅ Índices de tenant criados (5)'

-- Team
CREATE INDEX IF NOT EXISTS idx_team_tenant_id ON team(tenant_id);
CREATE INDEX IF NOT EXISTS idx_team_is_default ON team(is_default) WHERE is_default = true;
CREATE INDEX IF NOT EXISTS idx_team_created_by ON team(created_by);

\echo '  ✅ Índices de team criados (3)'

-- User
CREATE INDEX IF NOT EXISTS idx_user_email ON "user"(email);
CREATE INDEX IF NOT EXISTS idx_user_status ON "user"(status);

\echo '  ✅ Índices de user criados (2)'

-- Role
CREATE INDEX IF NOT EXISTS idx_role_code ON role(code);
CREATE INDEX IF NOT EXISTS idx_role_is_system ON role(is_system_role) WHERE is_system_role = true;
CREATE INDEX IF NOT EXISTS idx_role_parent ON role(parent_role_id);

\echo '  ✅ Índices de role criados (3)'

-- Permission
CREATE INDEX IF NOT EXISTS idx_permission_code ON permission(code);
CREATE INDEX IF NOT EXISTS idx_permission_category ON permission(category);

\echo '  ✅ Índices de permission criados (2)'

-- Memberships
CREATE INDEX IF NOT EXISTS idx_tenant_membership_user ON tenant_membership(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_membership_tenant ON tenant_membership(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_membership_role ON tenant_membership(role_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_user ON team_membership(user_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_team ON team_membership(team_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_role ON team_membership(role_id);

\echo '  ✅ Índices de memberships criados (6)'

-- Invitations (Partner)
CREATE INDEX IF NOT EXISTS idx_partner_invitation_token ON partner_invitation(token);
CREATE INDEX IF NOT EXISTS idx_partner_invitation_email ON partner_invitation(email);
CREATE INDEX IF NOT EXISTS idx_partner_invitation_tenant ON partner_invitation(tenant_id);
CREATE INDEX IF NOT EXISTS idx_partner_invitation_team ON partner_invitation(team_id);
CREATE INDEX IF NOT EXISTS idx_partner_invitation_status ON partner_invitation(status);
CREATE INDEX IF NOT EXISTS idx_partner_invitation_invited_by ON partner_invitation(invited_by);
CREATE INDEX IF NOT EXISTS idx_partner_invitation_created_at ON partner_invitation(created_at DESC);

-- Índice condicional otimizado
CREATE INDEX IF NOT EXISTS idx_partner_invitation_pending
ON partner_invitation(status, expires_at)
WHERE status = 'pending';

\echo '  ✅ Índices de partner_invitation criados (8)'

-- Invitations (Internal)
CREATE INDEX IF NOT EXISTS idx_internal_invitation_token ON internal_invitation(token);
CREATE INDEX IF NOT EXISTS idx_internal_invitation_email ON internal_invitation(email);
CREATE INDEX IF NOT EXISTS idx_internal_invitation_tenant ON internal_invitation(tenant_id);
CREATE INDEX IF NOT EXISTS idx_internal_invitation_status ON internal_invitation(status);
CREATE INDEX IF NOT EXISTS idx_internal_invitation_pending
ON internal_invitation(status, expires_at)
WHERE status = 'pending';

\echo '  ✅ Índices de internal_invitation criados (5)'

-- Pipeline & Stages
CREATE INDEX IF NOT EXISTS idx_pipeline_tenant ON pipeline(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_stage_pipeline ON pipeline_stage(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_stage_order ON pipeline_stage(pipeline_id, order_index);

\echo '  ✅ Índices de pipeline criados (3)'

-- Partner
CREATE INDEX IF NOT EXISTS idx_partner_tenant ON partner(tenant_id);
CREATE INDEX IF NOT EXISTS idx_partner_status ON partner(status);
CREATE INDEX IF NOT EXISTS idx_partner_tier ON partner(tier);
CREATE INDEX IF NOT EXISTS idx_partner_primary_contact ON partner(primary_contact_user_id);
CREATE INDEX IF NOT EXISTS idx_partner_update_partner ON partner_update(partner_id);

\echo '  ✅ Índices de partner criados (5)'

-- Lead
CREATE INDEX IF NOT EXISTS idx_lead_tenant ON lead(tenant_id);
CREATE INDEX IF NOT EXISTS idx_lead_pipeline ON lead(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_lead_stage ON lead(stage_id);
CREATE INDEX IF NOT EXISTS idx_lead_status ON lead(status);
CREATE INDEX IF NOT EXISTS idx_lead_assigned_to ON lead(assigned_to);
CREATE INDEX IF NOT EXISTS idx_lead_partner ON lead(partner_id);
CREATE INDEX IF NOT EXISTS idx_lead_created_at ON lead(created_at DESC);

\echo '  ✅ Índices de lead criados (7)'

-- Commission
CREATE INDEX IF NOT EXISTS idx_commission_tenant ON commission(tenant_id);
CREATE INDEX IF NOT EXISTS idx_commission_deal ON commission(deal_id);
CREATE INDEX IF NOT EXISTS idx_commission_partner ON commission(partner_id);
CREATE INDEX IF NOT EXISTS idx_commission_status ON commission(status);
CREATE INDEX IF NOT EXISTS idx_commission_approved_by ON commission(approved_by);

\echo '  ✅ Índices de commission criados (5)'

-- Clawback
CREATE INDEX IF NOT EXISTS idx_clawback_commission ON clawback(commission_id);
CREATE INDEX IF NOT EXISTS idx_clawback_status ON clawback(status);

\echo '  ✅ Índices de clawback criados (2)'

\echo '✨ ETAPA 6 concluída com sucesso! Total: 60 índices criados'
