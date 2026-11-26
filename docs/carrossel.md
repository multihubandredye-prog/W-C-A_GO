# Carrossel de imagens — experimental

`POST /send/carousel` monta uma única mensagem interativa de 2 a 10 cartões de imagem, na ordem informada. Cada cartão traz texto e um botão de resposta rápida. A estrutura e o envio são testados localmente; **a exibição real, a entrega e as respostas dos botões precisam ser testadas em aparelhos reais**. O limite de 10 é uma validação desta implementação, não uma garantia de que o WhatsApp aceitará todos os 10 cartões.

```json
{
  "phone": "5581999999999",
  "body": "Confira as imagens:",
  "image_paths": "/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image1.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image2.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image3.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image4.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image5.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image6.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image7.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image8.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image9.jpg,/storage/emulated/0/Download/arquivos-para-teste/arquivos-para-teste/Imagens/image10.jpg",
  "button_text": "Selecionar"
}
```

- `phone`: número com código do país ou JID suportado; o número acima é apenas exemplo.
- `image_paths`: obrigatório; texto único com 2 a 10 caminhos absolutos separados por vírgulas. A API lê os arquivos **no aparelho/ambiente onde o processo da API está rodando**; eles devem estar no armazenamento compartilhado `/storage/emulated/0/` e ser acessíveis para o processo. Como a vírgula é delimitadora, não use caminhos cujos nomes contenham vírgula. Não são aceitos URLs ou Base64 nesse campo.
- JPEG ou PNG válidos, até **5 MiB por imagem e 25 megapixels**. Todos os arquivos são validados antes do primeiro upload. As imagens originais não são sobrescritas; uma miniatura é gerada para a prévia.
- `body` é opcional, padrão `Confira as imagens:`. `button_text` é opcional, padrão `Selecionar` (máximo de 20 caracteres). Cada cartão recebe o mesmo texto do botão e identificador `carousel_1`, `carousel_2` etc., conforme a ordem dos caminhos.
- `card_titles` é opcional: um array de textos na mesma ordem e quantidade dos caminhos; sem ele, cada cartão mostra o nome do arquivo, incluindo a extensão, por exemplo `image1.jpg` e `image2.jpg`. Cada texto deve ter de 1 a 160 caracteres. `body` aceita até 1024 caracteres.
- `duration` e `is_forwarded` são opcionais e seguem as mesmas regras de `/send/buttons`.
- Use o mesmo acesso das demais rotas `/send`: Basic Auth se configurado, `X-Device-Id` quando necessário e `APP_BASE_PATH` se configurado.
- A resposta bem-sucedida contém `results.message_id` e `results.status`, mas **não prova entrega, renderização nem clique**. Se a resposta da rede for perdida, confira no WhatsApp antes de repetir para evitar duplicidade.
- As respostas interativas existentes podem expor `InteractiveReply.SelectedID`; é necessário testar se o WhatsApp entrega o ID de cada cartão ao webhook.
- **Segurança:** a rota pode ler imagens do armazenamento compartilhado acessíveis à API. Não exponha a API sem autenticação ou para clientes não confiáveis.

## Remoção

A implementação é isolada: remova o registro `app.Post("/send/carousel", rest.SendCarousel)` de `src/ui/rest/send.go` e remova `src/domains/send/carousel.go`, `src/usecase/send_carousel.go`, `src/usecase/send_carousel_test.go`, `src/ui/rest/send_carousel.go`, `src/ui/rest/send_carousel_test.go` e este documento. A classificação de respostas também acrescenta os arquivos `src/infrastructure/whatsapp/carousel_reply.go` e `carousel_reply_test.go` e três pontos em `event_message.go`; remova-os ao retirar a funcionalidade. Para reverter via Git, reverta primeiro os commits posteriores que a alteraram e depois o commit original, verificando eventuais conflitos. Não altera o contrato de `/send/buttons` nem o endpoint Pix.

## Resposta de seleção no webhook

A API reconhece respostas de botões com os IDs reservados `carousel_1` até `carousel_10` como seleção de carrossel. Para a resposta recebida, `Event` permanece `message`, enquanto `Payload.Type` passa a ser `CarouselResponseMessage`. Um exemplo ilustrativo do trecho de `Payload` é:

```json
{
  "Type": "CarouselResponseMessage",
  "InteractiveReply": {
    "Type": "carousel",
    "SelectedID": "carousel_2",
    "CardIndex": 2,
    "SelectedText": "Selecionar",
    "RepliedToID": "ID_DA_MENSAGEM_ORIGINAL"
  },
  "CarouselReply": {
    "Type": "carousel",
    "SelectedID": "carousel_2",
    "CardIndex": 2,
    "SelectedText": "Selecionar",
    "RepliedToID": "ID_DA_MENSAGEM_ORIGINAL"
  }
}
```

`SelectedText` e `RepliedToID` só aparecem quando a resposta recebida os contém. Esses dados identificam a posição, **não o arquivo nem seu caminho**; guarde a lista enviada e associe-a à mensagem original para recuperar essa informação, quando o ID original estiver disponível. Os IDs são reutilizados entre envios. A classificação depende da convenção de nomes, não de uma prova da origem: não use `carousel_1` a `carousel_10` em botões comuns. Respostas de botão comuns e listas mantêm seus tipos anteriores. O comportamento efetivo da resposta depende do cliente WhatsApp e deve ser validado em aparelho real. Atualize tanto a API quanto o Java do Receiver para usar os novos campos.

## Seleção de imagens na página da API

O formulário aceita selecionar 2 a 10 imagens JPEG/PNG no navegador, adicionar em etapas, remover e ajustar a ordem. Nesse modo os arquivos são transmitidos para a API via `multipart/form-data` em campos repetidos `images`; a API valida o conteúdo e os limites antes dos uploads ao WhatsApp. Os nomes dos arquivos são usados como títulos por padrão. O envio JSON com `image_paths` continua disponível e não foi alterado. Os arquivos selecionados não precisam estar no armazenamento local da API. A soma deve respeitar também o limite geral do corpo da requisição configurado na API.
