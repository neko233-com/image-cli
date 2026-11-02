package version

import "runtime"

var (
	// Version 由 release 构建通过 -ldflags 注入；本地开发默认为 dev。
	Version = "dev"
	Repo    = "neko233-com/image-cli"
	Binary  = "image-cli"
)

func String() string {
	return Version + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"
}
