# Central de Disparos WCA — Cena V2

Arquivo de trabalho: `central-de-disparos-wca-1.24.html`, com HTML, CSS e JavaScript no mesmo arquivo. Esta versão publicada ainda precisa de validação no Tasker/Android antes de substituir a cena anterior.

A seleção de imagens, vídeos e arquivos usa o gerenciador interno: grade ou lista, tamanhos, navegação de pastas, busca por tipo e toque prolongado para selecionar arquivos compatíveis em subpastas. A filtragem testa a existência de arquivos por extensão e mantém pastas inacessíveis visíveis, com falha no terminal. Pastas muito grandes podem ultrapassar o limite de seleção recursiva.

A aba Vídeos grava os vídeos para destinatários em `json-videos.json`, separados de `json-images.json`. Status continuam em `json-status.json`. Não é realizada migração automática; vídeos antigos em `json-images.json` não são exibidos na nova aba e geram aviso no terminal.

Envios de arquivo, imagem e vídeo usam `Tasker.httpRequest` com `fileToSend` e os endpoints multipart originais. É necessário validar no aparelho o formato da requisição, os arquivos e as respostas para cada tipo; um envio com resposta incerta não é repetido automaticamente.

A Cena V2 precisa da ponte JS e do acesso ao telefone. Os JSONs são lidos via bridge; arquivos ilegíveis ficam protegidos contra gravação e erros são mostrados no terminal. Faça uma cópia dos JSONs de `system/central-disparos-wca` antes do teste. O botão voltar requer o handler Tasker com `cache_modal` e `return_modal`. Arquivos com nomes contendo quebras de linha não são suportados pela listagem.

As prévias de vídeo preservam a proporção da imagem, sem exibir a reserva ao mesmo tempo; campos flexíveis usam limites de largura para evitar transbordamento.

As prévias dos editores reservam espaço no fluxo do modal para não sobrepor os campos seguintes.

## Áudios e seleção

A aba Áudios mantém itens em `json-audios.json`, separados dos arquivos anteriores. O envio imediato usa `/send/audio` via ponte HTTP nativa, com opção de mensagem de voz e encaminhamento. O suporte real de cada formato de áudio depende da API e do dispositivo.

O editor permite configurar dias e horários. O perfil adaptado é fornecido separadamente, fora do repositório; precisa ser instalado manualmente no lugar do código antigo e testado no Tasker para executar os agendamentos de áudio e vídeo.

Ao abrir imagens, vídeos ou áudios, o filtro correspondente fica ativo. A pasta Android permanece visível e não é verificada automaticamente pelo filtro. A seleção individual atualiza somente o cartão tocado, preservando as miniaturas já carregadas.

## Navegador em tela cheia

O navegador de arquivos ocupa a área disponível da cena, com cartões ajustáveis, miniatura maior e espaço separado para nome e tamanho. O retorno à pasta anterior é feito pelo botão com seta.

## Navegação e grade 1.11

A listagem e o resultado da filtragem de pastas são reutilizados durante a sessão do navegador, inclusive ao voltar. Para refletir arquivos alterados fora da cena, use o botão Atualizar pasta; o cache também é limpo ao reabrir o navegador. A grade exibe miniaturas quadradas, nomes abaixo e abreviação central para nomes longos.

## Seleção e duplicatas 1.12

A filtragem automática não varre a pasta Android nem suas subpastas; elas continuam navegáveis, sujeitas às permissões do aparelho. Ao adicionar, caminhos já existentes na mesma seção são ignorados e informados no terminal. Itens já duplicados anteriormente não são removidos. O mesmo caminho pode existir em seções diferentes. A seleção para ações em lote tem contorno e marcação mais visíveis.

## Destinatários em lote 1.13

Nas abas Imagens, Vídeos, Áudios, Mensagens, Enquetes e Arquivos, selecione os itens e toque em Destinatários. A opção padrão substitui mediante confirmação; Acrescentar preserva os destinatários existentes e evita duplicatas. Utilize Marcar todos para aplicar à aba inteira. A ação grava os destinatários e não envia mensagens; Status não participa.

## Seleção e progresso 1.14

A barra usa um único botão para marcar ou desmarcar todos. Arquivos e áudios exibem o nome do arquivo. O progresso mostra cada item, confirmações, falhas e ignorados; confirmação da API não significa entrega ao destinatário. O envio não é repetido automaticamente em caso de resultado incerto.

## Contatos antecipados e cabeçalhos 1.15

Contatos e grupos do primeiro dispositivo online selecionado são carregados em segundo plano ao iniciar, sem bloquear a interface. As listas ficam em memória durante a sessão, inclusive quando vazias; pesquisas reutilizam a requisição em andamento e os dados carregados. Para buscar alterações, use Atualizar contatos e grupos na busca; falhas são registradas no terminal. Quando o dispositivo utilizado muda, os dados de outra conta não são misturados. Os vídeos decorativos foram retirados da cena; o arquivo do aparelho não é apagado.

## Busca e edição em tela cheia 1.16

O editor de imagem/vídeo e a busca rápida ocupam a área disponível, com rolagem independente e controles fixos. O botão de atualizar contatos é representado por ícone; exibe carregamento, resultado ou falha, com detalhes no terminal.

## Integração do perfil 1.17

Para executar os agendamentos de áudio e vídeo, use a versão atualizada do código do perfil do Tasker em conjunto com a cena; o código do perfil não está incluído no repositório. O Java lê `json-videos.json` e `json-audios.json`, enviando multipart aos endpoints `/send/video` e `/send/audio` com os campos correspondentes e PTT; o arquivo de status permanece o mesmo. Foi compilado e exercitado com servidor HTTP simulado, mas não foi executado no Tasker nem testado com a API real. Não habilite o perfil antigo e o novo simultaneamente: risco de envios repetidos. Antes de ativar, faça backup dos dados e teste com um destinatário controlado. A lógica herdada de cache por item pode reenviar a destinatários já atendidos após falha parcial ou resposta incerta; em caso de falha, confira o histórico antes da próxima execução.

## Nome e compatibilidade

O nome visível é Central de Disparos WCA. Os dados ficam em `system/central-disparos-wca/`. A cena não move automaticamente os dados da pasta anterior; confirme que o diretório, os arquivos e o cache foram transferidos antes de atualizar o perfil e a cena. A alteração do HTML não atualiza automaticamente o perfil do Tasker.

## Diretório de dados 1.19

Os arquivos JSON e o histórico são lidos em `system/central-disparos-wca/`. O código do perfil fornecido separadamente usa o mesmo diretório para os JSONs e `send-cache`. Não existe migração automática; uma pasta ou arquivo ausente pode ser criado vazio pela cena. Preserve o cache original para evitar repetição de agendamentos já processados.

## Destinatários em lote 1.20

Após a confirmação da gravação, os itens que receberam a alteração são desmarcados. Em caso de cancelamento ou falha, a seleção permanece. A operação não envia mensagens.

## Editores em tela cheia 1.21

Todos os editores de status, imagens, vídeos, mensagens, enquetes, arquivos e áudios ocupam a área disponível. Os formulários de inclusão de status, mensagens e enquetes usam o mesmo padrão, com corpo rolável e ações fixas.

## Agendamento em lote 1.23

Selecione itens da mesma aba e toque em Agendamento para aplicar dias e horários iguais. Se algum item já tiver agendamento, a substituição é confirmada; cancelamento ou falha mantém a seleção. Após gravação confirmada, os itens são desmarcados. A operação apenas configura os itens e não envia mensagens; a execução automática depende do perfil instalado.

## Leitura de arquivos 1.24

A cena tenta novamente leituras malsucedidas e usa alternativas de leitura antes de bloquear o arquivo. Uma verificação temporariamente indisponível não impede a tentativa de leitura. Se todas falharem, a gravação permanece bloqueada para proteger os dados e o terminal indica que os itens não foram carregados; reabra a cena para tentar novamente.
