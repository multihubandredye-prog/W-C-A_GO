# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.1**
Arquivo: `guia-help-wca-1.1.html` — arquivo único, com HTML + CSS + JavaScript inline.

## Funcionamento atual

- Mantém as cinco abas do guia: **Introdução, Cadastro, Componentes, Envio e Receiver**;
- Preserva a navegação rápida, gestos laterais, listas expansíveis e botão para voltar ao topo;
- Exibe as **61 tarefas de envio** com suas variáveis e exemplos;
- Exibe os **20 grupos de variáveis do Receiver**;
- Carrega as últimas novidades a partir de `wca-resources/updates/changelog.txt` no repositório;
- Usa feedback visual `:active`, adequado ao uso no celular;
- O rodapé exibe somente o nome do desenvolvedor.

## Vídeos tutoriais

- **VER VÍDEOS TUTORIAIS** abre a lista dentro da própria cena;
- Tocar em um vídeo abre o reprodutor dentro de um segundo modal, sem tarefa auxiliar e sem fechar a cena;
- Links do YouTube usam `youtube-nocookie.com`, política `strict-origin-when-cross-origin`, identificação de origem e `widget_referrer` no `iframe`;
- O player oferece **ABRIR NO YOUTUBE** como alternativa caso o WebView local continue bloqueado pelo erro 153;
- Links diretos de mídia usam o elemento `<video>` nativo;
- O reprodutor oferece controles e o botão **TELA CHEIA**;
- Fechar o player interrompe e elimina a mídia carregada;
- Voltar no player retorna à lista de vídeos; voltar na lista fecha o modal.

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
