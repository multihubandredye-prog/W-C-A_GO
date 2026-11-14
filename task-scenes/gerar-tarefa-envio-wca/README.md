# Gerar Tarefa de Envio WCA — Cena V2 (WebView)

Projeto: `005-GERAR_TAREFA_ENVIO [wca]`
Versão atual: **1.13**
Arquivo: `gerar-tarefa-de-envio-wca-1.13.html` — **único arquivo**, com HTML + CSS + JS inline.

Seleção e geração das tarefas de envio: lista os dispositivos logados da API local e, ao escolher um número, abre a **seleção das tarefas do projeto `id_number.prj.xml`** (baixado do repositório), gera um **XML personalizado com apenas as tarefas marcadas** e grava em `system/tasks-send`. **Sem Java Code** — o desenho do ícone que era feito em Java (Canvas) agora roda dentro da cena.

## Seleção

- A lista exibe o **nome curto** das tarefas — só o que vem depois de `id-number-` (ex.: `001-id-number-wca-texto` → `wca-texto`); o nome completo fica guardado internamente (os avisos de dependência usam ele);
- **Todas as tarefas já vêm marcadas** por padrão;
- Um **único ícone** (Font Awesome, dupla checagem `fa-check-double` — sempre o mesmo) no **canto superior direito**, acima da lista, **alterna marcar/desmarcar todas**;
- O botão **GERAR PROJETO** abre uma **confirmação** (modal com o resumo: nº de tarefas + número/id) antes de gerar — CONFIRMAR gera, CANCELAR (ou voltar/clicar fora) fecha o modal e volta à seleção.

## Resultado limpo + contagem de instalação

- Ao gerar, a cena mostra apenas **"Gerado com sucesso"** e a **imagem do ícone gerado** (`icons/icon-id-<4>.png`) — **sem caminhos de arquivo** na tela;
- O título da seleção é **"Marque as tarefas que devem ser criadas."**;
- Começa uma **contagem regressiva de 10s** com o número numa **bola verde** (badge arredondado, número em negrito) à **direita** de "Iniciando o processo de instalação" (`10` → `1`): **sem clique**, ao zerar publica `project_created = true` e **encerra sozinha**;
- O botão **INSTALAR AGORA** fica **fora do cartão**, **logo abaixo dele** (o footer do desenvolvedor fica no rodapé da tela) — ao clicar, **cancela a contagem**, publica `project_created` e encerra na hora; a tarefa que mostra a cena instala o projeto;
- Voltar (ou sair da tela de resultado) **cancela a contagem** — sem auto-encerramento fora do resultado;
- Avisos de dependência (tarefa chamada via Perform Task mas não incluída) continuam aparecendo abaixo da imagem, quando houver.

## Botão voltar do Android

- Mesma lógica da **Central WCA**: a cena publica em **`cache_modal`** a tela do topo (`selecao` ou `resultado`; vazio = lista de dispositivos) e escuta **`return_modal`** para voltar à lista de dispositivos;
- Lado Tasker (já configurado): handler de **Back Pressed** da tela com **Cancelar Evento** dinâmico `"%cache_modal" != ""` (com aspas) e **Definir Variável** `return_modal = %TIMES`;
- Sem tela aberta (lista de dispositivos), o voltar fecha a cena pelo fluxo normal do Tasker.

## Bordas animadas estilo Central

- O cartão principal e a **lista de tarefas** usam o mesmo **gradiente cônico giratório** (`spinBorder`) da Central WCA;
- O fundo da lista é **opaco** — o anel aparece **só na borda** (sem brilho no centro);
- Na seleção em **tela inteira**, o anel do cartão é desenhado **para dentro** (`inset: 0`/`2px`) — visível nas **laterais** sem sair da tela;
- Entrada com `cardSlide`, brilho no topo do cabeçalho (`topShine`) e a lista de tarefas com borda animada própria (`spinBorder` 6s).

## Tarefas detectadas automaticamente

- A lista é montada a partir do **`id_number.prj.xml` baixado do repositório** (`main/task-project/`) a cada abertura da cena — **não é uma lista fixa**: qualquer tarefa nova adicionada ao XML aparece automaticamente na seleção (basta ter `<id>` e `<nme>`);
- Tarefas de infraestrutura (como a `003-PROCESSAR_REQUEST_MIDIA_WHATSAPP [wca]`) ficam fora da lista via `TAREFAS_OCULTAS` no código.

## Ícone sem duplicação

- A base do Canvas é sempre a primeira imagem **original** de `icons/` — arquivos `icon-id-*` (ícones já gerados) **nunca** são usados como base, então os dígitos nunca são desenhados duas vezes.

## Funcionamento

- **Busca dos dispositivos**: `GET http://localhost:3000/devices` (fetch direto; se falhar, uma tentativa pelo bridge com prazo de 15s). Só entram os `state = logged_in`;
- **Toque no dispositivo** → tela de seleção: baixa o `id_number.prj.xml` (bridge `httpRequest` com `fileToSave` direto em `system/tasks-send/`; fallback `fetch` + `writeFile`) e lista **todas as tarefas** com checkbox **já marcadas** (ordenadas por nome), com contagem "N de M", ícone de alternar todas no canto superior direito e botões **GERAR PROJETO / VOLTAR**;
- **GERAR PROJETO** abre a **confirmação** (modal) e, ao CONFIRMAR, monta o arquivo final:
  1. **Filtro**: remove os blocos `<Task>` não selecionados **e** os ids deles do `<tids>` do projeto (o XML continua **válido** para importação);
  2. **Substituições**: `id-number` → `id:<4 últimos dígitos>` (ex.: `001-id:7777-wca-texto`) em nomes de tarefas e do projeto; só o placeholder literal `5581XXXXXXXX` (com X) → número completo do dispositivo (**números reais ficam intactos**);
  3. **Ícone em Canvas** (réplica do Java): base = primeira imagem **original** de `icons/`, lida em base64 pelo shell e carregada via data URI; escala base 1000px com os mesmos cortes (−40/−50/−385/−20), fonte **bold 450**, relevo 3D de 15 camadas `#002233`, contorno `#004466` e degradê `#99FF00`→`#00CCFF` com os 4 dígitos; PNG gravado em `icons/icon-id-<4>.png` (via `.b64` + `base64 -d`) e o `>path-img<` do XML apontado para ele;
  4. **Validação**: DOMParser (quando disponível) rejeita XML malformado;
  5. **Gravação**: `system/tasks-send/id_number-<4>.prj.xml` (o original baixado permanece como cache);
- **Avisos no resultado**: dependências — tarefas chamadas via Perform Task (code 130) pelas selecionadas mas **não incluídas** (ex.: `003-PROCESSAR_REQUEST_MIDIA_WHATSAPP [wca]`); ícone não gerado (o XML mantém o placeholder);
- **INSTALAR AGORA** (logo abaixo do cartão, fora dele) cancela a contagem, publica `project_created` e encerra a cena pela **cascata completa** (dismissSceneV2 `1` → dismissLayout `1` → cancelScreen → tarefa `005-FECHAR_GERAR_TAREFA_ENVIO [wca]`); sem clique, a contagem de 10s encerra sozinha do mesmo jeito.

## Tela inteira (configuração do lado do Tasker)

Para a seleção ocupar a tela toda, a ação **Cena v2 Mostrar** desta cena deve usar **FullscreenWithResult** (Modo de Exibição), igual à cena 010 — em `DialogWithResult` a janela do diálogo é fixa e o CSS não a expande; manter **ID da Tela = 1** e Dispensar Automático vazio.

## Variáveis do Tasker consumidas pela cena

- **`cache_modal`** (escrita pela cena): tela do topo (`selecao`/`resultado`/vazio) — usada pelo handler de Back Pressed;
- **`return_modal`** (escrita pelo Tasker): comando de voltar — a cena escuta e retorna à lista de dispositivos;
- **`project_created`** (escrita pela cena): **`false`** ao abrir, **`true`** quando o XML é gerado com sucesso (reafirmado antes de encerrar pelo botão FECHAR) — a tarefa que mostra a cena lê `%project_created` ao dispensar.

**Nenhuma variável de tarefa é gravada** — a entrega é o **arquivo XML gerado** no aparelho.

## Rollback

A versão anterior (`1.12`) permanece no workspace para força-voltar em caso de erro.
