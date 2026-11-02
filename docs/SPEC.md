# 规范检查

`check` 在平台预算之外，可以按生产场景做轻量规则检查。规则只使用已经读出的宽高、格式、alpha、位深和文件大小，不额外解码。

```bash
image-cli check -domain game ./Textures
image-cli check -domain design ./Posters
image-cli check -domain film ./Plates
image-cli check -domain vfx -strict ./VFX
image-cli check -domain auto -profile auto -jobs 0 ./assets
```

| domain | 关注点 |
|--------|--------|
| `game` | 边长、偶数尺寸、JPEG 透明、法线 alpha、3D 贴图 NPOT，并保留平台预算 |
| `design` | 交付格式、超大边长、文件体积、极端宽高比 |
| `film` | 偶数像素、母版格式、8-bit 提示、低于 HD |
| `vfx` | 禁止 JPEG/GIF、缺少 alpha、奇数尺寸、实时边长 |
| `auto` | `_plate`/`_comp` 视为影视；`vfx` 视为特效；`ui`/`sprite` 视为设计；其余视为游戏 |
| `all` | 四套规则一起跑 |

目录默认递归。`-jobs` 省略或为 `0` 时使用 `runtime.NumCPU()`；正整数为固定 worker，上限 1024。`-continue` 在单张失败后继续。多文件时 stderr 会打印 `batch: files=N workers=...`。

`-strict` 在出现警告时以退出码 1 结束，适合 CI。
