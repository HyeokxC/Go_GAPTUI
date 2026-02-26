# Go_GAPTUI

**김치 프리미엄 실시간 모니터링 시스템** — Go 백엔드 + Python Textual TUI 프론트엔드

국내(업비트, 빗썸) / 해외(바이낸스, 바이빗, OKX, 비트겟) 거래소 간 실시간 시세를 수집하고, 김치 프리미엄·국내 갭·선물 베이시스를 계산하여 TUI 대시보드에 표시하는 하이브리드 아키텍처 애플리케이션입니다.

---

## 아키텍처

```
┌─────────────────────────────────────────────────────┐
│  Python Textual TUI (프론트엔드)                       │
│  - 대시보드 렌더링, 사용자 입력 처리                       │
└────────────────────┬────────────────────────────────┘
                     │ WebSocket (ws://127.0.0.1:9876/ws)
                     │ ↑ JSON Snapshot (200ms 주기)
                     │ ↓ 사용자 명령 (JSON)
┌────────────────────┴────────────────────────────────┐
│  Go 엔진 (백엔드)                                     │
│                                                      │
│  ┌──────────┐ ┌──────────┐ ┌───────────┐            │
│  │ Ticker   │ │ Orderbook│ │ Symbol    │            │
│  │ WS (7)   │ │ WS (5)   │ │ Fetcher   │            │
│  └────┬─────┘ └────┬─────┘ └─────┬─────┘            │
│       │             │             │                   │
│       ▼             ▼             ▼                   │
│  ┌──────────────────────────────────────┐            │
│  │         Kimchi Monitor               │            │
│  │  김프 / 국내갭 / 선물베이시스 계산       │            │
│  └────────────────┬─────────────────────┘            │
│                   ▼                                   │
│  ┌─────────────┐ ┌──────────────┐                    │
│  │  Scenario   │ │  Transfer    │                    │
│  │  Detector   │ │  Manager     │                    │
│  └─────────────┘ └──────────────┘                    │
│                                                      │
│  ┌──────────┐ ┌──────────┐ ┌───────────┐            │
│  │ IPC 서버  │ │ SQLite   │ │ Config    │            │
│  │ :9876    │ │ 로그 DB   │ │ TOML+ENV  │            │
│  └──────────┘ └──────────┘ └───────────┘            │
└──────────────────────────────────────────────────────┘
```

---

## 주요 기능

### 실시간 시세 수집
- **7개 거래소** WebSocket 동시 연결 (업비트, 빗썸, 바이낸스, 바이빗, OKX, 바이낸스선물, 바이빗선물)
- **5개 거래소** 호가창(Orderbook) WebSocket 수집 (업비트, 빗썸, 바이낸스, 바이빗, OKX)
- 자동 재연결 (3초 딜레이), 거래소별 Ping/Pong 관리

### 프리미엄 계산
- **김치 프리미엄(KIMP)**: `(국내가 - 해외USD × USDT_KRW) / (해외USD × USDT_KRW) × 100`
- **국내 갭(DOM-GAP)**: `(업비트가 - 빗썸가) / 빗썸가 × 100`
- **선물 베이시스(FUT%)**: `(현물USD - 선물USD) / 선물USD × 100`

### 시나리오 감지
| 시나리오 | 진입 임계값 | 변동 임계값 |
|---------|-----------|-----------|
| KIMP    | ±5.0%     | 3.0%p     |
| DOM-GAP | ±1.5%     | 0.5%p     |
| FUT%    | ±0.5%     | 0.1%p     |

- 최대 200개 스레드 관리, 비활성 스레드 자동 정리

### 호가창 분석
- 매수/매도 평균가 계산 (슬리피지 반영)
- 스프레드 계산
- 실질 김프 산출 (호가 기반)

### 송금 관리
- **20개 체인** 네트워크 매핑 (7개 거래소별 네트워크 명칭 자동 변환)
- **40+** 네트워크 별칭(alias) 자동 정규화
- 공통 네트워크 자동 탐색 및 최적 네트워크 선택
- 7단계 송금 상태 머신 (선택 → 주소확인 → 출금 → 확인대기 → 입금확인 → 완료/실패)

### IPC 프로토콜
- WebSocket 서버 (`ws://127.0.0.1:9876/ws`)
- 200ms 주기 JSON 스냅샷 브로드캐스트
- Python Textual TUI와 호환되는 snake_case JSON 포맷

---

## 기술 스택

| 구분 | 라이브러리 | 용도 |
|------|-----------|------|
| WebSocket | `gorilla/websocket` v1.5.3 | 거래소 WS + IPC 서버 |
| 정밀 소수점 | `shopspring/decimal` v1.4.0 | 금액 계산 (float 사용 금지) |
| 설정 파일 | `pelletier/go-toml/v2` v2.2.3 | TOML 설정 파싱 |
| 데이터베이스 | `modernc.org/sqlite` v1.46.1 | 로그 저장 (Pure Go, CGO 불필요) |
| UUID | `google/uuid` | 시나리오 스레드 / 송금 작업 ID |
| 로깅 | `rs/zerolog` v1.33.0 | 구조화 로깅 |
| 환경변수 | `joho/godotenv` v1.5.1 | `.env` 파일 로딩 |
| HTTP | `net/http` (표준 라이브러리) | REST API 호출 |
| 암호화 | `crypto/hmac`, `crypto/sha256`, `crypto/sha512` | 거래소 API 인증 서명 |

---

## 프로젝트 구조

```
Go_GAPTUI/
├── cmd/
│   └── kimchi/
│       └── main.go                  # 엔트리포인트 (설정 로딩, Runner 실행, 시그널 처리)
├── internal/
│   ├── background/
│   │   └── runner.go                # 런타임 오케스트레이터 (모든 고루틴 관리)
│   ├── config/
│   │   └── config.go                # AppConfig, API 키, 프록시 설정 (TOML + .env)
│   ├── db/
│   │   └── db.go                    # SQLite: 로그 저장/조회, 세션 기록
│   ├── exchanges/
│   │   ├── auth.go                  # 7개 거래소 인증 서명 (9개 함수)
│   │   ├── http_client.go           # HTTP 클라이언트 빌더 (프록시 지원)
│   │   ├── orderbook_processor.go   # 호가 분석: 슬리피지, 스프레드, 실질 김프
│   │   ├── orderbook_ws.go          # 5개 거래소 호가창 WebSocket 수집기
│   │   ├── symbol_fetcher.go        # 5개 거래소 심볼 자동 탐색 + 교집합 로직
│   │   └── ticker_ws.go             # 7개 거래소 시세 WebSocket 수집기
│   ├── ipc/
│   │   ├── protocol.go              # IPC JSON 스냅샷 타입 정의 (snake_case)
│   │   └── server.go                # WebSocket 서버 (:9876), 브로드캐스트, 명령 파싱
│   ├── models/
│   │   ├── coin_state.go            # CoinState (20개 필드, *decimal.Decimal)
│   │   ├── db.go                    # LogEntry, LogType 정의
│   │   ├── exchange.go              # Exchange 열거형 (9개), 분류 메서드
│   │   ├── orderbook.go             # OrderbookEntry, OrderbookInfo
│   │   ├── scenario.go              # 시나리오 타입/스레드/설정
│   │   ├── snapshot.go              # AppSnapshot (전체 상태)
│   │   ├── ticker.go                # TickerData (거래소, 심볼, 가격, 타임스탬프)
│   │   └── wallet.go                # 거래소별 입출금 상태
│   ├── monitor/
│   │   ├── kimchi_monitor.go        # 김프/국내갭/선물베이시스 계산 엔진
│   │   └── scenario_detector.go     # 시나리오 감지기 (3가지 타입, 스레드 관리)
│   └── transfer/
│       ├── address_book.go          # 주소록 조회/저장 (JSON 파일)
│       ├── chain_mapping.go         # 20개 체인 네트워크 매핑 + 40+ 별칭
│       ├── executor.go              # 송금 작업 생성/관리
│       └── state.go                 # 송금 상태 머신 (7단계)
├── config.toml                      # 설정 파일 (심볼, 거래소 옵션, 대시보드)
├── .env                             # API 키, 프록시 설정 (gitignore 대상)
├── go.mod
└── go.sum
```

**총 26개 Go 소스 파일, 4,899줄**

---

## 설치 및 실행

### 요구사항
- Go 1.24+
- Python 3.11+ (TUI 프론트엔드)

### 빌드

```bash
# 클론
git clone https://github.com/hyeokx/Go_GAPTUI.git
cd Go_GAPTUI

# 의존성 설치 및 빌드
go mod download
go build -o kimchi ./cmd/kimchi/
```

### 설정

1. **config.toml** 생성:

```toml
symbols = ["BTC", "ETH", "XRP", "SOL", "DOGE"]
usdt_symbol = "USDT"
channel_size = 1024
auto_discover_symbols = true

[exchange]
poll_interval_ms = 1000
request_timeout_secs = 10

[dashboard]
refresh_ms = 200
stale_display_secs = 30
```

2. **.env** 생성 (API 키 설정):

```env
# 업비트
UPBIT_ACCESS_KEY=your_key
UPBIT_SECRET_KEY=your_secret

# 빗썸
BITHUMB_ACCESS_KEY=your_key
BITHUMB_SECRET_KEY=your_secret

# 바이낸스
BINANCE_ACCESS_KEY=your_key
BINANCE_SECRET_KEY=your_secret

# 바이빗
BYBIT_ACCESS_KEY=your_key
BYBIT_SECRET_KEY=your_secret

# OKX
OKX_ACCESS_KEY=your_key
OKX_SECRET_KEY=your_secret
OKX_PASSPHRASE=your_passphrase

# 비트겟
BITGET_ACCESS_KEY=your_key
BITGET_SECRET_KEY=your_secret
BITGET_PASSPHRASE=your_passphrase

# 프록시 (선택)
# PROXY_HOST=127.0.0.1
# PROXY_PORT=1080
```

### 실행

```bash
# Go 엔진 실행
./kimchi

# 레이스 감지 모드 (개발 시 권장)
go run -race ./cmd/kimchi/
```

---

## 지원 거래소

| 거래소 | 코드 | 유형 | Ticker WS | Orderbook WS | API 인증 |
|--------|------|------|:---------:|:------------:|:--------:|
| 업비트 | UP | 국내 | ✅ | ✅ | JWT (HS256) |
| 빗썸 | BT | 국내 | ✅ | ✅ | JWT (HS256) |
| 바이낸스 | BN | 해외 | ✅ | ✅ | HMAC-SHA256 |
| 바이빗 | BB | 해외 | ✅ | ✅ | HMAC-SHA256 |
| OKX | OK | 해외 | ✅ | ✅ | HMAC-SHA256 |
| 비트겟 | BG | 해외 | — | — | HMAC-SHA256 |
| 바이낸스선물 | BNF | 선물 | ✅ | — | — |
| 바이빗선물 | BBF | 선물 | ✅ | — | — |
| 게이트 | GT | 해외 | — | — | HMAC-SHA512 |

---

## 백그라운드 작업 주기

| 작업 | 주기 |
|------|------|
| Ticker WebSocket (7개) | 상시 연결, 3초 재연결 |
| Orderbook WebSocket (5개) | 상시 연결, 3초 재연결 |
| Orderbook 일괄 처리 | 100ms |
| IPC 스냅샷 브로드캐스트 | 200ms |
| 환율 조회 (USDT/KRW) | 1초 |
| 지갑 상태 조회 | 60초 |
| 심볼 자동 탐색 | 시작 시 + 5분 주기 |

---

## 라이선스

Private repository.
