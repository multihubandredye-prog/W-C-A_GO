# Armazenamento Signal — cuidados de instalação

`WCA_EXPERIMENTAL_PRIMARY_SIGNAL_STORE=true` faz a API usar o banco principal também para as chaves e sessões Signal. O nome foi mantido para compatibilidade com a versão testada. Na ausência da opção, a API mantém o comportamento anterior; o modelo de instalação nova define a opção como `true`.

**Não há migração automática** de pré-chaves, sessões, identidades, chaves de grupos ou outros segredos do banco secundário. Alterar essa opção numa instalação existente pode interromper mensagens, grupos e chamadas. Não apague, sobrescreva, copie nem restaure bancos de uma sessão ativa sem analisar os riscos. Alterações de sessão depois de um backup podem tornar uma restauração antiga inconsistente.

O banco principal deve ser persistente e ficar em área privada com proteção adequada, incluindo arquivos SQLite WAL/SHM e backups. Usar banco em memória não é persistência. A Central 2.11 preserva essa opção ao salvar; versões antigas podem sobrescrevê-la. A reinstalação pode recriar o `.env`, exigindo cautela com sessões existentes.

Testes reais de duas chamadas após reinicialização foram bem-sucedidos em uma instalação do usuário; mensagens individuais, grupos, migração e todas as situações de rede ainda não estão garantidas. Antes de aplicar a contas em produção, avalie o armazenamento atual e teste em ambiente isolado.
