# image-cli

跨平台（Windows / macOS / Linux）图片处理 CLI，面向**游戏资源优化**与**图片后期**常见流程：格式转换、压缩、通道提取、调色板、NPOT/POT、精灵表切分、EXIF 信息等。

**游戏向默认策略：UI / 2D 不生成 mipmap**；3D 贴图可用 `game prepare` 按场景 profile 处理，**移动端默认 ASTC**（需 `astcenc`），**PC 端 BC7/BC5/BC6H**（需 `texconv`）。详见 [docs/GAME.md](docs/GAME.md)。

**多平台尺寸检测**：`check` / `info -budget` 覆盖 Windows、macOS、iOS、Android、WebGL、小游戏、Switch 等预算，见 [docs/BUDGET.md](docs/BUDGET.md)。

```bash
image-cli version          # 含 GitHub 仓库、Releases、安装脚本地址
image-cli version -json
image-cli check -profile ui button.png
```

基于 **Go 1.27**，单文件发布，支持 GitHub Release **懒更新提示**与 `image-cli update` 一键升级（内置 gh-proxy / ghfast / ghproxy 镜像探测）。

CI 在 Linux / Windows / macOS 上运行 `go vet` 与全量测试（含 CLI e2e、安装脚本 checksum 校验）。

## 安装

### 一键安装（推荐）

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.sh | bash
```

国内网络可使用 raw 加速：

```bash
curl -fsSL https://ghproxy.net/https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.ps1 | iex
```

国内推荐（自动镜像 + raw 加速）：

```powershell
irm https://ghproxy.net/https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.ps1 | iex
```

也可本地执行：`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install.ps1`

### 环境变量

| 变量 | 说明 |
|------|------|
| `IMAGE_CLI_UPDATE_MIRRORS` | 下载/更新镜像前缀列表（逗号分隔），默认多个 gh 加速前缀 |
| `IMAGE_CLI_DOWNLOAD_SOURCE` | `auto`（默认）/`github`/`mirror`；`install.ps1` 与 `install.sh` 均支持 |
| `IMAGE_CLI_PREFER_MIRROR=1` | `auto` 模式下优先镜像（适合国内） |
| `IMAGE_CLI_DIRECT_ONLY=1` | 仅直连 GitHub |
| `IMAGE_CLI_NO_UPDATE_CHECK=1` | 关闭每次命令后的懒更新提示（24h 节流） |
| `IMAGE_CLI_INSTALL_DIR` | 自定义安装目录 |

## 自动更新

- **懒更新**：执行任意子命令时，后台检查 GitHub Latest Release（24 小时内不重复）；有新版本则在 stderr 提示 `image-cli update`。
- **主动更新**：`image-cli update` 下载当前平台二进制并校验 `SHA256SUMS` 后替换自身。

## 命令速览

```text
image-cli info [-exif] [-json] picture.jpg
image-cli convert -format png -o out.png input.webp
image-cli batch-convert -format png -r ./assets
image-cli optimize -format webp -max-width 2048 hero.png
image-cli resize -w 512 -h 512 -mode fit sprite.png
image-cli rgb -out-dir ./channels -raw sprite.png
image-cli palette -n 32 concept.png
image-cli pow2 -square texture.png          # 游戏 POT 贴图
image-cli split -cols 4 -rows 4 sheet.png   # 精灵表
image-cli adjust -brightness 5 -gamma 1.1 photo.jpg

# 游戏资源（UI 默认无 mip / 无 ASTC）
image-cli game prepare -profile ui hud.png
image-cli game prepare -profile albedo -platform mobile albedo.png
image-cli game prepare -profile albedo -platform mobile -mipmap albedo.png
image-cli game prepare -profile normal -platform pc -mipmap -format bc5 normal.png
image-cli mipmap -out-dir ./mips texture.png
image-cli game profiles
```

支持读取：**PNG / JPEG / WebP / GIF / BMP / TIFF**；写出：**PNG / JPEG / WebP / GIF / BMP / TIFF**。

## 从源码构建

```bash
go build -o image-cli ./cmd/image-cli
```

## 发布

推送 tag（例如 `v0.1.0`）后 GitHub Actions 会构建六平台二进制并创建 Release。

## License

MIT
