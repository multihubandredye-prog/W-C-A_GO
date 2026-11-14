# Central WCA — Cena V2 (WebView)

Projeto: `001-CENTRAL_WHATS_CONNECT_API [wca]`
Versão atual: **1.47**
Arquivo: `central-wca-1.47.html` — **único arquivo**, com HTML + CSS + JS inline.

---

## Avisos do desenvolvedor acima do footer (1.47)

- **Banner destacado acima do footer** (fixo, com pulso e ícone de alerta) que lê o arquivo `wca-resources/updates/avisos` no repositório — mesmo padrão de fontes da cena Atualizações (API do GitHub primeiro, raw como fallback);
- O banner **só aparece quando há texto** no arquivo: arquivo vazio ou falha de rede = banner oculto, sem erro na tela;
- Para publicar um aviso, basta escrever o texto no arquivo `wca-resources/updates/avisos` do repositório — a Central exibe na próxima abertura.

## Chips de eventos: clique em qualquer lugar e +15% (1.46)

- **O chip INTEIRO do evento alterna** (antes só a caixinha): clique no fundo, no texto ou em qualquer área marca/desmarca; a caixinha alterna nativamente (sem dupla alternância) e o botão `?` **não desmarca** — só abre a explicação;
- Chips **~15% maiores**: texto 0.66→0.76rem, caixinha 13→15px, `?` 0.72→0.82rem, padding 8px 9px.

## Espaçamento interno dos inputs corrigido de vez (1.45)

- **Causa raiz** do espaçamento à esquerda dentro das caixas (visível desde as primeiras versões): a regra global `input[type="text"]` do painel tem `padding: .78rem .78rem .78rem 2.5rem` (espaço para o ícone dos campos do painel principal) e especificidade MAIOR que `.dev-input` — todos os inputs do painel dev herdavam o **padding esquerdo de ~40px**;
- Com o `text-align: center` dos campos numéricos, o texto centrava dentro do content box (que começa após esse padding) e ficava deslocado para a direita;
- **Correção**: as regras dos inputs do painel agora são `#devSettingsModal .dev-input` e `#devSettingsModal .dev-input-num` (especificidade maior que a global) — padding simétrico real (9px 12px no geral; 9px 6px no numérico) e o número centra de verdade na caixinha.

## Refinamento do design dos campos numéricos e do resultado (1.44)

- Os 3 campos numéricos (`AUTO_DELETE_MEDIA_DURATION`, `WHATSAPP_WEBHOOK_DELAY_MS`, `WHATSAPP_WEBHOOK_MAX_ATTEMPTS`) agora ficam em **uma única linha** (rótulo à esquerda + botão `?` + caixinha à direita);
- A caixinha numérica é **compacta** (110px) com o número **centralizado** dentro (placeholder do delay encurtado para `500` — a explicação completa continua no `?`);
- O modal de resultado do Verificar Premium ganhou **respiro entre a informação e o botão FECHAR** (`margin-top: 20px` nos botões).

## UX do premium check, Configurações da API e design (desde a 1.43)

- **Verificar Premium SEM toasts**: ao confirmar o SIM, abre um modal de **progresso** ("VERIFICANDO ASSINATURA" com spinner) e, ao concluir, o resultado aparece em um **modal pequeno** (ícone/título conforme o tipo, nome do usuário e "Vence em dd/mm/yyyy · N dias restantes / VENCE HOJE / ASSINATURA EXPIRADA") com botão FECHAR;
- **Fechar por qualquer via volta ao MENU** (nunca para a tela inicial): botão voltar (modais registrados no mapa de pais `WCA_MODAL_PAI`), clicar fora ou FECHAR — durante a verificação inclusive, o voltar reabre o menu e o resultado aparece quando a consulta terminar;
- **Configurações da API**: o item antes chamado "Configurações do desenvolvedor" agora se chama **"Configurações da API — Editar o arquivo .env do WCA"**, foi **movido para cima do item Ajuda** e, ao ser clicado, exibe um **modal de alerta formal** ("alterações incorretas podem afetar o funcionamento do WCA") antes de abrir o painel (CANCELAR/CONTINUAR); o título do painel também virou "CONFIGURAÇÕES DA API";
- **Voltar corrigido nos dois fluxos**: o painel de Configurações da API e os novos modais fecham para o MENU (o painel não estava registrado no mapa de pais e o voltar caía na tela inicial);
- **Design do painel**: inputs com texto **alinhado à esquerda** (o texto herdava o centralizado do modal — o "espaçamento esquisito") e os 3 campos numéricos (`AUTO_DELETE_MEDIA_DURATION`, `WHATSAPP_WEBHOOK_DELAY_MS`, `WHATSAPP_WEBHOOK_MAX_ATTEMPTS`) com **largura compacta** (max 170px) em vez da caixa gigante;

## Verificação Premium direta (desde a 1.42 — sem tarefa do Tasker)

- O botão **Verificar Premium** (menu → "Verificar Premium — Sincronizar assinatura") **não dispara mais tarefa nenhuma** (o antigo `task`=`5`/`task_exit`=`status_premium` foi extinto) — a verificação acontece **dentro da cena**, direto no link da Vercel;
- Fluxo replicado do código Java original: telefone do dispositivo **is_principal** do devices.json → **aquecimento** (5s) + pausa (2s) + **consulta real** (15s) em `https://wca-api-three-alpha.vercel.app?phone=...` → interpreta `nome`/`expiracao`/`status`;
- `status` **premium**/**expired** → PREMIUM com a expiração da resposta; qualquer outro (teste, `pending_activation`, cancelled) → **MODO TESTE** com a expiração guardada no `.sys_config.txt` ou +7 dias; **sem telefone/sem resposta** → repete o último estado guardado (nome ≠ "Teste" → PREMIUM);
- Grava **`.mode-system.log`** (linha no formato original: `[HH:mm:ss] TIPO | Usuário: NOME | Data de Expiração: dd/MM/yyyy | N dias restantes` / `VENCE HOJE` / `ASSINATURA EXPIRADA`) e **`.sys_config.txt`** (`{"nome","expiration","dias","lastCheck"}`) e **recarrega o log pelo pipeline oficial** (selo, nome, boas-vindas, foto);
- Feedback com toasts ("Sincronizando..." no início; PREMIUM ATIVO / MODO TESTE ATIVO / ASSINATURA EXPIRADA no fim). Nenhuma variável do Tasker é gravada.

## Desinstalação (desde a 1.41 — tudo dentro da cena)

- **Menu → Desinstalar WCA** abre a confirmação e, ao confirmar, roda um **assistente dentro da cena** (overlay dedicado, acima de tudo) com **barra de progresso e passo a passo**:
  1. **Etapa 1/3 — Desconectar os usuários**: consulta `GET /devices` da API (prazo de 3s); se a API **não estiver em execução**, a etapa é **pulada** com aviso; para cada dispositivo, o mesmo fluxo da lixeira (`/app/logout` + `DELETE /devices/{id}`) e limpeza do devices.json;
  2. **Etapa 2/3 — Eliminar os arquivos**: enumera os arquivos de `/data/data/net.dinglisch.android.taskerm/files/WCA` e `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api` via `find`, exibe **nome por nome** ("Eliminando (n de total): arquivo...") e apaga um a um com **100ms** de alternância; no fim, `rm -rf` nas duas pastas;
  3. **Etapa 3/3 — Concluído**: "Desinstalação concluída" e as **cenas se encerram em cascata** (`001-CENTRAL_WHATS_CONNECT_API [wca]V2` → `[wca]` → ID 1 → dismissLayout → cancelScreen);
- **Não usa mais a tarefa 006** (`uninstall-wca`) — a cena V2 executa tudo por conta própria (shell + API);
- **Configurações do desenvolvedor**: o botão saiu de junto ao menu e agora é um **item do próprio menu** ("Configurações do desenvolvedor — Parâmetros avançados (.env)") que abre o mesmo painel;

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

O ícone de **controles (sliders)** na barra de status, ao lado do menu, abre o painel que edita o arquivo `/data/data/net.dinglisch.android.taskerm/files/WCA/.env` — as configurações da API no dispositivo. O painel ocupa **a tela inteira** (cabeçalho no topo, lista de opções ocupando todo o espaço com rolagem, botões fixos no pé):

- **Aplicação**: **APP_OS** — o nome do app que o WhatsApp mostra como aparelho conectado. **Na instalação é gravado automaticamente** com o nome da sessão que você definir (o arquivo é criado na instalação com o padrão completo; sem nome definido, fica `WhatsConnectApi - Tasker`);
- **Mensagens**: resposta automática (texto), marcar como lida, recusar chamadas; **Mídia**: download automático e segundos antes de apagar; **Webhook**: atraso, tentativas de entrega, **ECO** de mensagens enviadas (`INCLUDE_OUTGOING`) e eventos; **Conta e presença**: validação, presença ao conectar e pulso de presença;
- Cada opção **ATIVADO/DESATIVADO** funciona como seleção (uma das duas); a **presença ao conectar** (available/unavailable/none) e os **tempos do pulso** são botões-chip com opções padrão (intervalo: 30m/1h/6h/12h/24h/48h; duração: 1m/5m/15m/30m/1h) — sem menus nativos do sistema;
- **Todos os ? do sistema abrem a mesma bolha flutuante** junto ao ícone, com título e explicação: as linhas do painel de desenvolvedor, cada evento do webhook (fecha com novo toque, toque fora, ao rolar a lista ou ao fechar o painel) e o ? do **nome da sessão** na instalação. A bolha fica em camada **acima do painel** e os toques nos ? são tratados por **delegação no documento** (um único listener), além de a bolha ser recriada sozinha se necessário;
- Os **eventos do webhook** são caixas de seleção com **todas as 19 opções** que a API emite, com botões **MARCAR TODAS / DESMARCAR TODAS** (todas desmarcadas = a API volta a enviar todos os eventos). **Cada evento tem o seu ?**: a explicação curta aparece numa **bolha flutuante** junto ao botão (fecha com novo toque, toque fora, ao rolar a lista ou ao fechar o painel). Destaques: `message` (mensagens enviadas e recebidas — o **ECO** controla se as enviadas entram), `chat_presence` (usuário digitando/gravando áudio), `message.ack` (mensagem **entregue/lida**), `status.response`;
- **Salvar pede confirmação**: SALVAR e SALVAR E REINICIAR abrem um modal de confirmação antes de gravar (o texto indica se a API será reiniciada); CANCELAR desiste;
- **Ícone "?" em cada item** abre um comentário curto explicando o que aquela configuração faz;
- **Cada salvamento regenera o arquivo inteiro** no formato padrão: as chaves fixas (porta, host, credenciais de acesso, banco de dados, endereço e segredo do webhook, `INSECURE_SKIP_VERIFY=false`, armazenamento de conversas) ficam sempre no padrão do sistema e não aparecem no painel;
- **Aplicar**: as mudanças valem **depois de reiniciar a API** — use **SALVAR E REINICIAR** (dispara o mesmo reinício do menu) ou **SALVAR** e reinicie depois;
- Validações impedem salvar valores inválidos (ex.: tentativas menores que 1, segundos não numéricos);
- O modal participa do botão voltar como os demais (fecha pelo voltar, X, tocar fora ou Esc).

**Arquivo inexistente**: sem o `.env`, a API sobe com os valores internos dela (que NÃO são o padrão do WCA — inclusive sem credenciais de acesso definidas). Por isso a instalação passa a **criar o arquivo automaticamente** com o padrão completo (e o `APP_OS` do nome da sessão — sem nome, `WhatsConnectApi - Tasker`). Antes da primeira instalação, o painel abre com os valores padrão sem ler nada.

## Como atualizar no Tasker

1. Abra a cena `001-CENTRAL_WHATS_CONNECT_API` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado no elemento Web
4. Teste os botões e acompanhe os Flash/log do `006-CONFIGURAR_SISTEMA`

## Política de versões

A partir da **1.29** as versões são **mantidas nesta pasta** — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda. As versões **1.0 a 1.28 foram removidas** (autorizações do usuário em 12/11/2026); o detalhamento de cada versão está nos **commits do repositório** (`git log task-scenes/central-wca`).

