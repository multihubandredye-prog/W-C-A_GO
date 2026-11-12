# Central WCA — Cena V2 (WebView)

Projeto: `001-CENTRAL_WHATS_CONNECT_API [wca]`
Versão atual: **1.22**
Arquivo: `Central WCA 1.22.html` — **único arquivo**, com HTML + CSS + JS inline (a partir desta versão o projeto não usa mais `index.css`/`index.js` separados nem `%project_path`).

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
| `taskOpenFaq` | `runTask` → 006 | `par1 = open-faq` |
| `taskTalkToDeveloper` | `runTask` → 006 | `par1 = talk-developer` |
| `taskSendPurchaseWhatsApp` | `runTask` → 006 | `par1 = buy-vip` |
| `taskUninstallWca` | `runTask` → 006 | `par1 = uninstall-wca` |
| `taskDeactivateWca` | `runTask` → 006 | `par1 = stop-wca` |
| `taskRestartWca` (também usado por `taskReactivateWca`) | `runTask` → 006 | `par1 = restart-wca` |
| `taskCheckUpdates` | `runTask` → 006 | `par1 = check_updates` |
| `taskReinstallApi` | `runTask` → 006 | `par1 = force-reinstall-api` |
| `alertReinstallTask` | `runTask` → 006 | `par1 = force-reinstall-task` |
| `taskEditProfilePicture` | `runTask` → 006 | `par1 = profile-picture` |

Exceções que **não** executam tarefa — apenas gravam variáveis na cena (`Tasker.setVariable`):

| Função | Variáveis gravadas |
|---|---|
| `taskInstall(data)` | `task = 4` e `exit_data = <dados da instalação>` |
| `taskPremiumNotFound()` | `task = 5` e `task_exit = status_premium` |

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

## Botão voltar do Android (Back Pressed)

Mecanismo: a cena publica em `cache_modal` o id do modal aberto no topo (vazio = nenhum) e escuta `return_modal`; ao receber, fecha APENAS o modal do topo (e reabre o modal de origem quando o fechado foi aberto a partir dele — Menu → confirmações, Premium, Avaliações etc.).

**Montagem validada no aparelho (evento Back Pressed da tela):**
1. **Cancelar Evento: Dinâmico** com a expressão **`"%cache_modal" != ""`** (COM as aspas — sem elas a avaliação falha com valor preenchido e a cena fecha junto);
2. Ação **Definir Variável**: `return_modal` = `%TIMES` (valor diferente a cada press garante o disparo).

**Os nomes das variáveis ficam no JavaScript da cena** (constantes `WCA_MODAL_STATE_VAR` = `cache_modal` e `WCA_RETURN_MODAL_VAR` = `return_modal`, no bloco "Botao voltar do Android"). Podem ser renomeados para qualquer nome (letras/números/_, sem `%`), desde que o MESMO nome seja usado nos dois lados: na constante do HTML e na expressão/ação do handler do Tasker. **Nomes padronizados (12/11/2026, para replicar em qualquer cena):** `cache_modal` (estado — qual modal está no topo, vazio = nenhum) + `return_modal` (comando do voltar). **Regras do `%`:** OBRIGATÓRIO na expressão do Cancelar Evento (`"%cache_modal" != ""`) e no valor `%TIMES` (são referências a variáveis); OPCIONAL no campo da variável da ação Definir Variável (com ou sem `%` funcionam — os exemplos do próprio Tasker usam sem). Fechamento da cena continua sendo SEMPRE pelo Tasker (o JS nunca faz dismiss).

## Como atualizar no Tasker

1. Abra a cena `001-CENTRAL_WHATS_CONNECT_API` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral de `Central WCA 1.0.html`
3. Confirme que **Allow Phone Access** continua habilitado no elemento Web
4. Teste os botões e acompanhe os Flash/log do `006-CONFIGURAR_SISTEMA`

## Política de versões

A partir da **1.21** as versões são **mantidas nesta pasta** — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda. As versões **1.0 a 1.20 foram removidas** (autorizações do usuário em 12/11/2026): o detalhamento de cada versão removida está nos commits do repositório, e a versão **1.21** é o ponto de retorno documentado (a 1.22 difere dela pelas correções do fluxo do modal de foto e pela proteção reforçada do `avatar-cache.url`).

## Histórico de versões

| Versão | Data | Mudanças |
|---|---|---|
| 1.22 | 12/11/2026 | **Modal de foto: confirmação de verdade + `avatar-cache.url` blindado**: o ícone de atualizar NÃO dispara mais a atualização ao ser clicado (restava um listener antigo da 1.13 no mesmo elemento — removido); o processo só começa no SIM da confirmação. Clicar em **RESETAR** ou no ícone agora **fecha o modal EDITAR FOTO** antes de abrir a confirmação (não ficam empilhados). **`avatar-cache.url` nunca mais eliminado**: a limpeza usa shell (`printf` cria-ou-trunca, nunca remove — era o `writeFile` vazio do bridge que apagava o arquivo no aparelho) e, ao carregar a foto, a cena recria o arquivo vazio caso algum fluxo externo o tenha removido |
| 1.21 | 12/11/2026 | **Correção de timing no alerta de servidor offline (bug do botão HABILITAR WCA)**: no arranque, o check de servidor e a leitura das conexões rodam em paralelo; se o status falhava ANTES das conexões chegarem, o alerta avaliava o cache vazio e mostrava "realize a primeira instalação" SEM o botão — ele só aparecia ao alternar abas (que re-executa o check). Agora, quando as conexões ou o log chegam depois de um check concluído offline, a UI é reaplicada na hora (mensagem correta + botão), sem depender de troca de aba. Mesmo grupo do bug da foto corrigido na 1.11 |

> O histórico detalhado das versões **1.0 a 1.20** (removidas com autorização do usuário em 12/11/2026) está nos **commits do repositório** (`git log task-scenes/central-wca`).
