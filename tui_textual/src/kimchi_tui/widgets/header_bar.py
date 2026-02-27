from __future__ import annotations

from textual import on
from textual.containers import Horizontal, Vertical
from textual.message import Message
from textual.widgets import Button, Input, Static

from ..colors import C, get_kimchi_color
from ..formatters import fmt_rate
from ..models import Snapshot


class KimchiHeader(Vertical):
    SELECTABLE_EXCHANGES: list[str] = ["UP", "BT", "BN", "BB", "OK"]

    class SearchChanged(Message):
        sender: "KimchiHeader"
        query: str

        def __init__(self, sender: "KimchiHeader", query: str) -> None:
            self.sender = sender
            self.query = query
            super().__init__()

    class ExchangesChanged(Message):
        sender: "KimchiHeader"
        exchange_a: str
        exchange_b: str

        def __init__(
            self, sender: "KimchiHeader", exchange_a: str, exchange_b: str
        ) -> None:
            self.sender = sender
            self.exchange_a = exchange_a
            self.exchange_b = exchange_b
            super().__init__()

    class TransferPressed(Message):
        sender: "KimchiHeader"

        def __init__(self, sender: "KimchiHeader") -> None:
            self.sender = sender
            super().__init__()

    def __init__(self) -> None:
        super().__init__(id="kimchi-header")
        self.exchange_a: str = "BT"
        self.exchange_b: str = "UP"

    def compose(self):
        with Horizontal(id="kimchi-header-row1"):
            yield Static("KIMCHI TERMINAL", id="header-title")
            yield Static("", id="header-tether-pill")
            yield Static("", id="header-row1-spacer")
            yield Static("", id="header-usdt-rate")
            yield Static("", id="header-usd-rate")
            yield Static("", id="header-coin-count")
        with Horizontal(id="kimchi-header-row2"):
            yield Input(placeholder="Search symbol / coin name", id="header-search")
            yield Static("", id="header-budget")
            yield Button("A: BT", id="header-exchange-a", classes="header-btn")
            yield Button("⇄", id="header-swap", classes="header-btn")
            yield Button("B: UP", id="header-exchange-b", classes="header-btn")
            yield Button("◐", id="header-theme", classes="header-btn")
            yield Button(
                "Transfer", id="header-transfer", classes="header-transfer-btn"
            )

    def update_data(self, snapshot: Snapshot, active_tab: str, dark_mode: bool) -> None:
        self.query_one("#header-tether-pill", Static).update(
            self._render_tether_premium(snapshot)
        )
        self.query_one("#header-usdt-rate", Static).update(
            f"[{C.TEXT_SECONDARY}]USDT[/] [{C.TEXT_PRIMARY}]KRW {fmt_rate(snapshot.usdt_krw)}[/]"
        )
        self.query_one("#header-usd-rate", Static).update(
            f"[{C.TEXT_SECONDARY}]USD[/] [{C.TEXT_PRIMARY}]KRW {fmt_rate(snapshot.usd_krw_forex)}[/]"
        )
        self.query_one("#header-coin-count", Static).update(
            f"[{C.TEXT_SECONDARY}]{len(snapshot.coin_states)} coins[/]"
        )
        self.query_one("#header-budget", Static).update(
            f"[{C.TEXT_SECONDARY}]Budget:[/] [{C.TEXT_PRIMARY}]{self._extract_budget_man(snapshot)}만[/]"
        )
        self.query_one("#header-exchange-a", Button).label = f"A: {self.exchange_a}"
        self.query_one("#header-exchange-b", Button).label = f"B: {self.exchange_b}"
        self.query_one("#header-theme", Button).label = (
            "◐ Dark" if dark_mode else "☀ Light"
        )
        transfer = self.query_one("#header-transfer", Button)
        transfer.variant = "default" if active_tab == "transfer-tab" else "primary"

    def clear_search(self) -> None:
        self.query_one("#header-search", Input).value = ""

    def focus_search(self) -> None:
        _ = self.query_one("#header-search", Input).focus()

    @on(Input.Changed, "#header-search")
    def _on_search_changed(self, event: Input.Changed) -> None:
        _ = self.post_message(self.SearchChanged(self, event.value.strip().upper()))

    @on(Button.Pressed, "#header-exchange-a")
    def _on_exchange_a_pressed(self, event: Button.Pressed) -> None:
        del event
        self.exchange_a = self._next_exchange(self.exchange_a)
        self.query_one("#header-exchange-a", Button).label = f"A: {self.exchange_a}"
        _posted = self.post_message(
            self.ExchangesChanged(self, self.exchange_a, self.exchange_b)
        )
        del _posted

    @on(Button.Pressed, "#header-exchange-b")
    def _on_exchange_b_pressed(self, event: Button.Pressed) -> None:
        del event
        self.exchange_b = self._next_exchange(self.exchange_b)
        self.query_one("#header-exchange-b", Button).label = f"B: {self.exchange_b}"
        _posted = self.post_message(
            self.ExchangesChanged(self, self.exchange_a, self.exchange_b)
        )
        del _posted

    @on(Button.Pressed, "#header-swap")
    def _on_swap_pressed(self, event: Button.Pressed) -> None:
        del event
        self.exchange_a, self.exchange_b = self.exchange_b, self.exchange_a
        self.query_one("#header-exchange-a", Button).label = f"A: {self.exchange_a}"
        self.query_one("#header-exchange-b", Button).label = f"B: {self.exchange_b}"
        _posted = self.post_message(
            self.ExchangesChanged(self, self.exchange_a, self.exchange_b)
        )
        del _posted

    @on(Button.Pressed, "#header-theme")
    def _on_theme_pressed(self, event: Button.Pressed) -> None:
        del event
        app = self.app
        action = getattr(app, "action_toggle_theme", None)
        if callable(action):
            action()

    @on(Button.Pressed, "#header-transfer")
    def _on_transfer_pressed(self, event: Button.Pressed) -> None:
        del event
        _posted = self.post_message(self.TransferPressed(self))
        del _posted

    def _render_tether_premium(self, snapshot: Snapshot) -> str:
        usdt = snapshot.usdt_krw
        usd = snapshot.usd_krw_forex
        if usdt is None or usd is None or usd == 0:
            return f"[bold {C.INACTIVE_TEXT} on {C.INACTIVE_BG}] Tether -- [/]"
        premium = (usdt / usd - 1.0) * 100
        bg_color = get_kimchi_color(premium)
        return f"[bold {C.PILL_TEXT} on {bg_color}] Tether {premium:+.2f}% [/]"

    def _next_exchange(self, current: str) -> str:
        if current not in self.SELECTABLE_EXCHANGES:
            return self.SELECTABLE_EXCHANGES[0]
        index = self.SELECTABLE_EXCHANGES.index(current)
        return self.SELECTABLE_EXCHANGES[(index + 1) % len(self.SELECTABLE_EXCHANGES)]

    def _extract_budget_man(self, snapshot: Snapshot) -> str:
        budget_man = 1000.0
        values: list[object] = [
            getattr(snapshot, "budget_man", None),
            getattr(snapshot, "budget", None),
            getattr(snapshot, "budget_krw", None),
            getattr(snapshot.transfer, "budget_man", None),
            getattr(snapshot.transfer, "budget", None),
            getattr(snapshot.transfer, "budget_krw", None),
        ]
        for value in values:
            parsed = self._to_float(value)
            if parsed is None or parsed <= 0:
                continue
            budget_man = parsed / 10000.0 if parsed > 10000 else parsed
            break
        return f"{budget_man:,.0f}" if budget_man.is_integer() else f"{budget_man:,.1f}"

    def _to_float(self, value: object) -> float | None:
        if isinstance(value, (int, float)):
            return float(value)
        if isinstance(value, str):
            try:
                return float(value)
            except ValueError:
                return None
        return None
