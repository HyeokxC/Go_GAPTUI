from __future__ import annotations
# pyright: reportImplicitOverride=false, reportUnusedCallResult=false

from typing import Any, cast

from textual import events
from textual.containers import Vertical
from textual.widgets import Static

from ..colors import C, SCENARIO_BADGE_LABELS
from ..indicators import scenario_badge, scenario_state_text
from ..models import LogThread, Snapshot


class ScenariosScreen(Vertical):
    FILTER_SEQUENCE = ["All", "GapThreshold", "DomesticGap", "FutBasis"]
    FILTER_LABELS = {
        "All": "All",
        "GapThreshold": "GapThreshold",
        "DomesticGap": "DomesticGap",
        "FutBasis": "FutBasis",
    }

    def __init__(self) -> None:
        super().__init__(id="scenarios-screen")
        self._snapshot = Snapshot()
        self._current_filter = "All"
        self._selected_index = 0
        self._expanded_ids: set[str] = set()

        self._editing_thresholds = False
        self._active_field = 0
        self._fields = ["", "", ""]
        self._editing_origin = (5.0, 1.5, 0.5)

    def compose(self):
        yield Static("━━━ Config ━━━", classes="panel-title")
        yield Static(id="scenario-config-bar")
        yield Static(id="scenario-screen-log")

    def on_mount(self) -> None:
        self._sync_from_snapshot(self._snapshot)
        self._refresh()

    def on_key(self, event: events.Key) -> None:
        key = event.key
        char = event.character or ""

        if self._editing_thresholds:
            if key in {"escape", "esc"}:
                self._cancel_editing()
                event.stop()
                return
            if key in {"tab", "right"}:
                self._active_field = (self._active_field + 1) % 3
                self._render_threshold_bar()
                event.stop()
                return
            if key == "enter":
                self._apply_thresholds()
                event.stop()
                return
            if key == "backspace":
                self._fields[self._active_field] = self._fields[self._active_field][:-1]
                self._render_threshold_bar()
                event.stop()
                return
            if char in {"q", "Q"}:
                event.stop()
                return
            if char.isdigit() or char == ".":
                current = self._fields[self._active_field]
                if char == "." and "." in current:
                    event.stop()
                    return
                self._fields[self._active_field] = current + char
                self._render_threshold_bar()
                event.stop()
                return
            return

        if char in {"t", "T"}:
            self._start_editing()
            event.stop()
            return
        if char in {"f", "F"}:
            self._cycle_filter()
            event.stop()
            return
        if char == "j":
            self._move_selection(1)
            event.stop()
            return
        if char == "k":
            self._move_selection(-1)
            event.stop()
            return
        if key == "enter":
            self._toggle_expanded_selected()
            event.stop()
            return

    def update_data(self, snapshot: Snapshot) -> None:
        self._snapshot = snapshot
        if not self._editing_thresholds:
            self._sync_from_snapshot(snapshot)
        self._refresh()

    def _sync_from_snapshot(self, snapshot: Snapshot) -> None:
        self._fields = [
            self._format_decimal(snapshot.scenario_config.gap_threshold_percent),
            self._format_decimal(snapshot.scenario_config.domestic_gap_threshold),
            self._format_decimal(snapshot.scenario_config.fut_basis_threshold),
        ]
        self._editing_origin = (
            snapshot.scenario_config.gap_threshold_percent,
            snapshot.scenario_config.domestic_gap_threshold,
            snapshot.scenario_config.fut_basis_threshold,
        )

    def _format_decimal(self, value: float) -> str:
        return f"{value:g}"

    def _refresh(self) -> None:
        self._render_threshold_bar()
        self._render_threads()

    def _render_threshold_bar(self) -> None:
        config = self.query_one("#scenario-config-bar", Static)
        config.update(self._build_threshold_line())

    def _render_threads(self) -> None:
        view = self.query_one("#scenario-screen-log", Static)
        all_threads = sorted(
            self._snapshot.scenario_threads,
            key=lambda thread: thread.main_timestamp,
            reverse=True,
        )
        filtered = self._get_filtered_threads(all_threads)

        if filtered:
            self._selected_index = max(0, min(self._selected_index, len(filtered) - 1))
        else:
            self._selected_index = 0

        active_count = sum(1 for thread in filtered if thread.is_active)
        lines: list[str] = [
            f"Active: {active_count} | Total: {len(filtered)} | Filter: {self.FILTER_LABELS[self._current_filter]}"
        ]

        if not filtered:
            lines.append(f"[{C.TEXT_MUTED}]No scenario threads[/]")
            view.update("\n".join(lines))
            return

        for idx, thread in enumerate(filtered):
            dot = "●" if thread.is_active else "○"
            badge = scenario_badge(thread.scenario)
            state = scenario_state_text(thread.is_active)
            thread_key = thread.key or "-"
            symbol = thread.symbol or "-"
            timestamp = thread.main_timestamp or "-"
            message = thread.main_message or "-"
            line = (
                f"{dot} {badge} "
                f"{symbol} {thread_key}: {message}  [{C.TEXT_MUTED}]{timestamp}[/]  {state}"
            )
            if idx == self._selected_index:
                line = f"[reverse]{line}[/]"
            lines.append(line)

            thread_id = self._thread_uid(thread)
            if thread_id in self._expanded_ids:
                for sub in reversed(thread.sub_entries[-6:]):
                    sub_message = sub.message or "-"
                    sub_time = sub.timestamp or "-"
                    lines.append(f"    └ {sub_message}  [{C.TEXT_MUTED}]{sub_time}[/]")

        view.update("\n".join(lines))

    def _build_threshold_line(self) -> str:
        field_labels = ["K", "D", "F"]
        field_colors = [C.BADGE_KIMP, C.BADGE_DOMGAP, C.BADGE_FUTBASIS]
        parts: list[str] = []
        for idx, (label, bg_color) in enumerate(zip(field_labels, field_colors)):
            value = self._fields[idx]
            if self._editing_thresholds and idx == self._active_field:
                shown = f"{value}|"
                parts.append(f"[bold white on {bg_color}] {label}:{shown} [/]")
            elif self._editing_thresholds:
                parts.append(f"[dim on {bg_color}] {label}:{value} [/]")
            else:
                parts.append(f"[bold white on {bg_color}] {label}:{value} [/]")
        return (
            "Thresholds | " + "  ".join(parts) + "  Tab/→:next  Enter:apply  Esc:cancel"
        )

    def _get_filtered_threads(self, threads: list[LogThread]) -> list[LogThread]:
        if self._current_filter == "All":
            return threads
        return [thread for thread in threads if thread.scenario == self._current_filter]

    def _move_selection(self, delta: int) -> None:
        filtered = self._get_filtered_threads(self._snapshot.scenario_threads)
        if not filtered:
            self._selected_index = 0
            self._render_threads()
            return
        next_index = self._selected_index + delta
        self._selected_index = max(0, min(next_index, len(filtered) - 1))
        self._render_threads()

    def _toggle_expanded_selected(self) -> None:
        filtered = self._get_filtered_threads(
            sorted(
                self._snapshot.scenario_threads,
                key=lambda thread: thread.main_timestamp,
                reverse=True,
            )
        )
        if not filtered:
            return
        self._selected_index = max(0, min(self._selected_index, len(filtered) - 1))
        thread_id = self._thread_uid(filtered[self._selected_index])
        if thread_id in self._expanded_ids:
            self._expanded_ids.remove(thread_id)
        else:
            self._expanded_ids.add(thread_id)
        self._render_threads()

    def _cycle_filter(self) -> None:
        current = self.FILTER_SEQUENCE.index(self._current_filter)
        self._current_filter = self.FILTER_SEQUENCE[
            (current + 1) % len(self.FILTER_SEQUENCE)
        ]
        self._selected_index = 0
        self._render_threads()

    def _start_editing(self) -> None:
        self._editing_thresholds = True
        self._active_field = 0
        self._fields = [
            self._format_decimal(self._editing_origin[0]),
            self._format_decimal(self._editing_origin[1]),
            self._format_decimal(self._editing_origin[2]),
        ]
        self._render_threshold_bar()

    def _cancel_editing(self) -> None:
        self._editing_thresholds = False
        self._fields = [
            self._format_decimal(self._editing_origin[0]),
            self._format_decimal(self._editing_origin[1]),
            self._format_decimal(self._editing_origin[2]),
        ]
        self._render_threshold_bar()

    def _apply_thresholds(self) -> None:
        values = [
            self._parse_value(value, fallback)
            for value, fallback in zip(self._fields, self._editing_origin)
        ]
        self._fields = [
            self._format_decimal(values[0]),
            self._format_decimal(values[1]),
            self._format_decimal(values[2]),
        ]
        self._editing_origin = (values[0], values[1], values[2])
        app = cast(Any, self.app)
        self.run_worker(
            app.ipc.send_command(
                "SetScenarioThreshold",
                kimp=values[0],
                domestic_gap=values[1],
                fut_basis=values[2],
            )
        )
        self._editing_thresholds = False
        self._render_threshold_bar()

    def _parse_value(self, raw: str, fallback: float) -> float:
        if not raw or raw == ".":
            return fallback
        try:
            return float(raw)
        except ValueError:
            return fallback

    def _thread_uid(self, thread: LogThread) -> str:
        return f"{thread.id}:{thread.scenario}:{thread.symbol}:{thread.key}:{thread.main_timestamp}"
