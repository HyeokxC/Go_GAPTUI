# Kimchi Textual TUI

Textual-based terminal UI for monitoring kimchi premium from a Rust backend over WebSocket IPC.

## Features

- Real-time monitor table with filtering and kimchi/domestic-gap highlighting
- Transfer workflow panel with exchange/network selectors and command dispatch
- Scenario threads, D/W status matrix, and full logs viewer
- Dark/light theme toggle (`l`)

## Run

```bash
python -m venv .venv
.venv/bin/pip install -e .
python -m kimchi_tui
```

Or run directly with Textual:

```bash
textual run kimchi_tui.app:KimchiApp
```

IPC endpoint defaults to `ws://127.0.0.1:9876`.
