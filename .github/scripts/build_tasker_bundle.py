#!/usr/bin/env python3
# Monta o pacote WCA-Tasker-Projetos com a ULTIMA versao (.html) de cada
# projeto de cena em task-scenes/*/ e as MEDIAS de wca-resources/media
# (achatadas na pasta icons/ do zip) — os projetos .prj.xml de task-project/
# NAO entram no zip (a cena 010-DOWNLOAD_VIEW extrai as midias do proprio
# zip para <projeto>/icons; nada e baixado fora do pacote).
# Roda no GitHub Actions antes da criacao da release; o zip sai na raiz do repo.
# A deteccao e dinamica: novas pastas/projetos entram no pacote automaticamente.
#
# MINIFICACAO (opcional): com MINIFICAR=yes/sim/true no ambiente, cada HTML
# das cenas e MINIFICADO antes de entrar no zip (apenas a COPIA do zip — os
# arquivos do repositorio permanecem intactos). O processo: CSS inline via
# rcssmin, JS inline via rjsmin (sem renomear nada), comentarios HTML fora e
# colapso de whitespace. Cada cena minificada e validada (tamanho menor,
# sintaxe JS via node --check quando disponivel, contagem de <script>/<style>
# preservada, zero % e zero console no JS); cena com problema entra ORIGINAL
# no zip (nunca quebra o pacote).
import os
import re
import sys
import shutil
import subprocess
import tempfile
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


def minificar_ativo():
    return os.environ.get('MINIFICAR', '').strip().lower() in ('1', 'true', 'yes', 'y', 'sim', 's')


def minificar_html(conteudo):
    import rjsmin
    import rcssmin

    def css_repl(m):
        return '<style>' + rcssmin.cssmin(m.group(1)) + '</style>'

    def js_repl(m):
        return '<script>' + rjsmin.jsmin(m.group(1)) + '</script>'

    saida = re.sub(r'<style>(.*?)</style>', css_repl, conteudo, flags=re.S)
    saida = re.sub(r'<script>(.*?)</script>', js_repl, saida, flags=re.S)
    saida = re.sub(r'<!--.*?-->', '', saida, flags=re.S)
    linhas = [l.strip() for l in saida.replace('\r\n', '\n').split('\n')]
    return ' '.join(l for l in linhas if l)


def validar_minificado(original, minificado):
    # Devolve lista de problemas (vazia = ok). Qualquer problema faz a cena
    # entrar ORIGINAL no zip.
    problemas = []
    blocos_min = re.findall(r'<script>(.*?)</script>', minificado, flags=re.S)
    blocos_ori = re.findall(r'<script>(.*?)</script>', original, flags=re.S)
    if len(blocos_min) != len(blocos_ori):
        problemas.append('contagem de <script> mudou')
    if minificado.count('<style>') != original.count('<style>'):
        problemas.append('contagem de <style> mudou')
    if len(minificado.encode('utf-8')) >= len(original.encode('utf-8')):
        problemas.append('tamanho nao diminuiu')
    for i, js in enumerate(blocos_min):
        rotulo = 'JS %d' % (i + 1)
        if '%' in js:
            problemas.append('% no ' + rotulo)
        if 'console' in js:
            problemas.append('console no ' + rotulo)
        if shutil.which('node'):
            tmp = tempfile.NamedTemporaryFile('w', suffix='.js', delete=False, encoding='utf-8')
            tmp.write(js)
            tmp.close()
            r = subprocess.run(['node', '--check', tmp.name], capture_output=True)
            os.unlink(tmp.name)
            if r.returncode != 0:
                problemas.append('sintaxe invalida no ' + rotulo)
    return problemas


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

    # Midias (video, imagens, audio): embutidas no zip achatadas em icons/
    # (sem subpastas e sem arquivos ocultos como .gitkeep); a cena
    # 010-DOWNLOAD_VIEW extrai essa pasta para <projeto>/icons no aparelho
    midias_raiz = raiz / 'wca-resources' / 'media'
    midias = sorted(
        f for f in midias_raiz.rglob('*')
        if f.is_file() and not f.name.startswith('.')
    ) if midias_raiz.is_dir() else []

    tag = os.environ.get('RELEASE_TAG', '').strip().replace('/', '-')
    nome_zip = f'WCA-Tasker-Projetos-{tag}.zip' if tag else 'WCA-Tasker-Projetos.zip'

    ativa = minificar_ativo()
    if ativa:
        try:
            import rjsmin  # noqa: F401
            import rcssmin  # noqa: F401
        except ImportError:
            print('ERRO: MINIFICAR ativo mas rjsmin/rcssmin nao instalados (pip install rjsmin rcssmin)')
            sys.exit(1)

    print('Cenas incluidas (ultima versao por projeto):')
    for pasta in sorted(escolhidas):
        f, _ = escolhidas[pasta]
        print(f'  task-scenes/{pasta}/{f.name}')
    if midias:
        print('Midias incluidas (wca-resources/media -> icons/):')
        for f in midias:
            print(f'  icons/{f.name}')
    print('Minificacao dos HTMLs: ' + ('ATIVA' if ativa else 'desativada (MINIFICAR=no)'))

    # ZIP sem compressao interna (ZIP_STORED): cada entrada e um recorte
    # direto dos bytes, o que permite a cena 010-DOWNLOAD_VIEW descompactar
    # o pacote por conta propria, sem depender de recursos de descompressao
    # do WebView do aparelho
    minificadas = 0
    with zipfile.ZipFile(nome_zip, 'w', zipfile.ZIP_STORED) as z:
        for pasta in sorted(escolhidas):
            f, _ = escolhidas[pasta]
            destino = f'task-scenes/{pasta}/{f.name}'
            conteudo = f.read_text(encoding='utf-8')
            if ativa:
                try:
                    reduzido = minificar_html(conteudo)
                    problemas = validar_minificado(conteudo, reduzido)
                    if problemas:
                        print(f'  AVISO: {destino} entra ORIGINAL no zip ({ "; ".join(problemas) })')
                    else:
                        z.writestr(destino, reduzido)
                        minificadas += 1
                        print(f'  minificada: {destino} ({len(conteudo.encode())} -> {len(reduzido.encode())} bytes)')
                        continue
                except Exception as e:
                    print(f'  AVISO: {destino} entra ORIGINAL no zip (erro: {e})')
            z.write(f, destino)
        for f in midias:
            z.write(f, f'icons/{f.name}')

    total_cenas = len(escolhidas)
    resumo_min = f', {minificadas} minificada(s)' if ativa else ''
    print(f'OK: {nome_zip} criado com {total_cenas} cena(s){resumo_min} e '
          f'{len(midias)} midia(s).')


if __name__ == '__main__':
    main()
