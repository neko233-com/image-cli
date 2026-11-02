# 图集与 GIF

## 合图 / 解图

三种描述文件可以互相独立使用。像素坐标在 PNG 里都是左上角为原点；写入 Unity `.meta` 时会换成 Unity 的左下角。

| `-format` | 合图产物 | 引擎 |
|-----------|----------|------|
| `unity` | `atlas.png` + `atlas.png.meta`（spriteMode: Multiple） | Unity Sprite Editor |
| `ue` | `atlas.png` + `atlas.json`（`image-cli.ue-atlas.v1`，左上角） | Unreal Paper2D 的 SourceUV / SourceDimension |
| `cocos` | `atlas.png` + `atlas.plist`（format 3） | Cocos2d-x、Cocos Creator |

```bash
image-cli atlas pack -format unity -padding 2 -pot -out-dir ./out ./sprites
image-cli atlas pack -format ue -out-dir ./out ./sprites
image-cli atlas pack -format cocos -out-dir ./out ./sprites
image-cli atlas unpack -out-dir ./sprites ./out/atlas.png.meta
image-cli atlas unpack -out-dir ./sprites ./out/atlas.json
image-cli atlas unpack -out-dir ./sprites ./out/atlas.plist
```

解图默认按扩展名识别。Cocos plist 的 format 2 和 format 3 都能读；`textureRotated` / `rotated` 为真时按顺时针 90° 转回。合图本身不旋转。

`-r` 会递归子目录。同名文件会自动加后缀。

## GIF

普通 `convert` 仍只处理第一帧。动画用 `gif`：

```bash
image-cli gif info anim.gif
image-cli gif extract -out-dir ./frames anim.gif
image-cli gif pack -o anim.gif -delay 10 ./frames
```

`-delay` 的单位是 1/100 秒。合帧使用 256 色调色板。
