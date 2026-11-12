# task-scenes — Projetos de cena (HTML)

Pasta central de todos os projetos de cena HTML (Scenes V2) do WCA.

Cada projeto fica em uma **subpasta própria**, contendo **todas as versões do HTML desde a 1.0** (nenhuma é removida a cada mudança) e o README do projeto.

| Projeto | Cena do Tasker | Pasta |
|---|---|---|
| Central WCA | `001-CENTRAL_WHATS_CONNECT_API [wca]` | `central-wca/` |

## Convenções

- **Arquivo único por versão**: HTML + CSS + JS inline (sem arquivos externos, sem `%project_path`)
- **Bridge V2**: `Tasker.runTask`, `Tasker.setVariable`, `Tasker.getVariable`, `Tasker.readFile`, `Tasker.writeFile`, `Tasker.listFiles`, `Tasker.deleteFile` — com prazo de 4s e nunca no caminho da inicialização
- **Nenhuma variável do Tasker embutida no JavaScript** — a cena V2 não substitui variáveis dentro de `<script>`; tudo é lido em runtime via `Tasker.getVariable`
- **Versões**: a cada alteração, um novo arquivo `Nome 1.x.html` é adicionado à pasta do projeto
- **README por projeto**: mapa de cliques, variáveis, arquivos e histórico — atualizado a cada versão
