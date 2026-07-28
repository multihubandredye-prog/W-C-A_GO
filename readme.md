# W-C-A GO — WhatsApp Cloud API (Go)

API REST em Go para WhatsApp multi-dispositivo, construída sobre [whatsmeow](https://github.com/tulir/whatsmeow).
Envia e recebe mensagens, mídia, enquetes, **botões interativos** e **listas**, gerencia grupos, contatos e múltiplos dispositivos.

> **Repositório de teste.** Este é o `Wca-Test`, usado para desenvolvimento e experimentação.

---

## Índice

- [Começando](#começando)
- [Autenticação](#autenticação)
- [Múltiplos dispositivos](#múltiplos-dispositivos)
- [Formato das respostas](#formato-das-respostas)
- **Endpoints**
  - [App / Sessão](#1-app--sessão)
  - [Dispositivos](#2-dispositivos)
  - [Envio de mensagens](#3-envio-de-mensagens)
  - [🔘 Botões interativos](#4--botões-interativos) ← *novo*
  - [📋 Listas interativas](#5--listas-interativas) ← *novo*
  - [Gerenciar mensagens](#6-gerenciar-mensagens)
  - [Conversas](#7-conversas)
  - [Usuário](#8-usuário)
  - [Grupos](#9-grupos)
  - [Newsletter](#10-newsletter)
- [Recebendo respostas de botões e listas](#recebendo-respostas-de-botões-e-listas)
- [Códigos de erro](#códigos-de-erro)

---

## Começando

### Docker (recomendado)

```bash
docker compose up -d
```

### Compilando do código-fonte

```bash
cd src
go mod download
go run . rest
```

A API sobe em `http://localhost:3000` por padrão.

### Variáveis de ambiente principais

| Variável | Padrão | Descrição |
|---|---|---|
| `APP_PORT` | `3000` | Porta HTTP |
| `APP_BASIC_AUTH` | — | Credenciais `user:senha` (separe múltiplas por vírgula) |
| `APP_BASE_PATH` | — | Prefixo de path, ex: `/api` |
| `APP_DEBUG` | `false` | Logs detalhados |
| `WHATSAPP_WEBHOOK` | — | URL(s) que recebem os eventos |
| `WHATSAPP_WEBHOOK_SECRET` | — | Segredo para assinar o webhook |
| `DB_URI` | `file:storages/whatsapp.db` | Banco (SQLite ou Postgres) |

---

## Autenticação

Se `APP_BASIC_AUTH` estiver configurado, envie o header HTTP Basic:

```bash
curl -u usuario:senha http://localhost:3000/app/status
```

---

## Múltiplos dispositivos

A API suporta várias sessões de WhatsApp simultâneas. Informe qual usar de duas formas:

```bash
# Via header (recomendado)
-H "X-Device-Id: SEU_DEVICE_ID"

# Ou via query string
?device_id=SEU_DEVICE_ID
```

Se houver apenas um dispositivo, ele é usado automaticamente.

---

## Formato das respostas

Todas as respostas seguem o mesmo formato:

```json
{
  "code": "SUCCESS",
  "message": "Send buttons success 5588999999999 (server timestamp: ...)",
  "results": {
    "message_id": "3EB0C767D26B8CA1B7F2",
    "status": "Send buttons success ..."
  }
}
```

Em caso de erro:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "buttons: maximum 3 buttons allowed, got 4.",
  "results": null
}
```

---

## 1. App / Sessão

### `GET /app/login`
Gera o QR Code para parear o WhatsApp.

```bash
curl -X GET http://localhost:3000/app/login
```

Retorna `qr_link` (imagem do QR) e `qr_duration` (segundos até expirar).

### `GET /app/login-with-code`
Login por código de 8 dígitos, sem QR.

```bash
curl -X GET "http://localhost:3000/app/login-with-code?phone=5588999999999"
```

### `GET /app/logout`
Encerra a sessão e apaga as credenciais.

```bash
curl -X GET http://localhost:3000/app/logout
```

### `GET /app/reconnect`
Reconecta usando a sessão já salva.

```bash
curl -X GET http://localhost:3000/app/reconnect
```

### `GET /app/devices`
Lista os dispositivos vinculados à conta do WhatsApp.

```bash
curl -X GET http://localhost:3000/app/devices
```

### `GET /app/status`
Status da conexão atual.

```bash
curl -X GET http://localhost:3000/app/status
```

### `GET /health`
Health check público, sem autenticação. Retorna `OK` ou 503.

```bash
curl -X GET http://localhost:3000/health
```

---

## 2. Dispositivos

Gerenciamento de múltiplas sessões.

### `GET /devices`
Lista todos os dispositivos cadastrados.

```bash
curl -X GET http://localhost:3000/devices
```

### `POST /devices`
Cria um novo dispositivo.

```bash
curl -X POST http://localhost:3000/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "Atendimento 01"}'
```

### `GET /devices/:device_id`
Detalhes de um dispositivo.

```bash
curl -X GET http://localhost:3000/devices/abc123
```

### `DELETE /devices/:device_id`
Remove o dispositivo e sua sessão.

```bash
curl -X DELETE http://localhost:3000/devices/abc123
```

### `GET /devices/:device_id/login`
QR Code daquele dispositivo específico.

```bash
curl -X GET http://localhost:3000/devices/abc123/login
```

### `POST /devices/:device_id/login/code`
Login por código para o dispositivo.

```bash
curl -X POST http://localhost:3000/devices/abc123/login/code \
  -H "Content-Type: application/json" \
  -d '{"phone": "5588999999999"}'
```

### `POST /devices/:device_id/logout`
Desconecta o dispositivo.

```bash
curl -X POST http://localhost:3000/devices/abc123/logout
```

### `POST /devices/:device_id/reconnect`
Reconecta o dispositivo.

```bash
curl -X POST http://localhost:3000/devices/abc123/reconnect
```

### `GET /devices/:device_id/status`
Status daquele dispositivo.

```bash
curl -X GET http://localhost:3000/devices/abc123/status
```

---

## 3. Envio de mensagens

> **Formato do telefone:** use apenas dígitos (`5588999999999`) para conversa individual
> ou o JID completo. Para **grupos**, use `120363XXXXXXXXXX@g.us`.

Campos comuns a todos os envios:

| Campo | Tipo | Descrição |
|---|---|---|
| `phone` | string | **Obrigatório.** Destinatário |
| `duration` | int | Mensagem temporária, em segundos (`86400`, `604800`, `7776000`) |
| `is_forwarded` | bool | Marca como encaminhada |

### `POST /send/message`
Envia texto.

```bash
curl -X POST http://localhost:3000/send/message \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "message": "Olá! Tudo bem?"
  }'
```

Para **responder** a uma mensagem, inclua `reply_message_id`:

```bash
curl -X POST http://localhost:3000/send/message \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "message": "Claro, pode sim!",
    "reply_message_id": "3EB0C767D26B8CA1B7F2"
  }'
```

### `POST /send/image`
Envia imagem (upload de arquivo).

```bash
curl -X POST http://localhost:3000/send/image \
  -F "phone=5588999999999" \
  -F "caption=Confira nossa promoção" \
  -F "image=@/caminho/foto.jpg" \
  -F "view_once=false" \
  -F "compress=true"
```

### `POST /send/json/image`
Mesma coisa, mas por URL ou base64 — sem upload.

```bash
curl -X POST http://localhost:3000/send/json/image \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "caption": "Confira nossa promoção",
    "image_url": "https://exemplo.com/foto.jpg"
  }'
```

### `POST /send/file`
Envia documento.

```bash
curl -X POST http://localhost:3000/send/file \
  -F "phone=5588999999999" \
  -F "caption=Segue o contrato" \
  -F "file=@/caminho/contrato.pdf"
```

### `POST /send/json/file`

```bash
curl -X POST http://localhost:3000/send/json/file \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "file_url": "https://exemplo.com/contrato.pdf",
    "caption": "Segue o contrato"
  }'
```

### `POST /send/video`

```bash
curl -X POST http://localhost:3000/send/video \
  -F "phone=5588999999999" \
  -F "caption=Veja o vídeo" \
  -F "video=@/caminho/video.mp4" \
  -F "compress=true"
```

### `POST /send/json/video`

```bash
curl -X POST http://localhost:3000/send/json/video \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "video_url": "https://exemplo.com/video.mp4",
    "caption": "Veja o vídeo"
  }'
```

### `POST /send/audio`
Envia áudio / mensagem de voz.

```bash
curl -X POST http://localhost:3000/send/audio \
  -F "phone=5588999999999" \
  -F "audio=@/caminho/audio.ogg"
```

### `POST /send/json/audio`

```bash
curl -X POST http://localhost:3000/send/json/audio \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "audio_url": "https://exemplo.com/audio.ogg"
  }'
```

### `POST /send/sticker`

```bash
curl -X POST http://localhost:3000/send/sticker \
  -F "phone=5588999999999" \
  -F "sticker=@/caminho/figurinha.webp"
```

### `POST /send/contact`
Compartilha um contato.

```bash
curl -X POST http://localhost:3000/send/contact \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "contact_name": "Suporte Técnico",
    "contact_phone": "5588988888888"
  }'
```

### `POST /send/link`
Envia link com pré-visualização.

```bash
curl -X POST http://localhost:3000/send/link \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "link": "https://exemplo.com",
    "caption": "Dá uma olhada nisso"
  }'
```

### `POST /send/json/link`
Versão com controle total da prévia.

```bash
curl -X POST http://localhost:3000/send/json/link \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "link": "https://exemplo.com",
    "caption": "Dá uma olhada",
    "title": "Título personalizado",
    "description": "Descrição personalizada"
  }'
```

### `POST /send/location`

```bash
curl -X POST http://localhost:3000/send/location \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "latitude": "-7.2306",
    "longitude": "-39.3153"
  }'
```

### `POST /send/poll`
Cria uma enquete.

```bash
curl -X POST http://localhost:3000/send/poll \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "question": "Qual o melhor horário para a reunião?",
    "options": ["09h", "14h", "16h"],
    "max_answer": 1
  }'
```

### `POST /send/presence`
Define seu status global (`available` / `unavailable`).

```bash
curl -X POST http://localhost:3000/send/presence \
  -H "Content-Type: application/json" \
  -d '{"type": "available"}'
```

### `POST /send/chat-presence`
Mostra "digitando..." ou "gravando áudio..." na conversa.

```bash
curl -X POST http://localhost:3000/send/chat-presence \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "action": "start"
  }'
```

`action`: `start` (digitando) ou `stop`.

---

## 4. 🔘 Botões interativos

Envia uma mensagem com **até 3 botões** clicáveis. Ideal para decisões rápidas: confirmar, escolher entre poucas opções, abrir um link ou ligar.

> **Precisa de mais de 3 opções?** Use [listas](#5--listas-interativas).

### `POST /send/buttons`

#### Campos

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `phone` | string | ✅ | Destinatário |
| `body` | string | ✅ | Texto principal da mensagem |
| `buttons` | array | ✅ | 1 a 3 botões |
| `title` | string | — | Cabeçalho acima do texto |
| `footer` | string | — | Rodapé em letra menor |
| `image_url` | string | — | Imagem no cabeçalho (URL http/https ou `data:image/...;base64,...`) |
| `duration` | int | — | Mensagem temporária, em segundos |
| `is_forwarded` | bool | — | Marca como encaminhada |

#### Tipos de botão

| `type` | O que faz | Campos exigidos |
|---|---|---|
| `reply` *(padrão)* | Resposta rápida — devolve o `id` no webhook | `title`, `id` |
| `cta_url` | Abre um link | `title`, `url` |
| `cta_call` | Inicia uma ligação | `title`, `phone_number` |
| `copy` | Copia um código | `title`, `copy_code` |

**Regras aplicadas automaticamente:**
- Máximo de **3 botões** — mais que isso retorna erro de validação
- `title` é truncado em **20 caracteres** (contando acentos e emoji corretamente)
- `id` vazio assume o valor do `title`
- IDs de botões `reply` devem ser únicos

#### Exemplo 1 — Botões de resposta rápida

```bash
curl -X POST http://localhost:3000/send/buttons \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "body": "Podemos confirmar seu agendamento para amanhã às 14h?",
    "footer": "Clínica Bem Estar",
    "buttons": [
      { "type": "reply", "title": "Confirmar", "id": "confirma_sim" },
      { "type": "reply", "title": "Remarcar",  "id": "remarcar" },
      { "type": "reply", "title": "Cancelar",  "id": "cancelar" }
    ]
  }'
```

#### Exemplo 2 — Botões mistos (link + ligação)

```bash
curl -X POST http://localhost:3000/send/buttons \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "title": "Seu pedido foi enviado!",
    "body": "Pedido #4821 saiu para entrega e chega hoje até as 18h.",
    "footer": "Loja XYZ",
    "buttons": [
      { "type": "cta_url",  "title": "Rastrear",     "url": "https://loja.com/rastreio/4821" },
      { "type": "cta_call", "title": "Falar conosco","phone_number": "5588988888888" },
      { "type": "reply",    "title": "Já recebi",    "id": "pedido_recebido" }
    ]
  }'
```

#### Exemplo 3 — Com imagem no cabeçalho e botão de copiar

```bash
curl -X POST http://localhost:3000/send/buttons \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "body": "Aproveite 20% de desconto na primeira compra!",
    "footer": "Válido até domingo",
    "image_url": "https://exemplo.com/banner-promo.jpg",
    "buttons": [
      { "type": "copy",    "title": "Copiar cupom", "copy_code": "BEMVINDO20" },
      { "type": "cta_url", "title": "Ir à loja",    "url": "https://loja.com" }
    ]
  }'
```

#### Resposta

```json
{
  "code": "SUCCESS",
  "message": "Send buttons success 5588999999999 (server timestamp: 2026-07-28 15:04:05 +0000 UTC)",
  "results": {
    "message_id": "3EB0C767D26B8CA1B7F2",
    "status": "Send buttons success ..."
  }
}
```

---

## 5. 📋 Listas interativas

Envia um **menu suspenso** com muito mais opções que os botões. As opções ficam organizadas em **seções**, cada uma com título e itens que podem ter descrição.

Perfeito para cardápios, catálogos, agendamentos e qualquer escolha com muitas alternativas.

### `POST /send/list`

#### Campos

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `phone` | string | ✅ | Destinatário |
| `description` | string | ✅ | Texto principal da mensagem |
| `sections` | array | ✅ | Grupos de opções |
| `button_text` | string | — | Rótulo do botão que abre a lista (padrão: `Select`) |
| `title` | string | — | Cabeçalho acima do texto |
| `footer` | string | — | Rodapé |
| `duration` | int | — | Mensagem temporária |
| `is_forwarded` | bool | — | Marca como encaminhada |

**Estrutura de `sections[]`:**

| Campo | Tipo | Descrição |
|---|---|---|
| `title` | string | Título da seção |
| `rows` | array | Itens selecionáveis |

**Estrutura de `rows[]`:**

| Campo | Tipo | Descrição |
|---|---|---|
| `title` | string | **Obrigatório.** Nome da opção (até 24 caracteres) |
| `row_id` | string | Identificador retornado no webhook. Padrão: o próprio `title` |
| `description` | string | Texto secundário (até 72 caracteres) |

**Limites aplicados automaticamente:**
- Até **10 linhas por seção**
- Até **30 linhas no total**
- `row_id` deve ser único entre **todas** as seções
- Títulos e descrições são truncados no limite

#### Exemplo 1 — Cardápio com várias seções

```bash
curl -X POST http://localhost:3000/send/list \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "title": "Pizzaria do Zé",
    "description": "Escolha o que deseja pedir hoje:",
    "footer": "Entrega em até 40 minutos",
    "button_text": "Ver cardápio",
    "sections": [
      {
        "title": "🍕 Pizzas Salgadas",
        "rows": [
          { "row_id": "pizza_marg", "title": "Margherita",  "description": "Molho, mussarela e manjericão — R$ 45" },
          { "row_id": "pizza_cala", "title": "Calabresa",   "description": "Calabresa e cebola — R$ 42" },
          { "row_id": "pizza_port", "title": "Portuguesa",  "description": "Presunto, ovo e ervilha — R$ 48" },
          { "row_id": "pizza_frang","title": "Frango c/ Catupiry", "description": "Frango desfiado — R$ 50" }
        ]
      },
      {
        "title": "🍰 Pizzas Doces",
        "rows": [
          { "row_id": "pizza_choco", "title": "Chocolate",   "description": "Chocolate ao leite — R$ 38" },
          { "row_id": "pizza_banana","title": "Banana Nevada","description": "Banana, canela e leite cond. — R$ 40" }
        ]
      },
      {
        "title": "🥤 Bebidas",
        "rows": [
          { "row_id": "coca_2l",  "title": "Coca-Cola 2L",  "description": "R$ 12" },
          { "row_id": "guarana",  "title": "Guaraná 2L",    "description": "R$ 10" },
          { "row_id": "suco_lar", "title": "Suco natural",  "description": "Laranja 500ml — R$ 8" }
        ]
      }
    ]
  }'
```

#### Exemplo 2 — Menu de atendimento simples

```bash
curl -X POST http://localhost:3000/send/list \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "description": "Olá! Como podemos ajudar você hoje?",
    "button_text": "Ver opções",
    "sections": [
      {
        "title": "Atendimento",
        "rows": [
          { "row_id": "menu_2via",     "title": "2ª via de boleto" },
          { "row_id": "menu_suporte",  "title": "Suporte técnico" },
          { "row_id": "menu_financeiro","title": "Financeiro" },
          { "row_id": "menu_cancelar", "title": "Cancelamento" },
          { "row_id": "menu_humano",   "title": "Falar com atendente" }
        ]
      }
    ]
  }'
```

#### Resposta

```json
{
  "code": "SUCCESS",
  "message": "Send list success 5588999999999 (server timestamp: ...)",
  "results": {
    "message_id": "3EB0C767D26B8CA1B7F2",
    "status": "Send list success ..."
  }
}
```

---

### Botões ou listas? Qual usar

| | Botões | Listas |
|---|---|---|
| Quantidade de opções | Até **3** | Até **30** (10 por seção) |
| Visual | Botões fixos abaixo da mensagem | Botão único que abre um menu |
| Agrupamento por categoria | ❌ | ✅ Seções com título |
| Descrição em cada opção | ❌ | ✅ |
| Abrir link / ligar / copiar | ✅ | ❌ (apenas seleção) |
| Imagem no cabeçalho | ✅ | ❌ |
| Melhor para | Confirmações, sim/não, ações diretas | Cardápios, catálogos, menus de atendimento |

---

## 6. Gerenciar mensagens

### `POST /message/:message_id/reaction`
Reage com emoji. Use `""` para remover a reação.

```bash
curl -X POST http://localhost:3000/message/3EB0C767D26B8CA1B7F2/reaction \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "emoji": "👍"
  }'
```

### `POST /message/:message_id/revoke`
Apaga para todos.

```bash
curl -X POST http://localhost:3000/message/3EB0C767D26B8CA1B7F2/revoke \
  -H "Content-Type: application/json" \
  -d '{"phone": "5588999999999"}'
```

### `POST /message/:message_id/delete`
Apaga apenas para você.

```bash
curl -X POST http://localhost:3000/message/3EB0C767D26B8CA1B7F2/delete \
  -H "Content-Type: application/json" \
  -d '{"phone": "5588999999999"}'
```

### `POST /message/:message_id/update`
Edita o texto de uma mensagem já enviada.

```bash
curl -X POST http://localhost:3000/message/3EB0C767D26B8CA1B7F2/update \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "message": "Texto corrigido"
  }'
```

### `POST /message/:message_id/read`
Marca como lida (dois tiques azuis).

```bash
curl -X POST http://localhost:3000/message/3EB0C767D26B8CA1B7F2/read \
  -H "Content-Type: application/json" \
  -d '{"phone": "5588999999999"}'
```

### `POST /message/:message_id/star` · `POST /message/:message_id/unstar`
Favorita / desfavorita.

```bash
curl -X POST http://localhost:3000/message/3EB0C767D26B8CA1B7F2/star \
  -H "Content-Type: application/json" \
  -d '{"phone": "5588999999999"}'
```

### `GET /message/:message_id/download`
Baixa a mídia de uma mensagem.

```bash
curl -X GET "http://localhost:3000/message/3EB0C767D26B8CA1B7F2/download?phone=5588999999999" \
  --output arquivo.jpg
```

### `POST /message/revoke_status_full`
Apaga todos os seus status publicados.

```bash
curl -X POST http://localhost:3000/message/revoke_status_full
```

---

## 7. Conversas

### `GET /chats`
Lista as conversas.

```bash
curl -X GET "http://localhost:3000/chats?limit=25&offset=0&search=maria"
```

| Parâmetro | Descrição |
|---|---|
| `limit` | Quantidade por página (padrão 25) |
| `offset` | Deslocamento |
| `search` | Busca por nome ou número |
| `has_media` | `true` para só conversas com mídia |

### `GET /chat/:chat_jid/messages`
Histórico de uma conversa.

```bash
curl -X GET "http://localhost:3000/chat/5588999999999@s.whatsapp.net/messages?limit=50"
```

| Parâmetro | Descrição |
|---|---|
| `limit` / `offset` | Paginação |
| `start_time` / `end_time` | Intervalo em ISO 8601 |
| `media_only` | Apenas mensagens com mídia |
| `is_from_me` | Filtra por remetente |
| `search` | Busca no conteúdo |

### `POST /chat/:chat_jid/pin`
Fixa ou desafixa a conversa.

```bash
curl -X POST http://localhost:3000/chat/5588999999999@s.whatsapp.net/pin \
  -H "Content-Type: application/json" \
  -d '{"pinned": true}'
```

### `POST /chat/:chat_jid/archive`
Arquiva ou desarquiva.

```bash
curl -X POST http://localhost:3000/chat/5588999999999@s.whatsapp.net/archive \
  -H "Content-Type: application/json" \
  -d '{"archived": true}'
```

### `POST /chat/:chat_jid/disappearing`
Define mensagens temporárias na conversa.

```bash
curl -X POST http://localhost:3000/chat/5588999999999@s.whatsapp.net/disappearing \
  -H "Content-Type: application/json" \
  -d '{"duration": 604800}'
```

Valores: `0` (desliga), `86400` (24h), `604800` (7 dias), `7776000` (90 dias).

---

## 8. Usuário

### `GET /user/info`
Informações de um número.

```bash
curl -X GET "http://localhost:3000/user/info?phone=5588999999999"
```

### `GET /user/avatar`
Foto de perfil.

```bash
curl -X GET "http://localhost:3000/user/avatar?phone=5588999999999&is_preview=false"
```

### `POST /user/avatar`
Troca sua foto de perfil.

```bash
curl -X POST http://localhost:3000/user/avatar \
  -F "avatar=@/caminho/foto.jpg"
```

### `POST /user/pushname`
Altera seu nome de exibição.

```bash
curl -X POST http://localhost:3000/user/pushname \
  -H "Content-Type: application/json" \
  -d '{"push_name": "Loja XYZ"}'
```

### `GET /user/check`
Verifica se um número tem WhatsApp.

```bash
curl -X GET "http://localhost:3000/user/check?phone=5588999999999"
```

### `GET /user/business-profile`
Dados do perfil comercial.

```bash
curl -X GET "http://localhost:3000/user/business-profile?phone=5588999999999"
```

### `GET /user/my/privacy`
Suas configurações de privacidade.

```bash
curl -X GET http://localhost:3000/user/my/privacy
```

### `GET /user/my/groups`
Grupos dos quais você participa.

```bash
curl -X GET http://localhost:3000/user/my/groups
```

### `GET /user/my/newsletters`
Canais que você segue.

```bash
curl -X GET http://localhost:3000/user/my/newsletters
```

### `GET /user/my/contacts`
Sua lista de contatos.

```bash
curl -X GET http://localhost:3000/user/my/contacts
```

---

## 9. Grupos

### `POST /group`
Cria um grupo.

```bash
curl -X POST http://localhost:3000/group \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Equipe de Vendas",
    "participants": ["5588999999999", "5588988888888"]
  }'
```

### `GET /group/info`
Informações do grupo.

```bash
curl -X GET "http://localhost:3000/group/info?group_id=120363XXXXXXXXXX@g.us"
```

### `GET /group/info-from-link`
Informações a partir de um convite.

```bash
curl -X GET "http://localhost:3000/group/info-from-link?link=https://chat.whatsapp.com/XXXX"
```

### `POST /group/join-with-link`
Entra em um grupo pelo link.

```bash
curl -X POST http://localhost:3000/group/join-with-link \
  -H "Content-Type: application/json" \
  -d '{"link": "https://chat.whatsapp.com/XXXX"}'
```

### `POST /group/leave`
Sai do grupo.

```bash
curl -X POST http://localhost:3000/group/leave \
  -H "Content-Type: application/json" \
  -d '{"group_id": "120363XXXXXXXXXX@g.us"}'
```

### `GET /group/participants`
Lista os membros.

```bash
curl -X GET "http://localhost:3000/group/participants?group_id=120363XXXXXXXXXX@g.us"
```

### `GET /group/participants/export`
Exporta os membros em CSV.

```bash
curl -X GET "http://localhost:3000/group/participants/export?group_id=120363XXXXXXXXXX@g.us" \
  --output membros.csv
```

### `POST /group/participants`
Adiciona membros.

```bash
curl -X POST http://localhost:3000/group/participants \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "participants": ["5588999999999"]
  }'
```

### `POST /group/participants/remove`
Remove membros.

```bash
curl -X POST http://localhost:3000/group/participants/remove \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "participants": ["5588999999999"]
  }'
```

### `POST /group/participants/promote` · `POST /group/participants/demote`
Promove a admin / rebaixa.

```bash
curl -X POST http://localhost:3000/group/participants/promote \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "participants": ["5588999999999"]
  }'
```

### `GET /group/participant-requests`
Lista pedidos pendentes de entrada.

```bash
curl -X GET "http://localhost:3000/group/participant-requests?group_id=120363XXXXXXXXXX@g.us"
```

### `POST /group/participant-requests/approve` · `.../reject`
Aprova ou rejeita pedidos.

```bash
curl -X POST http://localhost:3000/group/participant-requests/approve \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "participants": ["5588999999999"]
  }'
```

### `POST /group/name`
Renomeia o grupo.

```bash
curl -X POST http://localhost:3000/group/name \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "name": "Equipe de Vendas 2026"
  }'
```

### `POST /group/topic`
Altera a descrição.

```bash
curl -X POST http://localhost:3000/group/topic \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "topic": "Grupo oficial da equipe comercial"
  }'
```

### `POST /group/photo`
Troca a foto do grupo.

```bash
curl -X POST http://localhost:3000/group/photo \
  -F "group_id=120363XXXXXXXXXX@g.us" \
  -F "photo=@/caminho/foto.jpg"
```

### `POST /group/locked`
Só admins podem editar as informações.

```bash
curl -X POST http://localhost:3000/group/locked \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "locked": true
  }'
```

### `POST /group/announce`
Só admins podem enviar mensagens.

```bash
curl -X POST http://localhost:3000/group/announce \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "120363XXXXXXXXXX@g.us",
    "announce": true
  }'
```

### `GET /group/invite-link`
Obtém (ou renova) o link de convite.

```bash
curl -X GET "http://localhost:3000/group/invite-link?group_id=120363XXXXXXXXXX@g.us&reset=false"
```

---

## 10. Newsletter

### `POST /newsletter/unfollow`
Deixa de seguir um canal.

```bash
curl -X POST http://localhost:3000/newsletter/unfollow \
  -H "Content-Type: application/json" \
  -d '{"newsletter_id": "120363XXXXXXXXXX@newsletter"}'
```

---

## Recebendo respostas de botões e listas

Enviar o botão é só metade do trabalho — você precisa saber **no que o cliente clicou**. Quando ele responde, a API dispara um webhook para a URL configurada em `WHATSAPP_WEBHOOK`.

Para facilitar, o payload traz um campo unificado **`InteractiveReply`**, que funciona igual para botões e listas.

### Clique em botão

```json
{
  "Event": "message",
  "Payload": {
    "from": "5588999999999@s.whatsapp.net",
    "message_id": "3EB0...",
    "InteractiveReply": {
      "Type": "buttons",
      "SelectedID": "confirma_sim",
      "SelectedText": "Confirmar",
      "Name": "quick_reply",
      "ParamsJSON": "{\"display_text\":\"Confirmar\",\"id\":\"confirma_sim\"}"
    }
  }
}
```

### Seleção em lista

```json
{
  "Event": "message",
  "Payload": {
    "from": "5588999999999@s.whatsapp.net",
    "message_id": "3EB0...",
    "InteractiveReply": {
      "Type": "list",
      "SelectedID": "pizza_marg",
      "Title": "Margherita",
      "Description": "Molho, mussarela e manjericão — R$ 45"
    },
    "ListReply": { "...": "mesmo conteúdo" }
  }
}
```

### O campo que importa

**`SelectedID`** é a chave de tudo. Ele devolve exatamente o `id` (botão) ou `row_id` (lista) que você definiu ao enviar a mensagem.

Por isso vale usar identificadores descritivos:

```json
{ "row_id": "pizza_marg" }   ✅ fácil de tratar no código
{ "row_id": "1" }            ❌ vira um enigma depois
```

### Exemplo de tratamento

```javascript
app.post('/webhook', (req, res) => {
  const reply = req.body.Payload?.InteractiveReply;

  if (reply) {
    switch (reply.SelectedID) {
      case 'confirma_sim':
        // confirmar agendamento
        break;
      case 'pizza_marg':
        // adicionar Margherita ao pedido
        break;
      case 'menu_humano':
        // transferir para atendente
        break;
    }
  }

  res.sendStatus(200);
});
```

Campos adicionais disponíveis:

| Campo | Quando aparece |
|---|---|
| `InteractiveReply` | Sempre que há uma resposta interativa (unificado) |
| `ListReply` | Apenas em seleção de lista |
| `ButtonsReply` | Apenas em botões de formato legado |

---

## Códigos de erro

| Código | HTTP | Significado |
|---|---|---|
| `SUCCESS` | 200 | Deu certo |
| `VALIDATION_ERROR` | 400 | Payload inválido — a mensagem explica o campo |
| `INVALID_JID` | 400 | Número ou JID mal formatado |
| `AUTH_ERROR` | 401 | Credenciais inválidas |
| `SESSION_SAVED_ERROR` | 401 | Sessão não encontrada, faça login |
| `NOT_FOUND` | 404 | Recurso inexistente |
| `WA_CLI_ERROR` | 500 | WhatsApp não conectado |
| `INTERNAL_SERVER_ERROR` | 500 | Erro inesperado |

### Erros comuns nos endpoints interativos

| Mensagem | Causa |
|---|---|
| `buttons: maximum 3 buttons allowed, got 4.` | Passou de 3 botões — use lista |
| `buttons[1].url: cannot be blank for type cta_url.` | Faltou a URL |
| `buttons[0].id: duplicated value "x"` | Dois botões com o mesmo ID |
| `sections[0].rows: maximum 10 rows per section` | Divida em mais seções |
| `sections: maximum 30 rows in total` | Excedeu o total de linhas |

---

## Observações importantes sobre botões e listas

**Não é API oficial.** Mensagens interativas usam o protocolo NativeFlow do WhatsApp Web via engenharia reversa. Funcionam hoje, mas a Meta pode alterar o comportamento sem aviso prévio.

**Teste com um número descartável.** Enviar mensagens interativas em volume por conta pessoal aumenta o risco de bloqueio. Valide tudo antes de usar o número principal do negócio.

**Contas Business renderizam melhor.** Listas, em especial, podem não aparecer em algumas versões do WhatsApp quando enviadas de conta pessoal — e falham em silêncio, sem retornar erro.

**Estourar limites de caracteres não gera erro.** Se um título ultrapassar o limite, a mensagem pode simplesmente não renderizar no aparelho. Por isso a API já trunca automaticamente em 20 (botão), 24 (título de linha) e 72 (descrição) caracteres.

---

## Licença

Veja [LICENCE.txt](LICENCE.txt).
