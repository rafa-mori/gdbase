\echo '🚀 ETAPA 8: Inserindo dados iniciais (seed)...'

-- ============================================================================
-- CANALIZE PRM - Seed Data v1.0
-- ============================================================================
-- Dados iniciais para roles, permissions e org/tenant de exemplo
-- Data: 2025-11-10
-- Autor: Claude + Desenvolvedor
-- ============================================================================

-- ============================================================================
-- ETAPA 1: System Roles (Imutáveis)
-- ============================================================================

-- Roles principais do sistema (hierárquicos)
INSERT INTO role (id, code, display_name, description, is_system_role, parent_role_id) VALUES
-- Admin (topo da hierarquia)
('00000000-0000-0000-0000-000000000001', 'admin', 'Administrador', 'Acesso total ao sistema', true, NULL),

-- Gerencial (herda de admin)
('00000000-0000-0000-0000-000000000002', 'manager', 'Gerente', 'Gerenciamento de equipes e parceiros', true, '00000000-0000-0000-0000-000000000001'),

-- Parceiros (roles específicas)
('00000000-0000-0000-0000-000000000003', 'partner_admin', 'Administrador de Parceiro', 'Admin dentro da empresa parceira', true, NULL),
('00000000-0000-0000-0000-000000000004', 'partner_manager', 'Gerente de Parceiro', 'Gerente dentro da empresa parceira', true, '00000000-0000-0000-0000-000000000003'),
('00000000-0000-0000-0000-000000000005', 'partner_rep', 'Representante de Parceiro', 'Vendedor/SDR da empresa parceira', true, '00000000-0000-0000-0000-000000000004'),

-- Operacional
('00000000-0000-0000-0000-000000000006', 'finance', 'Financeiro', 'Gestão de comissões e pagamentos', true, NULL),
('00000000-0000-0000-0000-000000000007', 'cs', 'Customer Success', 'Atendimento e suporte a parceiros', true, NULL),

-- Visualização
('00000000-0000-0000-0000-000000000008', 'viewer', 'Visualizador', 'Apenas leitura', true, NULL);

-- ============================================================================
-- ETAPA 2: Permissions (Granulares)
-- ============================================================================

-- Permissions de Parceiros
INSERT INTO permission (code, display_name, description, category) VALUES
('partner.read', 'Visualizar Parceiros', 'Ver lista e detalhes de parceiros', 'partners'),
('partner.create', 'Criar Parceiros', 'Cadastrar novos parceiros', 'partners'),
('partner.update', 'Editar Parceiros', 'Atualizar dados de parceiros', 'partners'),
('partner.delete', 'Excluir Parceiros', 'Remover parceiros do sistema', 'partners'),
('partner.invite', 'Convidar Parceiros', 'Enviar convites para novos parceiros', 'partners');

-- Permissions de Leads/Deals
INSERT INTO permission (code, display_name, description, category) VALUES
('deal.read', 'Visualizar Negócios', 'Ver lista e detalhes de negócios', 'deals'),
('deal.create', 'Criar Negócios', 'Registrar novos negócios', 'deals'),
('deal.update', 'Editar Negócios', 'Atualizar dados de negócios', 'deals'),
('deal.delete', 'Excluir Negócios', 'Remover negócios do sistema', 'deals'),
('deal.assign', 'Atribuir Negócios', 'Atribuir negócios a parceiros/usuários', 'deals'),
('deal.close', 'Fechar Negócios', 'Marcar negócios como ganhos/perdidos', 'deals');

-- Permissions de Comissões
INSERT INTO permission (code, display_name, description, category) VALUES
('commission.read', 'Visualizar Comissões', 'Ver comissões calculadas', 'commissions'),
('commission.run', 'Executar Cálculo', 'Rodar processamento de comissões', 'commissions'),
('commission.approve', 'Aprovar Comissões', 'Aprovar comissões para pagamento', 'commissions'),
('commission.pay', 'Pagar Comissões', 'Marcar comissões como pagas', 'commissions'),
('commission.clawback', 'Estornar Comissões', 'Criar estornos de comissões', 'commissions');

-- Permissions de Pipelines
INSERT INTO permission (code, display_name, description, category) VALUES
('pipeline.read', 'Visualizar Pipelines', 'Ver funis de vendas', 'pipelines'),
('pipeline.create', 'Criar Pipelines', 'Criar novos funis', 'pipelines'),
('pipeline.update', 'Editar Pipelines', 'Modificar funis existentes', 'pipelines'),
('pipeline.delete', 'Excluir Pipelines', 'Remover funis do sistema', 'pipelines');

-- Permissions de Usuários/Times
INSERT INTO permission (code, display_name, description, category) VALUES
('user.read', 'Visualizar Usuários', 'Ver usuários do tenant', 'admin'),
('user.create', 'Criar Usuários', 'Cadastrar novos usuários', 'admin'),
('user.update', 'Editar Usuários', 'Atualizar dados de usuários', 'admin'),
('user.delete', 'Excluir Usuários', 'Remover usuários', 'admin'),
('user.invite', 'Convidar Usuários', 'Enviar convites internos', 'admin');

INSERT INTO permission (code, display_name, description, category) VALUES
('team.read', 'Visualizar Times', 'Ver times do tenant', 'admin'),
('team.create', 'Criar Times', 'Criar novos times', 'admin'),
('team.update', 'Editar Times', 'Modificar times existentes', 'admin'),
('team.delete', 'Excluir Times', 'Remover times', 'admin');

-- Permissions de Administração
INSERT INTO permission (code, display_name, description, category) VALUES
('settings.read', 'Visualizar Configurações', 'Ver configurações do tenant', 'admin'),
('settings.update', 'Editar Configurações', 'Modificar configurações', 'admin'),
('role.manage', 'Gerenciar Roles', 'Criar/editar roles e permissões', 'admin'),
('audit.read', 'Visualizar Auditoria', 'Ver logs de auditoria', 'admin');

-- ============================================================================
-- ETAPA 3: Role-Permission Matrix
-- ============================================================================

-- Admin (tem TODAS as permissões)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000001', id, true
FROM permission;

-- Manager (gestão completa, menos admin)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000002', id, true
FROM permission
WHERE category IN ('partners', 'deals', 'commissions', 'pipelines')
   OR code IN ('user.read', 'user.invite', 'team.read', 'settings.read');

-- Partner Admin (gestão dentro da empresa parceira)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000003', id, true
FROM permission
WHERE category IN ('deals', 'commissions')
   OR code IN ('partner.read', 'user.read', 'user.invite', 'team.read', 'team.create');

-- Partner Manager (gestão de deals do parceiro)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000004', id, true
FROM permission
WHERE category = 'deals'
   OR code IN ('partner.read', 'commission.read', 'user.read');

-- Partner Rep (vendedor do parceiro)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000005', id, true
FROM permission
WHERE code IN ('deal.read', 'deal.create', 'deal.update', 'commission.read', 'partner.read');

-- Finance (gestão de comissões)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000006', id, true
FROM permission
WHERE category = 'commissions'
   OR code IN ('deal.read', 'partner.read', 'user.read');

-- CS (suporte a parceiros)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000007', id, true
FROM permission
WHERE code IN ('partner.read', 'partner.update', 'deal.read', 'user.read', 'team.read');

-- Viewer (apenas leitura)
INSERT INTO role_permission (role_id, permission_id, value)
SELECT '00000000-0000-0000-0000-000000000008', id, true
FROM permission
WHERE code LIKE '%.read';

-- ============================================================================
-- ETAPA 4: Org e Tenant de Exemplo (para testes)
-- ============================================================================

-- Org de exemplo
INSERT INTO org (id, name) VALUES
('10000000-0000-0000-0000-000000000001', 'Canalize Holding');

-- Tenant de exemplo (empresa de teste)
INSERT INTO tenant (
    id,
    org_id,
    name,
    slug,
    domain,
    plan,
    is_active,
    is_trial,
    trial_ends_at
) VALUES (
    '20000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000001',
    'Canalize Demo Corp',
    'canalize-demo',
    'demo.canalize.app',
    'professional',
    true,
    false,
    NULL
);

-- Team padrão do tenant de exemplo
INSERT INTO team (
    id,
    tenant_id,
    name,
    description,
    is_default,
    is_active
) VALUES (
    '30000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'Time Principal',
    'Time padrão da empresa',
    true,
    true
);

-- ============================================================================
-- ETAPA 5: Usuário Admin de Exemplo (para testes)
-- ============================================================================

-- Usuário admin de teste - cria ou corrige hash/sinalizadores
INSERT INTO "user" (
    id,
    email,
    name,
    last_name,
    status,
    password_hash,
    force_password_reset,
    created_at
) VALUES (
    '40000000-0000-0000-0000-000000000001',
    'admin@canalize.demo',
    'Admin',
    'Demo',
    'active',
    '$2a$10$5ShyQwiXfc0Yu9g.AKJ2o.o5Tf4Vaw0fyxZmMNvTqT/zQ1pEoWzIa',
    false,
    NOW()
)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    last_name = EXCLUDED.last_name,
    status = EXCLUDED.status,
    password_hash = EXCLUDED.password_hash,
    force_password_reset = false,
    updated_at = NOW();

-- Vincular admin ao tenant com role admin
INSERT INTO tenant_membership (
    user_id,
    tenant_id,
    role_id,
    is_active
) VALUES (
    '40000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001', -- role admin
    true
);

-- Vincular admin ao team padrão
INSERT INTO team_membership (
    user_id,
    team_id,
    role_id,
    is_active
) VALUES (
    '40000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001', -- role admin
    true
);

-- ============================================================================
-- ETAPA 6: Pipeline de Exemplo
-- ============================================================================

-- Pipeline padrão
INSERT INTO pipeline (
    id,
    tenant_id,
    name,
    description,
    is_default
) VALUES (
    '50000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'Pipeline de Vendas B2B',
    'Funil padrão para vendas B2B',
    true
);

-- Stages do pipeline
INSERT INTO pipeline_stage (pipeline_id, name, order_index) VALUES
('50000000-0000-0000-0000-000000000001', 'Prospecção', 1),
('50000000-0000-0000-0000-000000000001', 'Qualificação', 2),
('50000000-0000-0000-0000-000000000001', 'Proposta', 3),
('50000000-0000-0000-0000-000000000001', 'Negociação', 4),
('50000000-0000-0000-0000-000000000001', 'Fechamento', 5);

-- ============================================================================
-- FIM DO SEED DATA v1.0
-- ============================================================================

-- Validar dados inseridos
DO $$
DECLARE
    role_count INT;
    permission_count INT;
    role_permission_count INT;
BEGIN
    SELECT COUNT(*) INTO role_count FROM role;
    SELECT COUNT(*) INTO permission_count FROM permission;
    SELECT COUNT(*) INTO role_permission_count FROM role_permission;

    RAISE NOTICE '✅ Seed concluído com sucesso!';
    RAISE NOTICE '   - % roles criadas', role_count;
    RAISE NOTICE '   - % permissions criadas', permission_count;
    RAISE NOTICE '   - % role-permission vinculadas', role_permission_count;
    RAISE NOTICE '   - 1 org de exemplo criada';
    RAISE NOTICE '   - 1 tenant de exemplo criado';
    RAISE NOTICE '   - 1 usuário admin criado (admin@canalize.demo)';
    RAISE NOTICE '   - 1 pipeline com 5 stages criado';
END $$;
