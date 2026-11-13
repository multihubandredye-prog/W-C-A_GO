# task-scenes — Projetos de cena (HTML)

Pasta central de todos os projetos de cena HTML (Scenes V2) do WCA.

Cada projeto fica em uma **subpasta própria**, contendo o **HTML da última versão** (o histórico completo de versões fica no ambiente de desenvolvimento; no repositório é mantida apenas a mais recente) e o README do projeto.

| Projeto | Cena do Tasker | Pasta |
|---|---|---|
| Central WCA | `001-CENTRAL_WHATS_CONNECT_API [wca]` | `central-wca/` |
| Instalação WCA | `002-INSTALACAO_PROJETO [wca]` | `instalacao-wca/` |
| Atualizações WCA | `006-ATUALIZACOES [wca]` | `atualizacoes-wca/` |
| Painel de Atividade WCA | `003-PAINEL_ATIVIDADE [wca]` | `painel-atividade-wca/` |
| Seletor de WhatsApp WCA | `004-WHATSAPP_SELETOR [wca]` | `seletor-whatsapp-wca/` |
| Gerar Tarefa de Envio WCA | `005-GERAR_TAREFA_ENVIO [wca]` | `gerar-tarefa-envio-wca/` |
| Contatos Pro WCA | Contatos Pro | `contatos-pro-wca/` |

## Convenções

- **Arquivo único por versão**: HTML + CSS + JS inline (sem arquivos externos, sem `%project_path`)
- **Bridge V2**: `Tasker.runTask`, `Tasker.setVariable`, `Tasker.getVariable`, `Tasker.readFile`, `Tasker.writeFile`, `Tasker.listFiles`, `Tasker.deleteFile` — com prazo de 4s e nunca no caminho da inicialização
- **Nenhuma variável do Tasker embutida no JavaScript** — a cena V2 não substitui variáveis dentro de `<script>`; tudo é lido em runtime via `Tasker.getVariable`
- **Versões**: a cada alteração, um novo arquivo é criado no padrão `nome-do-projeto-1.x.html` (minúsculas com hífen); no repositório permanece apenas a versão mais recente de cada projeto
- **README por projeto**: descrição do funcionamento atual (mapa de cliques, variáveis, arquivos) — atualizado a cada alteração
