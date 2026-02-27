from __future__ import annotations

import re

from rich.markup import escape
from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical, VerticalScroll
from textual.events import Click
from textual.message import Message
from textual.widgets import Button, Static

from ..colors import C, SCENARIO_BADGE_LABELS, get_kimchi_color, scenario_color
from ..models import CoinWalletStatus, ExchangeWalletStatus, LogThread, Snapshot


class ScenarioThreadCard(Static):
    thread_id: int
    closed: bool

    def __init__(self, thread_id: int, content: str, *, closed: bool) -> None:
        super().__init__(content, markup=True)
        self.thread_id = thread_id
        self.closed = closed

    def on_click(self, _: Click) -> None:
        parent = self.parent
        while parent is not None and not isinstance(parent, ScenarioPanel):
            parent = parent.parent
        if isinstance(parent, ScenarioPanel):
            parent.toggle_thread(self.thread_id)


class ScenarioPanel(Vertical):
    FILTER_BUTTON_TO_SCENARIO: dict[str, str] = {
        "scenario-kimp": "GapThreshold",
        "scenario-dom": "DomesticGap",
        "scenario-fut": "FutBasis",
    }

    class FilterChanged(Message):
        filters: set[str]
        sender: "ScenarioPanel"

        def __init__(self, sender: "ScenarioPanel", filters: set[str]) -> None:
            self.filters = filters
            self.sender = sender
            super().__init__()

    def __init__(self, id: str | None = None) -> None:
        super().__init__(id=id)
        self.active_filters: set[str] = {"GapThreshold", "DomesticGap", "FutBasis"}
        self.expanded_threads: set[int] = set()
        self._snapshot: Snapshot = Snapshot()

    def compose(self) -> ComposeResult:
        with Horizontal(id="scenario-header"):
            yield Static("Scenarios", classes="panel-title")
            yield Static("0 active / 0", id="scenario-count")
        with Horizontal(id="scenario-filter-row"):
            yield Button("KIMP", id="scenario-kimp", classes="scenario-filter-btn")
            yield Static("5.0%", id="scenario-threshold-kimp")
            yield Button("DOM-GAP", id="scenario-dom", classes="scenario-filter-btn")
            yield Static("1.5%", id="scenario-threshold-dom")
            yield Button("FUT%", id="scenario-fut", classes="scenario-filter-btn")
            yield Static("0.5%", id="scenario-threshold-fut")
        yield VerticalScroll(id="scenario-log")

    def on_mount(self) -> None:
        self._sync_buttons()
        self._refresh()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id not in self.FILTER_BUTTON_TO_SCENARIO:
            return

        scenario = self.FILTER_BUTTON_TO_SCENARIO[event.button.id]
        if scenario in self.active_filters:
            self.active_filters.remove(scenario)
        else:
            self.active_filters.add(scenario)

        self._sync_buttons()
        _ = self.post_message(self.FilterChanged(self, set(self.active_filters)))
        self._refresh()

    def toggle_thread(self, thread_id: int) -> None:
        if thread_id in self.expanded_threads:
            self.expanded_threads.remove(thread_id)
        else:
            self.expanded_threads.add(thread_id)
        self._refresh()

    def update_threads(self, snapshot: Snapshot) -> None:
        self._snapshot = snapshot
        valid_ids = {thread.id for thread in snapshot.scenario_threads}
        self.expanded_threads.intersection_update(valid_ids)
        self._refresh()

    def _sync_buttons(self) -> None:
        for button_id, scenario in self.FILTER_BUTTON_TO_SCENARIO.items():
            button = self.query_one(f"#{button_id}", Button)
            active = scenario in self.active_filters
            button.styles.background = (
                scenario_color(scenario) if active else C.INACTIVE_BG
            )
            button.styles.color = C.PILL_TEXT if active else C.INACTIVE_TEXT
            button.styles.text_style = "bold"

    def _refresh(self) -> None:
        self._update_thresholds()
        threads = self._filtered_threads(
            self._snapshot.scenario_threads, self._snapshot
        )
        active_threads = sorted(
            (thread for thread in threads if thread.is_active),
            key=lambda thread: thread.main_timestamp,
            reverse=True,
        )
        closed_threads = sorted(
            (thread for thread in threads if not thread.is_active),
            key=lambda thread: thread.main_timestamp,
            reverse=True,
        )

        total_threads = len(active_threads) + len(closed_threads)
        self.query_one("#scenario-count", Static).update(
            f"{len(active_threads)} active / {total_threads}"
        )

        visible_active = active_threads[:100]
        visible_closed = closed_threads[: max(0, 100 - len(visible_active))]

        container = self.query_one("#scenario-log", VerticalScroll)
        _ = container.remove_children()

        if not visible_active and not visible_closed:
            _ = container.mount(
                Static(f"[{C.TEXT_MUTED}]No scenario alerts yet[/]", markup=True)
            )
            return

        for thread in visible_active:
            _ = container.mount(
                ScenarioThreadCard(
                    thread.id, self._render_thread(thread, muted=False), closed=False
                )
            )

        if visible_closed:
            _ = container.mount(Static(f"[{C.TEXT_MUTED}]── Closed ──[/]", markup=True))
            for thread in visible_closed:
                _ = container.mount(
                    ScenarioThreadCard(
                        thread.id, self._render_thread(thread, muted=True), closed=True
                    )
                )

    def _update_thresholds(self) -> None:
        config = self._snapshot.scenario_config
        self.query_one("#scenario-threshold-kimp", Static).update(
            f"KIMP {config.gap_threshold_percent:.1f}%"
        )
        self.query_one("#scenario-threshold-dom", Static).update(
            f"DOM-GAP {config.domestic_gap_threshold:.1f}%"
        )
        self.query_one("#scenario-threshold-fut", Static).update(
            f"FUT% {config.fut_basis_threshold:.1f}%"
        )

    def _filtered_threads(
        self, threads: list[LogThread], snapshot: Snapshot
    ) -> list[LogThread]:
        filtered: list[LogThread] = []
        for thread in threads:
            if thread.scenario not in self.active_filters:
                continue
            if thread.is_active and not self._wallet_allows_thread(thread, snapshot):
                continue
            filtered.append(thread)
        return filtered

    def _wallet_allows_thread(self, thread: LogThread, snapshot: Snapshot) -> bool:
        if thread.scenario not in {"GapThreshold", "DomesticGap"}:
            return True

        wallet = snapshot.wallet_status.get(thread.symbol)
        if wallet is None:
            return False

        if thread.scenario == "GapThreshold":
            domestic_code, overseas_code = self._parse_gap_key(thread.key)
            if domestic_code is None or overseas_code is None:
                return False
            domestic = self._exchange_wallet(wallet, domestic_code)
            overseas = self._exchange_wallet(wallet, overseas_code)
            return bool(
                domestic is not None
                and domestic.deposit
                and overseas is not None
                and overseas.withdraw
            )

        if thread.key == "pos":
            source = wallet.bithumb
            target = wallet.upbit
        elif thread.key == "neg":
            source = wallet.upbit
            target = wallet.bithumb
        else:
            return False
        return bool(
            source is not None
            and source.withdraw
            and target is not None
            and target.deposit
        )

    def _parse_gap_key(self, key: str) -> tuple[str | None, str | None]:
        if "-" not in key:
            return None, None
        domestic_code, overseas_code = key.split("-", maxsplit=1)
        return domestic_code.upper(), overseas_code.upper()

    def _exchange_wallet(
        self, wallet: CoinWalletStatus, exchange_code: str
    ) -> ExchangeWalletStatus | None:
        mapping = {
            "UP": wallet.upbit,
            "BT": wallet.bithumb,
            "BN": wallet.binance,
            "BB": wallet.bybit,
            "OK": wallet.okx,
        }
        return mapping.get(exchange_code.upper())

    def _render_thread(self, thread: LogThread, *, muted: bool) -> str:
        badge_text = SCENARIO_BADGE_LABELS.get(thread.scenario, thread.scenario or "-")
        badge_bg = scenario_color(thread.scenario) if not muted else C.INACTIVE_BG
        badge = f"[bold {C.PILL_TEXT} on {badge_bg}] {escape(badge_text)} [/]"

        gap_bg = (
            get_kimchi_color(thread.last_logged_value) if not muted else C.INACTIVE_BG
        )
        gap_fg = C.PILL_TEXT if not muted else C.INACTIVE_TEXT
        gap_text = f"{thread.last_logged_value:+.1f}%"
        gap_pill = f"[bold {gap_fg} on {gap_bg}] {gap_text} [/]"

        symbol = thread.symbol or "-"
        korean_name = self._snapshot.korean_names.get(symbol, "")
        symbol_label = f"{symbol}: {korean_name}" if korean_name else symbol

        has_sub_entries = len(thread.sub_entries)
        is_expanded = thread.id in self.expanded_threads
        arrow = "▼" if is_expanded else "▶"

        bar_color = scenario_color(thread.scenario) if not muted else C.TEXT_MUTED
        symbol_color = C.TEXT_SECONDARY if muted else C.TEXT_PRIMARY
        message_color = C.TEXT_SECONDARY if muted else C.TEXT_PRIMARY

        lines = [
            (
                f"[{bar_color}]█[/] {arrow} {has_sub_entries}  "
                f"[{C.TEXT_SECONDARY}]{self._format_timestamp(thread.main_timestamp)}[/]  "
                f"{badge}  [bold {symbol_color}]{escape(symbol_label)}[/]  {gap_pill}"
            ),
            f"   [{message_color}]{escape(thread.main_message or '-')}[/]",
        ]

        if is_expanded:
            for sub_entry in thread.sub_entries:
                sub_line = f"   [{C.TEXT_MUTED}]└[/] [{C.TEXT_SECONDARY}]{self._format_timestamp(sub_entry.timestamp)}[/] [{message_color}]{escape(sub_entry.message or '-')}[/]"
                lines.append(sub_line)

        return "\n".join(lines)

    def _format_timestamp(self, raw: str) -> str:
        if not raw:
            return "--:--:--"
        match = re.search(r"(\d{2}:\d{2}:\d{2})", raw)
        if match:
            return match.group(1)
        return raw
