# Gerar Tarefa de Envio WCA — Cena V2 (WebView)

Projeto: `005-GERAR_TAREFA_ENVIO [wca]`
Versão atual: **1.2**
Arquivo: `gerar-tarefa-de-envio-wca-1.2.html` — **único arquivo**, com HTML + CSS + JS inline.

Seleção e geração das tarefas de envio: lista os dispositivos logados da API local e, ao escolher um número, abre a **seleção das tarefas do projeto `id_number.prj.xml`** (baixado do repositório), gera um **XML personalizado com apenas as tarefas marcadas** e grava em `system/tasks-send`. **Sem Java Code** — o desenho do ícone que era feito em Java (Canvas) agora roda dentro da cena.

## Seleção em tela inteira (1.2)

- Ao abrir a seleção das tarefas, a cena **expande para a tela inteira** (classe `modo-selecao`): a lista rola ocupando **todo o espaço** (sem o teto de altura do diálogo), os botões **MARCAR TODAS / DESMARCAR TODAS / GERAR XML / VOLTAR** ficam **fixos embaixo** e o rodapé continua visível;
- VOLTAR volta ao modo compacto da lista de dispositivos; abrir a seleção de novo expande de novo.

## Funcionamento

- **Busca dos dispositivos** (igual à 1.0): `GET http://localhost:3000/devices` (fetch direto; se falhar, uma tentativa pelo bridge com prazo de 15s). Só entram os `state = logged_in`;
- **Toque no dispositivo** → tela de seleção: baixa o `id_number.prj.xml` (bridge `httpRequest` com `fileToSave` direto em `system/tasks-send/`; fallback `fetch` + `writeFile`) e lista **todas as tarefas** com checkbox (ordenadas por nome), com contagem "N de M", **MARCAR TODAS / DESMARCAR TODAS / GERAR XML / VOLTAR**;
- **GERAR XML** monta o arquivo final:
  1. **Filtro**: remove os blocos `<Task>` não selecionados **e** os ids deles do `<tids>` do projeto (o XML continua **válido** para importação);
  2. **Substituições**: `id-number` → `id:<4 últimos dígitos>` (ex.: `001-id:7777-wca-texto`) em nomes de tarefas e do projeto; todos os números `5581XXXXXXXX` (12 dígitos) → número completo do dispositivo;
  3. **Ícone em Canvas** (réplica do Java): base = **primeira imagem** (alfabética) de `icons/`, lida em base64 pelo shell e carregada via data URI; escala base 1000px com os mesmos cortes (−40/−50/−385/−20), fonte **bold 450**, relevo 3D de 15 camadas `#002233`, contorno `#004466` e degradê `#99FF00`→`#00CCFF` com os 4 dígitos; PNG gravado em `icons/icon-id-<4>.png` (via `.b64` + `base64 -d`) e o `>path-img<` do XML apontado para ele;
  4. **Validação**: DOMParser (quando disponível) rejeita XML malformado;
  5. **Gravação**: `system/tasks-send/id_number-<4>.prj.xml` (o original baixado permanece como cache);
- **Avisos no resultado**: dependências — tarefas chamadas via Perform Task (code 130) pelas selecionadas mas **não incluídas** (ex.: `003-PROCESSAR_REQUEST_MIDIA_WHATSAPP [wca]`); ícone não gerado (o XML mantém o placeholder);
- **FECHAR** encerra a cena pela **cascata completa** (dismissSceneV2 `1` → dismissLayout `1` → cancelScreen → tarefa `005-FECHAR_GERAR_TAREFA_ENVIO [wca]`).

## Variáveis do Tasker consumidas pela cena

**Nenhuma.**

## Saída da cena (AMBIENTE DO TASKER)

**Nenhuma variável é gravada** (o antigo `create_task`/`json` foi extinto na 1.1) — a entrega é o **arquivo XML gerado** no aparelho.

## Rollback

A versão anterior (`1.0`) permanece no workspace para força-voltar em caso de erro.
