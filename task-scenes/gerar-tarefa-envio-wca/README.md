# Gerar Tarefa de Envio WCA — Cena V2 (WebView)

Projeto: `005-GERAR_TAREFA_ENVIO [wca]`
Versão atual: **1.0**
Arquivo: `gerar-tarefa-de-envio-wca-1.0.html` — **único arquivo**, com HTML + CSS + JS inline.

Seletor de dispositivo para gerar as tarefas de envio: consulta a API local, lista os dispositivos **logados** (nome + número completo) e, ao escolher um número, entrega à tarefa os dados no formato que o fluxo de criação espera. Estado inicial "Sincronizando...", com mensagens de resultado conforme a resposta da API.

## Funcionamento

- **Busca dos dispositivos**: mesma chamada do projeto original (`GET http://localhost:3000/devices` com `Accept: application/json` e `Authorization: Basic`). Se o acesso direto falhar (por exemplo, bloqueio do WebView a conexões locais), tenta **uma vez pelo bridge** (`Tasker.httpRequest`, prazo de 15s) antes de indicar problema;
- **Resultados**: `code = SUCCESS` → botões dos dispositivos com `state = logged_in`; resposta sem sucesso → "Erro na API."; busca sem resposta → "Servidor Offline."; lista vazia → "Nenhum dispositivo logado.";
- **Botão do dispositivo**: ícone do WhatsApp, `display_name` (ou "WhatsApp") e o número completo (parte do `jid` antes do `@`).

## Variáveis do Tasker consumidas pela cena

**Nenhuma.** Toda a informação vem da API local.

## Saída da cena (AMBIENTE DO TASKER)

| Ação do usuário | O que acontece |
|---|---|
| Toque no **dispositivo** | Grava as variáveis de tela `create_task = yes` e `json` e fecha a cena |

O `json` mantém o formato do projeto original (indentação de 4):

```json
{
    "Name": "nome do dispositivo",
    "Phone": "número completo",
    "FinalNumber": "últimos 4 dígitos"
}
```

As duas variáveis voltam para a tarefa que exibiu a cena quando ela é dismissada. O fechamento é a mesma cascata validada dos outros projetos: `Tasker.dismissSceneV2(ID da Tela)` → `dismissLayout` → `cancelScreen` → `runTask` na tarefa auxiliar `005-FECHAR_GERAR_TAREFA_ENVIO [wca]` (criar com a ação **Fechar cena v2** garante o fechamento em qualquer versão). A constante `WCA_SCENE_SCREEN_ID` (`'1'`) deve ser o **MESMO ID da Tela** configurado na ação *Cena v2 Mostrar*. Resposta dupla é ignorada (guarda interna — o primeiro toque vence).

**Receita no lado Tasker:** Show Scene v2 em modo **With Result** (Fullscreen With Result / Dialog With Result), **Screen ID `1`**, campo *Dispensar Automaticamente* vazio; após a ação, conferir `%create_task` (`yes`) e ler `%json`.

## Arquivos que a cena lê

**Nenhum.** A consulta é feita por requisição à API local (`localhost:3000`).

## Recursos externos

- Google Fonts (Inter) e Font Awesome CDN no `<head>`;
- Vídeo do cabeçalho: `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` (exige **Allow Phone Access** no elemento Web).

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** com o **JavaScript bridge** habilitado
- **Allow Phone Access** habilitado no elemento Web (vídeo via `file://` e fallback de busca pelo bridge)

## Como atualizar no Tasker

1. Abra a cena `005-GERAR_TAREFA_ENVIO` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado
4. Confira o **ID da Tela** da ação *Cena v2 Mostrar* — o mesmo valor deve estar na constante `WCA_SCENE_SCREEN_ID` do HTML
5. A resposta chega em `%create_task` + `%json` quando a cena fecha — o fluxo de criação da tarefa usa o `json` como antes

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.
