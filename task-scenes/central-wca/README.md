# Central WCA — Cena V2 (WebView)

Projeto: `001-CENTRAL_WHATS_CONNECT_API [wca]`

Versão atual: **2.19**

Arquivo atual: `central-wca-2.19.html`

A cena é um único arquivo HTML, com CSS e JavaScript inline, para uso no componente WebView das Scenes V2 do Tasker.


O cabeçalho não utiliza vídeo decorativo.

## Requisitos

- Tasker 6.7 ou superior, com Scenes V2;
- JavaScript Bridge habilitado no componente WebView;
- **Allow Phone Access** habilitado no elemento Web;
- arquivo completo da versão atual inserido no WebView da cena.

## Fontes de dados

| Arquivo | Uso |
|---|---|
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/connections/devices.json` | Lista e estado dos dispositivos, com cache em memória |
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/version` | Versão exibida no painel |
| `/storage/emulated/0/.system_cache/.mode-system.log` | Estado Premium, nome e expiração |
| `/storage/emulated/0/.system_cache/.sys_config.txt` | Último estado sincronizado da assinatura |
| `/data/data/net.dinglisch.android.taskerm/files/WCA/.env` | Configuração da API editada pela Central |
| `wca-resources/updates/avisos` | Fonte remota do aviso do desenvolvedor |
| `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/notices/notices.txt` | Cache local do aviso e horários em milissegundos |

Se o arquivo de versão existir, mas estiver vazio ou inválido, a Central consulta as releases públicas do repositório e usa a release real mais recente com `[UI]`, `[CORE]` ou `[FULL]` no título. Releases sem um desses rótulos são ignoradas. Nenhum token fica embutido na cena.

## Abas, dispositivos e instalação

- A aba **Instalação** é aberta imediatamente por padrão;
- `open_tab=two` promove a aba **Conexões** por listener e sondagem inicial;
- o listener recebe a variável depois da exibição da cena, preservando a compatibilidade com variáveis locais do Tasker;
- a escolha manual do usuário não é sobrescrita pela sondagem;
- dispositivos são lidos e gravados em `devices.json`; a leitura inicial do cache precede a sincronização e a renderização dos cartões;
- o primeiro dispositivo registrado é **PRIMÁRIO** e os seguintes são **AUXILIAR**; um principal já salvo é preservado e, quando registros existentes não têm principal, o primeiro válido é marcado;
- o fluxo de QR Code renova o código quando o contador chega a zero e libera o cartão ao fechar o modal;
- a instalação do primeiro número é direta ao Tasker: a cena valida o número, grava a sessão e entrega `number` e `session` em `exit_data` para `005-CONFIGURAR_SISTEMA [wca]` com `par1 = install-project`;
- ao digitar quando já há conexão no cache local de `devices.json`, o aviso informa que outros números devem ser adicionados na aba **Conexões**, via QR Code ou código; o botão de instalação fica bloqueado;
- antes de enviar a instalação, a cena aguarda a leitura local iniciada na abertura e verifica novamente se há conexão salva, preservando a instalação direta quando não há conexão registrada;
- o fluxo de instalação não inicia a API, não consulta `/devices` e não confere os arquivos do pacote — toda a configuração, inclusive a subida do serviço, acontece na tarefa do Tasker; a classificação usa os dados disponíveis no cache local;
- a trava `instalacaoEmAndamento` impede toque duplo enquanto o disparo está em curso;
- o retorno do disparo de `install-project` é exibido no campo e em toast, no lugar da falha silenciosa;
- a desinstalação tem quatro etapas: desconecta os usuários pela API, elimina os arquivos e as pastas do WCA, encerra o processo `whatsapp-wca-aarch64` pelo Shell e só então encerra as cenas em cascata;
- a limpeza das duas pastas do WCA é feita em lote, com tamanho aproximado em disco e tempo estimado quando houver dados suficientes; se o Shell não confirmar a remoção, a cena informa que a desinstalação não foi confirmada;
- o encerramento acontece depois da eliminação, por `runShell` com `ps -e -o PID,ARGS` + `grep` + `awk` e `kill -9`, sem depender de tarefa do Tasker;
- o filtro `awk -v me=$$` descarta o próprio shell do `runShell`, cuja linha de comando também contém o nome do binário;
- saída sem confirmação não desfaz a limpeza: o card informa o resultado e o encerramento das cenas acontece de qualquer forma.

## Premium

- Na tela inicial, um cartão **Pedido Premium pendente** mostra plano, valor e data e abre o gerenciamento ao tocar em **VER PEDIDO**. Ao voltar ou cancelar, retorna à tela inicial; quando aberto pelo menu, retorna ao menu. É exibido apenas quando há pendência para o número principal; o status é consultado ao identificar o número e a cada 60 segundos. Falhas de rede mantêm o último resumo confirmado durante a sessão;
- Os textos da Central usam **pedido** para o fluxo Premium, sem alterar os campos, endpoints ou identificadores da API;

- A verificação é feita diretamente pela cena, usando o telefone do dispositivo principal;
- o resultado atualiza `.mode-system.log` e `.sys_config.txt`;
- solicitações Premium pendentes permanecem individualizadas pelo `requestId`;
- edição ou cancelamento de uma solicitação não altera as demais;
- **ACELERAR PEDIDO** abre um seletor para enviar os dados da solicitação ao desenvolvedor pelo WhatsApp `+55 81 99757-3129` ou Telegram `@Andredye`;
- a mensagem começa com `*WATS-CONNECT-API*` e inclui nome, plano, telefone, valor e data da solicitação;
- cancelar ou voltar no seletor de canal retorna ao pedido Premium pendente;
- os modais de progresso e resultado retornam ao menu ao serem fechados.

## Avisos e interface

- O banner de avisos só aparece quando há conteúdo;
- O arquivo `system/notices/notices.txt` guarda em texto JSON o aviso (`texto`), o horário da última atualização bem-sucedida (`atualizadoEm`) e o da última tentativa sem sucesso (`tentativaEm`), em milissegundos, além do identificador atual e dos identificadores dispensados;
- O cache é exibido imediatamente e consultado novamente após 10 minutos; com a cena aberta, a próxima verificação é agendada conforme o horário salvo;
- A opção **Não exibir este aviso novamente** está somente no modal: ao marcar, salva o identificador no mesmo cache e só oculta após a gravação ser confirmada. Caso a gravação falhe, informa o erro e mantém o aviso;
- A identificação do aviso é calculada a partir de todo o conteúdo do arquivo remoto `wca-resources/updates/avisos`, incluindo espaços e quebras de linha. Qualquer alteração faz o aviso reaparecer após a próxima consulta, inclusive para quem dispensou o conteúdo anterior ou quando um texto antigo retorna. O intervalo de consulta continua sendo de 10 minutos. Na migração, um aviso já dispensado pode reaparecer uma vez se espaços ou quebras nas bordas não estiverem preservados no cache antigo.
- Avisos antigos sem ID usam uma identificação calculada a partir do texto, e caches da versão anterior continuam aceitos;
- Se o arquivo estiver ausente ou vazio, o aviso é buscado no repositório; a pasta só é criada se o arquivo não existir. Um aviso remoto vazio válido também é armazenado e oculta o banner;
- Em falha de rede, o último aviso é preservado e uma nova tentativa é limitada pelo intervalo de 10 minutos quando existe cache;
- quebras de linha são preservadas;
- o card compacto mostra uma prévia, o texto **Expandir** e uma borda animada discreta; tocar em qualquer parte abre o modal com o texto completo. No modal, o checkbox e o botão de fechamento têm espaçamento e o conteúdo rola em telas menores;
- todos os modais entram e saem com transição de foco, combinando ampliação e desfoque, incluindo confirmações, menu e telas modais;
- cancelar ou fechar um modal aberto pelo menu restaura automaticamente o menu de origem;
- o modal superior bloqueia a interação com o conteúdo de trás;
- a interface móvel usa feedback `:active`;
- o rodapé exibe somente o nome do desenvolvedor.

## Suporte técnico

A opção **Suporte Técnico** abre um modal para escolher o canal de atendimento:

- **WhatsApp:** `+55 81 99757-3129`;
- **Telegram:** `@Andredye`.

Os dois canais usam endereços HTTPS e são abertos fora do WebView por `Tasker.browseUrl`, com `Tasker.sendIntent` como alternativa. Assim, esquemas internos como `tg:resolve` não são carregados pela própria cena. Cancelar ou voltar no modal restaura o menu.

## Configurações da API

O item **Configurações da API — Editar o arquivo .env do WCA** abre um alerta e, após confirmação, apresenta o editor de tela cheia.

O painel usa nomes amigáveis para pessoas, sem expor as chaves técnicas como rótulos. Os eventos do webhook também aparecem com descrições legíveis, enquanto os nomes técnicos continuam sendo gravados corretamente no `.env`.

O painel permite editar:

- nome do dispositivo conectado (`APP_OS`);
- limite geral das requisições em MB (`APP_BODY_LIMIT`);
- resposta automática;
- marcação automática de leitura;
- rejeição automática de chamadas;
- recebimento de mídias: opção única de download com exclusão automática e duração em segundos, minutos, horas ou dias;
- limite geral de download, inclusive de links usados em envios;
- envio de mídias: limites para imagens, arquivos e vídeos;
- duração de retenção da mídia;
- tamanhos máximos, em MB, para imagem, arquivo, vídeo e download;
- atraso e tentativas do webhook;
- eventos enviados pelo webhook;
- inclusão de eventos de saída;
- validação da conta;
- presença inicial e pulso de presença.

Os 19 eventos de webhook podem ser selecionados individualmente. Cada opção e cada evento possui ajuda contextual. Deixar todos os eventos desmarcados mantém o comportamento da API de enviar todos.

### Padrão gerado

Cada salvamento regenera o `.env` inteiro. Os cinco limites abaixo são editáveis em MB na interface e convertidos automaticamente para bytes ao gravar. O padrão inicial recomendado é:

```dotenv
APP_PORT=3000
APP_HOST=0.0.0.0
APP_DEBUG=true
APP_BASIC_AUTH=user1:pass1,user2:pass2
APP_BASE_PATH=
APP_TRUSTED_PROXIES=0.0.0.0/0
APP_BODY_LIMIT=135000000
DB_URI=file:storages/whatsapp.db?_foreign_keys=on
DB_KEYS_URI=file::memory:?cache=shared&_foreign_keys=on
WHATSAPP_AUTO_DOWNLOAD_MEDIA=true
AUTO_DELETE_MEDIA=true
AUTO_DELETE_MEDIA_DURATION=30
WHATSAPP_MAX_IMAGE_SIZE=20000000
WHATSAPP_MAX_FILE_SIZE=50000000
WHATSAPP_MAX_VIDEO_SIZE=100000000
WHATSAPP_MAX_DOWNLOAD_SIZE=500000000
WHATSAPP_WEBHOOK=http://localhost:3129/Tasker
WHATSAPP_WEBHOOK_SECRET=tasker12345
WHATSAPP_WEBHOOK_INSECURE_SKIP_VERIFY=false
WHATSAPP_WEBHOOK_DELAY_MS=1000
WHATSAPP_WEBHOOK_INCLUDE_OUTGOING=true
WHATSAPP_ACCOUNT_VALIDATION=true
WHATSAPP_CHAT_STORAGE=true
```

Por exemplo, digitar `100` no campo de vídeo grava `WHATSAPP_MAX_VIDEO_SIZE=100000000`. Ao abrir um `.env` existente, a Central converte os bytes de volta para MB na interface.

Com o download automático desativado, o campo de tempo fica bloqueado e seu valor é preservado. Os limites de imagem, arquivo, vídeo e download permanecem editáveis porque também se aplicam a outros usos da API.

A instalação nova ativa `AUTO_DELETE_MEDIA=true` e usa 30 segundos. O seletor de unidade segue o visual da Central e, ao trocar de unidade, preenche 30 segundos, 1 minuto, 6 horas ou 1 dia; ao reabrir, mantém o tempo efetivamente salvo. A Central converte a duração escolhida em segundos e grava as chaves de download e exclusão com o mesmo estado. Assim, ao salvar uma configuração antiga em que o download estava ativo mas a exclusão não, a exclusão passa a ser ativada. A verificação é periódica e pode ocorrer alguns segundos depois do prazo.

O atraso padrão é **1000 ms** quando a chave não existe ou ao criar uma configuração padrão. Um valor existente, inclusive vazio, é preservado; vazio permanece sem atraso. A alteração passa a valer ao salvar o `.env` e reiniciar a API.

### Gravação e reinício

- **SALVAR** e **SALVAR E REINICIAR** pedem confirmação;
- depois da escrita, a Central relê o `.env` e exige igualdade exata com o conteúdo gerado;
- o reinício só é executado depois da confirmação da persistência;
- em caso de falha, a API não é reiniciada e o painel informa o erro;
- valores numéricos inválidos são bloqueados antes da gravação.


## Dependência do evento de digitação e gravação

Ao marcar **Digitando ou gravando áudio**, a Central abre uma confirmação explicando que o WhatsApp precisa manter o dispositivo WCA online. Se o usuário confirmar, a interface aplica automaticamente:

- **Presença ao conectar: ONLINE**;
- **Manter presença atualizada: DESATIVADO**.

O evento permanece marcado e o usuário deve salvar e reiniciar a API. Se cancelar ou fechar o aviso por qualquer caminho, o evento é desmarcado. A validação de salvamento também impede gravar `chat_presence` com uma combinação incompatível. O aviso informa que a visibilidade do status online depende das configurações de privacidade do WhatsApp.

Como `chat_presence` exige que a sessão permaneça online, ele começa desmarcado em novos arquivos `.env`; os outros 18 eventos permanecem selecionados. Ao abrir um `.env` antigo que tenha `chat_presence` marcado com presença incompatível, a Central também apresenta a confirmação para corrigir a configuração.

Ao desmarcar **Digitando ou gravando áudio**, a Central restaura automaticamente o comportamento padrão anterior: presença ao conectar em **OFFLINE**, pulso ativado, intervalo de `24h` e duração de `5m`. A mesma restauração ocorre ao cancelar ou fechar a confirmação sem ativar o evento.

## Reinstalação direta

As opções **Reinstalar API** e **Reinstalar Tasker/Projeto** fazem os downloads dentro da própria Central, usando os assets da release válida mais recente.

### Reinstalar API

1. localiza uma release válida que contenha `whatsapp-wca-aarch64.bin` e `ffmpeg-wca-aarch64.bin`;
2. baixa cada binário para um arquivo temporário `.update`;
3. acompanha o tamanho físico e mostra o progresso;
4. valida o tamanho, move o arquivo para o destino final e aplica permissão `700`;
5. depois dos dois downloads, executa o reinício existente com `par1 = restart-wca`.

A Central não usa mais `force-reinstall-api` para baixar os binários.

### Reinstalar Tasker/Projeto

1. localiza uma release válida com `W_C_A_API.prj.xml`;
2. baixa para `W_C_A_API.prj.xml.update`;
3. valida o tamanho e a presença de `<TaskerData` antes de substituir o XML atual;
4. depois da validação, exibe uma contagem regressiva de 10 até 0 e o botão **INSTALAR AGORA**;
5. somente quando o contador chega a zero ou o usuário toca no botão, executa `005-CONFIGURAR_SISTEMA [wca]` com `par1 = force-reinstall-task` e a variável local `bottom = yes`;
6. os dois caminhos usam a mesma continuação e enviam os mesmos valores;
7. uma trava impede disparos duplicados pelo botão e pelo contador.

Em caso de falha, o temporário é removido e a Central oferece **TENTAR NOVAMENTE**. O modal não pode ser fechado enquanto um download está em andamento.

## Botão voltar do Android

A cena publica o modal superior em `cache_modal` e escuta `return_modal`. O handler do evento **Back Pressed** deve:

1. cancelar dinamicamente o evento quando `"%cache_modal" != ""`;
2. definir `return_modal` como `%TIMES`.

O fechamento atinge somente o modal superior e restaura o modal de origem quando necessário.

## Integração com o Tasker

A cena utiliza:

- `Tasker.runTask`;
- `Tasker.browseUrl`;
- `Tasker.sendIntent`;
- `Tasker.setVariable`;
- `Tasker.getVariable`;
- `Tasker.readFile`;
- `Tasker.writeFile`;
- `Tasker.listFiles`;
- `Tasker.deleteFile`;
- `Tasker.runShell`;
- `Tasker.dismissSceneV2`.

Os comandos de sistema continuam sendo encaminhados à tarefa `005-CONFIGURAR_SISTEMA [wca]`, incluindo reinício, parada, atualização, reinstalação e instalação de projeto. O suporte agora abre diretamente o canal escolhido fora do WebView.

## Atualização no Tasker

1. Abra a cena `001-CENTRAL_WHATS_CONNECT_API` no editor;
2. substitua o conteúdo do WebView pelo HTML integral de `central-wca-2.3.html`;
3. confirme que **Allow Phone Access** continua habilitado;
4. valide abertura, modais, dispositivos, gravação do `.env` e reinício.

## Política de arquivos


O menu **Baixar cenas** pede confirmação antes de iniciar e, após confirmar, baixa o ZIP completo da release mais recente que o contém, valida e substitui apenas as cenas em `web-view/task-scenes/`. As mídias dentro do ZIP não são extraídas nem substituídas. Essa ação não reinstala o XML nem baixa binários. Em caso de falha, permite tentar novamente; arquivos já substituídos não são revertidos automaticamente.

A Central diferencia a disponibilidade da API do estado da conexão WhatsApp principal: consulta `state` em `/devices`, trata somente `logged_in` como conectado e apresenta estados de conexão, desconexão ou não verificado. O JID continua identificando a sessão salva e não comprova conexão ativa. A verificação periódica ocorre a cada 10 segundos.

A consulta de dispositivos não solicita reconexão e o botão de reconectar foi removido. A remoção usa logout para dispositivos online e DELETE para registros offline; este último remove apenas o registro da lista, preserva o cache de sessões e avisa que a sessão pode reaparecer. Estados intermediários ou desconhecidos não são removidos. Em caso de falha, informa que pode ter ocorrido remoção parcial; na remoção de todos, preserva os dispositivos não confirmados. Respostas de sucesso com lista vazia em `null` são aceitas.

Fechar o pareamento não executa logout. Listas vazias não apagam automaticamente o cache, inclusive durante inicialização. Dispositivos só são acrescentados ao cache quando a API informa estado logged_in. As escritas do arquivo de dispositivos são serializadas. Se a API não responder durante a remoção, o resultado é tratado como incerto.

O menu oferece **Reiniciar WCA**, com confirmação, para reiniciar o serviço pela tarefa do Tasker.

As confirmações `message.ack` possuem filtros independentes para entrega, leitura pelo destinatário e leitura pela própria conta. O evento geral precisa estar habilitado ou a lista geral de eventos deve estar vazia (o que permite todos); as três opções são gravadas no `.env` e aplicadas após reiniciar a API. Na configuração inicial da cena, o evento e os filtros ficam desativados. Em configurações existentes sem as chaves, é preservado o comportamento anterior quando o evento geral está permitido. A caixa geral controla a entrega; ao ativar uma leitura com a caixa desmarcada, é solicitada confirmação e a caixa é marcada. Se já estiver marcada, não há nova confirmação. Desmarcar a caixa ou desativar a última leitura desliga as duas leituras. As seções e a lista de eventos permanecem visíveis, sem recolhimento.

## Atendimento automático de chamadas

A versão 2.9 permite ativar o atendimento automático experimental de áudio individual e informar o MP3 privado na pasta `WCA/audio-call`. Ela grava `WHATSAPP_AUTO_ANSWER_CALL` e `WHATSAPP_AUTO_ANSWER_AUDIO` no `.env`; exige recusa automática desativada e reinício da API. Os valores são preservados ao salvar outras opções. A API precisa conter a implementação; chamadas reais e a chave de criptografia ainda precisam ser testadas.
