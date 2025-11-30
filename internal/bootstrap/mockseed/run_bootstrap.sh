#!/usr/bin/env bash
# shellcheck disable=SC2155,SC2207
# ==========================================
# CANALIZE SEED HYDRATION v0.0.1
# ==========================================
# Versão: 0.0.1
# Data: 2025-11-26
# Autores: Rafael Mori (Desenvolvedor) + ChatGPT (OpenAI)
# ==========================================
# Descrição:
#   Script executor do Seed Hydration Canalize v0.0.1
#   Inicializa dados para testes e demonstrações com lógica realista.
# ==========================================

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# ==========================================
# CONFIGURAÇÕES
# ==========================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${SCRIPT_DIR}/logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="${LOG_DIR}/bootstrap_${TIMESTAMP}.log"
JSON_LOG="${LOG_DIR}/bootstrap_${TIMESTAMP}.json"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
# CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ==========================================
# FUNÇÕES AUXILIARES
# ==========================================

log_info() {
  echo -e "${BLUE}ℹ️  [INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
  echo -e "${GREEN}✅ [SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
  echo -e "${YELLOW}⚠️  [WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
  echo -e "${RED}❌ [ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_step() {
  echo -e "${PURPLE}🚀 [STEP $1]${NC} $2" | tee -a "$LOG_FILE"
}

print_header() {
  echo "" | tee -a "$LOG_FILE"
  echo "========================================" | tee -a "$LOG_FILE"
  echo "$1" | tee -a "$LOG_FILE"
  echo "========================================" | tee -a "$LOG_FILE"
  echo "" | tee -a "$LOG_FILE"
}

# ==========================================
# VALIDAÇÕES PRÉ-EXECUÇÃO
# ==========================================

validate_prerequisites() {
  log_info "Validando pré-requisitos..."

  # Verificar se psql está disponível
  if ! command -v psql &> /dev/null; then
    log_error "psql não encontrado. Instale PostgreSQL client."
    exit 1
  fi

  # Verificar variável de ambiente DATABASE_URL
  if [ -z "${DATABASE_URL:-}" ]; then
    log_error "Variável DATABASE_URL não definida."
    log_info "Defina: export DATABASE_URL='postgres://user:pass@host:port/dbname'" # pragma: allowlist secret
    exit 1
  fi

  # Testar conexão
  if ! psql "$DATABASE_URL" -c "SELECT 1" &> /dev/null; then
    log_error "Não foi possível conectar ao banco de dados."
    exit 1
  fi

  log_success "Pré-requisitos validados"
}

# ==========================================
# EXECUTAR SEED HYDRATION
# ==========================================

hydration_seed_exec() {
  local step_num=$1
  local step_name=$2
  local step_file=$3
  local step_start=$(date +%s)

  log_step "$step_num" "$step_name"

  local full_path="${SCRIPT_DIR}/${step_file}"

  if [ ! -f "$full_path" ]; then
    log_error "Arquivo não encontrado: $full_path"
    return 1
  fi

  # Executar SQL com output redirecionado para log
  if psql "${DATABASE_URL:-}" -f "$full_path" >> "$LOG_FILE" 2>&1; then
    local step_end=$(date +%s)
    local duration=$((step_end - step_start))
    log_success "Hydration de seed $step_num concluída em ${duration}s"
    echo "$step_num,$step_name,$duration,success" >> "${LOG_DIR}/execution_summary_seed.csv"
    return 0
  else
    local step_end=$(date +%s)
    local duration=$((step_end - step_start))
    log_error "Falha no processo de hydration do seed $step_num após ${duration}s"
    echo "$step_num,$step_name,$duration,failed" >> "${LOG_DIR}/execution_summary_seed.csv"
    return 1
  fi
}

# ==========================================
# VALIDAÇÕES PÓS-EXECUÇÃO
# ==========================================

validate_installation() {
  local seed_file=$1

  log_info "Validando processos de Hydration..."

  local errors=0

  # Verificar se o arquivo de seed existe
  if [ ! -f "$seed_file" ]; then
    log_error "Arquivo de seed não encontrado: $seed_file"
    ((errors++))
  fi

  # shellcheck disable=SC2207
  local tables_seeded=( $(grep -sin 'insert into ' "$seed_file" | awk -F' ' '{ print $3 }' | uniq) )

  # Verificar relações de dados em tabelas críticas usando jq para manipular JSON
  local required_tables='{
    "org" : {
      "tenant" : "org_id",
      "team" : "org_id",
      "user" : "org_id"
    },
    "tenant" : {
      "team" : "tenant_id",
      "user" : "tenant_id"
    },
    "team" : {
      "user" : "team_id"
    },
    "user" : {
      "role" : "user_id"
    },
    "role" : {
      "permission" : "role_id"
    },
    "permission" : {
      "role" : "permission_id"
    },
    "role_permission" : {
      "role" : "role_id",
      "permission" : "permission_id"
    },
    "tenant_membership" : {
      "user" : "tenant_id"
    },
    "team_membership" : {
      "user" : "team_id"
    },
    "partner_invitation" : {
      "user" : "partner_id"
    },
    "internal_invitation" : {
      "user" : "internal_id"
    },
    "pipeline" : {
      "pipeline_stage" : "pipeline_id"
    },
    "pipeline_stage" : {
      "lead" : "stage_id"
    },
    "partner" : {
      "user" : "partner_id"
    },
    "lead" : {
      "user" : "lead_id"
    },
    "commission" : {
      "user" : "commission_id"
    },
    "clawback" : {
      "user" : "clawback_id"
    }
  }';

  # local required_tables_keys=$(echo "$required_tables" | jq -r 'keys[]');

  for table_seeded in "${tables_seeded[@]}"; do
    if ! echo "$required_tables" | jq -e --arg table "$table_seeded" '.[$table]' &> /dev/null; then
      log_warning "⚠️  Tabela '$table_seeded' não está na lista de validação crítica."
      continue
    fi

    # Busca as relações para a table_seeded onde ela é table principal
    local required_tables_keys=(
      $(echo "$required_tables" | jq -r --arg table "$table_seeded" '.[$table] | keys[]')
      $(echo "$required_tables" | jq -r 'to_entries[] | select(.value | has($table_seeded)) | .key')
    )

    for table in "${required_tables_keys[@]}"; do
      local relations=$(echo "$required_tables" | jq -r --arg table "$table" '.[$table] | to_entries[] | "\(.key):\(.value)"')
      local count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM $table")
      if [ "$count" -gt 0 ]; then
        for relation in $relations; do
          local rel_table=$(echo "$relation" | cut -d':' -f1)
          local rel_column=$(echo "$relation" | cut -d':' -f2)
          local rel_count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM $rel_table WHERE $rel_column IS NOT NULL")
          if [ "$rel_count" -gt 0 ]; then
            log_info "✓ Relação entre '$table' e '$rel_table' verificada"
          else
            log_error "✗ Relação entre '$table' e '$rel_table' NÃO encontrada"
            ((errors++))
          fi
        done
      else
        log_error "✗ Tabela '$table' está vazia"
        ((errors++))
      fi
    done
  done

  if [ $errors -eq 0 ]; then
    log_success "Validação de Hydration concluída sem erros"
    return 0
  else
    log_error "Validação de Hydration encontrou $errors erros"
    return 1
  fi
}

# ==========================================
# GERAR RELATÓRIO JSON
# ==========================================

generate_json_report() {
  local total_duration=$1
  local status=$2
  local seed_file=$3

  local tables_seeded=( $(grep -sin 'insert into ' "$seed_file" | awk -F' ' '{ print $3 }' | uniq) )
  local total_tables=${#tables_seeded[@]}

  local json_tables=""
  for table in "${tables_seeded[@]}"; do
    local count=$(psql "$DATABASE_URL" -tAc "SELECT COUNT(*) FROM $table")
    json_tables+="\"$table\": $count,"
  done
  json_tables=${json_tables%,} # Remove última vírgula

  cat > "$JSON_LOG" <<EOF
{
  "execution_timestamp": "$(date -Iseconds)",
  "database_url": "${DATABASE_URL//:*@/:***@}",
  "total_duration_seconds": $total_duration,
  "status": "$status",
  "log_file": "$LOG_FILE",
  "steps_executed": $(wc -l < "${LOG_DIR}/execution_summary_seed.csv"),
  "manifest_version": "0.0.1",
  "schema_version": "hybrid-v1.0",
  "features": {
    "seed_hydration": true
  },
  "tables_seeded": {
    "total_tables": $total_tables,
    "records_per_table": {
      $json_tables
    }
}
EOF

  log_info "Relatório JSON gerado: $JSON_LOG"
}

# ==========================================
# FUNÇÃO PRINCIPAL
# ==========================================

main() {
  local seed_pattern="${1:-simplemock}" # Padrão: simplemock, pode ser extendido futuramente
  local seeds_to_use=( $(find . -name "*$seed_pattern*" -exec readlink -f {} \;) )

  if [ ${#seeds_to_use[@]} -eq 0 ]; then
    log_error "Seed '$seed_pattern' não encontrada."
    exit 1
  fi

  local start_time=$(date +%s)

  # Criar diretório de logs
  mkdir -p "$LOG_DIR"

  # Header
  print_header "🚀 CANALIZE HYDRATION v0.0.1 - INICIANDO"

  log_info "Timestamp: $(date)"
  log_info "Database: ${DATABASE_URL//:*@/:***@}"
  log_info "Log: $LOG_FILE"

  # Validar pré-requisitos
  validate_prerequisites

  # Criar CSV de resumo
  echo "step,name,duration_seconds,status" > "${LOG_DIR}/execution_summary_seed.csv"

  local step_counter=1
  for seed_file in "${seeds_to_use[@]}"; do
    log_info "Usando seed: $seed_file"
    hydration_seed_exec "$step_counter" "$(basename "$seed_file")" "$seed_file"
    ((step_counter++))


    # Validar instalação
    if validate_installation "$seed_file"; then
      local end_time=$(date +%s)
      local total_duration=$((end_time - start_time))

      print_header "✅ BOOTSTRAP CONCLUÍDO COM SUCESSO"
      log_success "Tempo total: ${total_duration}s"
      log_info "Log completo: $LOG_FILE"

      generate_json_report "$total_duration" "success" "$seed_file"
      exit 0
    else
      local end_time=$(date +%s)
      local total_duration=$((end_time - start_time))

      print_header "⚠️  BOOTSTRAP CONCLUÍDO COM AVISOS"
      log_warning "Tempo total: ${total_duration}s"
      log_warning "Revise o log: $LOG_FILE"

      generate_json_report "$total_duration" "success_with_warnings" "$seed_file"
      exit 0
    fi

  done
}

# ==========================================
# EXECUTAR
# ==========================================

main "$@"
