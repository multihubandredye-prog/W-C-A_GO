# Instalação WCA — Cena V2 (WebView)

Projeto: `002-INSTALACAO_PROJETO [wca]`

Versão atual: **1.8**

Arquivo atual: `instalacao-wca-1.8.html`

A cena é um único arquivo HTML com CSS e JavaScript inline. Ela consulta as releases, baixa, acompanha, valida e substitui os arquivos diretamente pelo bridge do Tasker.

## Requisitos

- Tasker 6.7 ou superior, com Scenes V2;
- JavaScript Bridge habilitado no componente WebView;
- **Allow Phone Access** habilitado;
- ID da Tela V2 igual a `1`, ou ajuste correspondente em `SCREEN_ID`;
- acesso à internet e permissão de gravação nos destinos do WCA.

## Seleção do pacote

A variável local `type_project` define o pacote. A cena faz uma leitura inicial com prazo de quatro segundos e também registra um listener `(nome, valor)` antes do início do download.

| Valor | Arquivos |
|---|---|
| `project-api` | `whatsapp-wca-aarch64.bin` e `ffmpeg-wca-aarch64.bin` |
| `project-task` | `W_C_A_API.prj.xml` |
| `project-full` | os dois binários e depois o XML |
| `full` | alias de `project-full` |

Quando o valor está ausente ou é inválido, o padrão é `project-api`.

A variável `progress_install` não é usada. O progresso é calculado pela cena a partir do tamanho físico dos arquivos temporários.

## Localização da release

A cena consulta as 30 releases mais recentes do repositório e aceita publicações cujo título contenha `[UI]`, `[CORE]` ou `[FULL]`. O título exato `Teste` também é aceito.

Ela percorre as releases da mais recente para a mais antiga e escolhe a primeira que contenha todos os assets exigidos pelo modo atual.

## Download e progresso

Cada arquivo é baixado com `Tasker.httpRequest` para o destino final acrescido de `.update`. Durante o download, a cena consulta o tamanho físico aproximadamente a cada 450 milissegundos e atualiza:

- anel e percentual geral;
- quantidade baixada e tamanho esperado;
- cartão, ícone, cor e barra do arquivo atual;
- cronômetro da execução.

O download usa redirecionamento automático e prazo de até 600 segundos por arquivo.

## Destinos

| Arquivo | Destino |
|---|---|
| `whatsapp-wca-aarch64.bin` | `/data/data/net.dinglisch.android.taskerm/files/WCA/whatsapp-wca-aarch64.bin` |
| `ffmpeg-wca-aarch64.bin` | `/data/data/net.dinglisch.android.taskerm/files/WCA/ffmpeg-wca-aarch64.bin` |
| `W_C_A_API.prj.xml` | `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/system/tasks-send/W_C_A_API.prj.xml` |

## Validação e substituição

Antes de substituir qualquer arquivo, a cena exige que o temporário:

1. exista e não esteja vazio;
2. tenha o tamanho informado pelo asset, quando esse tamanho estiver disponível;
3. contenha `<TaskerData` quando for o projeto XML.

Somente depois da validação o `.update` substitui o destino final. Os binários recebem permissão `700`. Em caso de falha, o temporário é removido, o destino anterior é preservado e aparecem **TENTAR NOVAMENTE** e **FECHAR**.

## Continuação no Tasker

Todos os modos terminam executando:

```text
Tarefa: 006-CONFIGURAR_SISTEMA [wca]
par1: force-reinstall-task
bottom: yes
```

### Modos com XML

Em `project-task`, `project-full` e `full`, depois de validar o XML a cena:

1. inicia uma contagem de 10 até 0;
2. exibe **INSTALAR AGORA**;
3. executa a tarefa somente quando a contagem chega a zero ou o botão é tocado;
4. usa uma trava contra execução duplicada.

### Modo somente API

Em `project-api`, o botão **INSTALAR AGORA** não é exibido. Depois de validar os dois binários, a tarefa é executada automaticamente com os mesmos valores.

## Conclusão e fechamento

Após a chamada bem-sucedida da tarefa, a cena mostra `Instalação iniciada no Tasker.`, libera **FECHAR** e tenta encerrar a Tela V2 após 600 milissegundos.

A ordem de fechamento é:

1. `Tasker.dismissSceneV2("1")`;
2. `Tasker.dismissLayout("1")`;
3. `Tasker.cancelScreen("")`;
4. tarefa `003-FECHAR_INSTALACAO [wca]` com `par1=fechar-cena`.

O estado pode ser consultado por:

- `window.WCA_INSTALL_SCENE.isCompleted()`;
- `window.WCA_INSTALL_SCENE.percent()`;
- `window.WCA_INSTALL_SCENE.mode()`.

## Recursos visuais

O cabeçalho usa o vídeo local:

`file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4`

A interface mantém anel colorido, cartões por arquivo, barra individual, cronômetro e feedback móvel com `:active`.
