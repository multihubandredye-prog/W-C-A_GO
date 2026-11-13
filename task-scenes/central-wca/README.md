# Central WCA — Cena V2 (WebView)

Projeto: `001-CENTRAL_WHATS_CONNECT_API [wca]`
Versão atual: **1.33**
Arquivo: `Central WCA 1.33.html` — **único arquivo**, com HTML + CSS + JS inline.

---

## Regras permanentes

1. **Versão do painel**: SEMPRE lida do arquivo `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` (sem extensão) — nunca de variável do Tasker e nunca fictícia (`loadVersionFile`, ativa no `initialize`). Se o arquivo EXISTE porém vem **vazio/inválido**, a cena busca a versão da **última release REAL do repositório** na API do GitHub (lista `/releases` — o rótulo do tipo vem ENTRE COLCHETES no título: `[UI]`, `[CORE]` ou `[FULL]`, com ou sem espaços/emoji; a lista chega da mais recente para a mais antiga e a primeira com rótulo válido define a versão; releases sem o rótulo no título — ex.: testes — são IGNORADAS; remove um eventual `v` inicial e valida com o mesmo formato do arquivo), exibe nos badges e **GRAVA de volta no arquivo** (`writeFile`) — a requisição só acontece nesse caso; se nada for obtido (sem internet, por exemplo, ou nenhuma release com rótulo), a cena mantém o padrão. Sem token embutido (API pública; o fallback é raro e o limite de 60 requisições/hora por IP é suficiente) — **nunca** colocar o token dentro do HTML da cena;
2. **Conexões/dispositivos**: SEMPRE no arquivo `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/connections/devices.json` (leitura com cache em memória, escrita com cache + arquivo);
3. **Lógica e requisições**: manter a lógica e as 31 chamadas à API intactas — alterações de métodos se restringem à camada de cena;
4. **Premium/licença**: SEMPRE lido do arquivo `/storage/emulated/0/.system_cache/.mode-system.log` (`loadModeSystemLog`).

Valores fictícios existem **apenas** em `USER_NAME` e `SHOW_WELCOME_MESSAGE` enquanto durar o combinado — a leitura real via `Tasker.getVariable` está pronta e desligada. `Premium_LOG_DATA` é REAL (arquivo `.mode-system.log`), e o **nome real do usuário** também é extraído do próprio `.mode-system.log` (campo `Usuário: ...`), substituindo o fictício sempre que o arquivo existir.

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** na cena com o **JavaScript bridge** habilitado (padrão ativo)
- **Allow Phone Access** habilitado no elemento Web — obrigatório para arquivos e tarefas. Atenção: em algumas versões essa opção pode desmarcar sozinha ao reeditar a cena; confira antes de testar

## Mapa de cliques

Todos os botões de sistema despacham `006-CONFIGURAR_SISTEMA [wca]` com o comando embutido em `par1` (`Tasker.runTask`):

| Função da cena | Ação | Valor |
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
| `taskInstall(data)` | `runTask` → 006 | `par1 = install-project` + `exit_data` (JSON number/session) |

Exceções que **não** executam tarefa — apenas gravam variáveis na cena (`Tasker.setVariable`):

| Função | Variáveis gravadas |
|---|---|
| `taskPremiumNotFound()` | `task = 5` e `task_exit = status_premium` |

As variáveis gravadas com `setVariable` são variáveis **da tela**: quando a cena for fechada (Dismiss Scene v2), elas são devolvidas à tarefa que a exibiu.

## Variáveis do Tasker consumidas pela cena

A cena V2 **não substitui variáveis dentro do JavaScript**, então nenhuma
`%variável` fica embutida no código. **A cena usa VALORES
FICTÍCIOS fixos no CONFIG** (`USER_NAME`, `SHOW_WELCOME_MESSAGE`); a leitura real em runtime via `Tasker.getVariable` está
pronta em `loadTaskerVars` (com prazo de 4s), porém **desligada** (chamada
comentada no `initialize`) — para religar, basta descomentar essa linha:

`user_name`, `name_exist`, `photo_profile`, `mode_system`, `open_tab`

**`open_tab` (aba inicial):** define a aba inicial da cena. **A PRIMEIRA aba (Instalação) é o padrão IMEDIATO do arranque** — sem variável, vazia, não definida ou com valor desconhecido, a cena SEMPRE abre nela, sem depender do bridge. A variável só PROMOVE: `two` = aba **Conexões** quando o valor chegar. **Como o valor chega (padrão do projeto de instalação):** (1) **LISTENER** `Tasker.onVariableChanged('open_tab')` — captura o valor definido/trocado COM a cena aberta (recomendado: colocar a ação *Definir variável %open_tab* DEPOIS da ação *Cena v2 Mostrar* — igual ao `progress_install` da Instalação, publicado com a cena aberta; se redefinir com o MESMO valor, o evento pode não disparar — o `return_modal` usa `%TIMES` por isso); (2) **SONDAGEM inicial de 8s** (16 tentativas de 500ms com prazo individual) — cobre o valor definido ANTES de abrir, sem evento. `one` não muda nada (a cena já está nela). Nunca sobrescreve a escolha do usuário (flag `__wcaUserTouchedTab`). **Atenção ao escopo:** variáveis com nome todo minúsculo (como `open_tab`) são **LOCAIS da tarefa** no Tasker — invisíveis para a cena antes/depois se a leitura do bridge não as alcançar; a receita garantida é definir DEPOIS do Show Scene v2 (listener). **Verificação da versão no aparelho:** procure `1.31` no HTML da cena (comentário do bloco da versão do painel)

## Arquivos que a cena lê/escreve

| Arquivo | Uso |
|---|---|
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/connections/devices.json` | Lido no início (cache em memória); escrito a cada alteração de dispositivos |
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` | **Versão do painel** (sem extensão): lida no início e exibida nos badges de versão. Se encontrado vazio: busca no repositório a última release com título real (`UI`/`CORE`/`FULL` entre colchetes; releases de teste sem o rótulo são ignoradas), exibe e grava de volta no arquivo |
| `/storage/emulated/0/.system_cache/.mode-system.log` | **Premium/licença** (status, expiração): lido no início e aplicado em `processLicense`/`setWelcome` |
| `icons/video-system.mp4` (vídeo do cabeçalho) | Cascata com reinício controlado e vigia por polling: `file:///storage/…/icons/video-system.mp4` → `icons/video-system.mp4` → `icons/video-system.webm` (o erro do primeiro src pode disparar antes do script; por isso o load é reiniciado com os listeners já anexados) |
| Histórico de modais (`MODAL_HISTORY_PATH`) | Escrito a cada modal aberto |
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/user-profile-picture/` | Reset de foto: lista (`listFiles`) e apaga todos os arquivos (`deleteFile`) |

## API do bridge utilizada

`Tasker.runTask` · `Tasker.setVariable` · `Tasker.getVariable` · `Tasker.readFile` · `Tasker.writeFile` · `Tasker.listFiles` · `Tasker.deleteFile`

Referência oficial: tasker.joaoapps.com/userguide/en/scenes_v2/webview-actions.html (e seção *JavaScript Bridge* do manual Scenes V2).

Observação: o `listFiles` do bridge retorna objeto de variáveis de saída da ação (conteúdo da lista na chave `files`, separado por linha).

## Botão voltar do Android (Back Pressed)

Mecanismo: a cena publica em `cache_modal` o id do modal aberto no topo (vazio = nenhum) e escuta `return_modal`; ao receber, fecha APENAS o modal do topo (e reabre o modal de origem quando o fechado foi aberto a partir dele — Menu → confirmações, Premium, Avaliações etc.).

**Montagem validada no aparelho (evento Back Pressed da tela):**
1. **Cancelar Evento: Dinâmico** com a expressão **`"%cache_modal" != ""`** (COM as aspas — sem elas a avaliação falha com valor preenchido e a cena fecha junto);
2. Ação **Definir Variável**: `return_modal` = `%TIMES` (valor diferente a cada press garante o disparo).

**Os nomes das variáveis ficam no JavaScript da cena** (constantes `WCA_MODAL_STATE_VAR` = `cache_modal` e `WCA_RETURN_MODAL_VAR` = `return_modal`, no bloco "Botao voltar do Android"). Podem ser renomeados para qualquer nome (letras/números/_, sem `%`), desde que o MESMO nome seja usado nos dois lados: na constante do HTML e na expressão/ação do handler do Tasker. **Nomes padronizados (12/11/2026, para replicar em qualquer cena):** `cache_modal` (estado — qual modal está no topo, vazio = nenhum) + `return_modal` (comando do voltar). **Regras do `%`:** OBRIGATÓRIO na expressão do Cancelar Evento (`"%cache_modal" != ""`) e no valor `%TIMES` (são referências a variáveis); OPCIONAL no campo da variável da ação Definir Variável (com ou sem `%` funcionam — os exemplos do próprio Tasker usam sem). Fechamento da cena continua sendo SEMPRE pelo Tasker (o JS nunca faz dismiss).

## Tela do QR Code / código de pareamento

- O QR gira sozinho enquanto o modal está aberto: quando o contador chega a zero (30 segundos), um novo QR é buscado automaticamente (`qr_duration` da API);
- **Sair da tela do QR por QUALQUER caminho (X, CANCELAR, tocar fora, Esc ou o botão voltar) desbloqueia o dispositivo**: o card deixa de mostrar "AGUARDANDO" e os botões **QR Code** e **Com código** voltam imediatamente, permitindo gerar um novo QR — antes, sair pelo voltar travava o card em "Aguardando conexão" sem os botões;
- Se o QR for pedido de novo enquanto a sessão de pareamento anterior ainda está viva na API (janela de alguns minutos), a API pode não gerar na hora — o modal mostra a mensagem "Não foi possível gerar o QR Code. Aguarde alguns segundos e toque em QR Code de novo." em vez de ficar carregando para sempre; passada a janela, o novo QR sai de primeira;
- Fechar o modal do QR **sem** ter conectado também cancela o pareamento do lado da cena (o dispositivo criado volta ao estado normal; se ele já estava conectado, segue conectado — nada é deslogado nessa situação).

## Configurações de desenvolvedor (arquivo .env da API)

O ícone de **controles (sliders)** na barra de status, ao lado do menu, abre o painel que edita o arquivo `/data/data/net.dinglisch.android.taskerm/files/WCA/.env` — as configurações da API no dispositivo:

- **Só o que pode mudar aparece no painel**: Mensagens (resposta automática, marcar como lida, recusar chamadas), Mídia (download automático, segundos antes de apagar), Webhook (ignorar certificado, atraso, tentativas de entrega, **ECO** de mensagens enviadas, eventos), Conta e presença (validação, presença ao conectar, pulso de presença e intervalos);
- Cada opção **ATIVADO/DESATIVADO** funciona como seleção (uma das duas); os **eventos do webhook** são caixas de seleção com **todas as 18 opções** que a API emite, com botões **MARCAR TODAS / DESMARCAR TODAS** (todas desmarcadas = a API volta a enviar todos os eventos);
- **Ícone "?" em cada item** abre um comentário curto explicando o que aquela configuração faz;
- **Primeira instalação**: o arquivo não precisa existir — o painel abre com os valores padrão e o arquivo só é **criado quando você salva** (sempre completo, no formato padrão do sistema);
- **Cada salvamento regenera o arquivo inteiro**: as chaves fixas (porta, host, credenciais de acesso, banco de dados, endereço e segredo do webhook, armazenamento de conversas) ficam sempre no padrão do sistema e não aparecem no painel;
- **Aplicar**: as mudanças valem **depois de reiniciar a API** — use **SALVAR E REINICIAR** (dispara o mesmo reinício do menu) ou **SALVAR** e reinicie depois;
- Validações impedem salvar valores inválidos (ex.: tentativas menores que 1, duração de pulso sem unidade **s/m/h** — o formato que a API aceita);
- O modal participa do botão voltar como os demais (fecha pelo voltar, X, tocar fora ou Esc).

## Como atualizar no Tasker

1. Abra a cena `001-CENTRAL_WHATS_CONNECT_API` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado no elemento Web
4. Teste os botões e acompanhe os Flash/log do `006-CONFIGURAR_SISTEMA`

## Política de versões

A partir da **1.29** as versões são **mantidas nesta pasta** — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda. As versões **1.0 a 1.28 foram removidas** (autorizações do usuário em 12/11/2026); o detalhamento de cada versão está nos **commits do repositório** (`git log task-scenes/central-wca`).

