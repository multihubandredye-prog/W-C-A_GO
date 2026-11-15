# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.17**
Arquivo: `guia-help-wca-1.17.html` — arquivo único, com HTML + CSS + JavaScript inline.

- **Transição dos modais**: entrada e saída combinam `scale(1.15)` com `blur(14px)` até o conteúdo chegar ao tamanho normal e ficar nítido;

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

- **ABRIR PLAYLIST DE TUTORIAIS** abre diretamente a playlist oficial `PLEgPa9Eoqrdc` do YouTube;
- O player do YouTube é incorporado ao WebView em um overlay que ocupa `100vw × 100dvh`, sem baixar ou listar vídeos do repositório;
- A playlist usa `youtube-nocookie.com`, identificação de origem e política de referência compatível com o player incorporado;
- A reprodução não começa automaticamente: o usuário escolhe o vídeo e toca em Play no próprio player do YouTube;
- O botão voltar do Android encerra o player, descarrega o conteúdo incorporado e retorna ao Guia;
- Não há dependência da pasta `wca-resources/media/videos-tutoriais`, miniaturas locais, arquivos de vídeo ou estado de reprodução no Tasker.

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
