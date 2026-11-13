# Configuração do WebView

Projeto: `010-DOWNLOAD_VIEW [wca]`
Versão atual: **1.1**
Arquivo: `configuracao-webview-wca-1.1.html` — **único arquivo**, com HTML + CSS + JS inline.

## O que a cena faz

Baixa o pacote **WCA-Tasker-Projetos** (reúne as cenas HTML e os projetos do Tasker, gerado automaticamente pelo GitHub Actions em cada release) e salva o arquivo **zipado** na pasta do WebView do dispositivo:

```
/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/web-view
```

- **Release consultada na hora**: a cena pergunta à API do GitHub qual é a release mais recente que **já tem o pacote anexado** (procura o asset `WCA-Tasker-Projetos-<tag>.zip`, percorrendo da mais recente para a mais antiga) e mostra tag, nome do arquivo, tamanho em MB, data de publicação e a URL completa;
- **Pacote padrão**: se a API não responder (sem internet/limite de requisições), a cena usa o pacote embutido **1.3.36** e o badge muda para "Pacote padrão";
- **Vídeo do cabeçalho**: mesmo caminho das demais cenas (`icons/video-system.mp4`);
- **Container centralizado na tela** (vertical e horizontal).

## Fluxo da tela

1. Ao abrir: badge mostra "Verificando a última release..." e depois o resultado; o card do pacote (borda giratória) só aparece com os dados;
2. **BAIXAR PACOTE AGORA** abre o modal de confirmação ("Baixar o pacote?") com o nome do arquivo e a release;
3. **BAIXAR** no modal fecha o modal e dispara a tarefa de download; a linha de progresso mostra o estado (pedido enviado → iniciado → baixando);
4. **CANCELAR** (ou voltar/Esc) fecha o modal sem baixar;
5. **Ao receber `concluido`, a cena verifica fisicamente a pasta** (leitura via shell: a pasta precisa existir e ter pelo menos um arquivo gravado) e:
   - com arquivos confirmados: exibe apenas **"Download concluído."** e **se auto-fecha** sozinha (1,4s depois, via `dismissSceneV2` com o nome `010-DOWNLOAD_VIEW [wca]` ou o ID `1` — equivalentes; cascata completa como fallback);
   - sem arquivos: erro na linha de progresso + toast ("Nada foi gravado na pasta do WebView") e o botão volta a ficar habilitado para tentar de novo.

## Contrato com a tarefa `008-BAIXAR_PACOTE_WEBVIEW [wca]`

A cena dispara `runTask` com:

| Variável | Conteúdo |
|---|---|
| `par1` | URL completa do zip (browser_download_url do asset) |
| `par2` | Pasta de destino (`/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/web-view`) — **não é exibida na tela** |
| `par3` | Nome do arquivo (ex.: `WCA-Tasker-Projetos-1.3.36.zip`) |

A tarefa de download deve reportar o andamento gravando:

| Variável | Valores |
|---|---|
| `webview_status` | `iniciado` · `baixando` · `concluido` · `falhou` |
| `webview_progresso` | número de `0` a `100` (opcional; aparece como "n de 100") |

A cena recebe pelos dois caminhos (listener + sondagem de 1s com linha de base). Estados finais (`concluido`/`falhou`) não são sobrescritos por leituras atrasadas; a sondagem ignora o valor residual de uma sessão anterior (só o que mudar é aplicado).

**Verificação de salvamento**: após `concluido`, a cena executa um `runShell` que confirma pasta existente e não vazia antes de declarar o sucesso e fechar a cena.

## Botão voltar (Back Pressed)

Mesma lógica das demais cenas: a cena publica em **`cache_modal`** o id do modal do topo (vazio = nenhum; único modal: `confirmBaixarModal`, sem origem) e escuta **`return_modal`** com wrapper `(nome, valor)` para fechar apenas o topo. Lado Tasker: Cancel Event `"cache_modal" != ""` (com aspas) + Definir Variável `return_modal = %TIMES`.

## Exibição

Cena **`010-DOWNLOAD_VIEW [wca]`** (V2), **ID da Tela `1`** (nome e ID equivalentes), dispensa automática vazia. Sem modal aberto, o voltar fecha a cena pelo fluxo do Tasker; após o download concluído, a cena se fecha sozinha.

## Detalhes

- `Tasker.setVariable`/`getVariable`/`runTask` com prazo e tolerância a falha (Allow Phone Access obrigatório);
- Consulta de release com prazo de 20s (`Promise.race`) e fallback embutido;
- Toasts de sucesso/falha no mesmo estilo da Central;
- Sem variável do Tasker embutida no JavaScript (tudo lido em runtime), sem `console`.
