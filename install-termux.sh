#!/data/data/com.termux/files/usr/bin/bash
# ═══════════════════════════════════════════════════════════════
# GOWA Auto-Installer for Termux
# Installs, configures, and runs GOWA WhatsApp API on Android
# ═══════════════════════════════════════════════════════════════

# If piped (curl | bash), download to file and re-exec so reads work
if [ ! -t 0 ]; then
    TMPSCRIPT="$HOME/.gowa-install.sh"
    curl -sL "https://raw.githubusercontent.com/jonemp31/gowa/main/install-termux.sh" -o "$TMPSCRIPT"
    chmod +x "$TMPSCRIPT"
    exec bash "$TMPSCRIPT" "$@"
fi

set -e

# ───── Colors ─────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

print_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════╗"
    echo "║       GOWA Termux Auto-Installer         ║"
    echo "║       WhatsApp API for Android            ║"
    echo "╚══════════════════════════════════════════╝"
    echo -e "${NC}"
}

log_info()  { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step()  { echo -e "\n${BOLD}${CYAN}── $1 ──${NC}"; }

# ───── Validate environment ─────
if [ -z "$PREFIX" ] || [[ "$PREFIX" != *"com.termux"* ]]; then
    log_error "Este script deve ser executado no Termux!"
    exit 1
fi

print_banner

# ═══════════════════════════════════════════
# INTERACTIVE QUESTIONS
# ═══════════════════════════════════════════
log_step "Configuração"

# Question 1: Cell ID
while true; do
    echo -e "${BOLD}Qual é esse celular? (ex: cel1, cel2, cel3):${NC} "
    read -r CEL_ID
    if [[ "$CEL_ID" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        break
    fi
    log_error "ID inválido. Use apenas letras, números, - e _"
done

# Extract number from cel ID (cel1 → 1, cel2 → 2)
CEL_NUM=$(echo "$CEL_ID" | grep -oE '[0-9]+$' || echo "$CEL_ID")

# Question 2: Webhook URL
DEFAULT_WEBHOOK="https://webhook-dev.zapsafe.work/webhook/gowa-mobo"
echo -e "${BOLD}Manter webhook padrão?${NC}"
echo -e "  ${CYAN}${DEFAULT_WEBHOOK}?cel=${CEL_NUM}${NC}"
echo -e "${BOLD}[y/n]:${NC} "
read -r KEEP_WEBHOOK

if [[ "$KEEP_WEBHOOK" =~ ^[nN] ]]; then
    echo -e "${BOLD}Digite a URL da webhook (sem ?cel=):${NC} "
    read -r CUSTOM_WEBHOOK
    WEBHOOK_URL="${CUSTOM_WEBHOOK}?cel=${CEL_NUM}"
else
    WEBHOOK_URL="${DEFAULT_WEBHOOK}?cel=${CEL_NUM}"
fi

# Summary
echo ""
echo -e "${BOLD}═══ Resumo ═══${NC}"
echo -e "  Celular:  ${CYAN}${CEL_ID}${NC}"
echo -e "  Webhook:  ${CYAN}${WEBHOOK_URL}${NC}"
echo -e "  Tunnel:   ${CYAN}${CEL_ID}.autopilots.trade${NC}"
echo -e "  API:      ${CYAN}http://localhost:3000${NC}"
echo ""
echo -e "${BOLD}Confirma instalação? [y/n]:${NC} "
read -r CONFIRM
if [[ ! "$CONFIRM" =~ ^[yYsS] ]]; then
    log_warn "Instalação cancelada."
    exit 0
fi

# ═══════════════════════════════════════════
# STEP 1: Update & Install Dependencies
# ═══════════════════════════════════════════
log_step "1/10 — Atualizando pacotes"
pkg update -y && pkg upgrade -y

log_step "2/10 — Instalando dependências"
pkg install -y golang git ffmpeg libwebp tmux termux-api

# Cloudflared needs x11-repo
pkg install -y x11-repo 2>/dev/null || true
pkg install -y cloudflared 2>/dev/null
if ! command -v cloudflared &>/dev/null; then
    log_warn "cloudflared não encontrado no repo. Tentando instalação manual..."
    ARCH=$(uname -m)
    case "$ARCH" in
        aarch64) CF_ARCH="arm64" ;;
        armv7l|armv8l) CF_ARCH="arm" ;;
        *) log_error "Arquitetura não suportada: $ARCH"; exit 1 ;;
    esac
    curl -Lo "$PREFIX/bin/cloudflared" "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${CF_ARCH}"
    chmod +x "$PREFIX/bin/cloudflared"
fi

cloudflared version && log_info "cloudflared instalado" || { log_error "Falha ao instalar cloudflared"; exit 1; }

# ═══════════════════════════════════════════
# STEP 2: Wake Lock
# ═══════════════════════════════════════════
log_step "3/10 — Ativando wake-lock"
termux-wake-lock 2>/dev/null && log_info "Wake-lock ativado" || log_warn "Wake-lock falhou (instale Termux:API do F-Droid)"

# ═══════════════════════════════════════════
# STEP 3: Clone Repository
# ═══════════════════════════════════════════
log_step "4/10 — Clonando repositório"
if [ -d "$HOME/gowa" ]; then
    log_warn "Diretório ~/gowa já existe. Atualizando..."
    cd "$HOME/gowa"
    git pull || log_warn "git pull falhou, continuando com versão existente"
else
    cd "$HOME"
    git clone https://github.com/jonemp31/gowa.git
fi

# ═══════════════════════════════════════════
# STEP 4: Create Directories
# ═══════════════════════════════════════════
log_step "5/10 — Criando diretórios"
cd "$HOME/gowa/src"
mkdir -p statics/qrcode statics/senditems statics/media storages
log_info "Diretórios criados"

# ═══════════════════════════════════════════
# STEP 5: Compile
# ═══════════════════════════════════════════
log_step "6/10 — Compilando GOWA (pode demorar 5-15 min na primeira vez)"
cd "$HOME/gowa/src"
CGO_ENABLED=1 go build -ldflags="-w -s" -o "$HOME/gowa/whatsapp"
log_info "Binário compilado: ~/gowa/whatsapp"

# ═══════════════════════════════════════════
# STEP 6: Generate .env
# ═══════════════════════════════════════════
log_step "7/10 — Gerando .env"
cat > "$HOME/gowa/src/.env" << ENVEOF
# ═══════════════════════════════════════════
# APP
# ═══════════════════════════════════════════
APP_PORT=3000
APP_HOST=0.0.0.0
APP_DEBUG=false
APP_OS=Chrome
APP_BASIC_AUTH=admin:Gowa@2026!Pr0v
APP_BASE_PATH=
APP_TRUSTED_PROXIES=0.0.0.0/0

# ═══════════════════════════════════════════
# DATABASE (SQLite local)
# ═══════════════════════════════════════════
DB_URI=file:storages/whatsapp.db?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000
DB_KEYS_URI=

# ═══════════════════════════════════════════
# WHATSAPP
# ═══════════════════════════════════════════
WHATSAPP_LOG_LEVEL=ERROR
WHATSAPP_VERSION=2.3000.1035190227
WHATSAPP_PROXIES=
WHATSAPP_ACCOUNT_VALIDATION=false
WHATSAPP_PRESENCE_ON_CONNECT=unavailable
WHATSAPP_AUTO_REPLY=
WHATSAPP_AUTO_MARK_READ=true
WHATSAPP_AUTO_REJECT_CALL=true
WHATSAPP_AUTO_DOWNLOAD_MEDIA=false

# ═══════════════════════════════════════════
# WEBHOOK
# ═══════════════════════════════════════════
WHATSAPP_WEBHOOK=${WEBHOOK_URL}
WHATSAPP_WEBHOOK_SECRET=xstark1kk
WHATSAPP_WEBHOOK_INSECURE_SKIP_VERIFY=false
WHATSAPP_WEBHOOK_EVENTS=message,message.reaction,message.revoked,message.edited,message.ack,group.participants,connection,disconnection,login_success,device_removed,stream_replaced
WHATSAPP_WEBHOOK_INCLUDE_OUTGOING=false

# ═══════════════════════════════════════════
# CHAT STORAGE (SQLite local)
# ═══════════════════════════════════════════
WHATSAPP_CHAT_STORAGE=true

# ═══════════════════════════════════════════
# MEDIA LIMITS
# ═══════════════════════════════════════════
WHATSAPP_SETTING_MAX_IMAGE_SIZE=20000000
WHATSAPP_SETTING_MAX_FILE_SIZE=50000000
WHATSAPP_SETTING_MAX_VIDEO_SIZE=100000000
WHATSAPP_SETTING_MAX_DOWNLOAD_SIZE=500000000

# ═══════════════════════════════════════════
# CHATWOOT (desativado)
# ═══════════════════════════════════════════
CHATWOOT_ENABLED=false

# ═══════════════════════════════════════════
# TIMEZONE
# ═══════════════════════════════════════════
TZ=America/Sao_Paulo
ENVEOF
log_info ".env gerado com webhook: ${WEBHOOK_URL}"

# ═══════════════════════════════════════════
# STEP 7: Cloudflare Tunnel Setup
# ═══════════════════════════════════════════
log_step "8/10 — Configurando Cloudflare Tunnel"

TUNNEL_NAME="server_${CEL_ID}"
TUNNEL_HOSTNAME="${CEL_ID}.autopilots.trade"

# Login (if not already)
if [ ! -f "$HOME/.cloudflared/cert.pem" ]; then
    log_warn "Você precisa autenticar no Cloudflare."
    log_warn "Um link vai aparecer — abra no navegador do celular."
    echo ""
    cloudflared tunnel login
    echo ""
    if [ ! -f "$HOME/.cloudflared/cert.pem" ]; then
        log_error "Autenticação falhou. Rode 'cloudflared tunnel login' manualmente."
        exit 1
    fi
    log_info "Autenticação Cloudflare concluída"
else
    log_info "Cloudflare já autenticado"
fi

# Create tunnel (skip if exists)
EXISTING_TUNNEL=$(cloudflared tunnel list --output json 2>/dev/null | grep -o "\"id\":\"[^\"]*\"" | head -1 | cut -d'"' -f4 || echo "")
TUNNEL_ID=""

if cloudflared tunnel list 2>/dev/null | grep -q "$TUNNEL_NAME"; then
    log_warn "Tunnel '$TUNNEL_NAME' já existe"
    TUNNEL_ID=$(cloudflared tunnel list --output json 2>/dev/null | python3 -c "
import sys, json
tunnels = json.load(sys.stdin)
for t in tunnels:
    if t['name'] == '${TUNNEL_NAME}':
        print(t['id'])
        break
" 2>/dev/null || echo "")

    if [ -z "$TUNNEL_ID" ]; then
        # Fallback: parse from text output
        TUNNEL_ID=$(cloudflared tunnel list 2>/dev/null | grep "$TUNNEL_NAME" | awk '{print $1}')
    fi
else
    log_info "Criando tunnel '$TUNNEL_NAME'..."
    CREATE_OUTPUT=$(cloudflared tunnel create "$TUNNEL_NAME" 2>&1)
    echo "$CREATE_OUTPUT"
    TUNNEL_ID=$(echo "$CREATE_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -1)
fi

if [ -z "$TUNNEL_ID" ]; then
    log_error "Não foi possível obter o ID do tunnel. Configure manualmente."
    log_error "Execute: cloudflared tunnel list"
    exit 1
fi

log_info "Tunnel ID: ${TUNNEL_ID}"

# Route DNS
log_info "Configurando DNS: ${TUNNEL_HOSTNAME} → tunnel"
cloudflared tunnel route dns "$TUNNEL_NAME" "$TUNNEL_HOSTNAME" 2>&1 || log_warn "DNS route pode já existir"

# Generate config.yml
CREDS_FILE="$HOME/.cloudflared/${TUNNEL_ID}.json"
if [ ! -f "$CREDS_FILE" ]; then
    log_error "Arquivo de credenciais não encontrado: $CREDS_FILE"
    exit 1
fi

cat > "$HOME/.cloudflared/config.yml" << CFEOF
tunnel: ${TUNNEL_ID}
credentials-file: ${CREDS_FILE}

ingress:
  - hostname: ${TUNNEL_HOSTNAME}
    service: http://localhost:3000
  - service: http_status:404
CFEOF
log_info "config.yml gerado"

# ═══════════════════════════════════════════
# STEP 8: Watchdog Script
# ═══════════════════════════════════════════
log_step "9/10 — Criando watchdog e boot scripts"

cat > "$HOME/gowa/watchdog.sh" << 'WDEOF'
#!/data/data/com.termux/files/usr/bin/bash
while true; do
    cd ~/gowa/src
    ../whatsapp rest >> ~/gowa/gowa.log 2>&1
    echo "[$(date)] GOWA crashed, reiniciando em 5s..." >> ~/gowa/gowa.log
    sleep 5
done
WDEOF
chmod +x "$HOME/gowa/watchdog.sh"
log_info "watchdog.sh criado"

# Boot script
mkdir -p "$HOME/.termux/boot"
cat > "$HOME/.termux/boot/start-gowa.sh" << BOOTEOF
#!/data/data/com.termux/files/usr/bin/bash
termux-wake-lock
sleep 5
tmux new-session -ds gowa "$HOME/gowa/watchdog.sh"
tmux new-session -ds cf "cloudflared tunnel run ${TUNNEL_NAME}"
BOOTEOF
chmod +x "$HOME/.termux/boot/start-gowa.sh"
log_info "Boot script criado (auto-inicia no boot)"

# ═══════════════════════════════════════════
# STEP 9: Start Services
# ═══════════════════════════════════════════
log_step "10/10 — Iniciando serviços"

# Kill existing sessions if any
tmux kill-session -t gowa 2>/dev/null || true
tmux kill-session -t cf 2>/dev/null || true

# Start GOWA in tmux
tmux new-session -ds gowa "$HOME/gowa/watchdog.sh"
log_info "GOWA iniciada (tmux session: gowa)"

# Start Cloudflare tunnel in tmux
tmux new-session -ds cf "cloudflared tunnel run ${TUNNEL_NAME}"
log_info "Tunnel iniciado (tmux session: cf)"

# Wait for API to start
sleep 3

# ═══════════════════════════════════════════
# DONE
# ═══════════════════════════════════════════
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         INSTALAÇÃO CONCLUÍDA!            ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${BOLD}Celular:${NC}    ${CEL_ID}"
echo -e "  ${BOLD}API Local:${NC}  http://localhost:3000"
echo -e "  ${BOLD}API Tunnel:${NC} https://${TUNNEL_HOSTNAME}"
echo -e "  ${BOLD}Webhook:${NC}    ${WEBHOOK_URL}"
echo -e "  ${BOLD}Auth:${NC}       admin:Gowa@2026!Pr0v"
echo ""
echo -e "  ${BOLD}Comandos úteis:${NC}"
echo -e "    ${CYAN}tmux attach -t gowa${NC}    → Ver logs da API"
echo -e "    ${CYAN}tmux attach -t cf${NC}      → Ver logs do tunnel"
echo -e "    ${CYAN}cat ~/gowa/gowa.log${NC}    → Log completo"
echo -e "    ${CYAN}Ctrl+B, D${NC}              → Sair do tmux sem fechar"
echo ""
echo -e "  ${BOLD}Atualizar:${NC}"
echo -e "    ${CYAN}cd ~/gowa && git pull && cd src && go build -ldflags=\"-w -s\" -o ~/gowa/whatsapp${NC}"
echo -e "    ${CYAN}tmux kill-session -t gowa && tmux new-session -ds gowa ~/gowa/watchdog.sh${NC}"
echo ""
