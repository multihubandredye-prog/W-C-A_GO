#!/usr/bin/env python3
# Monta o pacote WCA-Tasker-Projetos com a ULTIMA versao (.html) de cada
# projeto de cena em task-scenes/*/ — o pacote da release leva SOMENTE as
# cenas; os projetos .prj.xml de task-project/ NAO entram no zip.
# Roda no GitHub Actions antes da criacao da release; o zip sai na raiz do repo.
# A deteccao e dinamica: novas pastas/projetos entram no pacote automaticamente.
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


def ultimas(arquivos):
    # Agrupa por pasta de projeto e mantem a versao mais alta de cada uma.
    grupos = {}
    for f in arquivos:
        grupos[f.parent.name] = escolher(grupos.get(f.parent.name), (f, versao_de(f.stem)))
    return grupos


def main():
    raiz = Path.cwd()
    cenas_raiz = raiz / 'task-scenes'
    if not cenas_raiz.is_dir():
        print('ERRO: pasta task-scenes nao encontrada')
        sys.exit(1)

    cenas = sorted(cenas_raiz.glob('*/*.html'))
    escolhidas = ultimas(cenas)
    if not escolhidas:
        print('ERRO: nenhuma cena encontrada em task-scenes/*/')
        sys.exit(1)

    tag = os.environ.get('RELEASE_TAG', '').strip().replace('/', '-')
    nome_zip = f'WCA-Tasker-Projetos-{tag}.zip' if tag else 'WCA-Tasker-Projetos.zip'

    print('Cenas incluidas (ultima versao por projeto):')
    for pasta in sorted(escolhidas):
        f, _ = escolhidas[pasta]
        print(f'  task-scenes/{pasta}/{f.name}')

    # ZIP sem compressao interna (ZIP_STORED): cada entrada e um recorte
    # direto dos bytes, o que permite a cena 010-DOWNLOAD_VIEW descompactar
    # o pacote por conta propria, sem depender de recursos de descompressao
    # do WebView do aparelho
    with zipfile.ZipFile(nome_zip, 'w', zipfile.ZIP_STORED) as z:
        for pasta in sorted(escolhidas):
            f, _ = escolhidas[pasta]
            z.write(f, f'task-scenes/{pasta}/{f.name}')

    print(f'OK: {nome_zip} criado com {len(escolhidas)} cena(s).')


if __name__ == '__main__':
    main()
