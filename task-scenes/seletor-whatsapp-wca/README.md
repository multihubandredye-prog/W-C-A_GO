# Seletor de WhatsApp WCA — Cena V2 (WebView)

Projeto: `004-WHATSAPP_SELETOR [wca]`
Versão atual: **1.2**
Arquivo: `seletor-de-whatsapp-wca-1.2.html` — **único arquivo**, com HTML + CSS + JS inline.

- **Feedback de toque (versão atual)**: todo o CSS usa **`:active`** no lugar de `:hover` — a cena é usada só em celular, então o feedback visual acontece **no toque**;

- **Rodapé (versão atual)**: exibe **só o nome** do desenvolvedor (sem a palavra "Desenvolvedor:").

Painel de seleção do WhatsApp: quando mais de um aplicativo WhatsApp é detectado no dispositivo, exibe o badge "Sistema Detectado", o texto informativo e os dois botões de escolha — **WhatsApp Messenger** (verde) e **WhatsApp Business** (azul), separados pelo divisor "OU". Conteúdo estático (badge, texto e contagem são fixos, como no original).

## Variáveis do Tasker consumidas pela cena

**Nenhuma.** A cena é estática — toda a informação está no HTML.

## Saída da cena (AMBIENTE DO TASKER)

| Ação do usuário | O que acontece |
|---|---|
| Botão **WhatsApp Messenger** | Grava a variável de tela `bottom = wa-message` e fecha a cena |
| Botão **WhatsApp Business** | Grava a variável de tela `bottom = wa-business` e fecha a cena |

A variável `bottom` volta para a tarefa que exibiu a cena quando ela é dismissada. O fechamento é a mesma cascata validada dos outros projetos: `Tasker.dismissSceneV2(ID da Tela)` → `dismissLayout` → `cancelScreen` → `runTask` na tarefa auxiliar `004-FECHAR_WHATSAPP_SELETOR [wca]` (criar com a ação **Fechar cena v2** garante o fechamento em qualquer versão). A constante `WCA_SCENE_SCREEN_ID` (`'1'`) deve ser o **MESMO ID da Tela** configurado na ação *Cena v2 Mostrar*. Resposta dupla é ignorada (guarda interna — o primeiro clique vence).

**Receita no lado Tasker:** Show Scene v2 em modo **With Result** (Fullscreen With Result / Dialog With Result), **Screen ID `1`**, campo *Dispensar Automaticamente* vazio; após a ação, ler `%bottom` (`wa-message` | `wa-business`). O modo With Result pausa a tarefa até a cena fechar e devolve a resposta na sequência do fluxo.

## Arquivos que a cena lê

**Nenhum.**

## Recursos externos

- Google Fonts (Inter) e Font Awesome CDN no `<head>`;
- Vídeo do cabeçalho: `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` (exige **Allow Phone Access** no elemento Web).

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** com o **JavaScript bridge** habilitado
- **Allow Phone Access** habilitado no elemento Web (vídeo via `file://`)

## Como atualizar no Tasker

1. Abra a cena `004-WHATSAPP_SELETOR` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado
4. Confira o **ID da Tela** da ação *Cena v2 Mostrar* — o mesmo valor deve estar na constante `WCA_SCENE_SCREEN_ID` do HTML
5. A resposta chega em `bottom` (`wa-message` | `wa-business`) quando a cena fecha

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.
