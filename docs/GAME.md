# 游戏资源处理指南

`image-cli` 以 **UI / 2D 默认不生成 mipmap** 为首要原则；3D 贴图、立方体、光照贴图等在需要时用 `-mipmap` 显式开启。

## 场景 profile

| profile | 典型用途 | 默认 mipmap | 移动端压缩 | PC 压缩 |
|---------|----------|-------------|------------|---------|
| `ui` | 界面、HUD | 否 | 无（预乘 alpha） | 无 |
| `sprite` / `atlas` | 2D 精灵、图集 | 否 | 无（预乘 alpha） | 无 |
| `vfx` | 粒子 / 特效序列 | 否 | 无（预乘 alpha，可 `-compress`） | 无 |
| `icon` | 应用图标 | 否 | 无 | 无 |
| `albedo` / `detail` | 基础色、细节色 | 否 | ASTC sRGB | BC7 |
| `normal` | 法线 | 否 | ASTC normal | BC5 |
| `emissive` | 自发光 | 否 | ASTC HDR | BC7 |
| `orm` / `mask` | 打包贴图、遮罩 | 否 | ASTC linear | BC7 |
| `roughness` / `metallic` / `ao` / `height` / `specular` / `opacity` | 3D 数据贴图 | 否 | ASTC linear | BC4 |
| `flow` / `noise` | 流动、噪声 | 否 | ASTC linear | BC7 |
| `hdr` / `lightmap` | HDR / 光照 | 否 | ASTC HDR | BC6H |
| `cubemap` | 环境立方体 | 否 | ASTC HDR | BC6H |

`game prepare` 与 `game batch-prepare` / `batch prepare` 使用同一套参数：`-mipmap`、`-pow2`、`-format`、`-block`、`-quality`、`-astc-mode`、`-max-edge`、`-compress` / `-no-compress`、`-out-dir`、`-mirror-tree`、`-jobs`、`-continue`。目录输入默认递归。`-jobs 0`（默认）等于逻辑 CPU 数，处理在收集完路径后并行执行。

`-profile auto` 按文件名识别（如 `_albedo`、`_n`、`_orm`、`_roughness`、`_vfx`、`_cubemap`）。识别失败时用 `-fallback`（默认 `ui`）。别名示例：`basecolor`→albedo，`particle`→vfx，`smoothness`→roughness（不反转数值）。

查看机器可读默认值：

```bash
image-cli game profiles
```

## 常用命令

**UI 图（默认）：仅缩放 / 预乘 alpha，不 mip、不 ASTC**

```bash
image-cli game prepare -profile ui button.png
```

**移动端 3D  albedo（ASTC，仍默认无 mip）**

```bash
image-cli game prepare -profile albedo -platform mobile texture.png
```

**移动端 3D + mipmap + ASTC 内置 mip 链**

```bash
image-cli game prepare -profile albedo -platform mobile -mipmap texture.png
```

**PC 端 BC7 + mipmap（需安装 [texconv](https://github.com/microsoft/DirectXTex/releases) 或设置 `IMAGE_CLI_TEXCONV`）**

```bash
image-cli game prepare -profile albedo -platform pc -mipmap -format bc7 texture.png
```

**仅导出 mipmap PNG 序列（不压缩）**

```bash
image-cli mipmap -out-dir ./mips hero.png
```

**ORM 通道打包（R=AO G=Rough B=Metal）**

```bash
image-cli game pack-orm -ao ao.png -roughness r.png -metallic m.png -o orm.png
image-cli game prepare -profile orm -platform mobile orm.png
```

**批量 UI 目录**

```bash
image-cli game batch-prepare -profile ui -no-compress ./Assets/UI
```

**一个模型目录：按文件名分别处理 albedo / 法线 / ORM / 特效**

```bash
image-cli game prepare -profile auto -fallback albedo -platform mobile -no-compress -out-dir ./prepared ./Textures
image-cli batch prepare -profile vfx -premultiply -no-compress ./VFX
image-cli game prepare -profile roughness -platform pc -mipmap -no-compress ground_roughness.png
```

## 外部工具

| 工具 | 环境变量 | 用途 |
|------|----------|------|
| [astcenc](https://github.com/ARM-software/astc-encoder) | `IMAGE_CLI_ASTCENC` | 移动端 ASTC（默认 4x4） |
| texconv | `IMAGE_CLI_TEXCONV` | PC BC 压缩为 DDS |

未安装外部工具时，`game prepare` 仍会输出 `*_prepared.png`（及可选 mip PNG），并在 manifest 中记录；压缩步骤会报错提示安装。

## 输出 manifest

每次 `game prepare` 默认生成 `<basename>.image-cli.json`，记录 profile、平台、输出文件列表与警告，便于 CI / Unity / Unreal 流水线消费。

`warnings` 中含各发行平台 **budget** 检测结果（`budget_overall` / `budget_summary` 在 `meta`）。详见 [BUDGET.md](BUDGET.md)。

## 尺寸是否合理

```bash
image-cli check -profile ui hud.png
image-cli check -platform ios,android,minigame pack.png
image-cli info -budget -json texture.png
```
