# 绿幕、影视和表情包

这些命令处理的是已经拍好或导出的画面，不依赖在线模型。

## 抠像

```bash
image-cli key -color green -o cut.png plate.png
image-cli key -color screen -similarity 0.32 -smoothness 0.12 -despill -o cut.png plate.png
image-cli key -sample 8,8 -bg scene.jpg -o comp.png actor.png
image-cli matte -choke 1 -feather 1 cut.png
image-cli composite -fg cut.png -x 40 -y 10 scene.jpg
image-cli gif key -color green -o cut.gif plate.gif
image-cli batch key -color green -out-dir ./cut ./plates
```

| 键色 | 含义 |
|------|------|
| `green` | 数字绿 `#00FF00` |
| `screen` | 常见棚拍绿 `#00B140` |
| `blue` | 蓝幕 |
| `magenta` | 品红幕 |
| `#RRGGBB` | 自定义 |

`-similarity` 越大，越接近键色的区域越透明。白墙和皮肤不会因为“有一点绿”被整片抠掉：只有键色通道明显高于其他通道时才会进入抠像。`-despill` 把残留的绿边或蓝边压回另外两个通道。

`-sample x,y` 用画面上的一个像素当键色。`-bg` 抠完直接叠到背景上。

GIF 会先按帧处置规则合成完整画面，再逐帧抠像，透明像素写进 GIF。

## 影视画面

```bash
image-cli letterbox -aspect 2.39 shot.png
image-cli letterbox -aspect 1.85 shot.png
image-cli vignette -strength 0.55 shot.png
```

亮度、对比和伽马仍用 `adjust`。

## 表情包和贴纸

```bash
image-cli sticker -color green -stroke 8 -stroke-color white -shadow -o sticker.png face.png
image-cli sticker -color none -stroke 6 -circle -o sticker.png already-cut.png
image-cli sticker -color green -text 好耶 -o sticker.png face.png
image-cli meme -top 上面一排 -bottom 下面一排 photo.jpg
image-cli caption -text "第一行\n第二行" -x 24 -y 64 photo.jpg
image-cli batch sticker -color green -stroke 6 -out-dir ./stickers ./faces
```

字幕默认找系统里的微软雅黑、黑体或对应平台的中文字体。也可以用 `-font` 或 `IMAGE_CLI_FONT` 指定 ttf、otf 或 ttc。

描边是白字黑边。贴纸默认先抠绿幕，再裁掉透明边，最后加白边。已经有透明通道的图用 `-color none`。
