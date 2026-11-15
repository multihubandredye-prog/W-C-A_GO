# Atualizações WCA — Cena V2 (WebView)

Projeto: `006-ATUALIZACOES [wca]`
Versão atual: **1.5**
Arquivo: `atualizacoes-wca-1.5.html` — arquivo único, com HTML + CSS + JavaScript inline.

Painel de atualização direta do WCA. A cena identifica a release real mais recente cujo título contém `[UI]`, `[CORE]` ou `[FULL]`, compara a tag com a versão instalada e executa o modo correspondente sem encaminhar para a cena de instalação.

## Modos

| Rótulo da release | Modo | Operação |
|---|---|---|
| `[CORE]` | `core` | Baixa e substitui `whatsapp-wca-aarch64.bin` e `ffmpeg-wca-aarch64.bin` |
| `[UI]` | `ui` / projeto | Baixa `W_C_A_API.prj.xml`, define `type_project = project-task` e encerra a cena para o Tasker continuar a importação |
| `[FULL]` | `full` | Atualiza os dois binários, baixa `W_C_A_API.prj.xml`, define `type_project = project-task` e encerra a cena |

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

Nos modos de projeto e full, antes do encerramento é definida a variável local:

```text
type_project = project-task
```

Ao encerrar a cena V2, a tarefa que a exibiu recebe a variável e pode executar a importação do projeto. O modo core não define `type_project`.

## Tratamento de falhas

- O arquivo atual só é substituído depois que o download temporário passa nas validações;
- Em caso de erro, o temporário é removido;
- O botão muda para **TENTAR NOVAMENTE**;
- A cena permanece aberta e mostra a mensagem do erro;
- Nenhuma tarefa instaladora é disparada pela cena.

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
4. Na tarefa que mostra a cena, após o retorno, trate `type_project = project-task` para importar o XML baixado.
