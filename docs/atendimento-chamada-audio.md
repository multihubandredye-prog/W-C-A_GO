# Atendimento de chamada de áudio recebida — experimental

`POST /call/answer` solicita que a API atenda uma chamada individual de áudio que ainda esteja tocando, reproduza um arquivo MP3 e desligue após terminar. Não atende vídeo nem chamadas de grupo. O envio de chamadas existente em `/send/call` continua separado.

## Preparação

- A API precisa estar atualizada com esta implementação e o dispositivo WhatsApp conectado.
- Desative a recusa automática (`WHATSAPP_AUTO_REJECT_CALL=false`) e reinicie a API se alterar a configuração. A recusa automática existente não foi alterada; se estiver ativa, ela pode recusar a chamada antes da solicitação.
- Configure o webhook e use o `CallID` do evento `call.offer`. O Java Receiver já expõe esse valor como `%call_id`; selecione na requisição o mesmo dispositivo que recebeu a chamada.
- O MP3 deve estar no próprio aparelho onde a API roda, em `/storage/emulated/0/`, ser legível pelo processo, válido e ter no máximo 20 MB (20.000.000 bytes). O arquivo original não é apagado nem alterado. O endpoint não aceita URL, arquivo remoto nem upload multipart.
- Atenda rapidamente, enquanto a chamada ainda toca. Pode ocorrer uma pequena corrida entre o webhook e o registro interno: se receber `409` indicando chamada indisponível, pode tentar novamente após uma pequena espera, enquanto ela tocar. Não repita após receber sucesso.

## Exemplo

Use o mesmo endereço, autenticação Basic (se configurada), prefixo `APP_BASE_PATH` e `X-Device-Id` das demais rotas da API:

```sh
curl -X POST 'http://127.0.0.1:3000/call/answer' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Id: ID_DO_DISPOSITIVO' \
  -u 'USUARIO:SENHA' \
  -d '{"call_id":"ID_DO_WEBHOOK","audio_path":"/storage/emulated/0/Download/aviso.mp3"}'
```

Substitua os valores pelo dispositivo, credenciais e caminho reais; omita `-u` se a autenticação não estiver habilitada. No Tasker, a ação HTTP Request pode enviar o mesmo JSON, substituindo `ID_DO_WEBHOOK` por `%call_id`. Garanta que os valores sejam codificados como JSON corretamente.

A resposta HTTP `200` com `results.status=answer_requested` confirma apenas que a solicitação de atendimento foi iniciada, **não** que a mídia conectou, que o MP3 terminou ou que a outra pessoa ouviu. Um `400` indica campos/arquivo inválidos ou erro no atendimento; `409` indica chamada indisponível, já em atendimento, modalidade não suportada ou recusa automática ativa. Os detalhes estão em `message`. Se a conexão de mídia ficar presa, há encerramento por segurança após cinco minutos; se o contato desligar, o arquivo é liberado. Falhas de decodificação durante a reprodução também encerram a chamada, mas a biblioteca usada não diferencia essa falha do fim normal no callback.

O atendimento e a transmissão dependem do WhatsApp, do estado de rede e da biblioteca meowcaller. **Ainda é necessário testar em aparelhos reais**, incluindo áudio audível, fim completo do arquivo, desligamento remoto, atraso entre webhook e atendimento, recusa automática e tentativas de chamada de vídeo/grupo. Os testes automatizados não comprovam o funcionamento da mídia no WhatsApp. Proteja a API com autenticação: clientes autorizados podem pedir a leitura de arquivos MP3 acessíveis no armazenamento compartilhado.
