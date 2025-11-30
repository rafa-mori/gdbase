# Canalize Schema Híbrido v3.0

Bootstrap automatizado do Schema Híbrido Canalize - Combina o melhor de sql_1 e sql_2.

## 🚀 Execução Rápida

```bash
# 1. Definir DATABASE_URL
export DATABASE_URL="postgres://canalize_adm:senha@localhost:5432/canalize_db"

# 2. Executar bootstrap
./run_bootstrap.sh

# 3. Ver logs
tail -f logs/bootstrap_*.log
```

## 📊 O que será criado

- **18 tabelas** (org, tenant, team, user, role, permission, etc.)
- **60 índices** otimizados
- **13 triggers** de auditoria
- **8 roles** do sistema
- **35 permissions** granulares
- **Dados de exemplo** para testes

## 📚 Documentação Completa

Ver `bootstrap.manifest.json` para detalhes técnicos.

## ✅ Validação

Após execução, verifique:
```sql
-- Ver tabelas
\dt

-- Ver roles
SELECT * FROM role WHERE is_system_role = true;

-- Ver tenant de exemplo
SELECT * FROM tenant WHERE slug = 'canalize-demo';
```

---

**Versão:** 3.0.0
**Data:** 2025-11-10
