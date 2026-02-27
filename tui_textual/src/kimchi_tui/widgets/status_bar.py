from __future__ import annotations

from textual.widgets import Static

from ..colors import C


TAB_HELP: dict[str, list[tuple[str, str]]] = {
    "monitor-tab": [
        ("j/k", "Nav"),
        ("v", "Table/Cards"),
        ("g/G", "Top/Bottom"),
        ("o/O", "Sort"),
        ("d", "D/W"),
        ("x", "Swap A/B"),
        ("f", "Fav"),
        ("e", "Expand"),
        ("l", "Theme"),
        ("/", "Search"),
        ("q", "Quit"),
    ],
    "transfer-tab": [
        ("s", "Source"),
        ("d", "Dest"),
        ("c", "Coin"),
        ("n", "Chain"),
        ("a", "Amount"),
        ("p", "Addr"),
        ("m", "Memo"),
        ("w", "Wallet"),
        ("r", "Refresh"),
        ("1-4", "Amount %"),
        ("j/k", "Hist Scroll"),
        ("Enter", "Start"),
        ("b/v", "Buy/Sell"),
        ("l", "Theme"),
        ("q", "Quit"),
    ],
    "scenarios-tab": [
        ("j/k", "Scroll"),
        ("Enter", "Expand"),
        ("f", "Filter"),
        ("t", "Thresholds"),
        ("l", "Theme"),
        ("q", "Quit"),
    ],
    "dw-tab": [
        ("j/k", "Scroll"),
        ("l", "Theme"),
        ("q", "Quit"),
    ],
    "logs-tab": [
        ("j/k", "Scroll"),
        ("G", "Bottom"),
        ("l", "Theme"),
        ("q", "Quit"),
    ],
}


class KimchiStatusBar(Static):
    def __init__(self) -> None:
        super().__init__("", id="kimchi-statusbar")

    def update_for_tab(self, tab_id: str) -> None:
        help_items = TAB_HELP.get(tab_id, TAB_HELP["monitor-tab"])
        text = "  ".join(
            f"[bold {C.ACCENT}]{key}[/]:{description}"
            for key, description in help_items
        )
        self.update(text)
