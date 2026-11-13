#!/usr/bin/env python3
# Monta o pacote WCA-Tasker-Projetos com:
#   - task-scenes/*/ -> a ULTIMA versao (.html) de cada projeto de cena
#   - task-project/  -> a ULTIMA versao de cada projeto do Tasker
# Roda no GitHub Actions antes da criacao da release; o zip sai na raiz do repo.
# A deteccao e dinamica: novos projetos/pastas entram no pacote automaticamente.
import os
import re
import sys
import zipfile
from pathlib import Path


def versao_de(stem):
    # Extrai a versao no fim do nome (ex.: "Central WCA 1.40" -> (1, 40)).
    m = re.search(r'(\d+(?:\.\d+)*)\s*$', stem)
    if not m:
        return None
    return tuple(int(p) for p in m.group(1).split('.'))


def chave_sem_versao(nome_arquivo):
    # Remove a versao do fim do nome mantendo a extensao completa
    # (ex.: "W_C_A_API 1.2.prj.xml" -> "W_C_A_API.prj.xml").
    return re.sub(r'\s*\d+(?:\.\d+)*\s*(?=\.)', '', nome_arquivo)


def escolher(atual, cand):
    # Mantém a versao mais alta; sem versao em nenhum dos dois, fica o primeiro.
    if atual is None:
        return cand
    va, vc = atual[1], cand[1]
    if va is None and vc is None:
        return atual
    if va is None:
        return cand
    if vc is None:
        return atual
    return cand if vc > va else atual


def ultimas(arquivos, por_pasta):
    grupos = {}
    for f in arquivos:
        chave = f.parent.name if por_pasta else chave_sem_versao(f.name)
        grupos[chave] = escolher(grupos.get(chave), (f, versao_de(f.stem)))
    return grupos


def main():
    raiz = Path.cwd()
    cenas_raiz = raiz / 'task-scenes'
    projetos_raiz = raiz / 'task-project'
    if not cenas_raiz.is_dir():
        print('ERRO: pasta task-scenes nao encontrada')
        sys.exit(1)

    cenas = sorted(cenas_raiz.glob('*/*.html'))
    escolhidas = ultimas(cenas, por_pasta=True)
    if not escolhidas:
        print('ERRO: nenhuma cena encontrada em task-scenes/*/')
        sys.exit(1)

    projetos = sorted(p for p in projetos_raiz.glob('*') if p.is_file()) if projetos_raiz.is_dir() else []
    escolhidos = ultimas(projetos, por_pasta=False)

    tag = os.environ.get('RELEASE_TAG', '').strip().replace('/', '-')
    nome_zip = f'WCA-Tasker-Projetos-{tag}.zip' if tag else 'WCA-Tasker-Projetos.zip'

    print('Cenas incluidas (ultima versao por projeto):')
    for pasta in sorted(escolhidas):
        f, v = escolhidas[pasta]
        print(f'  task-scenes/{pasta}/{f.name}')
    print('Projetos do Tasker incluidos:')
    for chave in sorted(escolhidos):
        f, v = escolhidos[chave]
        print(f'  task-project/{f.name}')

    # ZIP sem compressao interna (ZIP_STORED): cada entrada e um recorte
    # direto dos bytes, o que permite a cena 010-DOWNLOAD_VIEW descompactar
    # o pacote por conta propria, sem depender de recursos de descompressao
    # do WebView do aparelho
    with zipfile.ZipFile(nome_zip, 'w', zipfile.ZIP_STORED) as z:
        for pasta in sorted(escolhidas):
            f, _ = escolhidas[pasta]
            z.write(f, f'task-scenes/{pasta}/{f.name}')
        for chave in sorted(escolhidos):
            f, _ = escolhidos[chave]
            z.write(f, f'task-project/{f.name}')

    total = len(escolhidas) + len(escolhidos)
    print(f'OK: {nome_zip} criado com {total} arquivo(s) '
          f'({len(escolhidas)} cena(s) + {len(escolhidos)} projeto(s)).')


if __name__ == '__main__':
    main()
