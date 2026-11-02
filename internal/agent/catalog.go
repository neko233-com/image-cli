package agent

// Task 是给 Agent 的一条可执行配方。命令里的尖括号是必须替换的路径。
type Task struct {
	ID       string   `json:"id"`
	Audience string   `json:"audience"`
	When     string   `json:"when"`
	Command  string   `json:"command"`
	Notes    string   `json:"notes,omitempty"`
	Argv     []string `json:"argv"`
}

func Tasks() []Task {
	return []Task{
		{ID: "inspect", Audience: "common", When: "还不清楚素材尺寸、格式和透明通道", Command: "image-cli info -json <file>", Argv: []string{"image-cli", "info", "-json", "<file>"}, Notes: "先读信息再改文件。"},
		{ID: "version", Audience: "common", When: "确认工具版本和仓库", Command: "image-cli version -json", Argv: []string{"image-cli", "version", "-json"}},
		{ID: "game-ui", Audience: "game", When: "界面、HUD、图标，不要 mipmap，不要纹理压缩", Command: "image-cli game prepare -profile ui -no-compress -out-dir <out> <dir-or-file>", Argv: []string{"image-cli", "game", "prepare", "-profile", "ui", "-no-compress", "-out-dir", "<out>", "<dir-or-file>"}, Notes: "目录默认递归。UI 默认预乘 alpha。"},
		{ID: "game-ui-check", Audience: "game", When: "UI 入库前做规范检查", Command: "image-cli check -domain design -profile ui -jobs 0 <dir>", Argv: []string{"image-cli", "check", "-domain", "design", "-profile", "ui", "-jobs", "0", "<dir>"}, Notes: "CI 加 -strict，警告也会失败。"},
		{ID: "game-textures", Audience: "game", When: "一个模型目录里混有 albedo、法线、ORM、粗糙度等", Command: "image-cli game prepare -profile auto -fallback albedo -platform mobile -no-compress -jobs 0 -out-dir <out> <textures-dir>", Argv: []string{"image-cli", "game", "prepare", "-profile", "auto", "-fallback", "albedo", "-platform", "mobile", "-no-compress", "-jobs", "0", "-out-dir", "<out>", "<textures-dir>"}, Notes: "按文件名识别。PC 把 mobile 换成 pc。需要 ASTC/BC 时去掉 -no-compress，并保证 astcenc 或 texconv 在 PATH。"},
		{ID: "game-albedo-mobile", Audience: "game", When: "单张基础色，移动端，仍默认不生成 mipmap", Command: "image-cli game prepare -profile albedo -platform mobile -no-compress -out-dir <out> <file>", Argv: []string{"image-cli", "game", "prepare", "-profile", "albedo", "-platform", "mobile", "-no-compress", "-out-dir", "<out>", "<file>"}},
		{ID: "game-normal-pc", Audience: "game", When: "法线贴图，PC，要 mipmap 和 BC5", Command: "image-cli game prepare -profile normal -platform pc -mipmap -format bc5 <file>", Argv: []string{"image-cli", "game", "prepare", "-profile", "normal", "-platform", "pc", "-mipmap", "-format", "bc5", "<file>"}, Notes: "没有 texconv 时加 -no-compress，只出预处理 PNG。"},
		{ID: "game-vfx-texture", Audience: "game", When: "粒子、序列帧、特效贴图", Command: "image-cli game prepare -profile vfx -no-compress -premultiply -out-dir <out> <dir>", Argv: []string{"image-cli", "game", "prepare", "-profile", "vfx", "-no-compress", "-premultiply", "-out-dir", "<out>", "<dir>"}},
		{ID: "game-orm", Audience: "game", When: "把 AO、粗糙度、金属度合成 ORM", Command: "image-cli game pack-orm -ao <ao.png> -roughness <r.png> -metallic <m.png> -o orm.png", Argv: []string{"image-cli", "game", "pack-orm", "-ao", "<ao.png>", "-roughness", "<r.png>", "-metallic", "<m.png>", "-o", "orm.png"}, Notes: "通道是 R=AO G=Rough B=Metal。smoothness 不会自动取反。"},
		{ID: "game-atlas-unity", Audience: "game", When: "把散图打成 Unity 多精灵图集", Command: "image-cli atlas pack -format unity -padding 2 -pot -out-dir <out> <sprites-dir>", Argv: []string{"image-cli", "atlas", "pack", "-format", "unity", "-padding", "2", "-pot", "-out-dir", "<out>", "<sprites-dir>"}, Notes: "产出 atlas.png 和 atlas.png.meta。"},
		{ID: "game-atlas-ue", Audience: "game", When: "把散图打成 Unreal Paper2D 图集", Command: "image-cli atlas pack -format ue -padding 2 -pot -out-dir <out> <sprites-dir>", Argv: []string{"image-cli", "atlas", "pack", "-format", "ue", "-padding", "2", "-pot", "-out-dir", "<out>", "<sprites-dir>"}},
		{ID: "game-atlas-cocos", Audience: "game", When: "把散图打成 Cocos plist 图集", Command: "image-cli atlas pack -format cocos -padding 2 -pot -out-dir <out> <sprites-dir>", Argv: []string{"image-cli", "atlas", "pack", "-format", "cocos", "-padding", "2", "-pot", "-out-dir", "<out>", "<sprites-dir>"}},
		{ID: "game-atlas-unpack", Audience: "game", When: "把已有 Unity meta、UE JSON 或 Cocos plist 拆回散图", Command: "image-cli atlas unpack -out-dir <out> <descriptor>", Argv: []string{"image-cli", "atlas", "unpack", "-out-dir", "<out>", "<descriptor>"}},
		{ID: "game-check", Audience: "game", When: "3D 贴图或整包资源做平台预算和规范", Command: "image-cli check -domain game -profile auto -jobs 0 <dir>", Argv: []string{"image-cli", "check", "-domain", "game", "-profile", "auto", "-jobs", "0", "<dir>"}},
		{ID: "film-key", Audience: "film", When: "绿幕或蓝幕静帧抠像", Command: "image-cli key -color green -despill -o <out.png> <plate>", Argv: []string{"image-cli", "key", "-color", "green", "-despill", "-o", "<out.png>", "<plate>"}, Notes: "棚拍绿用 -color screen。蓝幕用 -color blue。这是色度键，不能把普通照片里的人物从复杂背景里分离。"},
		{ID: "film-key-sample", Audience: "film", When: "幕布颜色不纯，要用画面上的一点取样", Command: "image-cli key -sample <x,y> -despill -o <out.png> <plate>", Argv: []string{"image-cli", "key", "-sample", "<x,y>", "-despill", "-o", "<out.png>", "<plate>"}},
		{ID: "film-composite", Audience: "film", When: "抠完直接叠到背景", Command: "image-cli key -color green -bg <background> -o <out.png> <plate>", Argv: []string{"image-cli", "key", "-color", "green", "-bg", "<background>", "-o", "<out.png>", "<plate>"}},
		{ID: "film-matte", Audience: "film", When: "抠像边缘要收缩或羽化", Command: "image-cli matte -choke 1 -feather 1 -o <out.png> <cut.png>", Argv: []string{"image-cli", "matte", "-choke", "1", "-feather", "1", "-o", "<out.png>", "<cut.png>"}},
		{ID: "film-gif-key", Audience: "film", When: "绿幕 GIF 或序列动画要保留透明", Command: "image-cli gif key -color green -o <out.gif> <plate.gif>", Argv: []string{"image-cli", "gif", "key", "-color", "green", "-o", "<out.gif>", "<plate.gif>"}},
		{ID: "film-letterbox", Audience: "film", When: "出片要电影遮幅", Command: "image-cli letterbox -aspect 2.39 -o <out.png> <shot>", Argv: []string{"image-cli", "letterbox", "-aspect", "2.39", "-o", "<out.png>", "<shot>"}, Notes: "1.85 和 16:9 把 2.39 换成 1.85 或 1.778。"},
		{ID: "film-grade", Audience: "film", When: "简单调色：亮度、对比、伽马、暗角", Command: "image-cli adjust -brightness 0 -contrast 5 -gamma 1.05 -o <graded.png> <shot>", Argv: []string{"image-cli", "adjust", "-brightness", "0", "-contrast", "5", "-gamma", "1.05", "-o", "<graded.png>", "<shot>"}, Notes: "暗角另跑 image-cli vignette -strength 0.55。"},
		{ID: "film-check", Audience: "film", When: "检查母版是否偶数像素、位深和 HD", Command: "image-cli check -domain film -jobs 0 <dir>", Argv: []string{"image-cli", "check", "-domain", "film", "-jobs", "0", "<dir>"}},
		{ID: "sticker", Audience: "sticker", When: "绿幕人或道具做成带描边的表情贴纸", Command: "image-cli sticker -color green -stroke 8 -stroke-color white -shadow -o <sticker.png> <image>", Argv: []string{"image-cli", "sticker", "-color", "green", "-stroke", "8", "-stroke-color", "white", "-shadow", "-o", "<sticker.png>", "<image>"}, Notes: "已有透明通道时用 -color none。"},
		{ID: "meme", Audience: "sticker", When: "上下两行字幕的表情包", Command: "image-cli meme -top <上> -bottom <下> -o <out.png> <image>", Argv: []string{"image-cli", "meme", "-top", "<上>", "-bottom", "<下>", "-o", "<out.png>", "<image>"}, Notes: "中文需要系统字体，或 -font / IMAGE_CLI_FONT。"},
		{ID: "caption", Audience: "sticker", When: "在指定位置写字", Command: "image-cli caption -text <文字> -x 24 -y 64 -o <out.png> <image>", Argv: []string{"image-cli", "caption", "-text", "<文字>", "-x", "24", "-y", "64", "-o", "<out.png>", "<image>"}},
		{ID: "gif-extract", Audience: "sticker", When: "把 GIF 拆成完整帧再逐帧处理", Command: "image-cli gif extract -out-dir <frames> <file.gif>", Argv: []string{"image-cli", "gif", "extract", "-out-dir", "<frames>", "<file.gif>"}},
		{ID: "gif-pack", Audience: "sticker", When: "把序列帧合成循环 GIF", Command: "image-cli gif pack -delay 8 -o <out.gif> <frames-dir>", Argv: []string{"image-cli", "gif", "pack", "-delay", "8", "-o", "<out.gif>", "<frames-dir>"}, Notes: "delay 单位是 1/100 秒。"},
		{ID: "batch-convert", Audience: "common", When: "整目录换格式并保留子目录", Command: "image-cli batch convert -format png -jobs 0 -out-dir <out> <dir>", Argv: []string{"image-cli", "batch", "convert", "-format", "png", "-jobs", "0", "-out-dir", "<out>", "<dir>"}},
		{ID: "batch-sticker", Audience: "sticker", When: "一目录绿幕图都做成贴纸", Command: "image-cli batch sticker -color green -stroke 8 -jobs 0 -out-dir <out> <dir>", Argv: []string{"image-cli", "batch", "sticker", "-color", "green", "-stroke", "8", "-jobs", "0", "-out-dir", "<out>", "<dir>"}},
	}
}

func Find(id string) (Task, bool) {
	for _, task := range Tasks() {
		if task.ID == id {
			return task, true
		}
	}
	return Task{}, false
}

// Prompt 是给 Agent 的系统提示。运行 image-cli agent 得到同一份文本。
func Prompt() string {
	return prompt
}

const prompt = `你是图像生产代理。只使用 image-cli 处理游戏美术和影视特效素材。不要调用其他图像程序，不要编造不存在的参数。

开始一项任务前运行 image-cli agent -json 查看配方。选定 id 后，用 image-cli agent -task <id> 取出命令，把尖括号换成真实路径再执行。

硬规则：
- 先 image-cli info -json 或 check，再写文件。输出写到 -o 或 -out-dir，不覆盖用户没点名的原图。
- 目录默认递归。-jobs 0 表示按逻辑 CPU 数并行。大批量加 -continue，失败的文件跳过。
- 游戏 UI、图标、精灵默认不要 mipmap，不要 ASTC/BC。3D 贴图才用 albedo、normal、orm、roughness 等 profile。
- 混合贴图目录用 -profile auto。不要把整包贴图都当成 ui。
- 法线、ORM、粗糙度、金属度、AO、高度是数据图，不要当 sRGB 照片去调色。
- 绿幕和蓝幕用 key 或 gif key。这是色度键。普通照片里的人物、头发和复杂背景不要用 key 硬抠。
- 棚拍绿幕用 -color screen，纯数字绿用 -color green，蓝幕用 -color blue。键色不纯时用 -sample x,y。
- 抠像默认 -despill。边缘脏再 matte -choke 1 -feather 1，然后 composite 或 key -bg。
- 表情贴纸用 sticker。素材已经有透明通道时加 -color none。
- 中文用 meme 或 caption。找不到字体就停止，请用户提供 -font，不要用方框字凑合。
- Unity 图集用 atlas pack -format unity，Unreal 用 ue，Cocos 用 cocos。解图用 atlas unpack。
- 纹理压缩只有在用户明确要求，且机器上有 astcenc 或 texconv 时才去掉 -no-compress。否则保留 PNG，并在结果里说明跳过了压缩。
- check 给 CI 时加 -strict。游戏用 -domain game，设计稿用 design，影视母版用 film，特效用 vfx。
- 命令失败就把 stderr 原样返回，不要换一套参数重试超过两次。

标准流程：
1. 游戏 UI 目录：check -domain design -profile ui，然后 game prepare -profile ui -no-compress -out-dir。
2. 角色或场景贴图目录：game prepare -profile auto -fallback albedo -platform mobile|pc -no-compress -jobs 0 -out-dir。需要压缩时再按 profile 单独补 ASTC 或 BC。
3. 引擎图集：atlas pack -format unity|ue|cocos -padding 2 -pot -out-dir。
4. 影视绿幕静帧：key -color screen -despill -o。要进背景就加 -bg。边缘用 matte。出片用 letterbox -aspect 2.39，需要时 vignette。
5. 影视绿幕动画：gif key，或 gif extract 后逐帧 key，再 gif pack。
6. 表情包：sticker 出透明 PNG；需要字幕再用 meme 或 caption；动图用 gif pack。

禁止：
- 不要给 UI 生成 mipmap。
- 不要把 smoothness 数值取反。
- 不要对未知背景承诺“一键抠图”。
- 不要在没有 -out-dir 的情况下把整目录结果写回源文件，除非用户明确要求原地覆盖。
`
