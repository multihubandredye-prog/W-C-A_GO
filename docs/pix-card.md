# Cartão de chave Pix com botão de copiar

`POST /send/pix-card` envia uma imagem e um botão de cópia como uma mensagem interativa. Não gera cobrança, não transfere dinheiro e não verifica pagamento. O formato é uma imagem com botão, não o cartão Pix nativo.

```json
{
  "phone": "5511999999999",
  "image_url": "data:image/png;base64,BASE64_DA_IMAGEM",
  "pix_key": "5511888888888",
  "body": "Chave Pix"
}
```

- `phone`: destinatário da mensagem, com código do país ou JID suportado.
- `image_url`: obrigatório; URL HTTP/HTTPS ou `data:image/png;base64,` seguido da imagem em Base64. Limite de 5 MiB para a imagem decodificada; o limite geral de requisição também se aplica. Não aceita caminho de arquivo.
- `pix_key`: obrigatório; texto colocado no botão, independente da imagem. O cliente deve manter os dois coerentes. Espaços no início e no fim são removidos pelo envio de botões.
- `body`: opcional; padrão `Chave Pix` quando vazio.
- `duration` e `is_forwarded`: opcionais, com a mesma semântica de `/send/buttons`.

A API cria um botão `Copiar chave Pix` de tipo `copy`, que copia `pix_key`. Usa o mesmo dispositivo e autenticação que os demais endpoints `/send`: `X-Device-Id` quando necessário e autenticação HTTP Basic se configurada. O caminho também segue `APP_BASE_PATH`, se configurado.

Resposta bem-sucedida segue o formato padrão da API, com `results.message_id` e `results.status`; isso não confirma entrega, renderização, cópia nem pagamento. Ausência de imagem ou chave retorna HTTP 400; erros de validação e de envio seguem o tratamento padrão da API. `/send/buttons` permanece disponível sem mudanças de contrato.

Requer teste real após compilar e instalar: validar renderização, cópia, tamanho e proporção da imagem, aparelho remetente/destinatário e, se necessário, grupos. A apresentação é controlada pelo WhatsApp e pode mudar.
