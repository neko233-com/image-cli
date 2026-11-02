# 多平台贴图预算

`image-cli check` 与 `image-cli info -budget` 会按**保守工程预算**评估单张图片是否合理，覆盖：

| platform | 说明 |
|----------|------|
| `windows` / `mac` | PC 桌面，边长上限 8192，推荐 ≤4096 |
| `ios` / `android` | 移动端 ASTC，推荐 ≤2048，4 像素对齐 |
| `webgl` | 桌面浏览器，体积与 NPOT 敏感 |
| `webgl-mobile` | 移动浏览器 / WebView |
| `minigame` | 微信等小游戏，单文件推荐 ≤128KB |
| `switch` | Nintendo Switch，推荐场景贴图 1K–2K |

## 命令

```bash
# 全平台检测
image-cli check hero.png

# 仅小游戏 + 移动 WebGL
image-cli check -platform minigame,webgl-mobile ui_atlas.png

# 按 UI 场景关联平台
image-cli check -profile ui panel.png

# CI：警告也失败
image-cli check -strict -profile minigame asset.png

# 与 info 合并
image-cli info -budget -profile albedo texture.png --json
```

## 判定级别

- **ok**：边长、像素数、文件体积均在推荐范围内
- **warn**：超过推荐值但仍低于硬上限，或 NPOT / ASTC 对齐问题
- **fail**：超过平台硬上限（如小游戏 512KB 文件、边长 >2048）

`game prepare` 会自动把 budget 结果写入 manifest `warnings`，便于流水线拦截。

预算值为经验保守值，具体项目应以引擎 Quality Settings 与真机测为准；可通过 issue 讨论调整默认表。
