# Central WCA — Cena V2 (WebView)

Projeto: `001-CENTRAL_WHATS_CONNECT_API [wca]`
Versão atual: **1.15**
Arquivo: `Central WCA 1.15.html` — **único arquivo**, com HTML + CSS + JS inline (a partir desta versão o projeto não usa mais `index.css`/`index.js` separados nem `%project_path`).

---

## Regras permanentes

1. **Versão do painel**: SEMPRE lida do arquivo `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` (sem extensão) — nunca de variável do Tasker e nunca fictícia (`loadVersionFile`, ativa no `initialize`);
2. **Conexões/dispositivos**: SEMPRE no arquivo `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/connections/devices.json` (leitura com cache em memória, escrita com cache + arquivo);
3. **Lógica e requisições**: a lógica da V1 permanece intacta — a migração troca SOMENTE os métodos de cena (V1 → V2); as 31 chamadas à API são as mesmas;
4. **Premium/licença**: SEMPRE lido do arquivo `/storage/emulated/0/.system_cache/.mode-system.log` (`loadModeSystemLog`) — na V1 chegava pela variável `mode_system` do Tasker.

Valores fictícios existem **apenas** em `USER_NAME` e `SHOW_WELCOME_MESSAGE` enquanto durar o combinado — a leitura real via `Tasker.getVariable` está pronta e desligada. `Premium_LOG_DATA` voltou a ser REAL (arquivo `.mode-system.log`) a partir da 1.6, e a partir da 1.8 o **nome real do usuário** também é extraído do próprio `.mode-system.log` (campo `Usuário: ...`), substituindo o fictício sempre que o arquivo existir.

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** na cena com o **JavaScript bridge** habilitado (padrão ativo)
- **Allow Phone Access** habilitado no elemento Web — obrigatório para arquivos e tarefas. Atenção: em algumas versões essa opção pode desmarcar sozinha ao reeditar a cena; confira antes de testar

## Mapa de cliques (fluxo V2)

Todos os botões de sistema despacham `006-CONFIGURAR_SISTEMA [wca]` com o comando embutido em `par1` (`Tasker.runTask`):

| Função da cena | Ação V2 | Valor |
|---|---|---|
| `alertOpenFaq` | `runTask` → 006 | `par1 = open-faq` |
| `alertTalkToDeveloper` | `runTask` → 006 | `par1 = talk-developer` |
| `alertSendPurchaseWhatsApp` | `runTask` → 006 | `par1 = buy-vip` |
| `alertUninstallWca` | `runTask` → 006 | `par1 = uninstall-wca` |
| `alertDeactivateWca` | `runTask` → 006 | `par1 = stop-wca` |
| `alertRestartWca` (também usado por `alertReactivateWca`) | `runTask` → 006 | `par1 = restart-wca` |
| `alertCheckUpdates` | `runTask` → 006 | `par1 = check_updates` |
| `alertReinstallApi` | `runTask` → 006 | `par1 = force-reinstall-api` |
| `alertReinstallTask` | `runTask` → 006 | `par1 = force-reinstall-task` |
| `alertCustomProfilePicture` | `runTask` → 006 | `par1 = profile-picture` |

Exceções que **não** executam tarefa — apenas gravam variáveis na cena (`Tasker.setVariable`):

| Função | Variáveis gravadas |
|---|---|
| `alertInstall(data)` | `task = 4` e `exit_data = <dados da instalação>` |
| `alertPremiumNotFound()` | `task = 5` e `task_exit = status_premium` |

As variáveis gravadas com `setVariable` são variáveis **da tela**: quando a cena for fechada (Dismiss Scene v2), elas são devolvidas à tarefa que a exibiu.

## Variáveis do Tasker consumidas pela cena

A cena V2 **não substitui variáveis dentro do JavaScript**, então nenhuma
`%variável` fica embutida no código. **Por enquanto (1.4) a cena usa VALORES
FICTÍCIOS fixos no CONFIG** (`USER_NAME`, `SHOW_WELCOME_MESSAGE`,
`Premium_LOG_DATA`); a leitura real em runtime via `Tasker.getVariable` está
pronta em `loadTaskerVars` (com prazo de 4s), porém **desligada** (chamada
comentada no `initialize`) — para religar, basta descomentar essa linha:

`user_name`, `name_exist`, `photo_profile`, `mode_system`, `open_tab`

## Arquivos que a cena lê/escreve

| Arquivo | Uso |
|---|---|
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/connections/devices.json` | Lido no início (cache em memória); escrito a cada alteração de dispositivos |
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` | **Versão do painel** (sem extensão): lida no início e exibida nos badges de versão — não vem mais de variável do Tasker |
| `/storage/emulated/0/.system_cache/.mode-system.log` | **Premium/licença** (status, expiração): lido no início e aplicado em `processLicense`/`setWelcome` |
| `icons/video-system.mp4` (vídeo do cabeçalho) | Cascata com reinício controlado e vigia por polling: `file:///storage/…/icons/video-system.mp4` → `icons/video-system.mp4` → `icons/video-system.webm` (o erro do primeiro src pode disparar antes do script; por isso o load é reiniciado com os listeners já anexados) |
| Histórico de modais (`MODAL_HISTORY_PATH`) | Escrito a cada modal aberto |
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/user-profile-picture/` | Reset de foto: lista (`listFiles`) e apaga todos os arquivos (`deleteFile`) |

## API do bridge V2 utilizada

`Tasker.runTask` · `Tasker.setVariable` · `Tasker.getVariable` · `Tasker.readFile` · `Tasker.writeFile` · `Tasker.listFiles` · `Tasker.deleteFile`

Referência oficial: tasker.joaoapps.com/userguide/en/scenes_v2/webview-actions.html (e seção *JavaScript Bridge* do manual Scenes V2).

Observação: o `listFiles` do bridge retorna objeto de variáveis de saída da ação (conteúdo da lista na chave `files`, separado por linha).

## Mudanças da V1 para a V2 (1.0)

- **Arquivo único** — CSS e JS inline; sem `index.css`/`index.js` externos nem `%project_path`
- **`tk.*` → `Tasker.*`** — o bridge legado da V1 não se aplica à V2; todos os métodos viraram Promise (async/await)
- **`window.open` removido de todos os botões** — quem abre links/FAQ/contato agora é o 006 conforme o comando recebido em `par1`
- **`destroyScene` removido do JS** — o fechamento da cena passa a ser responsabilidade do lado Tasker (ex.: ação *Dismiss Scene v2* no fluxo do 006); o bridge V2 não destrói cena a partir do WebView
- **devices.json via cache em memória** — a V2 lê arquivos de forma assíncrona; o `initialize` aguarda o cache antes de renderizar, e todas as leituras sincronas passam a usar o cache (escritas atualizam cache + arquivo)
- Compatibilidade de comportamento mantida: mesmos nomes de funções, mesmos modais, card de atenção e demais fluxos de UI

## Como atualizar no Tasker

1. Abra a cena `001-CENTRAL_WHATS_CONNECT_API` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral de `Central WCA 1.0.html`
3. Confirme que **Allow Phone Access** continua habilitado no elemento Web
4. Teste os botões e acompanhe os Flash/log do `006-CONFIGURAR_SISTEMA`

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno a uma versão anterior sem perda.

## Histórico de versões

| Versão | Data | Mudanças |
|---|---|---|
| 1.0 | 11/11/2026 | Migração da cena V1 → V2; arquivo único HTML+CSS+JS; novo fluxo de cliques (006 + `par1` embutido); `tk.*` → `Tasker.*`; README criado |
| 1.1 | 11/11/2026 | Remove o card de atenção (aviso de desinstalação/reinstalação) — mensagem e lógica de exibição obsoletas; `%notification_caution` deixa de ser consumida; pasta do projeto no repositório renomeada para `task-scenes/` |
| 1.2 | 11/11/2026 | Correção de travamento: leituras do bridge (devices.json e versão) rodam em paralelo com prazo de 4s — a cena nunca mais fica presa em "CARREGANDO..." se o bridge não responder; a versão passa a ser lida do arquivo `system/version` (sem extensão) em vez da variável `%last_update`; badge estático do menu atualizado |
| 1.3 | 11/11/2026 | **Correção estrutural**: removidas TODAS as variáveis do Tasker embutidas no JavaScript (`%name_exist`/`%open_tab` "nus" quebravam a sintaxe do script inteiro na V2 e travavam a cena) — agora lidas em runtime via `Tasker.getVariable` e aplicadas no CONFIG/UI; vídeo do cabeçalho com fallback em cascata de caminhos; `task-scenes/README.md` criado no repositório |
| 1.15 | 11/11/2026 | **Botão voltar do Android fecha o modal do topo (voltar para o modal anterior)**: a cena publica em `wca_modal_aberto` o id do modal aberto no topo (vazio = nenhum) e escuta `return_modal` via `Tasker.onVariableChanged`. Montagem no Tasker (lado do usuário): evento **Back Pressed** da tela com **Cancel Event** = `wca_modal_aberto` e ação **SetVariable** `return_modal` = `%TIMES`. Com modal aberto: back consome e a cena fecha só o topo (Menu → confirmação → confirmação…, um por vez); sem modal: back fecha a cena normalmente. Pilha centralizada em `openModal`/`closeModal` (todas as 23 telas passam por elas — inclusive o Menu, que é o `helpModal`) + MutationObserver de segurança |
| 1.14 | 11/11/2026 | **`avatar-cache.url` nunca é eliminado (regra do usuário)**: no RESETAR e na limpeza de cache inválido, a IMAGEM (`avatar-cache.jpg`) continua sendo apagada como já era, mas o arquivo `avatar-cache.url` permanece na pasta com o conteúdo esvaziado (`writeFile` vazio no lugar de `deleteFile`) |
| 1.13 | 11/11/2026 | **Cache da foto do usuário (pedido do usuário)**: a URL do avatar devolvida pela API é temporária e NÃO dá erro quando a foto muda no WhatsApp (o CDN continua servindo a antiga até expirar) — por isso a detecção de troca é feita COMPARANDO a URL atual da API com a registrada. Funcionamento: ao obter o avatar pela API, a imagem é baixada para `system/user-profile-picture/avatar-cache.jpg` (via `httpRequest` com `fileToSave`) e a URL fica em `avatar-cache.url`; nas próximas aberturas a cena usa o arquivo local SEM nova requisição; se o arquivo local falhar, limpa o cache e busca na API automaticamente; botão **ATUALIZAR FOTO** no modal de editar foto (reconsulta a API: URL igual → "já está atualizada", URL nova → baixa e aplica). O RESETAR apaga também o cache |
| 1.12 | 11/11/2026 | **Reset da foto restaurado**: a reescrita da 1.9 havia cortado as ações finais do original em `resetProfilePicture` (fechar modal + toast "Imagem personalizada eliminada" + `requestProfileAvatar(3)`) — clicar em RESETAR não dava feedback nenhum. Restaurado: apaga os arquivos da pasta (se houver), fecha o modal, exibe o toast, limpa o caminho memorizado e re-busca o avatar pela API. A foto permanece **exclusiva do modo Premium** (como no original): sem Premium ativo/expirado ou boas-vindas desligadas, o bloco nem aparece e nenhuma requisição é feita |
| 1.11 | 11/11/2026 | **Foto fixada (bug de tempo)**: `resolveProfilePicture` roda no initialize e esconde o bloco da foto quando o Premium ainda não chegou — na V1 o estado Premium vinha injetado antes da cena abrir, na V2 ele chega DEPOIS (`.mode-system.log` assíncrono) e nada reaplicava a foto. Agora `loadModeSystemLog` também re-executa `resolveProfilePicture` quando o log chega. A pasta `user-profile-picture` estava vazia no teste (0 itens), então a foto vem da rota da API (`/user/avatar` → CDN do WhatsApp, https) |
| 1.10 | 11/11/2026 | **Foto do usuário restaurada (equivalente V2 do `%photo_profile`)**: na V1 o caminho da foto chegava pela variável Tasker; na V2 a cena localiza o arquivo direto em `system/user-profile-picture/` via `runShell` (`ls` + filtro de imagem) e exibe como `file://` (mesmo mecanismo do vídeo, exige Allow Phone Access). Sem arquivo na pasta, o fallback da API (`requestProfileAvatar` → `GET /user/avatar`) permanece como no original. O reset da foto também limpa o caminho memorizado |
| 1.9 | 11/11/2026 | **Camada de arquivos e requisições REFEITAS seguindo app de exemplo que funciona no aparelho (NOTAS GF)**: descoberto que o `Tasker.readFile` devolve lixo interno do bridge (era o "v scenev2_webview_action" no badge) — a leitura agora valida a resposta do `readFile` e, se inválida, lê via `Tasker.runShell` (`cat` com marcador de confirmação `__WCA_FILE__`), como no exemplo; `httpRequest` corrigido: headers são UMA STRING `Chave: valor` (não JSON) e a resposta é lida nas chaves reais `http_response_code`/`http_data`; TODAS as requisições http/https passam pelo bridge (fetch nativo só como fallback); gravações checam sinal de falha (`actionFailed`); listagem de fotos do perfil via `runShell` + `deleteFile`. Lógica/endpoints/regras intactos |
| 1.8 | 11/11/2026 | **Auditoria completa contra o original**: nada foi removido — `Avaliar o projeto`, `Desabilitar/Habilitar WCA` e `Renovar Premium` existem na cena e seguem as MESMAS regras de exibição do original (aparecem com conexão principal/servidor online; sumiam nos testes porque o servidor parecia offline e o devices.json não era lido — bugs corrigidos na 1.6/1.7). Os `window.open` do original eram o mecanismo de saída dos handlers, hoje equivalente ao `Tasker.runTask`. **Nome real do usuário** no welcome, extraído do `.mode-system.log`. **Certificação final**: 188 funções comuns ao original, 171 idênticas byte a byte; as 17 restantes diferem SOMENTE pela troca de método V1→V2 (`tk.*`/`window.open` → `Tasker.*`/`runTask`, mesmos valores de `par1`) e o `initialize` apenas GANHA o motor V2; as 16 funções novas são o motor da cena (bridge HTTP/arquivos, cascata de vídeo); 31 chamadas de rede intactas e na mesma ordem |
| 1.7 | 11/11/2026 | **Leitura de arquivos corrigida**: o `readFile` do bridge devolve a string pura do arquivo (o extrator antigo fatiava caractere por caractere — versão `1.3.34` virava `1`, devices e licença morriam no parse); `extractBridgeText` e `extrairRespostaBridge` agora aceitam string pura, chave `contents` e objetos de variáveis de saída |
| 1.6 | 11/11/2026 | **Camada HTTP via bridge**: a WebView V2 bloqueia HTTP puro (`http://localhost:3000`) — o `fetch` é redirecionado para `Tasker.httpRequest` (patch transparente; endpoints/headers/corpo idênticos, chamadas intactas), destravando status do servidor, devices e avatar; Premium/licença volta a ser REAL lendo `/storage/emulated/0/.system_cache/.mode-system.log`; `Premium_LOG_DATA` fictício removido |
| 1.5 | 11/11/2026 | Vídeo do cabeçalho: cascata reiniciada após a anexação dos listeners (o erro do src original disparava antes do script e a cascata nunca avançava) + vigia por polling que avança a cascata se nada carregar em 2s; auditoria confirmou as 31 chamadas fetch da V1 preservadas (devices, avatar, login/logout/reconnect, premium, reviews) |
| 1.4 | 11/11/2026 | Variáveis do Tasker passam a usar **valores fictícios** fixos no CONFIG (por enquanto): `USER_NAME="Usuário WCA"`, `SHOW_WELCOME_MESSAGE=true`, `Premium_LOG_DATA` com PREMIUM ativo expirando 31/12/2026; leitura real via `Tasker.getVariable` permanece pronta porém desligada (comentada no `initialize`) |
