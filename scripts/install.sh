#!/bin/bash
set -euo pipefail

# image-cli installer（支持 GitHub 镜像加速）
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.sh | bash
#   curl -fsSL https://ghproxy.net/https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.sh | bash
# Or: ... | bash -s -- v0.1.0

VERSION="${1:-latest}"
BINARY_NAME="image-cli"
REPO="${IMAGE_CLI_REPO:-neko233-com/image-cli}"

detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *)          echo "unsupported" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              echo "unsupported" ;;
    esac
}

route_urls() {
    local original="$1"
    printf '%s\n' "$original"
    if [ "${IMAGE_CLI_DIRECT_ONLY:-0}" = "1" ]; then
        return 0
    fi
    local mirrors="${IMAGE_CLI_UPDATE_MIRRORS-https://gh-proxy.com/,https://ghfast.top/,https://ghproxy.net/}"
    local old_ifs="$IFS"
    IFS=,
    for mirror in $mirrors; do
        [ -n "$mirror" ] || continue
        printf '%s\n' "${mirror%/}/$original"
    done
    IFS="$old_ifs"
}

fastest_route() {
    local target="$1"
    local best="" best_time="999999"
    while IFS= read -r candidate; do
        [ -n "$candidate" ] || continue
        local probe
        probe="$(curl -fsSL -L --range 0-65535 --max-time 8 -o /dev/null -w '%{http_code} %{time_total}' "$candidate" 2>/dev/null || true)"
        local code="${probe%% *}"
        local elapsed="${probe#* }"
        case "$code" in
            2*|3*)
                if awk "BEGIN { exit !($elapsed < $best_time) }"; then
                    best="$candidate"
                    best_time="$elapsed"
                fi
                ;;
        esac
    done <<EOF
$(route_urls "$target")
EOF
    [ -n "$best" ] || return 1
    printf '%s\n' "$best"
}

get_latest_version() {
    local api="https://api.github.com/repos/${REPO}/releases/latest"
    local route
    route="$(fastest_route "$api")" || return 1
    curl -fsSL --max-time 15 "$route" | sed -nE 's/.*"tag_name": *"v([^"]+)".*/\1/p'
}

normalize_version() {
    local v="$1"
    v="${v#v}"
    v="${v#V}"
    echo "$v"
}

install_binary() {
    local os="$1"
    local arch="$2"
    local ver="$3"

    local asset="${BINARY_NAME}-${os}-${arch}"
    local ext=""
    [ "$os" = "windows" ] && ext=".exe"

    local install_dir="${IMAGE_CLI_INSTALL_DIR:-/usr/local/bin}"
    local target="${BINARY_NAME}"

    if [ "$os" = "windows" ]; then
        install_dir="${LOCALAPPDATA:-$HOME/AppData/Local}/image-cli"
        target="${BINARY_NAME}${ext}"
        mkdir -p "$install_dir"
    fi

    local tmp_dir base downloaded expected actual entry name candidate
    tmp_dir=$(mktemp -d)
    trap 'rm -rf -- "$tmp_dir"; trap - RETURN' RETURN
    base="https://github.com/${REPO}/releases/download/v${ver}"

    local sums_route
    sums_route="$(fastest_route "$base/SHA256SUMS")" || { echo 'no reachable SHA256SUMS route' >&2; return 1; }
    local selected_base="${sums_route%/SHA256SUMS}"
    curl -fsSL --max-time 60 "$sums_route" -o "$tmp_dir/SHA256SUMS"

    downloaded="$tmp_dir/${asset}${ext}"
    curl -fsSL --max-time 300 "$selected_base/${asset}${ext}" -o "$downloaded"
    expected=$(awk -v name="${asset}${ext}" '$2==name || $2==("dist/" name) {print $1; exit}' "$tmp_dir/SHA256SUMS")
    [[ "$expected" =~ ^[a-f0-9]{64}$ ]] || { echo 'missing official artifact checksum' >&2; return 1; }
    if command -v sha256sum >/dev/null 2>&1; then
        actual=$(sha256sum "$downloaded" | awk '{print $1}')
    else
        actual=$(shasum -a 256 "$downloaded" | awk '{print $1}')
    fi
    [[ "$actual" == "$expected" ]] || { echo 'official artifact checksum mismatch' >&2; return 1; }

    name="$target"
    candidate="$install_dir/.$name.candidate.$$"
    if [ -w "$install_dir" ]; then
        cp "$downloaded" "$candidate"
        chmod 755 "$candidate"
        mv -f "$candidate" "$install_dir/$name"
    else
        sudo cp "$downloaded" "$candidate"
        sudo chmod 755 "$candidate"
        sudo mv -f "$candidate" "$install_dir/$name"
    fi
    echo "Installed verified artifact: $install_dir/$name"
    echo "Run: image-cli --help"
}

main() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
        echo "Unsupported operating system."
        exit 1
    fi

    if [ "$VERSION" = "latest" ] || [ -z "$VERSION" ]; then
        VERSION=$(get_latest_version)
    fi
    VERSION=$(normalize_version "$VERSION")
    [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][A-Za-z0-9._-]+)?$ ]] || { echo 'invalid release version' >&2; exit 1; }

    echo "Detected: ${OS}/${ARCH}"
    echo "Installing image-cli v${VERSION}..."
    install_binary "$OS" "$ARCH" "$VERSION"
    echo "Done."
}

main "$@"
