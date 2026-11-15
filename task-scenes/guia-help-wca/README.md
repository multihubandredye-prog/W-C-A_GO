# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.20**
Arquivo: `guia-help-wca-1.20.html` — arquivo único, com HTML + CSS + JavaScript inline.

## Funcionamento atual

- Mantém as cinco abas do guia: **Introdução, Cadastro, Componentes, Envio e Receiver**;
- Possui pesquisa incremental global, com resultados de variáveis, tarefas e informações de todas as abas;
- Ao tocar em um resultado, abre a aba e os detalhes correspondentes, posiciona o conteúdo e destaca temporariamente o destino;
- A pesquisa ignora diferenças entre letras maiúsculas/minúsculas e entre caracteres com ou sem acento;
- Preserva a navegação rápida, gestos laterais, listas expansíveis e botão para voltar ao topo;
- Exibe as **61 tarefas de envio** com suas variáveis e exemplos;
- Exibe os **20 grupos de variáveis do Receiver**;
- Documenta as variáveis atuais do receptor, incluindo diagnóstico da requisição, contexto e estados detalhados de presença;
- Explica métricas do sistema, recuperação de mídia expirada, inclusão de mensagens enviadas no webhook e persistência de conversas;
- Carrega as últimas novidades a partir de `wca-resources/updates/changelog.txt`, usando a API do GitHub como fonte principal e `raw.githubusercontent.com` como fallback;
- Usa feedback visual `:active`, adequado ao uso no celular;
- O rodapé exibe somente o nome do desenvolvedor.

## Vídeos tutoriais

- **ABRIR PLAYLIST DE TUTORIAIS** abre dentro do Guia uma lista dos vídeos públicos da playlist oficial `PLEgPa9Eoqrdc`;
- A listagem consulta primeiro o feed oficial do YouTube sem chave de API e, enquanto o feed ainda estiver atualizando, usa a própria página pública da playlist como fallback;
- As consultas priorizam `Tasker.httpRequest`, evitando o bloqueio CORS do WebView, e usam `fetch` somente como fallback;
- Os vídeos aparecem em uma lista vertical compacta, mantendo a ordem da playlist; cada linha mostra numeração, miniatura pequena, título e informação de publicação;
- Tocar em uma linha abre aquele vídeo diretamente no aplicativo ou página externa do YouTube;
- O modal também possui o botão **ABRIR PLAYLIST COMPLETA NO YOUTUBE**;
- O player incorporado não voltou: a reprodução continua no próprio YouTube, evitando o erro 153;
- O botão voltar do Android fecha primeiro a lista e retorna ao Guia;
- Não há dependência da pasta `wca-resources/media/videos-tutoriais`, arquivos de vídeo locais, iframe ou chave da YouTube Data API.

## Telegram

- Os perfis em **Agradecimentos Especiais** abrem diretamente `https://t.me/usuario` com `window.open`;
- Não usa mais os atalhos antigos `wca2`, `tk.setLocal` ou `destroyScene`;
- Esta integração será validada no WebView do aparelho. Se o Android bloquear a abertura externa, o fluxo poderá ser substituído posteriormente por uma tarefa do Tasker.

## Botão voltar

A cena publica o modal atual em `cache_modal` e escuta `return_modal` com o listener `(nome, valor)`. O evento Back Pressed do Tasker deve cancelar dinamicamente quando `"%cache_modal" != ""` e definir `return_modal = %TIMES`.

## Ação Show Scene V2

- Cena: `007-GUIA/HELP [wca]V2`;
- Modo recomendado: **FullscreenWithResult**;
- ID: `1`;
- Dispensar automaticamente: vazio.
