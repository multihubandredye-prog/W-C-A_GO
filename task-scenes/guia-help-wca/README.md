# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.18**
Arquivo: `guia-help-wca-1.18.html` — arquivo único, com HTML + CSS + JavaScript inline.

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

- **ABRIR PLAYLIST DE TUTORIAIS** abre diretamente a playlist oficial `PLEgPa9Eoqrdc` no YouTube;
- A playlist é aberta fora do player incorporado do WebView, evitando o erro 153 causado pela ausência de referência HTTP em páginas locais;
- O sistema tenta abrir o aplicativo ou a página externa do YouTube e usa navegação direta como fallback caso a nova janela seja bloqueada;
- O usuário escolhe e reproduz os vídeos no próprio YouTube;
- Não há dependência da pasta `wca-resources/media/videos-tutoriais`, miniaturas locais, arquivos de vídeo, iframe ou estado de reprodução no Tasker.

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
