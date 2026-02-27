#!/usr/bin/env bash
# Go_GAPTUI 실행 스크립트
# Go 백엔드 + Python Textual TUI 프론트엔드를 동시에 시작합니다.
#
# 사용법:
#   ./run.sh          # 일반 실행
#   ./run.sh --race   # Go race detector 활성화 (개발 시 권장)
#   ./run.sh --build  # 바이너리 빌드 후 실행
#   ./run.sh --tui    # TUI 프론트엔드만 실행 (백엔드는 별도 실행 중일 때)
#   ./run.sh --engine # Go 엔진만 실행 (TUI 없이)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

TUI_DIR="$SCRIPT_DIR/tui_textual"
VENV_PYTHON="$TUI_DIR/.venv/bin/python"
GO_PID=""
TUI_PID=""

# 색상
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

cleanup() {
    echo ""
    echo -e "${YELLOW}종료 중...${NC}"
    if [[ -n "$GO_PID" ]] && kill -0 "$GO_PID" 2>/dev/null; then
        # Kill the process and all its children (handles 'go run' spawning a subprocess)
        pkill -P "$GO_PID" 2>/dev/null || true
        kill "$GO_PID" 2>/dev/null || true
        wait "$GO_PID" 2>/dev/null || true
        echo -e "${GREEN}Go 엔진 종료됨${NC}"
    fi
    if [[ -n "$TUI_PID" ]] && kill -0 "$TUI_PID" 2>/dev/null; then
        kill "$TUI_PID" 2>/dev/null || true
        wait "$TUI_PID" 2>/dev/null || true
        echo -e "${GREEN}Textual TUI 종료됨${NC}"
    fi
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# 인자 파싱
USE_RACE=false
BUILD_FIRST=false
TUI_ONLY=false
ENGINE_ONLY=false

for arg in "$@"; do
    case "$arg" in
        --race)   USE_RACE=true ;;
        --build)  BUILD_FIRST=true ;;
        --tui)    TUI_ONLY=true ;;
        --engine) ENGINE_ONLY=true ;;
        -h|--help)
            echo "사용법: ./run.sh [옵션]"
            echo ""
            echo "옵션:"
            echo "  --race    Go race detector 활성화"
            echo "  --build   바이너리 빌드 후 실행"
            echo "  --tui     TUI 프론트엔드만 실행"
            echo "  --engine  Go 엔진만 실행"
            echo "  -h        도움말"
            exit 0
            ;;
    esac
done

# 사전 검증
if [[ "$TUI_ONLY" == false ]]; then
    if ! command -v go &>/dev/null; then
        echo -e "${RED}오류: Go가 설치되어 있지 않습니다.${NC}"
        exit 1
    fi

    if [[ ! -f "$SCRIPT_DIR/config.toml" ]]; then
        echo -e "${RED}오류: config.toml이 없습니다.${NC}"
        exit 1
    fi

    if [[ ! -f "$SCRIPT_DIR/.env" ]]; then
        echo -e "${YELLOW}경고: .env 파일이 없습니다. API 키 없이 실행됩니다.${NC}"
    fi
fi

if [[ "$ENGINE_ONLY" == false ]]; then
    if [[ ! -d "$TUI_DIR" ]]; then
        echo -e "${RED}오류: tui_textual 디렉토리가 없습니다.${NC}"
        echo "  symlink 생성: ln -s /path/to/kimchiCEX_tui/tui_textual ./tui_textual"
        exit 1
    fi

    if [[ ! -f "$VENV_PYTHON" ]]; then
        echo -e "${RED}오류: Python venv가 없습니다. ($VENV_PYTHON)${NC}"
        echo "  생성: cd tui_textual && python3 -m venv .venv && .venv/bin/pip install -e ."
        exit 1
    fi
fi

# Go 엔진 시작
start_engine() {
    echo -e "${CYAN}▶ Go 엔진 시작${NC}"

    if [[ "$BUILD_FIRST" == true ]]; then
        echo -e "${CYAN}  빌드 중...${NC}"
        if [[ "$USE_RACE" == true ]]; then
            go build -race -o "$SCRIPT_DIR/kimchi" ./cmd/kimchi/
        else
            go build -o "$SCRIPT_DIR/kimchi" ./cmd/kimchi/
        fi
        echo -e "${GREEN}  빌드 완료${NC}"
        "$SCRIPT_DIR/kimchi" &
        GO_PID=$!
    elif [[ "$USE_RACE" == true ]]; then
        go run -race ./cmd/kimchi/ &
        GO_PID=$!
    else
        go run ./cmd/kimchi/ &
        GO_PID=$!
    fi

    echo -e "${GREEN}  Go 엔진 PID: $GO_PID${NC}"
}

# WebSocket 서버 대기
wait_for_ws() {
    echo -e "${CYAN}  WebSocket 서버 대기 (127.0.0.1:9876)...${NC}"
    local max_wait=30
    local waited=0
    while ! nc -z 127.0.0.1 9876 2>/dev/null; do
        sleep 0.5
        waited=$((waited + 1))
        if [[ $waited -ge $((max_wait * 2)) ]]; then
            echo -e "${RED}  오류: WebSocket 서버가 ${max_wait}초 내에 시작되지 않았습니다.${NC}"
            return 1
        fi
        # Go 프로세스가 죽었는지 확인
        if [[ -n "$GO_PID" ]] && ! kill -0 "$GO_PID" 2>/dev/null; then
            echo -e "${RED}  오류: Go 엔진이 예기치 않게 종료되었습니다.${NC}"
            return 1
        fi
    done
    echo -e "${GREEN}  WebSocket 서버 준비됨${NC}"
}

start_tui() {
    echo -e "${CYAN}▶ Python Textual TUI 시작${NC}"
    cd "$TUI_DIR"
    PYTHONPATH="$TUI_DIR" "$VENV_PYTHON" -m kimchi_tui
}

# 실행
if [[ "$TUI_ONLY" == true ]]; then
    start_tui
elif [[ "$ENGINE_ONLY" == true ]]; then
    start_engine
    echo -e "${GREEN}Go 엔진만 실행 중 (Ctrl+C로 종료)${NC}"
    wait "$GO_PID"
else
    start_engine
    wait_for_ws || exit 1
    start_tui
fi
