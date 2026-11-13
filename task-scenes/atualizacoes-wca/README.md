# Atualizações WCA — Cena V2 (WebView)

Projeto: `006-ATUALIZACOES [wca]`
Versão atual: **1.3**
Arquivo: `atualizacoes-wca-1.3.html` — **único arquivo**, com HTML + CSS + JS inline.

Painel de atualizações: mostra quando há nova release **real** no repositório — a mais recente cujo TÍTULO traz o rótulo do tipo entre colchetes (`[UI]`, `[CORE]` ou `[FULL]`, com ou sem espaços/emoji; releases sem o rótulo no título, como testes, são **ignoradas**) — compara com a versão instalada (lida do arquivo `.../system/version`) e exibe o badge com o tipo e a tag: `core` = `ATUALIZAÇÃO DA API • tag` · `ui` = `ATUALIZAÇÃO DO TASKER • tag` · `full` = `ATUALIZAÇÃO COMPLETA • tag`. Exibe também as novidades (changelog do repositório, seção expansível — toque/clique FORA da seção recolhe a caixa) e o botão **Atualizar Agora** (sem ação quando não há atualização), que executa a `006-CONFIGURAR_SISTEMA [wca]` com o comando e o tipo da atualização.

## Saída da cena (AMBIENTE DO TASKER)

| Função | Ação | Valores |
|---|---|---|
| `taskRunUpdate(tipo)` | `runTask` → `006-CONFIGURAR_SISTEMA [wca]` | `par1 = installation-update` (SEMPRE o mesmo, para qualquer tipo) · `type_update = ui \| core \| full` conforme o rótulo da release |

Referência para o lado Tasker: no disparo da 006 com `par1 = installation-update`, ler `type_update` — `ui` = só projeto do Tasker (XML), `core` = só binários (whatsapp/ffmpeg), `full` = tudo (equivalente ao `project-full` da Instalação).

## Variáveis do Tasker consumidas pela cena

**Nenhuma.** A versão local vem do arquivo `.../system/version` e nada mais é lido do Tasker — a cena também não grava variáveis de tela (ao fechar, nada retorna à tarefa que a exibiu).

## Arquivos que a cena lê

| Arquivo | Uso |
|---|---|
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` | **Versão instalada** (base da comparação com a tag da release; escrita pela Central quando encontra o arquivo vazio) |

## Recursos externos

- `https://api.github.com/repos/multihubandredye-prog/W-C-A_GO/releases?per_page=30` (checagem da release, sem token);
- Novidades do card expansível (`CHANGELOG_FONTES`): **1ª fonte** = API do GitHub (`contents/wca-resources/updates/changelog.txt` com `Accept: application/vnd.github.raw` — mesmo domínio das releases); **fallback** = `raw.githubusercontent.com/multihubandredye-prog/W-C-A_GO/main/wca-resources/updates/changelog.txt`; se todas falharem, exibe "Não foi possível carregar as novidades.";
- Vídeo do cabeçalho: `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` (exige **Allow Phone Access** no elemento Web).

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** com o **JavaScript bridge** habilitado
- **Allow Phone Access** habilitado no elemento Web (vídeo via `file://` e leitura do arquivo de versão)

## Como atualizar no Tasker

1. Abra a cena `006-ATUALIZACOES` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado
4. Na `006-CONFIGURAR_SISTEMA [wca]`, tratar `par1 = installation-update` + `type_update` (`ui` | `core` | `full`)

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.

