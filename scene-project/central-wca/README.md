# Central WCA — Cena V2 (WebView)

Projeto: `001-CENTRAL_WHATS_CONNECT_API [wca]`
Versão atual: **1.0**
Arquivo: `Central WCA 1.0.html` — **único arquivo**, com HTML + CSS + JS inline (a partir desta versão o projeto não usa mais `index.css`/`index.js` separados nem `%project_path`).

---

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** na cena com o **JavaScript bridge** habilitado (padrão ativo)
- **Allow Phone Access** habilitado no elemento Web — obrigatório para arquivos e tarefas. Atenção: em algumas versões essa opção pode desmarcar sozinha ao reeditar a cena; confira antes de testar

## Mapa de cliques (fluxo V2)

Todos os botões de sistema despacham `006-CONFIGURAR_SISTEMA [wca]` com o comando embutido em `par1` (`Tasker.runTask`):

| Função da cena | Ação V2 | Valor |
|---|---|---|
| `alertOpenFaq` | `runTask` → 006 | `par1 = open-faq` |
| `alertTalkToDeveloper` | `runTask` → 006 | `par1 = talk-developer` |
| `alertSendPurchaseWhatsApp` | `runTask` → 006 | `par1 = buy-vip` |
| `alertUninstallWca` | `runTask` → 006 | `par1 = uninstall-wca` |
| `alertDeactivateWca` | `runTask` → 006 | `par1 = stop-wca` |
| `alertRestartWca` (também usado por `alertReactivateWca`) | `runTask` → 006 | `par1 = restart-wca` |
| `alertCheckUpdates` | `runTask` → 006 | `par1 = check_updates` |
| `alertReinstallApi` | `runTask` → 006 | `par1 = force-reinstall-api` |
| `alertReinstallTask` | `runTask` → 006 | `par1 = force-reinstall-task` |
| `alertCustomProfilePicture` | `runTask` → 006 | `par1 = profile-picture` |

Exceções que **não** executam tarefa — apenas gravam variáveis na cena (`Tasker.setVariable`):

| Função | Variáveis gravadas |
|---|---|
| `alertInstall(data)` | `task = 4` e `exit_data = <dados da instalação>` |
| `alertPremiumNotFound()` | `task = 5` e `task_exit = status_premium` |

As variáveis gravadas com `setVariable` são variáveis **da tela**: quando a cena for fechada (Dismiss Scene v2), elas são devolvidas à tarefa que a exibiu.

## Variáveis do Tasker consumidas pela cena

Resolvidas ao vivo pelo motor da V2 (com atualização sem recarregar a página):

`%user_name`, `%name_exist`, `%photo_profile`, `%last_update`, `%mode_system`, `%notification_caution`, `%open_tab`

## Arquivos que a cena lê/escreve

| Arquivo | Uso |
|---|---|
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/connections/devices.json` | Lido no início (`initialize` → cache em memória); escrito a cada alteração de dispositivos |
| Histórico de modais (`MODAL_HISTORY_PATH`) | Escrito a cada modal aberto |
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/user-profile-picture/` | Reset de foto: lista (`listFiles`) e apaga todos os arquivos (`deleteFile`) |

## API do bridge V2 utilizada

`Tasker.runTask` · `Tasker.setVariable` · `Tasker.readFile` · `Tasker.writeFile` · `Tasker.listFiles` · `Tasker.deleteFile`

Referência oficial: tasker.joaoapps.com/userguide/en/scenes_v2/webview-actions.html (e seção *JavaScript Bridge* do manual Scenes V2).

Observação: o `listFiles` do bridge retorna objeto de variáveis de saída da ação (conteúdo da lista na chave `files`, separado por linha).

## Mudanças da V1 para a V2 (1.0)

- **Arquivo único** — CSS e JS inline; sem `index.css`/`index.js` externos nem `%project_path`
- **`tk.*` → `Tasker.*`** — o bridge legado da V1 não se aplica à V2; todos os métodos viraram Promise (async/await)
- **`window.open` removido de todos os botões** — quem abre links/FAQ/contato agora é o 006 conforme o comando recebido em `par1`
- **`destroyScene` removido do JS** — o fechamento da cena passa a ser responsabilidade do lado Tasker (ex.: ação *Dismiss Scene v2* no fluxo do 006); o bridge V2 não destrói cena a partir do WebView
- **devices.json via cache em memória** — a V2 lê arquivos de forma assíncrona; o `initialize` aguarda o cache antes de renderizar, e todas as leituras sincronas passam a usar o cache (escritas atualizam cache + arquivo)
- Compatibilidade de comportamento mantida: mesmos nomes de funções, mesmos modais, card de atenção e demais fluxos de UI

## Como atualizar no Tasker

1. Abra a cena `001-CENTRAL_WHATS_CONNECT_API` (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral de `Central WCA 1.0.html`
3. Confirme que **Allow Phone Access** continua habilitado no elemento Web
4. Teste os botões e acompanhe os Flash/log do `006-CONFIGURAR_SISTEMA`

## Histórico de versões

| Versão | Data | Mudanças |
|---|---|---|
| 1.0 | 11/11/2026 | Migração da cena V1 → V2; arquivo único HTML+CSS+JS; novo fluxo de cliques (006 + `par1` embutido); `tk.*` → `Tasker.*`; README criado |
