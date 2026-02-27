from __future__ import annotations

from dataclasses import dataclass

from textual import on
from textual.containers import Horizontal
from textual.message import Message
from textual.widgets import Button, Static


class FilterBar(Horizontal):
    SORT_COLUMNS = ["Symbol", "Upbit", "Bithumb", "Binance", "Bybit", "Kimp"]

    @dataclass
    class State:
        query: str = ""
        dw_only: bool = False
        exchange_filter: str = "ALL"
        sort_column: str = "Kimp"
        sort_desc: bool = True

    class Changed(Message):
        def __init__(self, sender: "FilterBar", state: "FilterBar.State") -> None:
            self.state = state
            self.sender = sender
            super().__init__()

    EXCHANGE_ROTATION = ["ALL", "UP", "BT", "BN", "BB", "OK"]

    def __init__(self) -> None:
        super().__init__(id="filter-bar")
        self.state = FilterBar.State()

    def compose(self):
        yield Static("Filter", classes="filter-label")
        yield Button("D/W: OFF", id="filter-dw", classes="filter-btn")
        yield Button("EX: ALL", id="filter-ex", classes="filter-btn")
        yield Button("SORT: KIMP DESC", id="filter-sort", classes="filter-btn")
        yield Button("DIR: DESC", id="filter-dir", classes="filter-btn")

    def _emit_changed(self) -> None:
        self.post_message(self.Changed(self, self.state))

    def _refresh_sort_labels(self) -> None:
        direction = "DESC" if self.state.sort_desc else "ASC"
        self.query_one(
            "#filter-sort", Button
        ).label = f"SORT: {self.state.sort_column.upper()} {direction}"
        self.query_one("#filter-dir", Button).label = f"DIR: {direction}"

    def cycle_sort_column(self) -> None:
        idx = self.SORT_COLUMNS.index(self.state.sort_column)
        self.state.sort_column = self.SORT_COLUMNS[(idx + 1) % len(self.SORT_COLUMNS)]
        self._refresh_sort_labels()
        self._emit_changed()

    def toggle_sort_direction(self) -> None:
        self.state.sort_desc = not self.state.sort_desc
        self._refresh_sort_labels()
        self._emit_changed()

    def toggle_dw_only(self) -> None:
        self.state.dw_only = not self.state.dw_only
        self.query_one(
            "#filter-dw", Button
        ).label = f"D/W: {'ON' if self.state.dw_only else 'OFF'}"
        self._emit_changed()

    def clear_search(self) -> None:
        self.state.query = ""
        self._emit_changed()

    @on(Button.Pressed, "#filter-dw")
    def _on_dw_pressed(self, _: Button.Pressed) -> None:
        self.toggle_dw_only()

    @on(Button.Pressed, "#filter-ex")
    def _on_ex_pressed(self, _: Button.Pressed) -> None:
        idx = self.EXCHANGE_ROTATION.index(self.state.exchange_filter)
        self.state.exchange_filter = self.EXCHANGE_ROTATION[
            (idx + 1) % len(self.EXCHANGE_ROTATION)
        ]
        self.query_one("#filter-ex", Button).label = f"EX: {self.state.exchange_filter}"
        self._emit_changed()

    @on(Button.Pressed, "#filter-sort")
    def _on_sort_pressed(self, _: Button.Pressed) -> None:
        self.cycle_sort_column()

    @on(Button.Pressed, "#filter-dir")
    def _on_sort_dir_pressed(self, _: Button.Pressed) -> None:
        self.toggle_sort_direction()
