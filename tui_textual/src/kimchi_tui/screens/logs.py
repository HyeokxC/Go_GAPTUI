from __future__ import annotations

from textual import events
from textual.containers import Vertical
from textual.widgets import RichLog, Static

from ..colors import C
from ..indicators import log_level_badge
from ..models import Snapshot


class LogsScreen(Vertical):
    def __init__(self) -> None:
        super().__init__()
        self._last_log_lines: list[str] = []

    def compose(self):
        yield Static("System Logs", classes="panel-title")
        yield RichLog(id="logs-view", highlight=True, wrap=True, markup=True)

    def on_key(self, event: events.Key) -> None:
        key = event.key
        char = event.character or ""
        view = self.query_one("#logs-view", RichLog)

        if char == "j":
            view.scroll_down()
            event.stop()
            return
        if char == "k":
            view.scroll_up()
            event.stop()
            return
        if key == "G":
            view.scroll_end()
            event.stop()
            return

    def update_data(self, snapshot: Snapshot) -> None:
        view = self.query_one("#logs-view", RichLog)
        current_lines: list[str] = []
        for entry in snapshot.logs[-500:]:
            badge = log_level_badge(entry.log_type)
            current_lines.append(
                f"[{C.TEXT_SECONDARY}]{entry.timestamp}[/] {badge} [bold]{entry.symbol:<8}[/] {entry.message}",
            )
        if not current_lines:
            if self._last_log_lines:
                view.clear()
                self._last_log_lines = []
            return

        if current_lines[: len(self._last_log_lines)] == self._last_log_lines:
            for line in current_lines[len(self._last_log_lines) :]:
                view.write(line)
        else:
            view.clear()
            for line in current_lines:
                view.write(line)
        self._last_log_lines = current_lines
