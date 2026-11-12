# Instalação WCA — Cena V2 (WebView)

Projeto: `002-INSTALACAO_PROJETO [wca]`
Versão atual: **1.5**
Arquivo: `Instalação WCA 1.5.html` — **único arquivo**, com HTML + CSS + JS inline.

Cena de progresso da instalação: anel de progresso com percentual e cores variáveis, mensagem de status, tempo decorrido e card informativo por arquivo (temas por etapa).

## Funcionamento

1. **Progresso**: a cena escuta a variável `progress_install` nativamente via `Tasker.onVariableChanged('progress_install', ...)`. O objeto `PROGRESS_CONTROL` é mantido, então a injeção manual `javascript:window.PROGRESS_CONTROL.percent=...` continua funcionando para testes;
2. **Modo da instalação (`type_project`)**: padrão `project-api`; o valor real chega por `Tasker.getVariable('type_project')` (leitura inicial com prazo de 4s) + listener `type_project` no mesmo `onVariableChanged` — escolhe os cards exibidos;
3. **A cena fecha a SI MESMA ao atingir 100%** (exceção autorizada à regra "nunca se fechar no JS"): conclui a UI, exibe **"Instalação concluída."** e, após uma pausa curta (`AUTO_FECHAMENTO_DELAY_MS`, 600ms; `0` = imediato), tenta fechar-se em cascata silenciosa: `Tasker.dismissSceneV2(WCA_SCENE_SCREEN_ID)` (API de tarefa exposta pelo bridge — **CONFIRMADO FUNCIONANDO no aparelho em 12/11/2026: foi o método que fechou a cena no teste real**) → `Tasker.dismissLayout(...)` → `Tasker.cancelScreen("")` → `Tasker.runTask("003-FECHAR_INSTALACAO [wca]")`. O `WCA_SCENE_SCREEN_ID` (padrão `"1"`) deve ser o **MESMO ID da Tela** da ação Show Scene v2. Como o bridge oficial não documenta método de fechamento (verificado no `tasker.d.ts`), se **nenhuma** tentativa existir nesta versão do Tasker a cena permanece aberta — planos de fechamento no lado Tasker: (a) criar a tarefa auxiliar `003-FECHAR_INSTALACAO [wca]` com a ação **Fechar cena v2** (ID da Tela `1`) — usada automaticamente pela cascata; (b) no Java da tarefa de download: `tasker.setVariable("progress_install", "100"); ... tasker.dismissSceneV2("1");` (recebe o **ID da Tela**, não o nome). Consulta de estado de fora: `window.WCA_INSTALL_SCENE.isCompleted()` e `.percent()`;
4. **Valores iniciais**: se a cena abrir com o download já em andamento, o percentual atual é lido na inicialização (`getVariable` fire-and-forget — nada trava a abertura).

## Variáveis do Tasker consumidas pela cena

| Variável | Uso |
|---|---|
| `progress_install` | Percentual do download (0–99 publicado pelo Java; 100 conclui a UI). Recebida via `onVariableChanged` + leitura inicial |
| `type_project` | Modo dos cards: `project-task` = 1 card (PROJETO — baixando o `.prj.xml`); `project-api` = 2 cards (ARQUIVO 1 — whatsapp-wca, ARQUIVO 2 — ffmpeg-wca; **padrão** quando ausente); `project-full` = 3 cards (ARQUIVO 1 — whatsapp-wca, ARQUIVO 2 — ffmpeg-wca, ARQUIVO 3 — `.prj.xml`) na ordem de download do Java. O valor `full` também é aceito como alias (é o valor testado no Java de download) |

## Recursos externos

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
4. O progresso chega pelo listener nativo — não é necessária injeção `javascript:` na tarefa

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.

