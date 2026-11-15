# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.6**
Arquivo: `guia-help-wca-1.6.html` — arquivo único, com HTML + CSS + JavaScript inline.

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
- Tocar no vídeo abre imediatamente um overlay de reprodução que ocupa `100vw × 100dvh` de toda a área da cena, sem depender de `requestFullscreen()` do WebView;
- O player usa `object-fit: contain`, controles nativos e `playsinline`/`webkit-playsinline`; não há botão FECHAR — o botão voltar do Android encerra o vídeo;
- A lista também exibe o tamanho de cada arquivo retornado pelo GitHub;
- Ao abrir, o vídeo permanece pausado até o usuário tocar em Play; voltar interrompe e elimina a mídia carregada;
- Durante a reprodução, URL, título, posição e intenção de reprodução são persistidos somente em `localStorage`; nenhuma variável Tasker é usada para o vídeo;
- Se a rotação recriar o WebView, o player volta ao mesmo ponto: continua reproduzindo se estava em reprodução e permanece pausado se o usuário havia pausado; o estado expira após 10 minutos e é apagado ao voltar;
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
