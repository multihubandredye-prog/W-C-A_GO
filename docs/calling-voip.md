# Guia de Implementação: Chamadas de Voz VoIP (`POST /send/call`)

Este documento descreve detalhadamente a implementação, arquitetura, uso atual e instruções de monitoramento técnico da funcionalidade de **Chamadas de Voz VoIP** na **Whats Connect Api** (repositório de teste `Wca-Test`).

---

## 1. Visão Geral e Arquitetura

A rota `POST /send/call` foi implementada para permitir que a API realize chamadas de áudio no WhatsApp (1:1) utilizando a biblioteca open-source em Go **`meowcaller` (`github.com/purpshell/meowcaller`)**, que opera em conjunto com o cliente `whatsmeow`.

### Arquivos Modificados/Criados na API
- **`src/domains/send/call.go`**: Define os DTOs `CallRequest` (`phone`, `duration`, `audio_url`, `audio_path`) e `CallResponse`.
- **`src/domains/send/interfaces.go`**: Expõe a interface `ICallSender` integrada a `ISendUsecase`.
- **`src/usecase/send_call.go`**: Implementa a lógica de chamadas:
  - Cria o cliente `meowcaller.NewClient(client)`.
  - Origina a ligação de voz no WhatsApp para o número informado.
  - Registra o callback `call.OnReady` para identificar quando a mídia RTP/Relay está fluindo.
  - Baixa automaticamente áudios de URLs públicas (`audio_url`), decodifica strings Base64 (`audio_path`) ou abre arquivos locais do servidor.
  - Encerra a ligação programaticamente após o tempo configurado (`duration`, padrão 15 segundos).
- **`src/ui/rest/send.go`**: Registra o endpoint REST `POST /send/call`.
- **`src/validations/send_validation.go`**: Valida o número de telefone e garante que a duração da chamada esteja entre 1 e 3600 segundos.

---

## 2. Parâmetros da Rota e Exemplos cURL

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `phone` | string | **Sim** | Número do destinatário com DDI e DDD (ex: `5581999999999`) |
| `duration` | int | Não | Duração em segundos da chamada antes de encerrar (padrão 15s) |
| `audio_url` | string | Não | URL pública (http/https) do áudio MP3 para a API baixar e tocar na chamada |
| `audio_path` | string | Não | Áudio em Base64 (`data:audio/mp3;base64,...` ou string pura) OU caminho de arquivo MP3 no servidor |

### Exemplo 1: Toque de Alerta / Chamada Perdida
```bash
curl -X POST http://localhost:3000/send/call \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5581999999999",
    "duration": 15
  }'
```

### Exemplo 2: Chamada com Reprodução de Áudio (URL Pública)
```bash
curl -X POST http://localhost:3000/send/call \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5581999999999",
    "audio_url": "https://meusite.com/audio/aviso.mp3",
    "duration": 30
  }'
```

### Exemplo 3: Chamada com Reprodução de Áudio em Base64
```bash
curl -X POST http://localhost:3000/send/call \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5581999999999",
    "audio_path": "data:audio/mp3;base64,//uQx...",
    "duration": 30
  }'
```

---

## 3. Estágio Atual da Tecnologia e Comportamento ("Conectando...")

### O que funciona 100% hoje?
1. **Sinalização Completa (`KAT-verified`):** O servidor envia o comando de chamada (*Call Offer*), o celular do usuário toca imediatamente e o atendimento é detectado em tempo real pelo servidor.
2. **Sistema de Alerta Premium (Ring Alert):** Ideal para acordar administradores em alertas do Tasker/servidores ou notificar clientes urgentes via toque telefônico.

### Por que o WhatsApp no celular fica em "Conectando..." sem tocar o áudio?
Existem dois motivos técnicos fundamentais associados a chamadas VoIP/WebRTC no WhatsApp Web:
1. **Status Experimental do Módulo `live-relay` no `meowcaller`:**
   - No CHANGELOG oficial do projeto `meowcaller`, o módulo de transmissão de mídia de voz sobre servidores TURN da Meta (`ConnectRelayMedia`) é marcado como **`// NOT VALIDATED: live-relay only`**.
   - A biblioteca implementa a sinalização de forma homologada, mas a negociação do túnel UDP/RTP com os servidores TURN da Meta em produção ainda é um recurso de pesquisa em desenvolvimento pelo autor.
2. **Tráfego UDP e NAT em Contêineres Docker:**
   - Chamadas VoIP dependem de pacotes UDP bidirecionais (portas 3478/443).
   - Em ambientes Docker padrão (onde apenas portas TCP são expostas e há restrição de NAT), os pacotes UDP enviados pelos servidores TURN da Meta não conseguem entrar no contêiner. Sem receber pacotes de entrada (`rtpIn == 0`), o evento `OnReady` nunca é disparado e o WhatsApp permanece em "Conectando...".

---

## 4. Como Verificar se a Biblioteca/Mídia já está Disponível e Homologada

Como o projeto open-source `meowcaller` é atualizado constantemente pela comunidade de pesquisa, você pode monitorar e testar quando o áudio na linha for oficializado seguindo estes passos:

### Passo 1: Verifique o CHANGELOG do Projeto Oficial
1. Acesse o repositório oficial: **https://github.com/purpshell/meowcaller**
2. Abra o arquivo **`CHANGELOG.md`** ou **`engine_media.go`**.
3. Verifique se a anotação `// NOT VALIDATED: live-relay only` foi removida ou alterada para **`KAT-verified`** / **`implemented`** nas seções de mídia e retransmissão (*relay media transport*).

### Passo 2: Atualize a Dependência no Projeto em Go
Quando houver nova versão da biblioteca, execute dentro da pasta `src/` da API:
```bash
go get -u github.com/purpshell/meowcaller@latest
go mod tidy
go build ./...
```

### Passo 3: Verifique pelo Log do Servidor
Ao disparar uma requisição para `POST /send/call` e atender a ligação no celular, observe o terminal/log da API:
- **Status Atual (Aguardando Handshake):**  
  `INFO: Peer accepted call <ID>; relay negotiation in progress...`  
  *(Se o log parar nesta linha, o túnel TURN/UDP ainda está retido pela rede ou pela biblioteca).*
- **Status Homologado (Áudio na Linha Funcionando):**  
  `INFO: Call <ID> is ready (RTP/relay media flowing); starting duration timer (30s)`  
  *(Se esta linha aparecer no log, significa que o handshake UDP/TURN foi concluído com sucesso e o arquivo MP3 começou a ser reproduzido diretamente na chamada!)*
