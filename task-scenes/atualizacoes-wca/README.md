# Atualizações WCA — Cena V2 (WebView)

Projeto: `006-ATUALIZACOES [wca]`
Versão atual: **1.6**
Arquivo: `atualizacoes-wca-1.6.html` — arquivo único, com HTML + CSS + JavaScript inline.

Painel de atualização direta do WCA. A cena identifica a release real mais recente cujo título contém `[UI]`, `[CORE]` ou `[FULL]`, compara a tag com a versão instalada e executa o modo correspondente sem encaminhar para a cena de instalação.

## Modos

| Rótulo da release | Modo | Operação |
|---|---|---|
| `[CORE]` | `core` | Baixa e substitui os dois binários; usa `type_project = project-api` |
| `[UI]` | `ui` / projeto | Baixa `W_C_A_API.prj.xml`; usa `type_project = project-task` |
| `[FULL]` | `full` | Atualiza os dois binários e baixa `W_C_A_API.prj.xml`; usa `type_project = project-full` |

## Destinos

| Arquivo | Destino |
|---|---|
| `whatsapp-wca-aarch64.bin` | `/data/data/net.dinglisch.android.taskerm/files/WCA/whatsapp-wca-aarch64.bin` |
| `ffmpeg-wca-aarch64.bin` | `/data/data/net.dinglisch.android.taskerm/files/WCA/ffmpeg-wca-aarch64.bin` |
| `W_C_A_API.prj.xml` | `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/tasks-send/W_C_A_API.prj.xml` |

Os downloads usam os assets da própria release detectada. Cada arquivo é baixado primeiro com o sufixo `.update`, validado e movido para o destino final. Os binários recebem permissão `700`; o projeto precisa conter `<TaskerData`. Quando o tamanho do asset está disponível na API, ele também é conferido antes da substituição.

A versão instalada é gravada em:

```text
/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version
```

## Encerramento e continuidade no Tasker

Após a conclusão, a cena mostra uma contagem regressiva de 10 segundos. O botão permite encerrar imediatamente.

Após os downloads, antes do encerramento, a cena define a variável local `type_project` e executa:

```text
006-CONFIGURAR_SISTEMA [wca]
```

A tarefa recebe:

```text
par1 = installation-update
```

O valor de `type_project` segue o modo:

| Modo | `type_project` |
|---|---|
| Projeto/UI | `project-task` |
| Core | `project-api` |
| Full | `project-full` |

A tarefa é disparada antes da contagem regressiva e do encerramento da cena.

## Tratamento de falhas

- O arquivo atual só é substituído depois que o download temporário passa nas validações;
- Em caso de erro, o temporário é removido;
- O botão muda para **TENTAR NOVAMENTE**;
- A cena permanece aberta e mostra a mensagem do erro;
- `006-CONFIGURAR_SISTEMA [wca]` só é disparada depois que todos os arquivos do modo passam nas validações.

## Recursos externos

- `https://api.github.com/repos/multihubandredye-prog/W-C-A_GO/releases?per_page=30` — release, tipo e assets;
- API `contents` e `raw.githubusercontent.com` — changelog;
- URLs `browser_download_url` dos assets da release — downloads;
- `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` — vídeo do cabeçalho.

## Requisitos

- Tasker 6.7 ou superior;
- Cena V2 com JavaScript bridge;
- **Allow Phone Access** habilitado;
- Permissão do Tasker para acessar o armazenamento compartilhado;
- Os assets obrigatórios devem existir na release correspondente.

## Como atualizar no Tasker

1. Abra a cena `006-ATUALIZACOES [wca]`;
2. Substitua o HTML do WebView pelo conteúdo integral da versão atual;
3. Confirme que **Allow Phone Access** está habilitado;
4. Na `006-CONFIGURAR_SISTEMA [wca]`, trate `par1 = installation-update` com `type_project = project-task | project-api | project-full`.
