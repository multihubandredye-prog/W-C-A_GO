# VIP Access Manager — Cena V2

Projeto: `VIP Access Manager`
Versão atual: **1.10**
Arquivo: `vip-access-manager-1.10.html` — HTML único com CSS e JavaScript inline.

## Finalidade

Painel administrativo para consultar clientes Premium e gerenciar individualmente solicitações de ativação e renovação gravadas no Firebase.

## Fila de solicitações

Cada envio possui `requestId` e é armazenado em `pendingRequests`. Solicitações do mesmo cliente não são fundidas nem descartadas.

- a assinatura Premium aparece como cadastro principal;
- cada solicitação pendente aparece em uma linha/cartão próprio;
- itens do mesmo cliente são identificados como `Solicitação N de total`;
- planos iguais ou diferentes continuam independentes;
- o formato legado `pendingRequest` continua sendo reconhecido.

## Ações por solicitação

- **Aprovar:** ativa somente o `requestId` escolhido e mantém os demais pendentes;
- **Rejeitar:** encerra somente o item escolhido e registra a decisão;
- **Duplicada:** encerra somente o item escolhido como duplicado e registra a decisão.

Quando o cliente ainda possui Premium ativo, a renovação começa no vencimento atual. Uma aprovação posterior começa no vencimento já atualizado.

## Cena V2

- sem dependências externas de CSS, JavaScript, fontes ou ícones;
- feedback móvel com `:active`;
- bridge do Tasker para `cache_modal` e `return_modal`;
- listener `(nome, valor)` e sondagem de segurança para o botão voltar;
- sem `%` e sem `console` no JavaScript.

## Firebase e API

A cena lê a coleção `users` e usa os endpoints administrativos Premium. A API deve oferecer a fila `pendingRequests` e aceitar `requestId` na ativação/cancelamento. O documento legado é migrado na próxima gravação feita pela API.
