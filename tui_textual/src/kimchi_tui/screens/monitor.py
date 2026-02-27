from __future__ import annotations

from typing import Any, cast

from textual import events
from textual.containers import Horizontal, Vertical
from textual.widgets import Button, Input, RichLog, Static, TabbedContent

from ..colors import C
from ..formatters import fmt_rate_detailed
from ..indicators import dw_dots, log_level_badge
from ..models import Snapshot
from ..widgets.card_view import CardView
from ..widgets.coin_table import CoinTable
from ..widgets.detail_panel import CoinDetailPanel
from ..widgets.filter_bar import FilterBar
from ..widgets.header_bar import KimchiHeader
from ..widgets.scenario_panel import ScenarioPanel


class MonitorScreen(Vertical):
    def __init__(self) -> None:
        super().__init__(id="monitor-screen")
        self.filter_state = FilterBar.State()
        self.favorites: set[str] = set()
        self.expanded_coin: str | None = None
        self.card_view: bool = False
        self.exchange_a: str = "BT"
        self.exchange_b: str = "UP"
        self._snapshot = Snapshot()
        self._last_recent_log_lines: list[str] = []

    def compose(self):
        yield Static("USDT/KRW: - | USD/KRW: - | Ticker age: -", id="rate-bar")
        yield FilterBar()
        with Horizontal(id="monitor-main"):
            with Vertical(id="monitor-left"):
                yield CoinTable(id="coin-table")
                yield CardView()
                yield CoinDetailPanel()
            with Vertical(id="monitor-right"):
                yield ScenarioPanel(id="monitor-scenario")
                with Vertical(id="monitor-dw"):
                    yield Static("D/W Status", classes="panel-title")
                    yield Static("", id="dw-status-content")
                with Vertical(id="monitor-logs"):
                    yield Static("Recent Logs", classes="panel-title")
                    yield RichLog(
                        id="recent-log-view", highlight=True, wrap=True, markup=True
                    )

    def on_mount(self) -> None:
        self.query_one("#card-view", CardView).display = False
        self.query_one("#coin-detail-panel", CoinDetailPanel).display = False

    def on_key(self, event: events.Key) -> None:
        filter_bar = self.query_one(FilterBar)
        header = self.app.query_one(KimchiHeader)
        coin_table = self.query_one("#coin-table", CoinTable)
        focused = self.app.focused
        search_focused = isinstance(focused, Input) and focused.id == "header-search"
        key = event.key
        char = event.character or ""

        if key in {"escape", "esc"}:
            header.clear_search()
            self.expanded_coin = None
            self.update_data(self._snapshot)
            event.stop()
            return

        if search_focused:
            return

        if char == "j" and not self.card_view:
            coin_table.action_cursor_down()
            event.stop()
            return
        if char == "k" and not self.card_view:
            coin_table.action_cursor_up()
            event.stop()
            return
        if char == "g" and not self.card_view:
            if coin_table.row_count > 0:
                coin_table.move_cursor(row=0, column=0)
            event.stop()
            return
        if char == "G" and not self.card_view:
            if coin_table.row_count > 0:
                coin_table.move_cursor(row=coin_table.row_count - 1, column=0)
            event.stop()
            return
        if char == "o":
            filter_bar.cycle_sort_column()
            event.stop()
            return
        if char == "O":
            filter_bar.toggle_sort_direction()
            event.stop()
            return
        if char == "e" and not self.card_view:
            symbol = coin_table.selected_symbol()
            if symbol is not None:
                if self.expanded_coin == symbol:
                    self.expanded_coin = None
                else:
                    self.expanded_coin = symbol
                self.update_data(self._snapshot)
            event.stop()
            return
        if char == "x":
            self.exchange_a, self.exchange_b = self.exchange_b, self.exchange_a
            header.exchange_a = self.exchange_a
            header.exchange_b = self.exchange_b
            header.query_one(
                "#header-exchange-a", Button
            ).label = f"A: {self.exchange_a}"
            header.query_one(
                "#header-exchange-b", Button
            ).label = f"B: {self.exchange_b}"
            self.update_data(self._snapshot)
            event.stop()
            return
        if char == "d":
            filter_bar.toggle_dw_only()
            event.stop()
            return
        if char == "f":
            symbol = coin_table.selected_symbol()
            if symbol:
                if symbol in self.favorites:
                    self.favorites.remove(symbol)
                else:
                    self.favorites.add(symbol)
                self.update_data(self._snapshot)
            event.stop()
            return
        if char == "v":
            self.card_view = not self.card_view
            self.update_data(self._snapshot)
            event.stop()
            return
        if char == "/":
            header.focus_search()
            event.stop()
            return
        if key == "enter" and not self.card_view:
            event.stop()

    def on_filter_bar_changed(self, event: FilterBar.Changed) -> None:
        query = self.filter_state.query
        self.filter_state = event.state
        self.filter_state.query = query
        app = cast(Any, self.app)
        self.update_data(app.snapshot)

    def on_header_search_changed(self, event: KimchiHeader.SearchChanged) -> None:
        self.filter_state.query = event.query
        app = cast(Any, self.app)
        self.update_data(app.snapshot)

    def on_header_exchanges_changed(self, event: KimchiHeader.ExchangesChanged) -> None:
        self.exchange_a = event.exchange_a
        self.exchange_b = event.exchange_b
        app = cast(Any, self.app)
        self.update_data(app.snapshot)

    def on_header_transfer_pressed(self, _: KimchiHeader.TransferPressed) -> None:
        tabs = self.app.query_one("#main-tabs", TabbedContent)
        tabs.active = "transfer-tab"

    def update_data(self, snapshot: Snapshot) -> None:
        self._snapshot = snapshot
        usdt = fmt_rate_detailed(snapshot.usdt_krw)
        usd = fmt_rate_detailed(snapshot.usd_krw_forex)
        age = (
            f"{snapshot.last_ticker_age_ms}ms"
            if snapshot.last_ticker_age_ms is not None
            else "-"
        )
        self.query_one("#rate-bar", Static).update(
            f"USDT/KRW: [bold]{usdt}[/] | USD/KRW: [bold]{usd}[/] | Ticker age: [bold]{age}[/]"
        )

        coin_table = self.query_one("#coin-table", CoinTable)
        coin_table.update_rows(
            snapshot,
            query=self.filter_state.query,
            dw_only=self.filter_state.dw_only,
            exchange_filter=self.filter_state.exchange_filter,
            sort_column=self.filter_state.sort_column,
            sort_desc=self.filter_state.sort_desc,
            favorites=self.favorites,
            exchange_a=self.exchange_a,
            exchange_b=self.exchange_b,
        )

        card_view = self.query_one("#card-view", CardView)
        card_view.update_cards(
            snapshot,
            query=self.filter_state.query,
            dw_only=self.filter_state.dw_only,
            exchange_filter=self.filter_state.exchange_filter,
            sort_column=self.filter_state.sort_column,
            sort_desc=self.filter_state.sort_desc,
            favorites=self.favorites,
        )

        if self.expanded_coin and not coin_table.has_symbol(self.expanded_coin):
            self.expanded_coin = None

        coin_table.display = not self.card_view
        card_view.display = self.card_view

        detail_panel = self.query_one("#coin-detail-panel", CoinDetailPanel)
        detail_panel.display = bool(self.expanded_coin) and not self.card_view
        if self.expanded_coin and detail_panel.display:
            detail_panel.update_detail(self.expanded_coin, snapshot)

        scenario_panel = self.query_one("#monitor-scenario", ScenarioPanel)
        scenario_panel.update_threads(snapshot)

        dw_widget = self.query_one("#dw-status-content", Static)
        blocked_lines: list[str] = []
        for symbol, status in list(snapshot.wallet_status.items())[:60]:
            cells = []
            has_blocked = False
            for name, ex in [
                ("UP", status.upbit),
                ("BT", status.bithumb),
                ("BN", status.binance),
                ("BB", status.bybit),
                ("OK", status.okx),
            ]:
                if ex is None:
                    cells.append(f"{name} [{C.TEXT_MUTED}]-[/]")
                else:
                    if not ex.deposit or not ex.withdraw:
                        has_blocked = True
                    dot_str = dw_dots(ex.deposit, ex.withdraw)
                    blocked_chains: list[str] = []
                    if not ex.deposit and ex.deposit_blocked_chains:
                        blocked_chains.extend(ex.deposit_blocked_chains[:2])
                    if not ex.withdraw and ex.withdraw_blocked_chains:
                        blocked_chains.extend(ex.withdraw_blocked_chains[:2])
                    chain_text = ""
                    if blocked_chains:
                        chain_badges = " ".join(
                            f"[bold white on {C.DW_CHAIN_BADGE_BG}]{c}[/]"
                            for c in blocked_chains[:2]
                        )
                        chain_text = f" {chain_badges}"
                    cells.append(f"{name} {dot_str}{chain_text}")
            if has_blocked:
                blocked_lines.append(f"[bold]{symbol:<8}[/] " + "  ".join(cells))
        dw_widget.update(
            "\n".join(blocked_lines)
            if blocked_lines
            else f"[{C.TEXT_MUTED}]No wallet status yet[/]"
        )

        log_view = self.query_one("#recent-log-view", RichLog)
        current_lines: list[str] = []
        for entry in snapshot.logs[-80:]:
            badge = log_level_badge(entry.log_type)
            current_lines.append(
                f"[{C.TEXT_SECONDARY}]{entry.timestamp}[/] {badge} [bold]{entry.symbol}[/] {entry.message}",
            )
        if not current_lines:
            if self._last_recent_log_lines:
                log_view.clear()
                self._last_recent_log_lines = []
            return

        if (
            current_lines[: len(self._last_recent_log_lines)]
            == self._last_recent_log_lines
        ):
            for line in current_lines[len(self._last_recent_log_lines) :]:
                log_view.write(line)
        else:
            log_view.clear()
            for line in current_lines:
                log_view.write(line)
        self._last_recent_log_lines = current_lines
