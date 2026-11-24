# Seletor de WhatsApp WCA — Cena V2

Projeto `004-WHATSAPP_SELETOR [wca]V2` · versão **1.4** · HTML único `seletor-de-whatsapp-wca-1.4.html`, com CSS e JavaScript inline.

A cena apresenta as opções **WhatsApp Messenger** e **WhatsApp Business**. Ao selecionar, envia à ponte do Tasker a variável de tela `bottom`, respectivamente `wa-message` ou `wa-business`.

## Estabilidade

- Durante a operação os botões ficam temporariamente bloqueados contra toques duplicados. O progresso e eventuais erros aparecem na própria tela.
- A gravação de `bottom` tem prazo de 5 segundos. Se a ponte estiver indisponível, rejeitar ou demorar, a cena permanece aberta e informa como tentar novamente; não inicia o fechamento por erro de gravação.
- Uma gravação que ultrapassou o prazo ainda pode terminar posteriormente: enquanto estiver pendente, o usuário pode tentar novamente a mesma opção ou reabrir a tela, mas a outra opção fica bloqueada para evitar que uma resposta tardia sobrescreva uma escolha diferente.
- Após a ponte aceitar a gravação, a cena tenta fechar com `dismissSceneV2`, `dismissLayout` e, como alternativa, a tarefa auxiliar `004-FECHAR_WHATSAPP_SELETOR [wca]`, se disponível. As tentativas têm prazo e tratam também rejeições assíncronas. Não usa `cancelScreen`, pois encerra a cena com erro em vez de devolver o resultado.
- Se o fechamento não puder ser confirmado, a cena exibe uma mensagem e permite nova tentativa. A confirmação final do recebimento de `bottom` depende da execução da cena e da tarefa no dispositivo.

## Configuração no Tasker

- Na ação **Cena v2 Mostrar**, use `FullscreenWithResult` ou outro modo **With Result**, ID da Tela `1` e **Dispensar Automaticamente** vazio.
- Habilite o JavaScript bridge e **Allow Phone Access** no WebView. Após o fechamento, leia `%bottom` na tarefa; os valores continuam `wa-message` e `wa-business`.
- A tarefa auxiliar de fechamento é uma alternativa, não um requisito para os métodos de fechamento do bridge.
- A fonte e os ícones externos podem depender de rede.
