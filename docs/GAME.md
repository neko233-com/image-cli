# 游戏资源处理指南

`image-cli` 以 **UI / 2D 默认不生成 mipmap** 为首要原则；3D 贴图、立方体、光照贴图等在需要时用 `-mipmap` 显式开启。

## 场景 profile

| profile | 典型用途 | 默认 mipmap | 移动端压缩 | PC 压缩 |
|---------|----------|-------------|------------|---------|
| `ui` | 界面、HUD | 否 | 无 | 无 |
| `sprite` / `atlas` / `vfx` | 2D 精灵、图集、特效 | 否 | 无 | 无 |
| `icon` | 应用图标 | 否 | 无 | 无 |
| `albedo` | 基础色 | 否 | ASTC sRGB 4x4 | BC7 |
| `normal` | 法线 | 否 | ASTC normal | BC5 |
| `emissive` | 自发光 | 否 | ASTC HDR | BC7 |
| `orm` | AO/R/M 打包后贴图 | 否 | ASTC | BC7 |
| `lightmap` / `cubemap` | 光照 / 环境立方体 | 否 | ASTC HDR | BC6H |

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
image-cli game batch-prepare -profile ui -no-compress -r ./Assets/UI
```

## 外部工具

| 工具 | 环境变量 | 用途 |
|------|----------|------|
| [astcenc](https://github.com/ARM-software/astc-encoder) | `IMAGE_CLI_ASTCENC` | 移动端 ASTC（默认 4x4） |
| texconv | `IMAGE_CLI_TEXCONV` | PC BC 压缩为 DDS |

未安装外部工具时，`game prepare` 仍会输出 `*_prepared.png`（及可选 mip PNG），并在 manifest 中记录；压缩步骤会报错提示安装。

## 输出 manifest

每次 `game prepare` 默认生成 `<basename>.image-cli.json`，记录 profile、平台、输出文件列表与警告，便于 CI / Unity / Unreal 流水线消费。
