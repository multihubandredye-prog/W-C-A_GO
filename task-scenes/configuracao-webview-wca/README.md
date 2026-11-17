# Configuração do WebView

Cena `010-DOWNLOAD_VIEW [wca]V2` · versão **1.21** · arquivo único `configuracao-webview-wca-1.21.html` (HTML, CSS e JavaScript inline).

## Funcionamento

- Consulta as releases do repositório e seleciona a mais recente que contém o pacote `WCA-Tasker-Projetos-<tag>.zip`. Para completar o processo, essa mesma release deve conter `whatsapp-wca-aarch64.bin` e `ffmpeg-wca-aarch64.bin`.
- Baixa e verifica o ZIP. Extrai `task-scenes/` para `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/web-view` e `icons/` para `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons`, com proteção contra caminhos inseguros. Remove o ZIP após a extração bem-sucedida.
- Baixa os dois binários diretamente dos assets da release para `/data/data/net.dinglisch.android.taskerm/files/WCA`. Cada download é gravado em um arquivo temporário `.update`, verificado quanto ao tamanho e à existência, marcado como executável e só então substitui o arquivo correspondente. O tamanho é comparado com o informado pela release quando disponível.
- Caso a consulta à release, a localização dos dois binários, o download ou a gravação falhe, exibe erro e permite tentar novamente, sem anunciar conclusão ou reiniciar a Central. Arquivos já substituídos não são revertidos se um download posterior falhar.
- Após a conclusão, exibe contagem de 5 a 1, executa `001-CENTRAL_WHATS_CONNECT_API [wca]` e fecha a cena. O fluxo não instala o número nem inicia a API diretamente.
- O download e a extração usam a ponte Tasker e shell sem root; a consulta de releases usa `fetch`. Acesso do telefone deve estar habilitado no WebView.
- O botão voltar usa `cache_modal` e `return_modal`; os modais usam transição de escala e desfoque e o feedback de toque usa `:active`.
