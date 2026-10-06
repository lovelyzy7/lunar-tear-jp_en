# 多平台构建指南（Windows / Linux / macOS）

本项目使用 [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite)（**纯 Go** 的 SQLite 实现），
因此**不需要 CGO**，可直接交叉编译到其它平台 / 架构。

## 1. 依赖

| 用途 | 依赖 |
|---|---|
| 编译 | Go **1.25+**（`server/go.mod` 声明 1.25.8） |
| 数据库迁移 | [goose](https://github.com/pressly/goose)：`go install github.com/pressly/goose/v3/cmd/goose@latest` |
| 重新生成 proto（可选） | `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc`（`make proto`） |
| 运行期 | `server/assets/` 资源目录、`db/` 数据库文件 |

## 2. 目标矩阵

| 目标平台 | GOOS | GOARCH | 附加 |
|---|---|---|---|
| Windows x64 | `windows` | `amd64` | 产物带 `.exe` |
| Windows ARM64 | `windows` | `arm64` | 产物带 `.exe` |
| Linux x64 | `linux` | `amd64` | |
| Linux ARM64 | `linux` | `arm64` | 常见于云主机 / 树莓派 4+ |
| Linux ARMv7 | `linux` | `arm` | 需 `GOARM=7`（树莓派 2/3） |
| macOS ARM64 | `darwin` | `arm64` | Apple Silicon |
| macOS x64 | `darwin` | `amd64` | Intel |

需要构建的可执行文件（`server/cmd/<name>`）：`lunar-tear`、`octo-cdn`、`auth-server`、
`wizard`、`wizard-restore`（可选：`dev`、`register-account`、`claim-account`、`import-snapshot`）。

## 3. 一键交叉编译（Linux / macOS / WSL / Git Bash）

```bash
cd server
BINS="lunar-tear octo-cdn auth-server wizard wizard-restore"
for target in windows/amd64 windows/arm64 linux/amd64 linux/arm64 linux/arm darwin/arm64 darwin/amd64; do
  os=${target%/*}; arch=${target#*/}
  ext=""; [ "$os" = "windows" ] && ext=".exe"
  out="dist/${os}-${arch}"
  mkdir -p "$out"
  for c in $BINS; do
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" GOARM=7 \
      go build -trimpath -ldflags "-s -w" -o "$out/$c$ext" "./cmd/$c" || exit 1
  done
  (cd dist && tar -czf "lunar-tear-server-${os}-${arch}.tar.gz" "${os}-${arch}")
  echo "==> dist/lunar-tear-server-${os}-${arch}.tar.gz"
done
```

> Windows 上用 Git Bash / WSL 可直接跑上面的脚本；纯 PowerShell 见第 4 节。

## 4. 一键交叉编译（Windows PowerShell）

```powershell
cd server
$bins = 'lunar-tear','octo-cdn','auth-server','wizard','wizard-restore'
$targets = @(@('windows','amd64'), @('windows','arm64'), @('linux','amd64'), @('linux','arm64'), @('darwin','arm64'))
foreach ($t in $targets) {
  $os, $arch = $t
  $ext = if ($os -eq 'windows') { '.exe' } else { '' }
  $out = "dist\$os-$arch"
  New-Item -ItemType Directory -Force $out | Out-Null
  $env:CGO_ENABLED = '0'; $env:GOOS = $os; $env:GOARCH = $arch
  foreach ($c in $bins) {
    go build -trimpath -ldflags '-s -w' -o "$out\$c$ext" ".\cmd\$c"
    if ($LASTEXITCODE -ne 0) { throw "build failed: $c ($os/$arch)" }
  }
  tar -czf "dist\lunar-tear-server-$os-$arch.tar.gz" -C dist "$os-$arch"
  Write-Host "==> dist\lunar-tear-server-$os-$arch.tar.gz"
}
```

## 5. Makefile（本机平台）

```bash
cd server
make build          # lunar-tear
make build-cdn      # octo-cdn
make build-auth     # auth-server
make build-all      # bin/dev、bin/auth-server、bin/octo-cdn、bin/lunar-tear
make clean          # 清理 bin/
make migrate        # goose 应用数据库迁移（db/game.db）
make proto          # 重新生成 gen/proto/
```

> Makefile 会自动在 Windows 上为产物加 `.exe`。

## 6. 单目标手工命令（等价写法）

```bash
CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o dist/lunar-tear-linux-arm64          ./cmd/lunar-tear
CGO_ENABLED=0 GOOS=linux   GOARCH=arm   GOARM=7 go build -trimpath -ldflags "-s -w" -o dist/lunar-tear-linux-armv7   ./cmd/lunar-tear
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o dist/lunar-tear.exe                  ./cmd/lunar-tear
```

## 7. 运行

```bash
# 1) 首次初始化（生成配置 / 数据库）
./wizard

# 2) 手工启动（示例）
./lunar-tear  --listen 0.0.0.0:8003 --public-addr <你的IP>:8003 --db db/game.db \
              --octo-url http://<你的IP>:8080 --auth-url http://0.0.0.0:3000
./auth-server --listen 0.0.0.0:3000 --db db/auth.db --game-db db/game.db
./octo-cdn    --listen 0.0.0.0:8080 --public-addr <你的IP>:8080
```

## 8. Docker（多架构）

```bash
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile      -t lunar-tear:latest       .
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile.auth -t lunar-tear-auth:latest  .
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile.cdn  -t lunar-tear-cdn:latest   .
```

## 9. 常见问题

| 现象 | 原因 / 解决 |
|---|---|
| `cgo: C compiler not found` | 本项目**不需要** CGO：设置 `CGO_ENABLED=0` 后再构建 |
| ARMv7 运行报非法指令 | 必须带 `GOARM=7`（默认可能是 5/6） |
| 交叉编译产物启动报缺资源 | 产物不含 `assets/`、`db/`，需与二进制一起分发 |
| `goose: command not found` | `go install github.com/pressly/goose/v3/cmd/goose@latest` 并确保 `$GOPATH/bin` 在 PATH |
| 客户端提示数据过期 | 确认 `internal/service/data.go` 的 mtime 版本逻辑生效（本 fork 已修复） |
| 本地平台上一键全量 | `make build-all`（输出到 `server/bin/`） |

### 官方流水线对照

`.github/workflows/release.yml` 使用 **Go 1.25 矩阵**构建（Ubuntu runner）：

| goos | goarch | 归档 |
|---|---|---|
| linux | amd64 | tar.gz |
| linux | arm64 | tar.gz |
| darwin | amd64 | tar.gz |
| darwin | arm64 | tar.gz |
| windows | amd64 | zip |

本指南在此之上额外给出 **windows/arm64** 与 **linux/arm (GOARM=7)** 两个目标；
归档命名沿用 `lunar-tear-server-<os>-<arch>.tar.gz`（Windows 用 `.zip`）。
