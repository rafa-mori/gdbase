-- ============================================================
-- SEED CANALIZE - V0.1 (Minimal, Realistic, Coherent)
-- ============================================================

-- ================
-- ORG PRINCIPAL
-- ================
insert into orgs (
   id,
   name,
   created_at
) values ( '01JHX4ZCP3G2T4JC3P8VBA5N3A',
           'Canalize PRM',
           now() );

-- ================
-- TENANT PRINCIPAL
-- ================
insert into tenants (
   id,
   org_id,
   name,
   slug,
   status,
   created_at
) values ( '01JHX4ZH8Z9AGF1AFM7V0E7EAQ',
           '01JHX4ZCP3G2T4JC3P8VBA5N3A',
           'Canalize HQ',
           'canalize-hq',
           'active',
           now() );

-- ================
-- PERMISSIONS (ESSENCIAIS)
-- ================
insert into permissions (
   id,
   code,
   description
) values ( '01JHX50R3GGVQHH3M8RP6N3N8Y',
           'dashboard.view',
           'Acessar dashboards' ),( '01JHX50R3MPD8ZT7NPB0D9G7V3',
                                    'users.manage',
                                    'Gerenciar usuários' ),( '01JHX50R3W8HS14QW9PGN4VRE0',
                                                             'settings.manage',
                                                             'Gerenciar configurações' ),( '01JHX50R3Z0B5ZQ5FJQ96A6BYT',
                                                                                           'partners.view',
                                                                                           'Ver parceiros' ),( '01JHX50R43DW4TN9CP8NVQ7CFR'
                                                                                           ,
                                                                                                               'leads.view',
                                                                                                               'Visualizar leads'
                                                                                                               ),( '01JHX50R46MSN74Z8VMEGXSY9D'
                                                                                                               ,
                                                                                                                                  'leads.manage'
                                                                                                                                  ,
                                                                                                                                  'Gerenciar leads'
                                                                                                                                  )
                                                                                                                                  ;

-- ================
-- ROLES BÁSICOS
-- ================
insert into roles (
   id,
   code,
   label,
   description
) values ( '01JHX52373KT7NPKDD6M5WBR4M',
           'admin',
           'Administrador',
           'Acesso total' ),( '01JHX5237F45G3P6PRADAH9J8K',
                              'manager',
                              'Gestor',
                              'Coordenação e supervisão' ),( '01JHX5237J1TQ5J0V8B4CWVZ4N',
                                                             'viewer',
                                                             'Visualizador',
                                                             'Acesso mínimo' );

-- ================
-- ROLE PERMISSIONS
-- ================
-- Admin -> tudo
insert into role_permissions (
   role_id,
   permission_id
)
   select '01JHX52373KT7NPKDD6M5WBR4M',
          id
     from permissions;

-- Manager -> subset
insert into role_permissions (
   role_id,
   permission_id
)
   select '01JHX5237F45G3P6PRADAH9J8K',
          id
     from permissions
    where code in ( 'dashboard.view',
                    'leads.view',
                    'partners.view' );

-- Viewer -> mínimo
insert into role_permissions (
   role_id,
   permission_id
) values ( '01JHX5237J1TQ5J0V8B4CWVZ4N',
           '01JHX50R3GGVQHH3M8RP6N3N8Y' ); -- dashboard.view

-- ================
-- USUÁRIOS
-- ================
insert into users (
   id,
   email,
   full_name,
   password_hash,
   created_at
) values
  -- Usuário principal (você)
 ( '01JHX54TJEGW95WNH9TEG10M4H',
           'rafael@canalize.app',
           'Rafael Mori',
           crypt(
              'canalize123',
              gen_salt('bf')
           ),
           now() ),

  -- Usuário secundário (Thiago)
           ( '01JHX54TJQBV8Z7RPKT4HR8X7N',
                     'thiago@canalize.app',
                     'Thiago CTO',
                     crypt(
                        'canalize123',
                        gen_salt('bf')
                     ),
                     now() );

-- ================
-- MEMBERSHIPS (ligação usuário ↔ tenant ↔ role)
-- ================
insert into memberships (
   id,
   user_id,
   tenant_id,
   role_id,
   created_at
) values ( '01JHX56GNNGQVPNB6WNXCHQH8W',
           '01JHX54TJEGW95WNH9TEG10M4H',
           '01JHX4ZH8Z9AGF1AFM7V0E7EAQ',
           '01JHX52373KT7NPKDD6M5WBR4M',
           now() ), -- admin
           ( '01JHX56GP2AY39WXD2T3XYR94M',
                     '01JHX54TJQBV8Z7RPKT4HR8X7N',
                     '01JHX4ZH8Z9AGF1AFM7V0E7EAQ',
                     '01JHX5237F45G3P6PRADAH9J8K',
                     now() ); -- manager

-- ================
-- TEAMS (mínimo necessário)
-- ================
insert into teams (
   id,
   tenant_id,
   name,
   created_at
) values ( '01JHX58W89V5Z2N4AWR4FSXCT3',
           '01JHX4ZH8Z9AGF1AFM7V0E7EAQ',
           'Equipe Comercial',
           now() );

-- ================
-- TEAMS MEMBERS
-- ================
insert into teams_members (
   team_id,
   user_id
) values ( '01JHX58W89V5Z2N4AWR4FSXCT3',
           '01JHX54TJEGW95WNH9TEG10M4H' ),( '01JHX58W89V5Z2N4AWR4FSXCT3',
                                            '01JHX54TJQBV8Z7RPKT4HR8X7N' );

-- ================
-- INVITE DE EXEMPLO
-- (pra testar magic link -> signup)
-- ================
insert into invites (
   id,
   email,
   role_id,
   tenant_id,
   created_by,
   status,
   created_at
) values ( '01JHX5AHDV7H8T6C2911CKGHEC',
           'novo-user@canalize.app',
           '01JHX5237J1TQ5J0V8B4CWVZ4N', -- viewer
           '01JHX4ZH8Z9AGF1AFM7V0E7EAQ',
           '01JHX54TJEGW95WNH9TEG10M4H',
           'pending',
           now() );
-- ============================================================
