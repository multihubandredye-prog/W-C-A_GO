# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.3**
Arquivo: `guia-help-wca-1.3.html` — arquivo único, com HTML + CSS + JavaScript inline.

## Funcionamento atual

- Mantém as cinco abas do guia: **Introdução, Cadastro, Componentes, Envio e Receiver**;
- Preserva a navegação rápida, gestos laterais, listas expansíveis e botão para voltar ao topo;
- Exibe as **61 tarefas de envio** com suas variáveis e exemplos;
- Exibe os **20 grupos de variáveis do Receiver**;
- Carrega as últimas novidades a partir de `wca-resources/updates/changelog.txt` no repositório;
- Usa feedback visual `:active`, adequado ao uso no celular;
- O rodapé exibe somente o nome do desenvolvedor.

## Vídeos tutoriais

- **VER VÍDEOS TUTORIAIS** consulta dinamicamente a pasta `wca-resources/media/videos-tutoriais` pela API `contents` do GitHub;
- Todos os arquivos `.mp4`, `.webm`, `.ogg`, `.ogv` e `.m4v` presentes na pasta são listados automaticamente, em ordem natural pelo nome;
- O título exibido é gerado a partir do nome do arquivo, sem extensão e com hífens/sublinhados convertidos em espaços;
- Tocar em um vídeo usa o `download_url` retornado pelo GitHub e abre o elemento `<video>` nativo dentro de um segundo modal, sem tarefa auxiliar e sem fechar a cena;
- A lista é consultada novamente sempre que o modal é aberto, permitindo adicionar ou remover vídeos sem atualizar o HTML;
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
