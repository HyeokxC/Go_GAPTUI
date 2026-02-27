from __future__ import annotations

from typing import override

from rich.markup import escape
from textual import events
from textual.containers import Vertical, VerticalScroll
from textual.widgets import Static

from ..colors import C
from ..models import CoinWalletStatus, Snapshot


class DWStatusScreen(Vertical):
    def __init__(self) -> None:
        super().__init__()
        self._snapshot: Snapshot = Snapshot()
        self._blocked_symbols: list[str] = []
        self._cursor_index: int = 0

    @override
    def compose(self):
        yield Static("D/W Status", classes="panel-title", id="dw-title")
        yield VerticalScroll(id="dw-list")

    def on_key(self, event: events.Key) -> None:
        char = event.character or ""
        scroll = self.query_one("#dw-list", VerticalScroll)

        if char == "j":
            if self._blocked_symbols:
                self._cursor_index = min(
                    self._cursor_index + 1, len(self._blocked_symbols) - 1
                )
                self.update_data(self._snapshot)
                scroll.scroll_down()
            else:
                scroll.scroll_down()
            _ = event.stop()
            return

        if char == "k":
            if self._blocked_symbols:
                self._cursor_index = max(self._cursor_index - 1, 0)
                self.update_data(self._snapshot)
                scroll.scroll_up()
            else:
                scroll.scroll_up()
            _ = event.stop()
            return

    def update_data(self, snapshot: Snapshot) -> None:
        self._snapshot = snapshot
        blocked_items: list[tuple[str, CoinWalletStatus]] = []
        for symbol in sorted(snapshot.wallet_status.keys()):
            status = snapshot.wallet_status[symbol]
            if self._coin_has_blocked(status):
                blocked_items.append((symbol, status))

        blocked_count = len(blocked_items)
        self._blocked_symbols = [symbol for symbol, _ in blocked_items]
        if self._blocked_symbols:
            self._cursor_index = min(self._cursor_index, len(self._blocked_symbols) - 1)
        else:
            self._cursor_index = 0

        title_color = C.RED if blocked_count > 0 else C.GREEN
        self.query_one("#dw-title", Static).update(
            f"[bold {title_color}]D/W Status | Blocked: {blocked_count}[/]"
        )

        container = self.query_one("#dw-list", VerticalScroll)
        _ = container.remove_children()

        if not blocked_items:
            _ = container.mount(
                Static(f"[{C.TEXT_MUTED}]All exchanges D/W available[/]", markup=True)
            )
            return

        for index, (symbol, status) in enumerate(blocked_items):
            _ = container.mount(
                Static(
                    self._render_coin_block(
                        symbol=symbol,
                        status=status,
                        is_selected=index == self._cursor_index,
                    ),
                    markup=True,
                )
            )

    @staticmethod
    def _coin_has_blocked(status: CoinWalletStatus) -> bool:
        return any(
            exchange is not None and (not exchange.deposit or not exchange.withdraw)
            for exchange in (
                status.upbit,
                status.bithumb,
                status.binance,
                status.bybit,
                status.okx,
            )
        )

    def _render_coin_block(
        self, *, symbol: str, status: CoinWalletStatus, is_selected: bool
    ) -> str:
        symbol_line = f"[bold {C.ACCENT}]◉ {escape(symbol)}[/]"
        if is_selected:
            symbol_line = f"[reverse]{symbol_line}[/]"

        lines = [symbol_line]
        for exchange_name, ex_status in (
            ("Upbit", status.upbit),
            ("Bithumb", status.bithumb),
            ("Binance", status.binance),
            ("Bybit", status.bybit),
            ("OKX", status.okx),
        ):
            d_status = self._dw_status_text(
                prefix="D",
                is_ok=ex_status.deposit if ex_status else None,
                chains=ex_status.deposit_blocked_chains if ex_status else [],
            )
            w_status = self._dw_status_text(
                prefix="W",
                is_ok=ex_status.withdraw if ex_status else None,
                chains=ex_status.withdraw_blocked_chains if ex_status else [],
            )
            lines.append(
                f"  [{C.TEXT_SECONDARY}]{exchange_name:<8}[/]  {d_status}  {w_status}"
            )

        return "\n".join(lines)

    def _dw_status_text(
        self, *, prefix: str, is_ok: bool | None, chains: list[str]
    ) -> str:
        if is_ok is None:
            return f"[{C.TEXT_MUTED}]{prefix}:-[/]"
        if is_ok:
            return f"[{C.GREEN}]{prefix}:ok[/]"

        chain_names = ", ".join(escape(chain) for chain in chains)
        chain_suffix = f" [{chain_names}]" if chain_names else ""
        return f"[bold white on {C.RED}] {prefix}:blocked{chain_suffix} [/]"
