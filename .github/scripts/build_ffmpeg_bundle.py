#!/usr/bin/env python3
"""
build_ffmpeg_bundle.py
======================
Baixa o ffmpeg e todas as suas dependências do Termux (aarch64),
empacota em res.tar.zst para uso com o wrapper run-android-executable.

Estratégia:
  1. Baixar o índice de pacotes do Termux (com fallback .gz / sem extensão)
  2. Resolver dependências recursivamente (ffmpeg + todas as deps)
  3. Baixar e extrair todos os .deb
  4. Usar readelf para descobrir as SONAME exatas que o binário ffmpeg precisa
  5. Construir mapa nome→arquivo real (desreferenciando symlinks)
  6. Copiar ffmpeg + TODAS as .so necessárias como arquivos físicos (sem symlinks)
  7. Gerar res.tar.zst e available_commands.txt
"""

import urllib.request
import gzip
import lzma
import re
import os
import shutil
import subprocess
import sys


def download_with_fallback(urls_with_compression):
    """Tenta cada (url, compressão) em ordem; retorna bytes descomprimidos."""
    for url, compression in urls_with_compression:
        req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
        try:
            with urllib.request.urlopen(req, timeout=60) as resp:
                raw = resp.read()
            if compression == "gz":
                return gzip.decompress(raw)
            elif compression == "xz":
                return lzma.decompress(raw)
            else:
                return raw
        except Exception as e:
            print(f"  ⚠️  Falhou ({url}): {e}")
    return None


def parse_packages(text):
    packages = {}
    current = {}
    for line in text.split("\n"):
        if not line.strip():
            if "Package" in current:
                packages[current["Package"]] = current
            current = {}
            continue
        if ":" in line:
            k, v = line.split(":", 1)
            current[k.strip()] = v.strip()
    if "Package" in current:
        packages[current["Package"]] = current
    return packages


def resolve_deps(packages, roots, blacklist):
    queue = list(roots)
    visited = set()
    result = []
    while queue:
        name = queue.pop(0)
        if name in visited or name in blacklist:
            continue
        visited.add(name)
        pkg = packages.get(name)
        if not pkg:
            print(f"  ⚠️  Pacote não encontrado: {name}")
            continue
        result.append(pkg)
        deps_str = pkg.get("Depends", "")
        if deps_str:
            for dep in deps_str.split(","):
                dep = re.sub(r"\(.*?\)", "", dep).strip()
                if "|" in dep:
                    dep = dep.split("|")[0].strip()
                if dep:
                    queue.append(dep)
    return result


def get_sonames_needed(binary_path):
    needed = set()
    try:
        out = subprocess.check_output(
            ["readelf", "-d", binary_path], stderr=subprocess.DEVNULL, text=True
        )
        for line in out.splitlines():
            if "NEEDED" in line:
                m = re.search(r"\[(.+?)\]", line)
                if m:
                    needed.add(m.group(1))
    except Exception as e:
        print(f"  ⚠️  readelf falhou em {binary_path}: {e}")
    return needed


def build_somap(lib_dirs):
    """Constrói mapa soname→caminho_físico_real a partir dos diretórios extraídos."""
    somap = {}
    for lib_dir in lib_dirs:
        if not os.path.isdir(lib_dir):
            continue
        for root, _, files in os.walk(lib_dir):
            for fname in files:
                fpath = os.path.join(root, fname)
                real = os.path.realpath(fpath)
                if os.path.isfile(real) and ".so" in fname:
                    somap.setdefault(fname, real)
    return somap


def collect_all_sonames(binary_path, somap):
    """Resolve recursivamente TODAS as sonames necessárias (transitivas)."""
    all_needed = set()
    queue = get_sonames_needed(binary_path)
    visited = set()
    while queue:
        soname = queue.pop()
        if soname in visited:
            continue
        visited.add(soname)
        all_needed.add(soname)
        real_path = somap.get(soname)
        if real_path and os.path.isfile(real_path):
            for t in get_sonames_needed(real_path):
                if t not in visited:
                    queue.add(t)
        else:
            print(f"  ⚠️  SONAME não mapeado: {soname}")
    return all_needed


def main():
    workspace = os.getcwd()
    extracted_dir = os.path.join(workspace, "_tmp_termux_extracted")
    bundle_dir    = os.path.join(workspace, "_tmp_ffmpeg_bundle")
    bin_dir       = os.path.join(bundle_dir, "bin")
    lib_dir       = os.path.join(bin_dir, "lib")

    for path in [extracted_dir, bundle_dir]:
        if os.path.exists(path):
            shutil.rmtree(path)
    os.makedirs(lib_dir, exist_ok=True)

    # 1. Índice de pacotes
    BASE = "https://packages-cf.termux.dev/apt/termux-main/dists/stable/main/binary-aarch64/Packages"
    print("📥 Baixando índice de pacotes do Termux (aarch64)...")
    data = download_with_fallback([
        (BASE,          None),
        (BASE + ".gz",  "gz"),
        (BASE + ".xz",  "xz"),
    ])
    if data is None:
        print("❌ Falha ao baixar índice de pacotes.")
        sys.exit(1)
    print("✅ Índice baixado.")

    packages = parse_packages(data.decode("utf-8"))

    # 2. Resolver dependências
    BLACKLIST = {"bash", "coreutils", "grep", "sed", "ncurses", "readline", "sh"}
    print("🔍 Resolvendo dependências de ffmpeg...")
    deps = resolve_deps(packages, ["ffmpeg"], BLACKLIST)
    print(f"  → {len(deps)} pacotes.")

    # 3. Baixar e extrair .deb
    os.makedirs(extracted_dir, exist_ok=True)
    pkg_lib_dirs = []
    ffmpeg_bin = None

    for pkg in deps:
        pkg_name = pkg["Package"]
        filename  = pkg.get("Filename", "")
        if not filename:
            continue
        url = f"https://packages-cf.termux.dev/apt/termux-main/{filename}"
        deb_path = os.path.join(workspace, f"_dl_{pkg_name}.deb")

        print(f"📦 Baixando {pkg_name}...")
        try:
            req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
            with urllib.request.urlopen(req, timeout=120) as resp, open(deb_path, "wb") as out:
                out.write(resp.read())
        except Exception as e:
            print(f"  ⚠️  Erro ao baixar {pkg_name}: {e}")
            continue

        pkg_dir = os.path.join(extracted_dir, pkg_name)
        os.makedirs(pkg_dir, exist_ok=True)
        subprocess.run(["dpkg-deb", "-x", deb_path, pkg_dir], check=True, capture_output=True)
        os.remove(deb_path)

        lib_search = os.path.join(pkg_dir, "data", "data", "com.termux", "files", "usr", "lib")
        if os.path.isdir(lib_search):
            pkg_lib_dirs.append(lib_search)

        if pkg_name == "ffmpeg":
            candidate = os.path.join(pkg_dir, "data", "data", "com.termux", "files", "usr", "bin", "ffmpeg")
            if os.path.isfile(candidate) or os.path.islink(candidate):
                ffmpeg_bin = candidate
                print(f"  ✨ ffmpeg encontrado: {ffmpeg_bin}")

    if ffmpeg_bin is None:
        print("❌ Binário ffmpeg não encontrado.")
        sys.exit(1)

    # 4. Construir mapa soname→real e descobrir todas as sonames necessárias
    print("🔬 Construindo mapa de bibliotecas...")
    somap = build_somap(pkg_lib_dirs)
    print(f"  → {len(somap)} arquivos .so mapeados.")

    # O ffmpeg pode ser um symlink; resolve para uso no readelf
    real_ffmpeg = os.path.realpath(ffmpeg_bin)

    print("🔬 Analisando dependências de runtime via readelf...")
    all_needed = collect_all_sonames(real_ffmpeg, somap)
    print(f"  → {len(all_needed)} sonames necessárias:")
    for s in sorted(all_needed):
        found_mark = "✅" if s in somap else "❌"
        print(f"     {found_mark} {s}")

    # 5. Copiar ffmpeg como arquivo físico
    dst_ffmpeg = os.path.join(bin_dir, "ffmpeg")
    shutil.copy2(real_ffmpeg, dst_ffmpeg)
    os.chmod(dst_ffmpeg, 0o755)
    print(f"\n✅ ffmpeg copiado ({os.path.getsize(dst_ffmpeg)//1024} KB)")

    # 6. Copiar todas as .so necessárias como arquivos físicos
    missing = []
    for soname in sorted(all_needed):
        dst = os.path.join(lib_dir, soname)
        real_path = somap.get(soname)

        if real_path and os.path.isfile(real_path):
            shutil.copy2(real_path, dst)
            os.chmod(dst, 0o755)
        else:
            # fuzzy: tenta encontrar arquivo que começa com o mesmo prefixo
            base_prefix = soname.split(".so")[0] + ".so"
            found = None
            for k, v in somap.items():
                if k.startswith(base_prefix) and os.path.isfile(v):
                    found = (k, v)
                    break
            if found:
                shutil.copy2(found[1], dst)
                os.chmod(dst, 0o755)
                print(f"  🔗 {soname} ← fuzzy match: {found[0]}")
            else:
                print(f"  ❌ AUSENTE: {soname}")
                missing.append(soname)

    # 7. Gerar aliases de versão major
    #    Ex: libavdevice.so.62.3.100 → também cria libavdevice.so.62
    print("🔗 Gerando aliases de versão...")
    for fname in list(os.listdir(lib_dir)):
        fpath = os.path.join(lib_dir, fname)
        if not os.path.isfile(fpath):
            continue
        parts = fname.split(".so.")
        if len(parts) == 2:
            sub = parts[1].split(".")
            if len(sub) > 1:
                # alias só com major (ex: libfoo.so.6)
                alias = parts[0] + ".so." + sub[0]
                alias_path = os.path.join(lib_dir, alias)
                if not os.path.exists(alias_path):
                    shutil.copy2(fpath, alias_path)
                    print(f"     alias: {alias} ← {fname}")

    # 8. Empacotar
    if missing:
        print(f"\n⚠️  {len(missing)} sonames ausentes: {missing}")
    else:
        print("\n✅ Todas as dependências resolvidas!")

    print("📦 Gerando res.tar.zst ...")
    tar_cmd = f'tar -cf - -C "{bundle_dir}" bin | zstd -19 -T0 > "{os.path.join(workspace, "res.tar.zst")}"'
    subprocess.run(tar_cmd, shell=True, check=True)

    with open(os.path.join(workspace, "available_commands.txt"), "w") as f:
        f.write("ffmpeg,")

    for path in [extracted_dir, bundle_dir]:
        if os.path.exists(path):
            shutil.rmtree(path)

    out_size = os.path.getsize(os.path.join(workspace, "res.tar.zst")) // 1024
    print(f"\n🎉 res.tar.zst gerado ({out_size} KB) — pronto!")


if __name__ == "__main__":
    main()
