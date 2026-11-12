# Instalação WCA — Cena V2 (WebView)

Projeto: `002-INSTALACAO_PROJETO [wca]`
Versão atual: **1.0**
Arquivo: `Instalação WCA 1.0.html` — **único arquivo**, com HTML + CSS + JS inline (sem `%project_path`, sem `index.css`/`index.js` externos).

Cena de progresso da instalação: anel de progresso com percentual e cores variáveis, mensagem de status, tempo decorrido e card informativo por arquivo (temas por etapa).

## Conversão V1 → V2 (só os métodos mudaram)

A lógica, o visual e os comportamentos são os mesmos da V1. As diferenças:

1. **Progresso sem injeção de JavaScript**: na V1 a tarefa injetava `javascript:window.PROGRESS_CONTROL.percent=%progress_install;void(0);`. Na V2 a cena escuta a variável **nativamente** via `Tasker.onVariableChanged('progress_install', ...)` — a etapa de injeção **não é mais necessária** (e pode ser removida da tarefa). O objeto `PROGRESS_CONTROL` foi mantido, então a injeção manual continua funcionando para testes;
2. **Modo da instalação sem variável embutida**: a V1 tinha `window.INSTALL_MODE = "%type_project";` dentro do JS (quebraria a cena inteira na V2). Agora o padrão é `project-api` e o valor real chega por `Tasker.getVariable('type_project')` (leitura inicial com prazo de 4s) + listener `type_project` no mesmo `onVariableChanged`;
3. **A cena NUNCA se destrói**: a V1 chamava `tk.destroyScene(...)` ao atingir 100% — na V2 isso não existe (regra do projeto: fechamento é SEMPRE responsabilidade do Tasker). Ao chegar em 100 a cena apenas conclui a UI (anel cheio/verde). **Feche a cena no lado Tasker** quando `progress_install` chegar a 100 (ex.: evento *Variable Set* em `%progress_install` com valor 100 → *Close Scene v2*, ou no fim do fluxo Java após o sucesso);
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

A partir da **1.0** as versões são **mantidas nesta pasta** — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.

## Histórico de versões

| Versão | Data | Mudanças |
|---|---|---|
| 1.0 | 12/11/2026 | Conversão da cena V1 → V2: arquivo único HTML+CSS+JS; progresso via `Tasker.onVariableChanged('progress_install')` (sem injeção javascript:); modo via `Tasker.getVariable` + listener `type_project` (sem `%type_project` embutido); `tk.destroyScene` removido — fechamento da cena passou a ser responsabilidade do Tasker; `PROGRESS_CONTROL` mantido para compatibilidade |
