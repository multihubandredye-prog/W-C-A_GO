# Central WCA — Cena V2 (WebView)

Projeto: `001-CENTRAL_WHATS_CONNECT_API [wca]`

Versão atual: **1.73**

Arquivo atual: `central-wca-1.73.html`

A cena é um único arquivo HTML, com CSS e JavaScript inline, para uso no componente WebView das Scenes V2 do Tasker.

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
| `wca-resources/updates/avisos` | Aviso do desenvolvedor exibido acima do rodapé |

Se o arquivo de versão existir, mas estiver vazio ou inválido, a Central consulta as releases públicas do repositório e usa a release real mais recente com `[UI]`, `[CORE]` ou `[FULL]` no título. Releases sem um desses rótulos são ignoradas. Nenhum token fica embutido na cena.

## Abas, dispositivos e instalação

- A aba **Instalação** é aberta imediatamente por padrão;
- `open_tab=two` promove a aba **Conexões** por listener e sondagem inicial;
- o listener recebe a variável depois da exibição da cena, preservando a compatibilidade com variáveis locais do Tasker;
- a escolha manual do usuário não é sobrescrita pela sondagem;
- dispositivos são lidos e gravados em `devices.json`;
- o fluxo de QR Code renova o código quando o contador chega a zero e libera o cartão ao fechar o modal;
- a instalação do número é direta ao Tasker: a cena valida o número, grava a sessão e entrega `number` e `session` em `exit_data` para `006-CONFIGURAR_SISTEMA [wca]` com `par1 = install-project`;
- a cena não inicia a API, não consulta `/devices` e não confere os arquivos do pacote — toda a configuração, inclusive a subida do serviço, acontece na tarefa do Tasker;
- a trava `instalacaoEmAndamento` impede toque duplo enquanto o disparo está em curso;
- o retorno do disparo de `install-project` é exibido no campo e em toast, no lugar da falha silenciosa;
- a desinstalação tem quatro etapas: desconecta os usuários pela API, elimina os arquivos e as pastas do WCA, encerra o processo `whatsapp-wca-aarch64` pelo Shell e só então encerra as cenas em cascata;
- o encerramento acontece depois da eliminação, por `runShell` com `ps -e -o PID,ARGS` + `grep` + `awk` e `kill -9`, sem depender de tarefa do Tasker;
- o filtro `awk -v me=$$` descarta o próprio shell do `runShell`, cuja linha de comando também contém o nome do binário;
- saída sem confirmação não desfaz a limpeza: o card informa o resultado e o encerramento das cenas acontece de qualquer forma.

## Premium

- A verificação é feita diretamente pela cena, usando o telefone do dispositivo principal;
- o resultado atualiza `.mode-system.log` e `.sys_config.txt`;
- solicitações Premium pendentes permanecem individualizadas pelo `requestId`;
- edição ou cancelamento de uma solicitação não altera as demais;
- **ACELERAR PEDIDO** abre um seletor para enviar os dados da solicitação ao desenvolvedor pelo WhatsApp `+55 81 99757-3129` ou Telegram `@Andredye`;
- a mensagem começa com `*WATS-CONNECT-API*` e inclui nome, plano, telefone, valor e data da solicitação;
- cancelar ou voltar no seletor de canal retorna à solicitação Premium pendente;
- os modais de progresso e resultado retornam ao menu ao serem fechados.

## Avisos e interface

- O banner de avisos só aparece quando há conteúdo;
- quebras de linha são preservadas;
- o banner compacto possui rolagem interna e abre o texto completo em modal;
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
- download automático de mídia;
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
WHATSAPP_MAX_IMAGE_SIZE=20000000
WHATSAPP_MAX_FILE_SIZE=50000000
WHATSAPP_MAX_VIDEO_SIZE=100000000
WHATSAPP_MAX_DOWNLOAD_SIZE=500000000
WHATSAPP_WEBHOOK=http://localhost:3129/Tasker
WHATSAPP_WEBHOOK_SECRET=tasker12345
WHATSAPP_WEBHOOK_INSECURE_SKIP_VERIFY=false
WHATSAPP_WEBHOOK_INCLUDE_OUTGOING=true
WHATSAPP_ACCOUNT_VALIDATION=true
WHATSAPP_CHAT_STORAGE=true
```

Por exemplo, digitar `100` no campo de vídeo grava `WHATSAPP_MAX_VIDEO_SIZE=100000000`. Ao abrir um `.env` existente, a Central converte os bytes de volta para MB na interface.

`AUTO_DELETE_MEDIA_DURATION` define apenas a duração de retenção. A chave não ativa a exclusão automática por si só; a API também precisa estar configurada para habilitar essa função.

`WHATSAPP_WEBHOOK_DELAY_MS` pode permanecer vazio. As mudanças entram em vigor após o reinício da API.

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
5. somente quando o contador chega a zero ou o usuário toca no botão, executa `006-CONFIGURAR_SISTEMA [wca]` com `par1 = force-reinstall-task` e a variável local `bottom = yes`;
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

Os comandos de sistema continuam sendo encaminhados à tarefa `006-CONFIGURAR_SISTEMA [wca]`, incluindo reinício, parada, atualização, reinstalação e instalação de projeto. O suporte agora abre diretamente o canal escolhido fora do WebView.

## Atualização no Tasker

1. Abra a cena `001-CENTRAL_WHATS_CONNECT_API` no editor;
2. substitua o conteúdo do WebView pelo HTML integral de `central-wca-1.73.html`;
3. confirme que **Allow Phone Access** continua habilitado;
4. valide abertura, modais, dispositivos, gravação do `.env` e reinício.

## Política de arquivos

O workspace mantém as versões históricas a partir da 1.29. O repositório publicado mantém somente o arquivo da versão atual. O histórico das alterações permanece no Git.
