# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.32**
Arquivo: `guia-help-wca-1.32.html` — arquivo único, com HTML + CSS + JavaScript inline.


O cabeçalho não utiliza vídeo decorativo.

## Funcionamento atual

- Mantém as cinco abas do guia: **Introdução, Cadastro, Componentes, Envio e Receiver**;
- Possui pesquisa incremental global, com resultados de variáveis, tarefas e informações de todas as abas;
- Ao tocar em um resultado, abre a aba e os detalhes correspondentes, posiciona o conteúdo e destaca temporariamente o destino;
- A pesquisa ignora diferenças entre letras maiúsculas/minúsculas e entre caracteres com ou sem acento;
- Preserva a navegação rápida, gestos laterais, listas expansíveis e botão para voltar ao topo;
- Exibe as **62 tarefas de envio** com suas variáveis, incluindo a tarefa 062 de cartão Pix;
- Exibe os **21 grupos de variáveis do Receiver**;
- Documenta as variáveis atuais do receptor, incluindo diagnóstico da requisição, contexto e estados detalhados de presença;
- Explica métricas do sistema, recuperação de mídia expirada, inclusão de mensagens enviadas no webhook e persistência de conversas;
- Carrega as últimas novidades a partir de `wca-resources/updates/changelog.txt`, usando a API do GitHub como fonte principal e `raw.githubusercontent.com` como fallback;
- Usa feedback visual `:active`, adequado ao uso no celular;
- O rodapé exibe somente o nome do desenvolvedor.

## Vídeos tutoriais

- **ABRIR PLAYLIST DE TUTORIAIS** abre dentro do Guia uma lista dos vídeos públicos da playlist oficial `PLEgPa9Eoqrdc`;
- A listagem consulta tanto o feed oficial do YouTube quanto a página pública da playlist e combina os resultados sem repetir IDs;
- Essa combinação preserva a ordem da página da playlist e inclui vídeos comuns, Shorts e itens não listados que estejam visíveis na playlist pública; o feed complementa data, título e identificação dos Shorts públicos;
- As consultas priorizam `Tasker.httpRequest`, evitando o bloqueio CORS do WebView, e usam `fetch` somente como fallback;
- Os vídeos aparecem em uma lista vertical compacta; cada linha mostra numeração ou indicação **SHORT**, miniatura pequena, título real e informação de publicação;
- Tocar em uma linha abre o vídeo em tela cheia dentro do próprio WebView do Guia;
- Para evitar o erro 153 das páginas locais, o Guia carrega um intermediário HTTPS mínimo a partir de uma URL imutável do RawCDN/GitHack, vinculada ao commit que contém `wca-resources/youtube-player.html`; esse intermediário fornece a referência HTTP exigida pelo player incorporado do YouTube sem depender do cache do branch principal;
- O player usa `youtube-nocookie.com`, mantém controles, tela cheia e suporte à playlist, sem sair para a página ou aplicativo do YouTube;
- Não existe botão X sobre o player: o fechamento é feito pelo botão voltar do Android;
- Durante a reprodução, o intermediário envia periodicamente a posição atual ao Guia. ID, título, tipo, posição aproximada e estado ficam preservados em `localStorage` e na variável local `video_rotation_state`;
- Quando a rotação recria o WebView, o Guia reabre a lista e o mesmo vídeo em tela cheia, retomando do ponto aproximado salvo;
- O modal também possui o botão **ABRIR PLAYLIST COMPLETA NO YOUTUBE** como alternativa externa;
- O botão voltar do Android fecha primeiro o player e retorna à lista; um segundo voltar fecha a lista e retorna ao Guia;
- O estado expira após 10 minutos e é eliminado ao fechar o player normalmente;
- Não há dependência da pasta `wca-resources/media/videos-tutoriais`, arquivos de vídeo locais ou chave da YouTube Data API.

## Telegram

- Cada pessoa em **Agradecimentos Especiais** mantém um link HTTPS completo para `https://t.me/usuario`;
- Nome, função e identificador fazem parte da mesma área clicável, não somente o texto iniciado por `@`;
- No Tasker, o toque usa `Tasker.browseUrl` para abrir o endereço fora do WebView; se essa ação não estiver disponível, usa `Tasker.sendIntent` com `android.intent.action.VIEW` e destino `activity`;
- A navegação interna é cancelada quando a ponte externa está disponível, evitando que o WebView tente carregar o esquema `tg:resolve` e apresente `ERR_UNKNOWN_URL_SCHEME`;
- Fora do Tasker, o link HTTPS nativo permanece como fallback;
- Não usa os atalhos antigos `wca2`, `tk.setLocal` ou `destroyScene`.

## Botão voltar

A cena publica o modal atual em `cache_modal` e escuta `return_modal` com o listener `(nome, valor)`. O evento Back Pressed do Tasker deve cancelar dinamicamente quando `"%cache_modal" != ""` e definir `return_modal = %TIMES`.

## Ação Show Scene V2

- Cena: `007-GUIA/HELP [wca]V2`;
- Modo recomendado: **FullscreenWithResult**;
- ID: `1`;
- Dispensar automaticamente: vazio.

## Resposta de carrossel no Receiver

A aba Receiver inclui as variáveis `carousel_reply`, `carousel_selected_id`, `carousel_selected_index`, `carousel_selected_text`, `carousel_replied_to_id` e `carousel_reply_json`, com exemplo de condição, convenção dos IDs e limites de identificação. Requer atualização da API e instalação separada do Java Receiver; o clique real deve ser validado no aparelho.
