# Instalação WCA — Cena V2 (WebView)

Projeto: `002-INSTALACAO_PROJETO [wca]`
Versão atual: **1.4**
Arquivo: `Instalação WCA 1.4.html` — **único arquivo**, com HTML + CSS + JS inline (sem `%project_path`, sem `index.css`/`index.js` externos).

Cena de progresso da instalação: anel de progresso com percentual e cores variáveis, mensagem de status, tempo decorrido e card informativo por arquivo (temas por etapa).

## Conversão V1 → V2 (só os métodos mudaram)

A lógica, o visual e os comportamentos são os mesmos da V1. As diferenças:

1. **Progresso sem injeção de JavaScript**: na V1 a tarefa injetava `javascript:window.PROGRESS_CONTROL.percent=%progress_install;void(0);`. Na V2 a cena escuta a variável **nativamente** via `Tasker.onVariableChanged('progress_install', ...)` — a etapa de injeção **não é mais necessária** (e pode ser removida da tarefa). O objeto `PROGRESS_CONTROL` foi mantido, então a injeção manual continua funcionando para testes;
2. **Modo da instalação sem variável embutida**: a V1 tinha `window.INSTALL_MODE = "%type_project";` dentro do JS (quebraria a cena inteira na V2). Agora o padrão é `project-api` e o valor real chega por `Tasker.getVariable('type_project')` (leitura inicial com prazo de 4s) + listener `type_project` no mesmo `onVariableChanged`;
3. **A cena fecha a SI MESMA ao atingir 100%** (pedido do usuário na 1.3 — exceção autorizada à regra "nunca se fechar no JS", replicando o comportamento da V1): conclui a UI, exibe **"Instalação concluída."** e, após uma pausa curta (`AUTO_FECHAMENTO_DELAY_MS`, 600ms; `0` = imediato), tenta fechar-se em cascata silenciosa: `Tasker.dismissSceneV2(WCA_SCENE_SCREEN_ID)` (API de tarefa exposta pelo bridge em algumas versões) → `Tasker.dismissLayout(...)` → `Tasker.cancelScreen("")` → `Tasker.runTask("003-FECHAR_INSTALACAO [wca]")`. O `WCA_SCENE_SCREEN_ID` (padrão `"1"`) deve ser o **MESMO ID da Tela** da ação Show Scene v2. Como o bridge oficial não documenta método de fechamento (verificado no `tasker.d.ts`), se **nenhuma** tentativa existir nesta versão do Tasker a cena permanece aberta — planos de fechamento no lado Tasker: (a) criar a tarefa auxiliar `003-FECHAR_INSTALACAO [wca]` com a ação **Fechar cena v2** (ID da Tela `1`) — usada automaticamente pela cascata; (b) no Java da tarefa de download: `tasker.setVariable("progress_install", "100"); ... tasker.dismissSceneV2("1");` (recebe o **ID da Tela**, não o nome). Consulta de estado de fora: `window.WCA_INSTALL_SCENE.isCompleted()` e `.percent()`;
4. **Valores iniciais**: se a cena abrir com o download já em andamento, o percentual atual é lido na inicialização (`getVariable` fire-and-forget — nada trava a abertura).

## Variáveis do Tasker consumidas pela cena

| Variável | Uso |
|---|---|
| `progress_install` | Percentual do download (0–99 publicado pelo Java; 100 conclui a UI). Recebida via `onVariableChanged` + leitura inicial |
| `type_project` | Modo dos cards: `project-task` = 1 card (PROJETO — baixando o `.prj.xml`); `project-api` = 2 cards (ARQUIVO 1 — whatsapp-wca, ARQUIVO 2 — ffmpeg-wca; **padrão** quando ausente) |

## Recursos externos (mantidos da V1)

- Google Fonts (Orbitron/Exo 2) e Tailwind CDN no `<head>`;
- Vídeo do cabeçalho: `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` (exige **Allow Phone Access** no elemento Web).

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** com o **JavaScript bridge** habilitado
- **Allow Phone Access** habilitado no elemento Web (vídeo via `file://`)

## Como atualizar no Tasker

1. Abra a cena `002-INSTALACAO_PROJETO` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado
4. Remova a etapa de injeção `javascript:window.PROGRESS_CONTROL...` da tarefa (substituída pelo listener nativo)

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.

## Histórico de versões

| Versão | Data | Mudanças |
|---|---|---|
| 1.4 | 12/11/2026 | **`Tasker.dismissSceneV2` incluído na cascata de autofechamento (sugestão do usuário após a 1.3 não fechar)**: passa a ser a PRIMEIRA tentativa (antes de dismissLayout/cancelScreen) — se o bridge da versão do usuário expuser o método, a cena fecha direto; se não existir, o try/catch segue para as próximas. O `runTask` da tarefa auxiliar ganhou `.catch` para não gerar rejeição silenciosa quando a 003 não existir. Lembrete: criar a tarefa `003-FECHAR_INSTALACAO [wca]` (ação Fechar cena v2, ID da Tela 1) garante o fechamento em qualquer versão |
| 1.3 | 12/11/2026 | **A cena fecha a SI MESMA ao concluir (pedido do usuário — comportamento da V1)**: ao atingir 100%, exibe "Instalação concluída." e, após 600ms (ajustável em `AUTO_FECHAMENTO_DELAY_MS`), tenta o autofechamento em cascata silenciosa: `dismissLayout` (ID da Tela da constante `WCA_SCENE_SCREEN_ID`) → `cancelScreen` → `runTask` para a tarefa auxiliar `003-FECHAR_INSTALACAO [wca]` (criar com a ação Fechar cena v2, ID da Tela 1). Se nenhuma via existir nesta versão do Tasker, a cena permanece aberta (fechamento no lado Tasker como antes). Exceção autorizada à regra "nunca se fechar no JS", válida para esta cena |
| 1.2b | 12/11/2026 | **README: receita definitiva do fechamento no 100%** — o `tasker.dismissSceneV2` é API de TAREFA (Java Code) e recebe o **ID da Tela** (não o nome da cena): `tasker.dismissSceneV2("1")` após publicar `progress_install = 100`. O bridge da cena não possui dismiss (verificado no `tasker.d.ts` oficial) — por isso a cena nunca se fecha sozinha |
| 1.2 | 12/11/2026 | **Remoção de TODAS as mensagens de console (pedido do usuário)**: as 2 ocorrências de `console.warn` (falha do listener `onVariableChanged` e da leitura inicial `getVariable`) foram eliminadas — os `catch` seguem capturando os erros silenciosamente, sem mudança de lógica |
| 1.1 | 12/11/2026 | **A cena permanece aberta ao concluir (pedido do usuário)**: reforço do comportamento sem `destroyScene` — ao atingir 100% a UI conclui, a mensagem de status passa a **"Instalação concluída."** e a cena permanece aberta até ser fechada pelo lado Tasker (fechamento opcional, no momento que o usuário escolher). Adicionado método público `window.WCA_INSTALL_SCENE` (`isCompleted()` / `percent()`) para consulta de estado via injeção de teste/lado Tasker |
| 1.0 | 12/11/2026 | Conversão da cena V1 → V2: arquivo único HTML+CSS+JS; progresso via `Tasker.onVariableChanged('progress_install')` (sem injeção javascript:); modo via `Tasker.getVariable` + listener `type_project` (sem `%type_project` embutido); `tk.destroyScene` removido — fechamento da cena passou a ser responsabilidade do Tasker; `PROGRESS_CONTROL` mantido para compatibilidade |
