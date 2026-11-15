# Guia / Help WCA — Cena V2

Projeto: `007-GUIA/HELP [wca]V2`
Versão atual: **1.14**
Arquivo: `guia-help-wca-1.14.html` — arquivo único, com HTML + CSS + JavaScript inline.

- **Transição dos modais**: entrada e saída combinam `scale(1.15)` com `blur(14px)` até o conteúdo chegar ao tamanho normal e ficar nítido;

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
- O player usa `object-fit: contain`, controles nativos, `playsinline`/`webkit-playsinline` e `preload="auto"`; inicia o buffer ao abrir, permanece pausado e mantém um poster preto embutido até a reprodução começar; não há botão FECHAR — o botão voltar do Android encerra o vídeo;
- O título completo fica oculto acima dos controles inferiores, em faixa preta com texto branco e fonte normal; aceita várias linhas, captura o toque antes dos controles nativos e some automaticamente após 3,2 segundos;
- Todos os cards usam diretamente a imagem local `icon-system.png`; nenhum elemento de vídeo é criado para gerar thumbnails, evitando requisições concorrentes e consumo desnecessário;
- A lista também exibe o tamanho de cada arquivo retornado pelo GitHub;
- Ao abrir, o vídeo permanece pausado até o usuário tocar em Play; voltar interrompe e elimina a mídia carregada;
- Durante a reprodução, URL, título, posição e intenção de reprodução são persistidos em `localStorage` e, automaticamente, na variável local interna `video_rotation_state`; esse fallback é necessário quando a rotação recria o WebView com outro armazenamento;
- Se a rotação recriar o WebView, o player volta ao mesmo vídeo e ponto aproximado, sempre pausado; não há tentativa de Play automático nem clique simulado, e o usuário inicia novamente pelo controle nativo; o estado expira após 10 minutos e é apagado ao voltar;
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
