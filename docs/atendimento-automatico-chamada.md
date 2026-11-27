# Atendimento automático de chamadas de áudio — experimental

A API pode tentar atender automaticamente chamadas **individuais de áudio**, reproduzir um MP3 e desligar ao terminar. Não usa Tasker, webhook ou `/call/answer` para iniciar o atendimento. A configuração vale para **todos os dispositivos conectados à mesma API e todos os números**; não há filtros, agenda, grupos nem vídeo nesta versão. O webhook existente continua funcionando. A rota manual de atendimento foi removida; `/send/call` permanece disponível para originar chamadas.

No `.env` da API (no aparelho), configure:

```env
WHATSAPP_AUTO_REJECT_CALL=false
WHATSAPP_AUTO_ANSWER_CALL=true
WHATSAPP_AUTO_ANSWER_AUDIO=/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call/2026_11_26_21_14_59.mp3
```

O atendimento é **desativado por padrão**. Substitua o caminho se necessário. O arquivo MP3 deve existir dentro de `/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call/`, ser legível pelo processo da API e ter no máximo 20.000.000 bytes. O original não é alterado. A API não copia o áudio para lá: crie a pasta e coloque o arquivo nela com uma ferramenta que tenha acesso à pasta privada do Tasker. Caminhos fora dela, inclusive links simbólicos externos, são rejeitados. URLs e upload não são aceitos. Reinicie a API para aplicar as mudanças.

Se `WHATSAPP_AUTO_REJECT_CALL=true`, a recusa tem prioridade e a API não inicia atendimento automático. Se o MP3 estiver inválido ou indisponível, não atende. A API tenta atender somente após a biblioteca registrar a chamada e a oferta ser identificada como individual e de áudio; falha de chave ou ausência da oferta impossibilita o atendimento. Não são alterados os bancos de dados ou a configuração de chaves.

O player é conectado após a solicitação de atendimento, sem esperar `OnReady` ou áudio do outro lado. Isso acompanha a ordem básica do exemplo da biblioteca, mas **não garante** que o início do MP3 seja audível: o estado de mídia e os sinais do WhatsApp precisam ser validados em ligação real. A biblioteca pode iniciar mídia antes da resposta, perder sinais durante a espera ou falhar ao conectar. `Answer` não comprova que o telefone foi atendido ou que a pessoa ouviu. O callback de fim também pode representar erro de decodificação. Há atraso aproximado de 500 ms antes do desligamento e prazo de segurança de cinco minutos, que pode interromper áudios mais longos. Remoto desligando encerra e libera o arquivo.

A Central WCA 2.8 recria o `.env` ao salvar configurações, removendo linhas desconhecidas; a Central 2.9 reconhece os dois novos campos. Versões antigas da Central ou reinstalação podem apagá-los ou sobrescrevê-los. A funcionalidade não altera automaticamente o `.env` existente.

## Teste real recomendado

Depois de instalar uma API que contenha este código, ative a opção, verifique o MP3 e ligue de outro WhatsApp sem depender do webhook. Confira se a chamada é atendida, se todo o áudio chega e se desliga. Teste também vídeo, grupo, desligamento remoto, MP3 ausente e recusa ativa. Se falhar, consulte logs com `[meowcaller]` e `Automatic audio call`; não compartilhe chaves, números ou dados sensíveis. O erro de pré-chave ausente relatado anteriormente continua possível; a implementação não o corrige. Não refaça pareamento nem apague bancos sem análise e autorização.
