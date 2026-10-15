# Proxy WCA (opcional)

Proxy que redireciona o webhook da API para o Tasker: escuta em `127.0.0.1:8081` (rota `/Tasker`), enfileira os eventos e reenvia para `127.0.0.1:3129` no formato que o receiver do Tasker entende (`recebido_em` / `origem` / `metodo` / `url_original` / `headers` / `body`), com espaçamento de 500 ms entre entregas.

> A API atual já envia o webhook diretamente no formato do Tasker, então o proxy **não é mais necessário no fluxo padrão** (e o binário não é mais gerado nos releases). Este projeto fica mantido aqui para caso seja necessário utilizá-lo no futuro.

## Compilar manualmente (Android aarch64)

```bash
GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o go_port_wca-aarch64.bin proxy/proxy.go
```

## 📂 Instalação (Proxy WCA)

Coloque o binário `go_port_wca-aarch64.bin` em: `/data/data/net.dinglisch.android.taskerm/files/WCA/`

## ⚙️ Executar no Tasker (Java Code)

```java
Runtime rt = Runtime.getRuntime(); String[] cmd = new String[]{"/system/bin/sh","-c","cd /data/data/net.dinglisch.android.taskerm/files/WCA && nohup /system/bin/linker64 /data/data/net.dinglisch.android.taskerm/files/WCA/go_port_wca-aarch64.bin </dev/null >/dev/null 2>&1 &"}; rt.exec(cmd);
```

## 🔍 Verificar Proxy (Run Shell)

```bash
sleep 0.2
pgrep -f go_port_wca-aarch64 >/dev/null && echo "proxy running" || echo "proxy not running"
```

## 🛑 Parar Proxy (Run Shell)

```bash
pid=$(ps -e -o PID,ARGS | grep "go_port_wca-aarch64" | grep -v grep | awk '{print $1}') && [ -n "$pid" ] && kill -9 $pid || echo "Process terminated"
```
