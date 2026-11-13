# Contatos Pro WCA — Cena V2 (WebView)

Projeto: `Contatos Pro [wca]` (use o nome/ID da cena como está no seu Tasker)
Versão atual: **1.13**
Arquivo: `Contatos Pro WCA 1.13.html` — **único arquivo**, com HTML + CSS + JS inline.

Central de contatos e grupos: escolhe um dispositivo logado e lista **Contatos**, **Grupos** e **Minha Agenda** em abas, com busca, seleção múltipla (toque longo), modais de detalhes com cópia rápida e **LID do contato em segundo plano**. As cópias vão para a **área de transferência** do dispositivo (nomes, números/JIDs, ou tudo, um por linha).

## Funcionamento

- **Tela de dispositivos**: busca `GET /devices` na API local e lista os com `state = logged_in` (nome, número e selo "Online"); botão **Atualizar Dispositivos** refaz a busca;
- **Tela de contatos/grupos**: busca `GET /user/my/contacts?filter=chatted` (agenda sincronizada no WhatsApp) e `GET /user/my/groups` (todos os grupos que você participa) com `X-Device-Id` do dispositivo escolhido. Contatos válidos descartam JIDs inválidos (`0@…`, `status@…`); grupos ficam só os `@g.us`;
- **Ordenação (com nome primeiro)**: quem tem nome salvo fica no início em ordem alfabética; quem só tem número (sem nome na agenda) vai para o fim, também em ordem — vale para contatos e grupos;
- **Abas**: Contatos/Grupos/Minha Agenda com contadores e indicador deslizante; **arrastar para o lado** na lista alterna entre as três abas (esquerda = próxima, direita = anterior) — rolagem vertical e modo de seleção não trocam;
- **Minha Agenda (aba própria)**: os contatos da agenda do aparelho chegam pela variável **`contacts_address`** do Tasker (JSON em array com `number` e `display_name`, como o do projeto). A cena lê no padrão das outras cenas: listener + sondagem de 16 leituras a cada 500 ms (prazo de 1,5 s por leitura); se o valor chegar depois, o listener cobre. **Variável inexistente ou vazia: a aba fica vazia ("Nenhum contato na agenda") e as outras abas seguem normalmente.** **Números normalizados para 12 dígitos** (`55` + DDD + 8, sem `+` e sem o nono dígito): a correção tira o `+`, zeros e prefixo de operadora (`015…`, `041…`), o `55` quando presente e o 9º dígito; contato sem DDD completo (só 9 dígitos) não tem como formar 12 e fica de fora. Ordenação com nome primeiro (A-Z, nomes estilizados normalizados) e quem só tem número no fim; busca por nome ou número; tocar num contato abre o mesmo modal (copiar nome/número, LID e foto — com o número já corrigido);
- **Busca**: filtra por nome (sem acentos/pontuação) ou número/JID em tempo real, atualizando os contadores;
- **Seleção múltipla**: toque longo (0,5s, com vibração) entra no modo de seleção; toques alternam itens; **Todos** seleciona/desfaz a lista atual; **Limpar** zera; barra inferior com o total e cópias em lote (**Nome** / **Número ou JID** / **Tudo**) e pré-visualização em modal (primeiros 60 itens);
- **Modais**: contato (copiar nome, número, nome + número e **LID** — copia só o número do LID, sem o sufixo) e grupo (copiar nome, JID ou ambos); fecham pelo X, tocando fora ou tecla Esc; a linha do LID busca `GET /user/info?phone=<número>` (campo `resolved_lid`) em segundo plano, com cache por JID e indicação "Buscando LID… / LID não sincronizado / LID indisponível";
- **Acesso à API**: mesmo caminho do projeto original (fetch direto em `http://127.0.0.1:3000` com `Authorization: Basic` + `X-Device-Id`); se o acesso direto falhar, cada busca tenta **uma vez pelo bridge** (`Tasker.httpRequest`, prazo de 15s) antes de indicar falha de conexão;
- **Avatares (fotos de perfil)**: a cena busca a foto real de cada **dispositivo conectado** na tela inicial e, ao abrir o modal, do **contato** ou **grupo**; a consulta é `GET /user/avatar?phone=<jid>&is_preview=false&is_community=false` (a mesma usada pela Central), com `X-Device-Id` do próprio dispositivo; a URL devolvida aponta para os servidores do WhatsApp e é aplicada direto na imagem, que substitui o ícone do WhatsApp quando carrega; cache em memória por JID evita repetir a consulta (o botão **Atualizar Dispositivos** limpa o cache e refaz as fotos); se a busca falhar, demorar ou o item não tiver foto, o ícone do WhatsApp permanece;
- **Layout centralizado**: o container (card + rodapé do desenvolvedor) fica centralizado verticalmente na tela quando há espaço sobrando (como na tela inicial de dispositivos); com conteúdo mais alto que a tela (lista de contatos/grupos grande), volta ao alinhamento do topo com rolagem normal, sem cortar nada;
- **Listas longas**: a animação de entrada tem teto nos primeiros itens (evita itens que "sumiam" ao rolar, quando o atraso somado passava de segundos) e a busca só re-renderiza após uma pequena pausa (150 ms), com montagem da lista em fragmento — um único lote no DOM;
- **Visual da Central (borda e avatar animados)**: o card principal tem a **borda giratória** da Central (anel cônico azul/verde em rotação contínua) e o avatar dos modais de contato e grupo tem o **anel giratório** ao redor da foto (mesmas cores e animação do projeto central-wca);
- **Foto ampliada**: tocar no avatar do modal de contato ou grupo abre a foto **grande e arredondada** (com o mesmo anel giratório), com fundo escurecido **por cima de tudo** (a sobreposição fica fora do card, no fim da página, acima dos modais e dos toasts); fecha com um toque em qualquer lugar, tecla Esc ou o botão voltar (o voltar fecha a foto ampliada primeiro e o modal permanece); **os avatares dos dispositivos conectados não têm esse recurso**;
- **Números com fonte normal**: os números/JIDs exibidos nas linhas (contatos, grupos e agenda) e nos modais usam a fonte padrão do app (DM Sans), sem a fonte de títulos;
- **Container na largura da tela**: o card ocupa a largura disponível da tela (com a margem de respiro do app);
- **Toasts** de confirmação para cópias e ações de seleção.

## Variáveis do Tasker consumidas pela cena

- **`contacts_address`** — JSON em array com os contatos da agenda do aparelho (`number`, `display_name`, `account_type`, `_id`). Se for **local** (nome em minúscula), defina-a na tarefa que abre a cena **DEPOIS** da ação Show Scene v2 (mesmo esquema do Painel de Atividade); se for global, basta existir quando a cena abrir. Vazia ou inexistente: a aba **Minha Agenda** fica vazia e nada mais é afetado. Os valores são lidos por listener + sondagem (a cena cobre valor definido antes ou logo depois do Show).

Toda a informação das abas Contatos e Grupos vem da API local; a autenticação e o host ficam no próprio código.

## Botão voltar do Android (Back Pressed)

A cena publica em `cache_modal` qual "tela" está no topo e escuta `return_modal` (mesmo padrão das outras cenas do projeto; nomes nas constantes `WCA_MODAL_STATE_VAR` e `WCA_RETURN_MODAL_VAR`, bloco "Botao voltar do Android" do HTML):

| Estado publicado em `cache_modal` | Quando |
|---|---|
| `bulk-modal` | Modal de cópia em lote aberto (topo máximo) |
| `contact-modal` | Modal de detalhes do contato aberto |
| `group-modal` | Modal de detalhes do grupo aberto |
| `contacts-panel` | Painel de contatos/grupos visível |
| *(vazio)* | Painel de dispositivos (raiz) |

Ao **fechar** um modal (voltar, X ou tocando fora) o estado volta ao nível atual — o painel de contatos, quando é o caso — e NÃO ao vazio: o vazio existe apenas na raiz (tela inicial). É por isso que a cena só "solta" o botão voltar (deixando o Tasker fechar a cena) quando você está na tela inicial; nos demais níveis cada voltar é consumido pela cena (fecha modal, cancela seleção ou volta ao painel de dispositivos).

Ao receber `return_modal`, a cena fecha o topo: modais fecham; no painel de contatos, cancela o modo seleção (com toast) e, num novo voltar, retorna ao painel de dispositivos; na raiz ela não interfere (o voltar segue o padrão e a cena é fechada pelo lado Tasker). **Ordem completa do voltar com o painel de contatos aberto:** 1º fecha o modal aberto (se houver) · 2º cancela a seleção (se ativa) · 3º volta ao painel de dispositivos · só então o voltar seguinte fecha a cena.

**Montagem validada no evento Back Pressed da tela (igual às outras cenas):**
1. **Cancelar Evento: Dinâmico** com a expressão **`"%cache_modal" != ""`** (COM as aspas);
2. **Definir Variável**: `return_modal` = `%TIMES`.

## Saída da cena

**Nenhuma variável de tela.** As cópias vão para a área de transferência do dispositivo. A cena **não se fecha sozinha** — o fechamento é feito pelo lado Tasker (fluxo/back do seu projeto), como nos outros projetos.

## Arquivos que a cena lê

**Nenhum.** As consultas são requisições à API local (`127.0.0.1:3000`).

## Recursos externos

- Google Fonts (Syne + DM Sans) e Font Awesome CDN no `<head>`;
- Vídeos do cabeçalho: `file:///storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons/video-system.mp4` (exige **Allow Phone Access** no elemento Web).

## Requisitos da cena V2

- Tasker 6.7 ou superior (Scenes V2 / "Modern (V2)")
- Componente **WebView** com o **JavaScript bridge** habilitado
- **Allow Phone Access** habilitado no elemento Web (vídeos via `file://` e fallback de busca pelo bridge)

## Como atualizar no Tasker

1. Abra a cena do Contatos Pro (V2) no editor
2. No componente **WebView**, substitua o HTML pelo conteúdo integral do arquivo da versão atual
3. Confirme que **Allow Phone Access** continua habilitado
4. A cena é consulta/cópia pura: nada precisa ser lido depois dela no fluxo

## Política de versões

Todas as versões são **mantidas nesta pasta** desde a 1.0 — a cada alteração um novo arquivo é adicionado e nenhum anterior é removido, permitindo correções ou retorno sem perda.
