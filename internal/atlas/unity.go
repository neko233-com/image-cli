package atlas

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// WriteUnityMeta 写出 Unity TextureImporter 多精灵 .meta。rect 原点在贴图左下角。
func WriteUnityMeta(pngName string, sheet Sheet) (string, error) {
	guid, err := newGUID()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, `fileFormatVersion: 2
guid: %s
TextureImporter:
  serializedVersion: 13
  textureType: 8
  textureShape: 1
  spriteMode: 2
  spritePixelsToUnits: 100
  alphaIsTransparency: 1
  sRGBTexture: 1
  alphaUsage: 1
  enableMipMap: 0
  maxTextureSize: 8192
  textureSettings:
    serializedVersion: 2
    filterMode: 1
    wrapU: 1
    wrapV: 1
  spriteSheet:
    serializedVersion: 2
    sprites:
`, guid)
	for i, sp := range sheet.Sprites {
		uy := sheet.Height - sp.Y - sp.H
		fmt.Fprintf(&b, `    - serializedVersion: 2
      name: %s
      rect:
        serializedVersion: 2
        x: %d
        y: %d
        width: %d
        height: %d
      alignment: 0
      pivot: {x: 0.5, y: 0.5}
      border: {x: 0, y: 0, z: 0, w: 0}
      internalID: %d
`, yamlScalar(sp.Name), sp.X, uy, sp.W, sp.H, int64(i+1))
	}
	b.WriteString("  spritePackingTag: \n  userData: \n  assetBundleName: \n  assetBundleVariant: \n")
	return b.String(), nil
}

// ParseUnityMeta 读取 .meta 里的 sprite rect，并把左下角原点换成 PNG 左上角。
func ParseUnityMeta(text string, atlasHeight int) ([]Sprite, error) {
	if atlasHeight <= 0 {
		return nil, fmt.Errorf("unity meta requires atlas height")
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	inSprites := false
	inRect := false
	var cur *Sprite
	var out []Sprite
	flush := func() {
		if cur == nil || cur.Name == "" || cur.W <= 0 || cur.H <= 0 {
			cur = nil
			return
		}
		cur.Y = atlasHeight - cur.Y - cur.H
		cur.SourceW, cur.SourceH = cur.W, cur.H
		out = append(out, *cur)
		cur = nil
	}
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "sprites:" {
			inSprites = true
			continue
		}
		if !inSprites {
			continue
		}
		if strings.HasPrefix(trim, "outline:") || strings.HasPrefix(trim, "physicsShape:") || strings.HasPrefix(trim, "secondaryTextures:") {
			if !strings.HasPrefix(line, "      ") && !strings.HasPrefix(line, "\t") {
				inSprites = false
			}
		}
		if strings.HasPrefix(trim, "- serializedVersion:") || strings.HasPrefix(trim, "- name:") {
			flush()
			cur = &Sprite{}
			inRect = false
			if name, ok := yamlValue(trim, "name"); ok {
				cur.Name = name
			}
			continue
		}
		if cur == nil {
			continue
		}
		if name, ok := yamlValue(trim, "name"); ok && cur.Name == "" {
			cur.Name = name
			continue
		}
		if trim == "rect:" || strings.HasPrefix(trim, "rect:") {
			inRect = true
			continue
		}
		if !inRect {
			continue
		}
		switch {
		case strings.HasPrefix(trim, "x:"):
			cur.X = atoiValue(trim)
		case strings.HasPrefix(trim, "y:"):
			cur.Y = atoiValue(trim)
		case strings.HasPrefix(trim, "width:"):
			cur.W = atoiValue(trim)
		case strings.HasPrefix(trim, "height:"):
			cur.H = atoiValue(trim)
		default:
			if !strings.HasPrefix(trim, "serializedVersion:") {
				inRect = false
			}
		}
	}
	flush()
	if len(out) == 0 {
		return nil, fmt.Errorf("unity meta: no sprites")
	}
	return out, nil
}

func yamlValue(trim, key string) (string, bool) {
	prefixes := []string{"- " + key + ":", key + ":"}
	for _, p := range prefixes {
		if strings.HasPrefix(trim, p) {
			return strings.Trim(strings.TrimSpace(trim[len(p):]), `"'`), true
		}
	}
	return "", false
}

func atoiValue(trim string) int {
	_, v, _ := strings.Cut(trim, ":")
	n, _ := strconv.Atoi(strings.TrimSpace(v))
	return n
}

func yamlScalar(s string) string {
	if s == "" || strings.ContainsAny(s, ":#{}[]&*!|>'\"%@`") || strings.Contains(s, " ") {
		return strconv.Quote(s)
	}
	return s
}

func newGUID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}
