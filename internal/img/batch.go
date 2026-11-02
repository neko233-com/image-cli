package img

import (
	"path/filepath"
	"strings"
)

var rasterExts = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".webp": {}, ".gif": {}, ".bmp": {}, ".tif": {}, ".tiff": {},
}

func isRaster(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := rasterExts[ext]
	return ok
}
