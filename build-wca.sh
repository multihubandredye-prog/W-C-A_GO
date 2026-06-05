#!/data/data/com.termux/files/usr/bin/bash
# =============================================================
#  WCA - Build WhatsApp WCA
#  Gera: whatsapp-wca.bin (GoWA com ffmpeg hardcoded)
#  Saída: /storage/emulated/0/Download/a-binario-wca/
# =============================================================

set -e

GOWA_SRC="/data/data/com.termux/files/home/gowa/src"
OUTPUT_DIR="/storage/emulated/0/Download/a-binario-wca"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

step()  { echo -e "
${BOLD}${GREEN}▶ $1${NC}"; }
ok()    { echo -e "${GREEN}✓ $1${NC}"; }
warn()  { echo -e "${YELLOW}⚠  $1${NC}"; }
error() { echo -e "${RED}✗ ERRO: $1${NC}"; exit 1; }
info()  { echo -e "${CYAN}  → $1${NC}"; }

echo -e "${CYAN}"
echo "  ╔══════════════════════════════════════════════════╗"
echo "  ║  WCA - Build WhatsApp WCA                        ║"
echo "  ║  whatsapp-wca.bin com ffmpeg hardcoded           ║"
echo "  ╚══════════════════════════════════════════════════╝"
echo -e "${NC}"

# -------------------------------------------------------
# ETAPA 1 — Dependências
# -------------------------------------------------------
step "[1/4] Verificando dependências..."

pkg update -y 2>&1 | tail -3

for dep in golang git binutils; do
    if command -v "$dep" &>/dev/null; then
        ok "$dep OK"
    else
        info "Instalando $dep..."
        yes | pkg install -y "$dep" 2>&1 | tail -2
        ok "$dep instalado"
    fi
done

if ! command -v termux-elf-cleaner &>/dev/null; then
    yes | pkg install -y termux-elf-cleaner 2>&1 | tail -2 || warn "termux-elf-cleaner indisponível"
else
    ok "termux-elf-cleaner OK"
fi

[ -d "$GOWA_SRC" ] || error "Projeto GoWA não encontrado em $GOWA_SRC"
ok "Projeto GoWA encontrado"

# -------------------------------------------------------
# ETAPA 2 — Diretório de saída
# -------------------------------------------------------
step "[2/4] Criando diretório de saída..."

if [ ! -d "/storage/emulated/0" ]; then
    error "Storage indisponível — execute 'termux-setup-storage' antes"
fi

mkdir -p "$OUTPUT_DIR"
ok "Saída: $OUTPUT_DIR"

# -------------------------------------------------------
# ETAPA 3 — Patch inteligente no GoWA
# -------------------------------------------------------
step "[3/4] Aplicando patch inteligente em todos os .go..."

# Limpar patches anteriores
info "Removendo patches anteriores..."
find "$GOWA_SRC" -name "wca_ffmpeg_exec.go" -delete 2>/dev/null || true
ok "Patches anteriores removidos"

# ========== CORREÇÃO: REMOVER BACKUPS PARA GARANTIR A VERSÃO ATUAL ==========
info "Removendo arquivos .bak antigos para garantir uso da versão atual..."
find "$GOWA_SRC" -name "*.bak" -delete
# ============================================================================

# Detectar arquivos com chamadas ffmpeg
info "Varrendo projeto em busca de chamadas ffmpeg..."
FFMPEG_FILES=$(grep -rl 'exec\.LookPath("ffmpeg")\|exec\.Command("ffmpeg"\|exec\.CommandContext.*"ffmpeg"\|wcaFFmpegCmd\|wcaFFmpegBin\|wcaFFmpegCmdCtx' "$GOWA_SRC" --include="*.go" 2>/dev/null | grep -v "wca_ffmpeg_exec\.go" | grep -v "/vendor/" | grep -v "\.bak" | grep -v "Whats-Connect-Api-main" || true)

if [ -z "$FFMPEG_FILES" ]; then
    warn "Nenhum arquivo com chamadas ffmpeg encontrado"
else
    FILE_COUNT=$(echo "$FFMPEG_FILES" | wc -l)
    ok "$FILE_COUNT arquivo(s) detectado(s):"

    echo "$FFMPEG_FILES" | while read -r FILE; do
        DIR=$(dirname "$FILE")
        PKG=$(grep '^package ' "$FILE" | head -1 | awk '{print $2}')
        EXEC_FILE="$DIR/wca_ffmpeg_exec.go"

        info "  $(realpath --relative-to="$GOWA_SRC" "$FILE") (package $PKG)"

        # Criar wca_ffmpeg_exec.go
        if [ ! -f "$EXEC_FILE" ]; then
            cat > "$EXEC_FILE" << GOEXEC
package $PKG

import (
	"context"
	"os/exec"
)

const (
	wcaLinker64  = "/system/bin/linker64"
	wcaFFmpegBin = "/data/data/net.dinglisch.android.taskerm/files/WCA/ffmpeg-wca.bin"
)

func wcaFFmpegCmd(args ...string) *exec.Cmd {
	full := append([]string{wcaFFmpegBin, "ffmpeg"}, args...)
	return exec.Command(wcaLinker64, full...)
}

func wcaFFmpegCmdCtx(ctx context.Context, args ...string) *exec.Cmd {
	full := append([]string{wcaFFmpegBin, "ffmpeg"}, args...)
	return exec.CommandContext(ctx, wcaLinker64, full...)
}
GOEXEC
            ok "  wca_ffmpeg_exec.go criado"
        fi

        # Criar backup para esta rodada (Garanti que o arquivo original está limpo)
        cp -f "$FILE" "${FILE}.bak"

        # Contar ocorrências
        BEFORE=$(grep -c 'exec\.LookPath("ffmpeg")\|exec\.Command("ffmpeg"\|exec\.CommandContext.*"ffmpeg"' "$FILE" || true)

        # Aplicar patch
        sed -i 's/exec\.LookPath("ffmpeg")/func() (string, error) { return wcaFFmpegBin, nil }()/g' "$FILE"
        sed -i 's/exec\.Command("ffmpeg",/wcaFFmpegCmd(/g' "$FILE"
        sed -i 's/exec\.Command("ffmpeg")/wcaFFmpegCmd()/g' "$FILE"
        sed -i 's/exec\.CommandContext(\([^,]*\), "ffmpeg",/wcaFFmpegCmdCtx(\1,/g' "$FILE"

        ok "  $BEFORE ocorrências patcheadas"
    done
fi

# -------------------------------------------------------
# ETAPA 4 — Compilar whatsapp-wca.bin
# -------------------------------------------------------
step "[4/4] Compilando whatsapp-wca.bin..."

cd "$GOWA_SRC"

info "go mod tidy..."
go mod tidy 2>&1 | tail -5

info "Compilando GoWA..."
if ! CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o whatsapp-wca .; then

    warn "Compilação falhou — restaurando originais..."
    find "$GOWA_SRC" -name "*.bak" | while read -r bak; do
        mv -f "$bak" "${bak%.bak}"
    done
    find "$GOWA_SRC" -name "wca_ffmpeg_exec.go" -delete 2>/dev/null || true
    error "Compilação falhou."
fi

[ -f "$GOWA_SRC/whatsapp-wca" ] || error "whatsapp-wca não gerado"
termux-elf-cleaner whatsapp-wca 2>&1 || true
ok "whatsapp-wca compilado: $(du -sh $GOWA_SRC/whatsapp-wca | cut -f1)"

cp -f "$GOWA_SRC/whatsapp-wca" "$OUTPUT_DIR/whatsapp-wca.bin"
ok "whatsapp-wca.bin → $OUTPUT_DIR"

# ========== LIMPEZA: RESTAURAR ORIGINAIS APÓS BUILD ==========
info "Restaurando arquivos originais..."
find "$GOWA_SRC" -name "*.bak" | while read -r bak; do
    mv -f "$bak" "${bak%.bak}"
done
# ==============================================================

echo ""
echo -e "${CYAN}══════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${GREEN}  WCA compilado!${NC}"
echo -e "${CYAN}══════════════════════════════════════════════════════${NC}"
echo ""
