#!/usr/bin/env python3
import urllib.request
import lzma
import re
import os
import shutil
import subprocess

def main():
    # 1. Configurar caminhos (usando pastas locais no workspace)
    workspace = os.getcwd()
    extracted_dir = os.path.join(workspace, "_tmp_termux_extracted")
    bundle_dir = os.path.join(workspace, "_tmp_ffmpeg_bundle")
    bin_dir = os.path.join(bundle_dir, "bin")
    lib_dir = os.path.join(bin_dir, "lib")
    
    # Limpar diretórios anteriores
    for path in [extracted_dir, bundle_dir]:
        if os.path.exists(path):
            shutil.rmtree(path)
    
    os.makedirs(lib_dir, exist_ok=True)
    
    # 2. Baixar o índice Packages do Termux aarch64
    url_packages = "https://packages-cf.termux.dev/apt/termux-main/dists/stable/main/binary-aarch64/Packages"
    print("📥 Baixando índice de pacotes do Termux (aarch64)...")
    req = urllib.request.Request(url_packages, headers={'User-Agent': 'Mozilla/5.0'})
    try:
        with urllib.request.urlopen(req) as response:
            data = response.read()
    except Exception as e:
        print(f"❌ Erro ao baixar índice de pacotes: {e}")
        exit(1)
        
    packages_text = data.decode('utf-8')

    # 3. Fazer o parse do arquivo Packages
    print("🔍 Analisando dependências...")
    packages = {}
    current_pkg = {}
    for line in packages_text.split('\n'):
        if not line.strip():
            if 'Package' in current_pkg:
                packages[current_pkg['Package']] = current_pkg
            current_pkg = {}
            continue
        if ':' in line:
            key, val = line.split(':', 1)
            current_pkg[key.strip()] = val.strip()
    if 'Package' in current_pkg:
        packages[current_pkg['Package']] = current_pkg

    # 4. Resolver dependências recursivamente para o ffmpeg
    to_resolve = ['ffmpeg']
    resolved = set()
    dependencies = []

    # Lista de pacotes que não queremos resolver dependências de build/sistema
    # para evitar pacotes gigantescos como python, perl, etc.
    blacklist = ['bash', 'coreutils', 'grep', 'sed', 'ncurses', 'readline']

    while to_resolve:
        pkg_name = to_resolve.pop(0)
        if pkg_name in resolved or pkg_name in blacklist:
            continue
        resolved.add(pkg_name)
        
        pkg = packages.get(pkg_name)
        if not pkg:
            continue
        
        dependencies.append(pkg)
        
        # Obter e limpar a string de dependências
        depends_str = pkg.get('Depends', '')
        if depends_str:
            # Substitui expressões de versão como (>= 1.2) e divide por vírgula
            deps = [re.sub(r'\(.*?\)', '', d).strip() for d in depends_str.split(',')]
            for dep in deps:
                if '|' in dep:
                    # Alternativas (ex: libssl1.1 | libssl3) - pegar a primeira
                    dep = dep.split('|')[0].strip()
                if dep and dep not in resolved and dep not in blacklist:
                    to_resolve.append(dep)

    print(f"✅ Resolvidos {len(dependencies)} pacotes necessários para o bundle.")

    # 5. Baixar debs e extrair
    os.makedirs(extracted_dir, exist_ok=True)
    
    for pkg in dependencies:
        pkg_name = pkg['Package']
        filename = pkg['Filename']
        url = f"https://packages-cf.termux.dev/apt/termux-main/{filename}"
        deb_path = os.path.join(workspace, f"_tmp_{pkg_name}.deb")
        
        print(f"📥 Baixando {pkg_name}...")
        try:
            req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
            with urllib.request.urlopen(req) as response, open(deb_path, "wb") as out:
                out.write(response.read())
        except Exception as e:
            print(f"⚠️ Erro ao baixar {pkg_name}: {e}. Tentando continuar...")
            continue
            
        print(f"📂 Extraindo {pkg_name}...")
        pkg_extract_dir = os.path.join(extracted_dir, pkg_name)
        os.makedirs(pkg_extract_dir, exist_ok=True)
        subprocess.run(["dpkg-deb", "-x", deb_path, pkg_extract_dir], check=True)
        os.remove(deb_path)
        
        # 6. Copiar arquivos necessários para o diretório do bundle
        # Se for o ffmpeg, copiamos o executável
        if pkg_name == 'ffmpeg':
            ffmpeg_src = os.path.join(pkg_extract_dir, "data/data/com.termux/files/usr/bin/ffmpeg")
            if os.path.exists(ffmpeg_src):
                shutil.copy2(ffmpeg_src, os.path.join(bin_dir, "ffmpeg"))
                print(f"✨ Executável do FFmpeg copiado.")
                
        # Copiar todas as bibliotecas .so do pacote para bin/lib
        usr_lib_dir = os.path.join(pkg_extract_dir, "data/data/com.termux/files/usr/lib")
        if os.path.exists(usr_lib_dir):
            copied_libs = 0
            for item in os.listdir(usr_lib_dir):
                src_item = os.path.join(usr_lib_dir, item)
                # Copiar arquivos .so e desreferenciar links simbólicos (salvando como arquivos físicos)
                if item.endswith(".so") or ".so." in item:
                    if os.path.islink(src_item):
                        real_path = os.path.realpath(src_item)
                        if os.path.exists(real_path) and os.path.isfile(real_path):
                            shutil.copy2(real_path, os.path.join(lib_dir, item))
                            copied_libs += 1
                    else:
                        shutil.copy2(src_item, os.path.join(lib_dir, item))
                        copied_libs += 1
            if copied_libs > 0:
                print(f"✨ Copiadas {copied_libs} bibliotecas de {pkg_name}.")

    # 7. Gerar os arquivos finais para a compilação
    print("📦 Compactando bundle com zstd...")
    tar_cmd = f"tar -cf - -C {bundle_dir} bin | zstd -19 -T0 > {os.path.join(workspace, 'res.tar.zst')}"
    subprocess.run(tar_cmd, shell=True, check=True)
    
    with open(os.path.join(workspace, "available_commands.txt"), "w") as f:
        f.write("ffmpeg,")
        
    # Limpar pastas temporárias locais
    for path in [extracted_dir, bundle_dir]:
        if os.path.exists(path):
            shutil.rmtree(path)
            
    print("🎉 Geração de res.tar.zst e available_commands.txt concluída com sucesso!")

if __name__ == "__main__":
    main()
