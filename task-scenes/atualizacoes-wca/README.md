# Atualizações WCA — Cena V2 (WebView)

Projeto: `006-ATUALIZACOES [wca]`
Versão atual: **1.1**
Arquivo: `Atualizações WCA 1.1.html` — **único arquivo**, com HTML + CSS + JS inline (sem `%project_path`, sem `index.css`/`index.js` externos).

Painel de atualizações: mostra quando há nova release **real** no repositório (badge com o tipo e a versão), as novidades (changelog do repo, seção expansível) e o botão **Atualizar Agora** — que executa a `006-CONFIGURAR_SISTEMA [wca]` com o comando e o tipo da atualização.

## O que mudou da V1 para a V2 (1.0)

A estrutura, o visual e a lógica de comparação são os mesmos da V1. As diferenças:

1. **Arquivo único** — CSS e JS inline; o CSS/HTML/JS da V1 foram mantidos (mesmos ids, classes, temas e textos);
2. **Versão local SEM variável** — a V1 lia `%last_update`; a V2 lê do **arquivo** `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` (o mesmo arquivo da Central), com validação de formato, fallback de leitura via shell e prazo de 15s (`lerVersaoLocal`); arquivo vazio/ausente = base de comparação vazia (qualquer release real conta como nova — mesmo comportamento da V1 quando a variável antiga chegava vazia);
3. **Release considerada = última REAL** — a lista `/releases` da API do GitHub é filtrada pelo **rótulo entre colchetes no título** (`[UI]`, `[CORE]` ou `[FULL]`, com ou sem espaços/emoji); releases sem o rótulo (ex.: testes) são **ignoradas** — mesma regra da Central 1.31. O tipo sai do próprio rótulo, já em minúsculas;
4. **Nada de sair da cena por conta própria** — REMOVIDOS da V1: `tk.setLocal('task'/'new_version')`, `window.open` e `tk.destroyScene`. A cena não grava variável de tela nenhuma e não abre nada: apenas **executa a tarefa** e quem conduz o resto (download, navegação e fechamento da cena) é o lado Tasker;
5. **Badge por tipo**: `core` = `ATUALIZAÇÃO DA API • tag`; `ui` = `ATUALIZAÇÃO DO TASKER • tag`; `full` = `ATUALIZAÇÃO COMPLETA • tag` (novo — a V1 só tinha API/TASKER);
6. **Zero console** (regra permanente das cenas V2).

## Saída da cena (AMBIENTE DO TASKER)

| Função | Ação | Valores |
|---|---|---|
| `taskRunUpdate(tipo)` | `runTask` → `006-CONFIGURAR_SISTEMA [wca]` | `par1 = installation-update` (SEMPRE o mesmo, para qualquer tipo) · `type_update = ui \| core \| full` conforme o rótulo da release |

Referência para o lado Tasker: no disparo da 006 com `par1 = installation-update`, ler `type_update` — `ui` = só projeto do Tasker (XML), `core` = só binários (whatsapp/ffmpeg), `full` = tudo (equivalente ao `project-full` da Instalação).

## Variáveis do Tasker consumidas pela cena

**Nenhuma.** A versão local vem do arquivo (item 2) e nada mais é lido do Tasker — a cena também não grava variáveis de tela (ao fechar, nada retorna à tarefa que a exibiu).

## Arquivos que a cena lê

| Arquivo | Uso |
|---|---|
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` | **Versão instalada** (base da comparação com a tag da release; escrita pela Central quando encontra o arquivo vazio) |

## Recursos externos

- `https://api.github.com/repos/multihubandredye-prog/W-C-A_GO/releases?per_page=30` (checagem da release, sem token);
- `https://raw.githubusercontent.com/multihubandredye-prog/W-C-A_GO/main/wca-resources/updates/changelog.txt` (novidades do card expansível);
- Vídeo do cabeçalho: `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` (exige **Allow Phone Access** no elemento Web).

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** com o **JavaScript bridge** habilitado
- **Allow Phone Access** habilitado no elemento Web (vídeo via `file://` e leitura do arquivo de versão)

## Como atualizar no Tasker

1. Abra a cena `006-ATUALIZACOES` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado
4. O fluxo antigo que esperava `task`/`new_version` após a cena não é mais necessário — passe a tratar `par1 = installation-update` + `type_update` na 006

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.

## Histórico de versões

| Versão | Data | Mudanças |
|---|---|---|
| 1.1 | 16/09/2026 | **Correção: "Últimas Novidades" não expandia no aparelho**. A conversão 1.0 estava fiel à V1 (CSS com 0 diferenças e mesmo código de toggle — auditado), mas o WebView V2 não gerava o evento `click` de forma confiável no toque sobre aquele DIV (na V1 o script era `type="module"` no WebView antigo; nos projetos V2 validados os alvos de clique são `<button>`). A expansão agora trata o **toque direto** (`touchstart`/`touchend` no header: tap detectado pela distância percorrida — rolagem não dispara; `preventDefault` no touchend evita gerar clique duplicado) **além** do `click` normal, com guarda de 400ms para nunca alternar duas vezes na mesma interação; ao expandir aplica também `maxHeight` inline (`20rem`) como garantia visual, além da classe `expanded`; seta do changelog gira via `classList.toggle('rotated')` |
| 1.0 | 16/09/2026 | Conversão da cena V1 → V2 (arquivo único HTML+CSS+JS): versão local lida do **arquivo** `.../system/version` (não mais da variável antiga); release real filtrada pelo rótulo `[UI]`/`[CORE]`/`[FULL]` no título (testes ignorados); botão **Atualizar Agora** executa a `006-CONFIGURAR_SISTEMA [wca]` com `par1 = installation-update` + `type_update = ui \| core \| full` (valores em minúsculas, escolhido pelo usuário) — removidos `setLocal`, `window.open` e `destroyScene` da V1; novo badge `ATUALIZAÇÃO COMPLETA` para o tipo `full`; changelog e expansão idênticos à V1 |
