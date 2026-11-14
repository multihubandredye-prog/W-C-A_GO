# Configuração do WebView

Projeto: `010-DOWNLOAD_VIEW [wca]`
Versão atual: **1.18**
Arquivo: `configuracao-webview-wca-1.18.html` — **único arquivo**, com HTML + CSS + JS inline.

- **Rodapé (versão atual)**: exibe **só o nome** do desenvolvedor (sem a palavra "Desenvolvedor:").

## O que a cena faz

Baixa o pacote **WCA-Tasker-Projetos** (reúne as cenas HTML e as mídias (ícones, áudio e vídeo), gerado automaticamente pelo GitHub Actions em cada release — **tudo dentro do zip da release, nada é baixado direto do repositório**) e salva o arquivo **zipado** na pasta do WebView do dispositivo:

```
/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/web-view
```

- **Release consultada na hora**: a cena pergunta à API do GitHub qual é a release mais recente que **já tem o pacote anexado** (procura o asset `WCA-Tasker-Projetos-<tag>.zip`, percorrendo da mais recente para a mais antiga) e mostra tag, nome do arquivo, tamanho em MB (**o zip já inclui as mídias**), data de publicação — a URL não é exibida;
- **Pacote padrão**: se a API não responder (sem internet/limite de requisições), a cena usa o pacote embutido **1.3.36** e o badge muda para "Pacote padrão";
- **Vídeo do cabeçalho**: mesmo caminho das demais cenas (`icons/video-system.mp4`);
- **Container centralizado na tela** (vertical e horizontal);
- **Cabeçalho sem vídeos**: só o título "WHATS CONNECT API" centrado (sem os círculos de vídeo ao lado — exclusividade visual desta cena);
- **Download feito INTEIRO pela cena, sem nenhuma tarefa do Tasker**, em cadeia de tentativas:
  1. `Tasker.httpRequest` com `fileToSave` e **header como UMA STRING** (`User-Agent: Tasker` — formato que o Tasker espera);
  2. o mesmo GET no **mínimo de opções** (sem header);
  3. reserva: `fetch` dentro do WebView convertendo os bytes em **base64** gravado via `writeFile` e decodificado no shell (`base64 -d`) — se falhar, o `.b64` parcial é apagado;
  4. **Descompactação automática (tudo na cena)**: o zip é lido em **base64** pelo shell, o **diretório central é parseado em JavaScript** (estrutura PK) e cada arquivo é regravado via base64 + shell, seguindo **rotas por pasta raiz do zip** — `task-scenes/` → pasta do WebView (mantendo a estrutura) e `icons/` → `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons` (as **mídias do pacote**, FORA do web-view; pasta raiz desconhecida é rejeitada); o **zip é apagado** no fim — resta só o conteúdo (o zip só permanece se a descompactação falhar, para permitir nova tentativa sem baixar de novo); guardas contra **zip-slip** e **contagem `find | wc -l` por destino**;
  5. o pacote do build é publicado **sem compressão interna** (ZIP_STORED), então cada entrada é um recorte direto dos bytes — sem depender de descompressão do WebView (entrada comprimida é inflada via `DecompressionStream`, se existir);
6. **mídias dentro do zip da release**: o build embute `wca-resources/media/` no pacote (achatado em `icons/`, quantos arquivos existirem, sem ocultos), então vídeo, ícones e imagens chegam no **mesmo download** — a cena extrai tudo do zip, **sem consultar nem baixar nada do repositório**.

## Fluxo da tela

1. Ao abrir: badge mostra "Verificando a última release..." e depois o resultado; o card do pacote (borda giratória) só aparece com os dados;
2. **BAIXAR PACOTE AGORA** abre o modal de confirmação ("Baixar o pacote?") com o nome do arquivo e a release;
3. **BAIXAR** no modal fecha o modal e dispara a tarefa de download; a linha de progresso mostra o estado (pedido enviado → iniciado → baixando);
4. **CANCELAR** (ou voltar/Esc) fecha o modal sem baixar;
5. **Ao receber `concluido`, a cena verifica fisicamente a pasta** (leitura via shell: a pasta precisa existir e ter pelo menos um arquivo gravado) e:
   - com arquivos confirmados: exibe apenas **"Download concluído."** e **se auto-fecha** sozinha (1,4s depois, via `dismissSceneV2` com o nome `010-DOWNLOAD_VIEW [wca]` ou o ID `1` — equivalentes; cascata completa como fallback);
   - sem arquivos: erro na linha de progresso + toast ("Nada foi gravado na pasta do WebView") e o botão volta a ficar habilitado para tentar de novo.

## Download (tudo dentro da cena — sem tarefas)

1. `createDirectory` (bridge) garante a pasta de destino — com `runShell mkdir -p` como reserva (o caminho **não é exibido na tela**);
2. `Tasker.httpRequest` com `method: GET`, `url` (browser_download_url do asset), **`fileToSave`** (pasta + nome do arquivo), `User-Agent: Tasker`, `timeout: 90` (segundos) e `autoRedirect: true` — segue o redirecionamento do GitHub e grava os bytes no disco (prazo total de 120s);
3. **Verificação física** via `runShell` (`wc -c`): o arquivo precisa existir, não estar vazio e ter **exatamente o tamanho** informado pela API da release (quando disponível);
4. Verificado → descompactação → **"Download concluído."** (centralizado) e auto-fechamento; se a descompactação falhar, o aviso informa e **o zip é mantido** na pasta;
5. Erro de rede/HTTP → mensagem "Não foi possível baixar o pacote. Tente novamente." com o botão de volta.

## Botão voltar (Back Pressed)

Mesma lógica das demais cenas: a cena publica em **`cache_modal`** o id do modal do topo (vazio = nenhum; único modal: `confirmBaixarModal`, sem origem) e escuta **`return_modal`** com wrapper `(nome, valor)` para fechar apenas o topo. Lado Tasker: Cancel Event `"cache_modal" != ""` (com aspas) + Definir Variável `return_modal = %TIMES`.

## Exibição

Cena **`010-DOWNLOAD_VIEW [wca]`** (V2), **ID da Tela `1`** (nome e ID equivalentes), dispensa automática vazia. Sem modal aberto, o voltar fecha a cena pelo fluxo do Tasker; após o download concluído, a cena se fecha sozinha.

## Detalhes

- Ações de bridge usadas: `httpRequest` (download com `fileToSave`), `runShell` (pasta + verificação), `deleteFile` (limpeza de parcial), `setVariable`/`onVariableChanged` (botão voltar) — todas com prazo e tolerância a falha (Allow Phone Access obrigatório);
- Consulta de release com prazo de 20s (`Promise.race`) e fallback embutido;
- **Etapas com tempo mínimo de exibição** (Baixando ≥ 1,8s; Descompactando ≥ 1,2s; tela final ≥ 2,6s antes de fechar) para o usuário acompanhar as fases — o fechamento automático continua acontecendo;
- **A URL do download não é exibida** na tela (fica só internamente);
- **Progresso anunciando cada arquivo**: "Baixando o pacote zip (x MB)..." e "Descompactando (n de total): nome..." — incluindo cada mídia extraída; o usuário vê exatamente o que está sendo processado a cada momento;
- **Contagem regressiva antes de encerrar**: após extrair e apagar o zip, exibe **"Download concluído. Reiniciando em 5..."** e conta 5 → 4 → 3 → 2 → 1 (1 segundo por número) e aí **executa a tarefa `001-CENTRAL_WHATS_CONNECT_API [wca]`** (reinicializa/reabre a Central com as cenas atualizadas) e **em seguida** se auto-encerra (substituiu o delay fixo de 2,6s);
- **Limpeza segura**: qualquer arquivo de limpeza (parcial `.b64` ou zip) só é apagado **depois de confirmar que existe** (shell `if [ -f ]`) — o `deleteFile` do Tasker mostra um toast nativo de erro quando o arquivo não existe, o que aparecia em falha de download sem internet;
- **Auto-encerramento com o identificador REAL da cena**: `010-DOWNLOAD_VIEW [wca]V2` (com o sufixo V2, como está no Tasker) — antes usava `010-DOWNLOAD_VIEW [wca]` sem o sufixo e o encerramento pelo nome não encontrava a cena; mantidos os fallbacks — e a cascata agora dispara **todos os métodos** (não para no primeiro que não lança erro, pois o Tasker pode aceitar em silêncio um identificador que não corresponde à tela: nome com V2, ID 1, dismissLayout e cancelScreen vão todos);
- **Textos centralizados**: o texto de progresso ("Baixando...", "Download concluído.", avisos) ocupa a linha inteira da caixa (flex:1) para o centrado valer de verdade — antes o span encolhia à esquerda da caixa flex mesmo com text-align: center; o parágrafo explicativo do card também ficou centrado;
- Toasts de sucesso/falha no mesmo estilo da Central;
- Sem variável do Tasker embutida no JavaScript (tudo lido em runtime), sem `console`.
