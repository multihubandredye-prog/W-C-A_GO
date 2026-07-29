<p align="center">
  <img src="src/views/assets/wca-logo.svg" width="140" alt="Whats Connect Api Logo">
</p>

<h1 align="center">Whats Connect Api</h1>

<p align="center">
  API REST para WhatsApp multi-dispositivo, escrita em Go.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.25">
  <img src="https://img.shields.io/badge/Fiber-v2-00C7B7?style=flat-square" alt="Fiber v2">
  <img src="https://img.shields.io/badge/whatsmeow-multidevice-25D366?style=flat-square&logo=whatsapp&logoColor=white" alt="whatsmeow">
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker ready">
</p>

---

## Visão geral

**Whats Connect Api** expõe o WhatsApp como uma API REST convencional. Toda a
comunicação com o protocolo multi-dispositivo é feita pela biblioteca
[whatsmeow](https://github.com/tulir/whatsmeow), sem depender da API oficial
da Meta nem de emulação de navegador.

| Recurso | Descrição |
|---|---|
| **Mensagens** | Texto, imagem, vídeo, áudio, documento, figurinha, contato, link e localização |
| **Interativo** | Botões de ação e listas de seleção, com captura das respostas |
| **Enquetes** | Criação e leitura dos votos |
| **Conversas** | Histórico, busca, fixar, arquivar e mensagens temporárias |
| **Grupos** | Criação, participantes, permissões, convites e solicitações |
| **Multi-sessão** | Vários números conectados simultaneamente na mesma instância |
| **Webhooks** | Eventos de entrada e saída entregues em tempo real |

---

## Índice

**Primeiros passos**
- [Instalação](#instalação)
- [Configuração](#configuração)
- [Autenticação](#autenticação)
- [Múltiplos dispositivos](#múltiplos-dispositivos)
- [Formato das respostas](#formato-das-respostas)

**Referência da API**
1. [App / Sessão](#1-app--sessão)
2. [Dispositivos](#2-dispositivos)
3. [Envio de mensagens](#3-envio-de-mensagens)
4. [Botões interativos](#4-botões-interativos)
5. [Listas interativas](#5-listas-interativas)
6. [Gerenciar mensagens](#6-gerenciar-mensagens)
7. [Conversas](#7-conversas)
8. [Usuário](#8-usuário)
9. [Grupos](#9-grupos)
10. [Newsletter](#10-newsletter)

**Integração**
- [Recebendo respostas de botões e listas](#recebendo-respostas-de-botões-e-listas)
- [Códigos de erro](#códigos-de-erro)
- [Observações sobre mensagens interativas](#observações-sobre-mensagens-interativas)

---

## Instalação

### Docker

```bash
docker compose up -d
```

### A partir do código-fonte

Requer Go 1.25 ou superior.

```bash
cd src
go mod download
go run . rest
```

O serviço fica disponível em `http://localhost:3000`.

---

## Configuração

As opções são lidas de variáveis de ambiente ou de um arquivo `.env` na raiz
do projeto.

#### Aplicação

| Variável | Padrão | Descrição |
|---|---|---|
| `APP_PORT` | `3000` | Porta HTTP do servidor |
| `APP_HOST` | `0.0.0.0` | Interface de escuta |
| `APP_DEBUG` | `false` | Habilita log detalhado das requisições |
| `APP_OS` | Não | Nome exibido em *Aparelhos conectados* no WhatsApp |
| `APP_BASIC_AUTH` | Não | Credenciais `usuario:senha`, separadas por vírgula |
| `APP_BASE_PATH` | Não | Prefixo das rotas, por exemplo `/api` |
| `APP_TRUSTED_PROXIES` | Não | Faixas de IP confiáveis atrás de proxy reverso |

#### Banco de dados

| Variável | Padrão | Descrição |
|---|---|---|
| `DB_URI` | `file:storages/whatsapp.db` | Conexão principal — SQLite ou PostgreSQL |
| `DB_KEYS_URI` | Não | Armazenamento das chaves de criptografia |

#### Webhook

| Variável | Padrão | Descrição |
|---|---|---|
| `WHATSAPP_WEBHOOK` | Não | URLs que recebem os eventos, separadas por vírgula |
| `WHATSAPP_WEBHOOK_SECRET` | Não | Segredo usado para assinar o payload |
| `WHATSAPP_WEBHOOK_EVENTS` | *todos* | Filtra quais eventos são enviados |
| `WHATSAPP_WEBHOOK_INCLUDE_OUTGOING` | `false` | Também notifica mensagens enviadas por você |
| `WHATSAPP_WEBHOOK_INSECURE_SKIP_VERIFY` | `false` | Ignora validação de certificado TLS |

#### Comportamento do WhatsApp

| Variável | Padrão | Descrição |
|---|---|---|
| `WHATSAPP_AUTO_REPLY` | Não | Resposta automática para toda mensagem recebida |
| `WHATSAPP_AUTO_MARK_READ` | `false` | Marca as mensagens como lidas ao recebê-las |
| `WHATSAPP_AUTO_REJECT_CALL` | `false` | Rejeita chamadas automaticamente |
| `WHATSAPP_AUTO_DOWNLOAD_MEDIA` | `true` | Baixa as mídias recebidas para o disco |
| `AUTO_DELETE_MEDIA_DURATION` | Não | Segundos até apagar a mídia baixada |
| `WHATSAPP_ACCOUNT_VALIDATION` | `true` | Verifica se o número existe no WhatsApp antes de enviar |
| `WHATSAPP_CHAT_STORAGE` | `true` | Persiste o histórico de conversas |

#### Presença

| Variável | Padrão | Descrição |
|---|---|---|
| `WHATSAPP_PRESENCE_ON_CONNECT` | `unavailable` | Estado ao conectar: `available`, `unavailable` ou `none` |
| `WHATSAPP_PRESENCE_PULSE_ENABLED` | `true` | Habilita o pulso periódico de presença |
| `WHATSAPP_PRESENCE_PULSE_INTERVAL` | `24h` | Intervalo entre cada pulso |
| `WHATSAPP_PRESENCE_PULSE_DURATION` | `5m` | Duração de cada pulso |

> O pulso de presença fica *online* apenas pelo tempo definido em
> `PULSE_DURATION`, a cada `PULSE_INTERVAL`, e volta a *offline* em seguida.
> Com os valores padrão, isso equivale a 5 minutos online por dia. Para nunca
> aparecer online, defina `WHATSAPP_PRESENCE_PULSE_ENABLED=false` e
> `WHATSAPP_PRESENCE_ON_CONNECT=unavailable`.

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

### `POST /send/call`
Realiza uma chamada de voz (VoIP) no WhatsApp para o contato. Ideal para alertas de sistema, avisos de emergência ou URA.

```bash
# Exemplo 1: Chamada simples de alerta (faz tocar e encerra após o tempo definido)
curl -X POST http://localhost:3000/send/call \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "duration": 15
  }'
```

```bash
# Exemplo 2: Chamada com reprodução de áudio (toca o arquivo MP3 ao atender)
curl -X POST http://localhost:3000/send/call \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "audio_path": "/caminho/para/alerta.mp3",
    "duration": 30
  }'
```

| Campo | Tipo | Descrição |
|---|---|---|
| `phone` | string | **Obrigatório.** Número do destinatário com DDI e DDD |
| `duration` | int | *Opcional.* Duração da chamada em segundos antes de desligar (padrão 15s) |
| `audio_path` | string | *Opcional.* Caminho de um arquivo MP3 no servidor para reproduzir quando o usuário atender |

---

## 4. Botões interativos

Envia uma mensagem com **até 3 botões** clicáveis. Ideal para decisões rápidas: confirmar, escolher entre poucas opções, abrir um link ou ligar.

> **Precisa de mais de 3 opções?** Use [listas](#5-listas-interativas).

### `POST /send/buttons`

#### Campos

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `phone` | string | Sim | Destinatário |
| `body` | string | Sim | Texto principal da mensagem |
| `buttons` | array | Sim | 1 a 3 botões |
| `title` | string | Não | Cabeçalho acima do texto |
| `footer` | string | Não | Rodapé em letra menor |
| `image_url` | string | Não | Imagem no cabeçalho (URL http/https ou `data:image/...;base64,...`) |
| `duration` | int | Não | Mensagem temporária, em segundos |
| `is_forwarded` | bool | Não | Marca como encaminhada |

#### Tipos de botão

| `type` | O que faz | Campos exigidos |
|---|---|---|
| `reply` *(padrão)* | Resposta rápida — devolve o `id` no webhook | `title`, `id` |
| `cta_url` | Abre um link | `title`, `url` |
| `cta_call` | Inicia uma ligação | `title`, `phone_number` |
| `copy` | Copia um código | `title`, `copy_code` |

| `type` | Campo obrigatório além do `title` |
|---|---|
| `reply` | `id` |
| `cta_url` | `url` |
| `cta_call` | `phone_number` |
| `copy` | `copy_code` |

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

#### Exemplo 4 — Cobrança com chave PIX copiável

O botão `copy` é ideal para PIX: o cliente copia a chave com um toque, sem
risco de errar ao digitar.

```bash
curl -X POST http://localhost:3000/send/buttons \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "body": "Para realizar o pagamento, use a chave PIX abaixo:\n\n*Chave PIX (Telefone):*\n(81) 84752-564\n\n*Nome:* João Silva\n*Banco:* Nubank",
    "footer": "Clínica Bem Estar",
    "buttons": [
      { "type": "copy",  "title": "Copiar PIX",  "copy_code": "558184752564" },
      { "type": "reply", "title": "Já paguei",   "id": "ja_paguei" },
      { "type": "reply", "title": "Comprovante", "id": "enviar_comprovante" }
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

## 5. Listas interativas

Envia um **menu suspenso** com muito mais opções que os botões. As opções ficam organizadas em **seções**, cada uma com título e itens que podem ter descrição.

Perfeito para cardápios, catálogos, agendamentos e qualquer escolha com muitas alternativas.

### `POST /send/list`

#### Campos

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `phone` | string | Sim | Destinatário |
| `description` | string | Sim | Texto principal da mensagem |
| `sections` | array | Sim | Grupos de opções |
| `button_text` | string | Não | Rótulo do botão que abre a lista (padrão: `Select`) |
| `title` | string | Não | Cabeçalho acima do texto |
| `footer` | string | Não | Rodapé |
| `duration` | int | Não | Mensagem temporária |
| `is_forwarded` | bool | Não | Marca como encaminhada |

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
- Até **10 linhas no total, somando todas as seções (limite do WhatsApp)**
- `row_id` deve ser único entre **todas** as seções
- Títulos e descrições são truncados no limite

> Para catálogos maiores que 10 itens, use a [paginação](#paginação-de-listas).

#### Paginação de listas

O WhatsApp entrega no máximo 10 linhas por mensagem. Para oferecer um catálogo
maior, envie todos os itens de uma vez com `paginate` e a API divide em páginas,
reservando a última linha de cada uma para a navegação.

| Campo | Padrão | Descrição |
|---|---|---|
| `paginate` | `false` | Divide o catálogo em páginas |
| `page_size` | `9` | Itens por página, sem contar a linha de navegação (máx. 9) |
| `pagination_label` | `Ver mais` | Texto da linha de navegação |
| `forward_pagination` | `false` | Notifica o webhook quando o cliente navega |

Com 30 itens e `page_size: 9`:

| Página | Itens | Linha de navegação |
|---|---|---|
| 1 | 1 a 9 | Ver mais — Página 2 de 4 |
| 2 | 10 a 18 | Ver mais — Página 3 de 4 |
| 3 | 19 a 27 | Ver mais — Página 4 de 4 |
| 4 | 28 a 30 | — |

Quando o cliente toca na navegação, a API envia a próxima página
automaticamente. O catálogo fica disponível por **7 dias** e sobrevive a
reinícios do servidor.

```bash
curl -X POST http://localhost:3000/send/list \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5588999999999",
    "title": "Moda & Estilo",
    "description": "Escolha a peça que deseja comprar hoje:",
    "footer": "Entrega em até 3 dias úteis",
    "button_text": "Ver produtos",
    "paginate": true,
    "page_size": 9,
    "pagination_label": "Ver mais",
    "forward_pagination": true,
    "sections": [
      {
        "title": "Tipos de Roupa",
        "rows": [
          { "row_id": "camiseta_basica_preta",  "title": "Camiseta Básica Preta",  "description": "Algodão 100% — R$ 49,90" },
          { "row_id": "camiseta_basica_branca", "title": "Camiseta Básica Branca", "description": "Algodão 100% — R$ 49,90" }
        ]
      }
    ]
  }'
```

**Restrições:** a paginação aceita uma única seção e `page_size` de no máximo 9,
já que a décima linha é reservada para a navegação.

**Notificação da navegação:** com `forward_pagination: true`, o clique em
"Ver mais" gera um webhook com `IsPagination: true`, a página aberta e os itens
exibidos — veja
[respostas de botões e listas](#recebendo-respostas-de-botões-e-listas). Com
`false`, a próxima página é enviada do mesmo jeito, mas sem notificar.

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
| Quantidade de opções | Até **3** | Até **10** no total |
| Visual | Botões fixos abaixo da mensagem | Botão único que abre um menu |
| Agrupamento por categoria | Não | Sim, seções com título |
| Descrição em cada opção | Não | Sim |
| Abrir link, ligar ou copiar | Sim | Não, apenas seleção |
| Imagem no cabeçalho | Sim | Não |
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

Enviar o botão resolve apenas metade do fluxo: é preciso identificar qual opção
o destinatário escolheu. Quando ele responde, a API dispara um webhook para a
URL configurada em `WHATSAPP_WEBHOOK`.

O payload traz o campo unificado **`InteractiveReply`**, com a mesma estrutura
para botões e listas, evitando tratamento condicional no consumidor.

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

### Navegação em listas paginadas

Quando a lista foi enviada com `forward_pagination: true`, o clique em
"Ver mais" também gera um webhook. Ele traz `IsPagination: true` e o conteúdo
da página que acabou de ser entregue:

```json
{
  "Payload": {
    "Type": "ListResponseMessage",
    "InteractiveReply": {
      "SelectedID": "__wca_page_2",
      "IsPagination": true,
      "Page": 2,
      "TotalPages": 4,
      "TotalRows": 30
    },
    "PaginationSent": {
      "MessageID": "3EB0C7...",
      "Page": 2,
      "TotalPages": 4,
      "RowsCount": 10,
      "HasMore": true,
      "Rows": [
        { "RowID": "camiseta_polo_azul", "Title": "Camiseta Polo Azul", "Description": "R$ 79,90" }
      ]
    }
  }
}
```

Use `IsPagination` para separar navegação de escolha real:

```javascript
const reply = req.body.Payload?.InteractiveReply;

if (reply?.IsPagination) {
  return res.sendStatus(200);   // navegação: a API já enviou a próxima página
}

processarPedido(reply.SelectedID);
```

Com `forward_pagination: false` (padrão), esse evento não é enviado e o webhook
recebe apenas as escolhas de produto.

### Consultando o que foi enviado

Quando `WHATSAPP_WEBHOOK_INCLUDE_OUTGOING=true`, o webhook das mensagens que
você envia inclui as opções oferecidas, em `ButtonsSent` (botões) ou
`ListSent` (listas):

```json
{
  "Payload": {
    "Type": "ButtonsMessage",
    "ButtonsSent": {
      "Body": "Para realizar o pagamento, use a chave PIX abaixo:",
      "Footer": "Clínica Bem Estar",
      "ButtonsCount": 3,
      "Buttons": [
        { "Name": "cta_copy",    "Title": "Copiar PIX",  "CopyCode": "558184752564" },
        { "Name": "quick_reply", "Title": "Já paguei",   "ID": "ja_paguei" },
        { "Name": "quick_reply", "Title": "Comprovante", "ID": "enviar_comprovante" }
      ]
    }
  }
}
```

### Tipos reportados no campo `Type`

| Situação | `Type` |
|---|---|
| Botões enviados por você | `ButtonsMessage` |
| Lista enviada por você | `ListMessage` |
| Clique em botão | `ButtonsResponseMessage` |
| Seleção em lista | `ListResponseMessage` |

### Botões que não geram resposta

Apenas botões do tipo `reply` retornam evento. Os demais executam uma ação
local no aparelho do destinatário, sem enviar nada ao servidor:

| Tipo | Retorna evento |
|---|---|
| `reply` | Sim |
| `copy` | Não — o texto é copiado localmente |
| `cta_url` | Não — abre o navegador |
| `cta_call` | Não — abre o discador |

Essa é uma característica do protocolo do WhatsApp. Para saber que o cliente
copiou uma chave PIX, por exemplo, ofereça também um botão `reply` como
"Já paguei" e use o clique nele como confirmação.

### O campo que importa

**`SelectedID`** é a chave de tudo. Ele devolve exatamente o `id` (botão) ou `row_id` (lista) que você definiu ao enviar a mensagem.

Por isso vale usar identificadores descritivos:

```json
{ "row_id": "pizza_marg" }   // recomendado: legível no código
{ "row_id": "1" }            // evite: perde o significado
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
| `sections: maximum 10 rows in total` | Excedeu o total de linhas permitido |

---

## Observações sobre mensagens interativas

**Recurso não oficial.** As mensagens interativas utilizam o protocolo NativeFlow do WhatsApp Web. O comportamento pode mudar sem aviso prévio a critério da Meta.

**Valide em um número de testes.** O envio de mensagens interativas em volume a partir de uma conta pessoal aumenta o risco de bloqueio. Homologue o fluxo antes de aplicá-lo ao número principal.

**Contas Business têm melhor compatibilidade.** As listas podem não ser renderizadas em determinadas versões do aplicativo quando enviadas por conta pessoal, sem que a API retorne erro.

**Exceder os limites de caracteres não gera erro.** A mensagem simplesmente deixa de ser renderizada no dispositivo. Para evitar esse cenário, a API trunca automaticamente em 20 caracteres (título de botão), 24 (título de linha) e 72 (descrição).

---

## Direitos Autorais e Uso

Todos os direitos reservados a **multihubandredye-prog**.  
Este software, denominado **Whats Connect Api**, é mantido na modalidade Premium.  
Consulte [LICENCE.txt](LICENCE.txt) para os termos completos.
