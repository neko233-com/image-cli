package atlas

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// WriteCocosPlist 写出 Cocos2d-x / Cocos Creator 可导入的 plist format 3。坐标原点为左上角。
func WriteCocosPlist(textureName string, sheet Sheet) (string, error) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	b.WriteString(`<plist version="1.0"><dict>` + "\n")
	b.WriteString("<key>frames</key>\n<dict>\n")
	for _, sp := range sheet.Sprites {
		key := sp.Name
		if !strings.Contains(key, ".") {
			key += ".png"
		}
		fmt.Fprintf(&b, "<key>%s</key>\n<dict>\n", xmlText(key))
		fmt.Fprintf(&b, "<key>spriteOffset</key><string>{%d,%d}</string>\n", sp.TrimX, sp.TrimY)
		fmt.Fprintf(&b, "<key>spriteSize</key><string>{%d,%d}</string>\n", sp.W, sp.H)
		fmt.Fprintf(&b, "<key>spriteSourceSize</key><string>{%d,%d}</string>\n", nz(sp.SourceW, sp.W), nz(sp.SourceH, sp.H))
		fmt.Fprintf(&b, "<key>textureRect</key><string>{{%d,%d},{%d,%d}}</string>\n", sp.X, sp.Y, sp.W, sp.H)
		fmt.Fprintf(&b, "<key>textureRotated</key><%t/>\n", sp.Rotated)
		b.WriteString("</dict>\n")
	}
	b.WriteString("</dict>\n<key>metadata</key>\n<dict>\n")
	b.WriteString("<key>format</key><integer>3</integer>\n")
	b.WriteString("<key>pixelFormat</key><string>RGBA8888</string>\n")
	b.WriteString("<key>premultiplyAlpha</key><false/>\n")
	fmt.Fprintf(&b, "<key>realTextureFileName</key><string>%s</string>\n", xmlText(textureName))
	fmt.Fprintf(&b, "<key>size</key><string>{%d,%d}</string>\n", sheet.Width, sheet.Height)
	fmt.Fprintf(&b, "<key>textureFileName</key><string>%s</string>\n", xmlText(textureName))
	b.WriteString("</dict>\n</dict></plist>\n")
	return b.String(), nil
}

// ParseCocosPlist 读取 format 2 与 format 3。
func ParseCocosPlist(data []byte) (string, []Sprite, error) {
	root, err := decodePlist(data)
	if err != nil {
		return "", nil, err
	}
	frames, _ := root["frames"].(map[string]any)
	if len(frames) == 0 {
		return "", nil, fmt.Errorf("cocos plist: no frames")
	}
	meta, _ := root["metadata"].(map[string]any)
	texture, _ := meta["textureFileName"].(string)
	if texture == "" {
		texture, _ = meta["realTextureFileName"].(string)
	}
	var out []Sprite
	for name, raw := range frames {
		dict, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		sp := Sprite{Name: strings.TrimSuffix(name, ".png")}
		if rect, ok := dict["textureRect"].(string); ok {
			x, y, w, h, err := parseRect(rect)
			if err != nil {
				return "", nil, fmt.Errorf("%s: %w", name, err)
			}
			sp.X, sp.Y, sp.W, sp.H = x, y, w, h
		} else if frame, ok := dict["frame"].(string); ok {
			x, y, w, h, err := parseRect(frame)
			if err != nil {
				return "", nil, fmt.Errorf("%s: %w", name, err)
			}
			sp.X, sp.Y, sp.W, sp.H = x, y, w, h
		}
		if v, ok := dict["textureRotated"].(bool); ok {
			sp.Rotated = v
		} else if v, ok := dict["rotated"].(bool); ok {
			sp.Rotated = v
		}
		if size, ok := dict["spriteSize"].(string); ok {
			w, h, err := parsePair(size)
			if err != nil {
				return "", nil, err
			}
			sp.W, sp.H = w, h
		}
		if size, ok := dict["spriteSourceSize"].(string); ok {
			w, h, err := parsePair(size)
			if err != nil {
				return "", nil, err
			}
			sp.SourceW, sp.SourceH = w, h
		} else if size, ok := dict["sourceSize"].(string); ok {
			w, h, err := parsePair(size)
			if err != nil {
				return "", nil, err
			}
			sp.SourceW, sp.SourceH = w, h
		}
		if sp.Rotated && dict["spriteSize"] == nil {
			sp.W, sp.H = sp.H, sp.W
		}
		if sp.SourceW == 0 {
			sp.SourceW = sp.W
		}
		if sp.SourceH == 0 {
			sp.SourceH = sp.H
		}
		out = append(out, sp)
	}
	if len(out) == 0 {
		return "", nil, fmt.Errorf("cocos plist: no frames")
	}
	return texture, out, nil
}

var rectPattern = regexp.MustCompile(`\{\{(-?\d+),(-?\d+)\},\{(-?\d+),(-?\d+)\}\}`)
var pairPattern = regexp.MustCompile(`\{(-?\d+),(-?\d+)\}`)

func parseRect(s string) (int, int, int, int, error) {
	m := rectPattern.FindStringSubmatch(strings.ReplaceAll(s, " ", ""))
	if m == nil {
		return 0, 0, 0, 0, fmt.Errorf("bad rect %q", s)
	}
	return atoi(m[1]), atoi(m[2]), atoi(m[3]), atoi(m[4]), nil
}

func parsePair(s string) (int, int, error) {
	m := pairPattern.FindStringSubmatch(strings.ReplaceAll(s, " ", ""))
	if m == nil {
		return 0, 0, fmt.Errorf("bad pair %q", s)
	}
	return atoi(m[1]), atoi(m[2]), nil
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func nz(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}

func xmlText(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func decodePlist(data []byte) (map[string]any, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil, fmt.Errorf("plist: missing dict")
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "plist" {
			continue
		}
		val, err := nextValue(dec)
		if err != nil {
			return nil, err
		}
		m, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("plist: root is not a dict")
		}
		return m, nil
	}
}

func nextValue(dec *xml.Decoder) (any, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			return readValue(dec, t)
		case xml.EndElement:
			return nil, fmt.Errorf("plist: unexpected </%s>", t.Name.Local)
		}
	}
}

func readValue(dec *xml.Decoder, se xml.StartElement) (any, error) {
	switch se.Name.Local {
	case "dict":
		m := map[string]any{}
		var key string
		for {
			tok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "key" {
					text, err := textUntilEnd(dec)
					if err != nil {
						return nil, err
					}
					key = text
					continue
				}
				val, err := readValue(dec, t)
				if err != nil {
					return nil, err
				}
				m[key] = val
			case xml.EndElement:
				return m, nil
			}
		}
	case "array":
		var arr []any
		for {
			tok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			switch t := tok.(type) {
			case xml.StartElement:
				val, err := readValue(dec, t)
				if err != nil {
					return nil, err
				}
				arr = append(arr, val)
			case xml.EndElement:
				return arr, nil
			}
		}
	case "string", "integer":
		return textUntilEnd(dec)
	case "true":
		return finishEmpty(dec, true)
	case "false":
		return finishEmpty(dec, false)
	default:
		return nil, fmt.Errorf("plist: unsupported <%s>", se.Name.Local)
	}
}

func textUntilEnd(dec *xml.Decoder) (string, error) {
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			b.Write(t)
		case xml.EndElement:
			return b.String(), nil
		}
	}
}

func finishEmpty(dec *xml.Decoder, v bool) (bool, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return false, err
		}
		if _, ok := tok.(xml.EndElement); ok {
			return v, nil
		}
	}
}
