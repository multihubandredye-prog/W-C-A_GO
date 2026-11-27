# Atendimento automático de chamadas de áudio — experimental

A API pode tentar atender automaticamente chamadas **individuais de áudio**, reproduzir um MP3 e desligar ao terminar. Não usa Tasker, webhook ou `/call/answer` para iniciar o atendimento. A configuração vale para **todos os dispositivos conectados à mesma API e todos os números**; não há filtros, agenda, grupos nem vídeo nesta versão. O webhook existente continua funcionando. A rota manual de atendimento foi removida; `/send/call` permanece disponível para originar chamadas.

No `.env` da API (no aparelho), configure:

```env
WHATSAPP_AUTO_REJECT_CALL=false
WHATSAPP_AUTO_ANSWER_CALL=true
WHATSAPP_AUTO_ANSWER_AUDIO=/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call/2026_11_26_21_14_59.mp3
```

O atendimento é **desativado por padrão**. Substitua o caminho se necessário. O arquivo MP3 deve existir dentro de `/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call/`, ser legível pelo processo da API e ter no máximo 20.000.000 bytes. Na Central 2.25, informe o caminho absoluto de um MP3 acessível ao Tasker: ao salvar, a Central utiliza o shell do Tasker para copiar os bytes para `auto-call.mp3` na pasta privada e o `.env` recebe esse caminho. A cópia depende das permissões do Tasker no Android. A cópia não modifica o original. Se a cópia falhar, a Central não grava o `.env` nem reinicia a API. Se a gravação posterior do `.env` falhar, a cópia pode já ter sido atualizada. Caminhos fora dela, inclusive links simbólicos externos, são rejeitados. URLs e upload não são aceitos. Reinicie a API para aplicar as mudanças.

Se `WHATSAPP_AUTO_REJECT_CALL=true`, a recusa tem prioridade e a API não inicia atendimento automático. Se o MP3 estiver inválido ou indisponível, não atende. A API tenta atender somente após a biblioteca registrar a chamada e a oferta ser identificada como individual e de áudio; falha de chave ou ausência da oferta impossibilita o atendimento. O atendimento não migra dados nem chaves de uma sessão existente.

O player é conectado após a solicitação de atendimento, sem esperar `OnReady` ou áudio do outro lado. Isso acompanha a ordem básica do exemplo da biblioteca, mas **não garante** que o início do MP3 seja audível: o estado de mídia e os sinais do WhatsApp precisam ser validados em ligação real. A biblioteca pode iniciar mídia antes da resposta, perder sinais durante a espera ou falhar ao conectar. `Answer` não comprova que o telefone foi atendido ou que a pessoa ouviu. O callback de fim também pode representar erro de decodificação. Há atraso aproximado de 500 ms antes do desligamento e prazo de segurança de cinco minutos, que pode interromper áudios mais longos. Remoto desligando encerra e libera o arquivo.

A Central WCA 2.11 preserva o estado de `WCA_EXPERIMENTAL_PRIMARY_SIGNAL_STORE` no `.env` existente e cria novos arquivos com essa opção ativada. Versões antigas podem removê-la. Essa opção usa o banco principal persistente para o estado Signal; não migra chaves nem sessões do banco secundário existente. Versões antigas da Central ou reinstalação podem apagá-los ou sobrescrevê-los. A funcionalidade não altera automaticamente o `.env` existente.

## Teste real recomendado

Depois de instalar uma API que contenha este código, ative a opção, verifique o MP3 e ligue de outro WhatsApp sem depender do webhook. Confira se a chamada é atendida, se todo o áudio chega e se desliga. Teste também vídeo, grupo, desligamento remoto, MP3 ausente e recusa ativa. Se falhar, consulte logs com `[meowcaller]` e `Automatic audio call`; não compartilhe chaves, números ou dados sensíveis. O uso do banco principal persistente ajudou em testes reais após reinício, mas não garante a recuperação de mensagens ou sessões existentes. Não refaça pareamento nem apague bancos sem análise e autorização.

## Regras por WhatsApp — Central 2.25

Cada WhatsApp conectado pode selecionar todos, somente selecionados ou ignorar selecionados. Contatos individuais vêm da sessão selecionada; números que não puderem ser resolvidos não são listados e podem ser adicionados manualmente com código do país. Se um filtro estiver ativo e o número da chamada não puder ser identificado, a API não atende automaticamente. Chamadas de grupos e vídeos continuam fora do atendimento. A recusa automática continua tendo prioridade.

Dias e horas são opcionais e individuais por WhatsApp, com fuso IANA explícito; início inclusivo e fim exclusivo. Em horário que atravessa a meia-noite, o dia escolhido é o dia de início. Fora do período, a API não atende nem recusa por causa dessa regra.

As regras ficam em `/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call/auto-call-policies.json`, com marcador ao lado. São dados sensíveis. Antes de configurar regras, mantém-se o comportamento anterior de atender todos. Após configurar, arquivo ausente com marcador ou arquivo inválido impede atendimento automático; não apague nem edite os arquivos manualmente sem backup e análise. Na Central 2.25, abra **Gerenciar caixa postal** nas configurações da API. O WhatsApp principal conectado é selecionado automaticamente; se não estiver conectado, é selecionado o primeiro conectado. A escolha pode ser alterada antes de salvar na janela dedicada. Em Permitidos ou Ignorados, a tela exibe até dois contatos da lista atual; **Exibir todos** abre a lista completa e permite marcar e remover vários. Para aplicar as remoções é necessário salvar o atendimento automático. Cada WhatsApp tem listas independentes de permitidos e ignorados e um único modo ativo: somente a lista do modo escolhido é aplicada. No modo Todos, ambas permanecem guardadas sem aplicação. Ao carregar configuração antiga, a lista antiga é atribuída ao modo filtrado salvo; se o modo antigo era Todos, os números ficam não classificados até o usuário escolher a lista. A Central 2.25 requer uma API que suporte `lists_version: 1` e não grava listas em versões antigas. Após atualizar e salvar listas separadas, evite voltar a uma API antiga, que pode descartar os campos novos ao gravar. A operação grava as regras do dispositivo e depois atualiza apenas ativação e caminho do áudio no `.env`; não é uma transação única, portanto uma falha posterior pode deixar a cópia do áudio ou regras já alteradas. O caminho do MP3 e a ativação global continuam no `.env`.

## Outros formatos na Central

A partir da Central 2.27, a interface aceita arquivos locais MP3, OPUS, OGG, WAV, M4A, AAC e FLAC de até 20.000.000 bytes. MP3 continua sendo copiado como antes; os demais são convertidos pelo FFmpeg do projeto para o MP3 privado antes de salvar. Falha na conversão não grava o .env. A API de caixa postal continua lendo o MP3 privado; é necessário testar codecs e permissões no Android.

## Tempo para atender por WhatsApp

A política inclui `answer_delay_seconds` de 0 a 10 (padrão 0 = sem espera). A contagem inicia na primeira observação da chamada pela API; após a espera, filtros e horários são reavaliados, e o desligamento remoto cancela a espera. A resposta e a mídia dependem da rede e da biblioteca. A Central 2.28 exige uma API que informe esse campo; clientes antigos que não enviam o campo preservam o valor existente ao salvar. Após instalar uma API nova, evite voltar a uma versão antiga, pois ela pode descartar esse campo ao gravar políticas.
