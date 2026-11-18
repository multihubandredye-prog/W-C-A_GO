# Configuração do WebView

Cena `010-DOWNLOAD_VIEW [wca]V2` · versão **1.23** · arquivo único `configuracao-webview-wca-1.23.html` (HTML, CSS e JavaScript inline).

## Funcionamento

- Consulta as releases e seleciona a mais recente com `WCA-Tasker-Projetos-<tag>.zip`. Se a consulta falhar, utiliza o pacote padrão configurado na cena.
- Antes do download verifica, via shell sem root, a existência dos arquivos `whatsapp-wca-aarch64.bin` e `ffmpeg-wca-aarch64.bin` em `/data/data/net.dinglisch.android.taskerm/files/WCA`. Arquivos existentes são preservados e não são baixados novamente, independentemente da versão ou do tamanho.
- Quando um ou ambos faltam, procura apenas os assets necessários na mesma release do ZIP. Se a consulta ou um asset necessário não estiver disponível, informa o erro e permite nova tentativa. Uma falha na verificação não é confundida com arquivo ausente.
- Baixa e valida o ZIP, extrai `task-scenes/` para `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/web-view` e `icons/` para `/storage/emulated/0/Tasker/projects_task/projects/whats-connect-api/icons`, com proteção contra caminhos inseguros. Remove o ZIP após extração bem-sucedida.
- Verifica novamente a presença dos binários e baixa somente os que ainda faltarem. Cada download usa arquivo temporário `.update`, é verificado como não vazio e comparado ao tamanho informado pela release quando disponível; só então recebe permissão de execução e é colocado na pasta WCA. Uma falha impede a mensagem de conclusão e o reinício, permitindo nova tentativa.
- Com os arquivos já presentes, o processo de ZIP prossegue sem solicitar os binários. A verificação é apenas de existência: não detecta corrupção, arquivo vazio ou versão antiga dos binários existentes.
- Após a conclusão, exibe contagem de 5 a 1, executa `001-CENTRAL_WHATS_CONNECT_API [wca]` e fecha a cena. Não instala o número nem inicia a API diretamente.
- Download e extração usam a ponte Tasker; a consulta de releases usa `fetch`. O acesso do telefone deve estar habilitado no WebView. O botão voltar usa `cache_modal` e `return_modal`; os modais usam transição de escala e desfoque e o feedback de toque usa `:active`.

O campo **Total a baixar** soma o ZIP e os binários ausentes com tamanho informado pela mesma release. Se não for possível verificar arquivos ou obter todos os tamanhos, exibe **Indisponível**. A verificação é refeita ao iniciar o download.
