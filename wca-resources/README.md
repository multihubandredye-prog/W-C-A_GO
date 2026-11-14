# WCA Resources

Recursos estáticos do projeto WCA (WhatsApp Client API).

## Estrutura

```
wca-resources/
├── media/
│   ├── images/             # Imagens do sistema (ícones, screenshots, exemplos)
│   ├── videos/             # Vídeos do sistema incluídos no pacote do Tasker
│   ├── videos-tutoriais/   # Tutoriais reproduzidos online pelo Guia / Help
│   └── audio/              # Áudios do sistema
└── updates/                # Atualizações e changelogs
```

A pasta `media/videos-tutoriais/` é consultada dinamicamente pelo Guia / Help e fica excluída do arquivo `WCA-Tasker-Projetos-*.zip`. As demais mídias continuam sendo achatadas em `icons/` dentro do pacote.
