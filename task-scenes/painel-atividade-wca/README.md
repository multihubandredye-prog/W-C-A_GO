# Painel de Atividade WCA — Cena V2 (WebView)

Projeto: `003-PAINEL_ATIVIDADE [wca]`
Versão atual: **1.2**
Arquivo: `painel-de-atividade-wca-1.2.html` — **único arquivo**, com HTML + CSS + JS inline.

Fundo da cena: **transparente** (`--bg-body: transparent`) — a cena não pinta o fundo; o que aparece atrás é a superfície da própria cena (janela/dialogo) e o fundo configurado **no componente WebView** no editor da cena (para tirar o cinza, o fundo do componente WebView deve estar transparente lá).

Painel de atividade/informações: exibe uma mensagem enviada pelo Tasker, com cor do texto e posição vertical configuráveis e botões opcionais **Não/Sim**. Toda a informação exibida vem de variáveis do Tasker lidas em runtime — nenhuma variável fica embutida no código (a cena V2 não substitui variáveis dentro do `<script>`).

## Variáveis do Tasker consumidas pela cena

| Variável | Uso |
|---|---|
| `msg` | Conteúdo da mensagem. `&`, `<` e `>` são escapados e o texto é aplicado com quebra de linha preservada (`pre-wrap`); vazio/ausente = "(mensagem vazia)" |
| `color_text` | Cor do texto da mensagem (qualquer valor CSS de cor; vazio = cor padrão do painel) |
| `position_mode` | Posição vertical do painel: `top` · `center` · `bottom` (padrão `center` quando não chega ou vem inválido) |
| `bottom_active` | `true` = exibe a linha de botões **Não/Sim**; ausente/`false` = esconde |

**Como os valores chegam (padrão validado nos outros projetos):** (1) **LISTENER** `Tasker.onVariableChanged` nos 4 nomes — captura valores definidos/trocados COM a cena aberta; (2) **SONDAGEM inicial de 8s** (16 tentativas de 500ms, prazo individual de 1,5s por variável, para quando as 4 são obtidas) — cobre valores definidos ANTES de abrir. **Receita garantida no Tasker:** definir as variáveis DEPOIS da ação *Cena v2 Mostrar* (igual ao `open_tab` da Central e ao `progress_install` da Instalação). Escopo: nomes todo minúsculo são variáveis **LOCAIS** da tarefa.

## Saída da cena (AMBIENTE DO TASKER)

| Ação do usuário | O que acontece |
|---|---|
| Botão **Sim** | Grava a variável de tela `bottom = yes` e fecha a cena |
| Botão **Não** | Grava a variável de tela `bottom = no` e fecha a cena |

A variável `bottom` volta para a tarefa que exibiu a cena quando ela é dismissada. O fechamento é a mesma cascata validada da Instalação: `Tasker.dismissSceneV2(ID da Tela)` → `dismissLayout` → `cancelScreen` → `runTask` na tarefa auxiliar `003-FECHAR_PAINEL_ATIVIDADE [wca]` (criar com a ação **Fechar cena v2** garante o fechamento em qualquer versão). A constante `WCA_SCENE_SCREEN_ID` (`'1'`) deve ser o **MESMO ID da Tela** configurado na ação *Cena v2 Mostrar*. Resposta dupla é ignorada (guarda interna).

## Arquivos que a cena lê

**Nenhum.** A cena consome apenas variáveis do Tasker (tabela acima).

## Recursos externos

- Google Fonts (Inter) e Font Awesome CDN no `<head>`;
- Vídeo do cabeçalho: `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` (exige **Allow Phone Access** no elemento Web).

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** com o **JavaScript bridge** habilitado
- **Allow Phone Access** habilitado no elemento Web (vídeo via `file://`)

## Como atualizar no Tasker

1. Abra a cena `003-PAINEL_ATIVIDADE` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado
4. Confira o **ID da Tela** da ação *Cena v2 Mostrar* — o mesmo valor deve estar na constante `WCA_SCENE_SCREEN_ID` do HTML
5. Defina `msg`, `color_text`, `position_mode` e `bottom_active` DEPOIS do *Cena v2 Mostrar*; a resposta chega em `bottom` (`yes` | `no`) após a cena fechar

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.
