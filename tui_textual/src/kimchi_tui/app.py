from __future__ import annotations
# pyright: reportUnannotatedClassAttribute=false

from typing import override

from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.message import Message
from textual.widgets import TabPane, TabbedContent

from .ipc_client import IpcClient
from .models import Snapshot, parse_snapshot
from .colors import set_theme
from .screens.dw_status import DWStatusScreen
from .screens.logs import LogsScreen
from .screens.monitor import MonitorScreen
from .screens.scenarios import ScenariosScreen
from .screens.transfer import TransferScreen
from .widgets.header_bar import KimchiHeader
from .widgets.status_bar import KimchiStatusBar


class KimchiApp(App[None]):
    CSS_PATH = ["styles/dark.tcss", "styles/light.tcss"]
    TITLE = "Kimchi Premium Monitor"
    SUB_TITLE = "Rust backend IPC via WebSocket"

    BINDINGS = [
        Binding("l", "toggle_theme", "Theme"),
        Binding("q", "quit", "Quit"),
    ]

    class SnapshotArrived(Message):
        data: dict[str, object]

        def __init__(self, data: dict[str, object]) -> None:
            self.data = data
            super().__init__()

    class ConnectionChanged(Message):
        connected: bool

        def __init__(self, connected: bool) -> None:
            self.connected = connected
            super().__init__()

    def __init__(self) -> None:
        super().__init__()
        self.ipc: IpcClient = IpcClient()
        self.snapshot: Snapshot = Snapshot()
        self.dark_mode: bool = True

    @override
    def compose(self) -> ComposeResult:
        yield KimchiHeader()
        with TabbedContent(id="main-tabs"):
            with TabPane("Monitor", id="monitor-tab"):
                yield MonitorScreen()
            with TabPane("Transfer", id="transfer-tab"):
                yield TransferScreen()
            with TabPane("Scenarios", id="scenarios-tab"):
                yield ScenariosScreen()
            with TabPane("D/W Status", id="dw-tab"):
                yield DWStatusScreen()
            with TabPane("Logs", id="logs-tab"):
                yield LogsScreen()
        yield KimchiStatusBar()

    def on_mount(self) -> None:
        _ = self.add_class("dark-theme")
        tabs = self.query_one("#main-tabs", TabbedContent)
        active_id = tabs.active or "monitor-tab"
        self.query_one(KimchiStatusBar).update_for_tab(active_id)
        self.update_ui()
        self.ipc.on_snapshot = self._on_snapshot
        self.ipc.on_connect = self._on_connect
        self.ipc.on_disconnect = self._on_disconnect
        _ = self.run_worker(self.ipc.connect(), name="ipc-worker", exclusive=True)

    def on_unmount(self) -> None:
        self.ipc.stop()

    def _on_snapshot(self, data: dict[str, object]) -> None:
        _ = self.post_message(self.SnapshotArrived(data))

    def _on_connect(self) -> None:
        _ = self.post_message(self.ConnectionChanged(True))

    def _on_disconnect(self) -> None:
        _ = self.post_message(self.ConnectionChanged(False))

    def on_kimchi_app_connection_changed(self, event: ConnectionChanged) -> None:
        del event

    def on_kimchi_app_snapshot_arrived(self, event: SnapshotArrived) -> None:
        self.snapshot = parse_snapshot(event.data)
        self.update_ui()

    def on_tabbed_content_tab_activated(
        self, event: TabbedContent.TabActivated
    ) -> None:
        active_id = event.tab.id or "monitor-tab"
        self.query_one(KimchiStatusBar).update_for_tab(active_id)
        self.update_ui()

    def on_kimchi_header_search_changed(
        self, event: KimchiHeader.SearchChanged
    ) -> None:
        self.query_one(MonitorScreen).on_header_search_changed(event)

    def on_kimchi_header_exchanges_changed(
        self, event: KimchiHeader.ExchangesChanged
    ) -> None:
        self.query_one(MonitorScreen).on_header_exchanges_changed(event)

    def on_kimchi_header_transfer_pressed(
        self, event: KimchiHeader.TransferPressed
    ) -> None:
        self.query_one(MonitorScreen).on_header_transfer_pressed(event)

    def update_ui(self) -> None:
        tabs = self.query_one("#main-tabs", TabbedContent)
        active_id = tabs.active or "monitor-tab"
        self.query_one(KimchiHeader).update_data(
            self.snapshot, active_id, self.dark_mode
        )

        if active_id == "monitor-tab":
            self.query_one(MonitorScreen).update_data(self.snapshot)
        elif active_id == "transfer-tab":
            self.query_one(TransferScreen).update_data(self.snapshot)
        elif active_id == "scenarios-tab":
            self.query_one(ScenariosScreen).update_data(self.snapshot)
        elif active_id == "dw-tab":
            self.query_one(DWStatusScreen).update_data(self.snapshot)
        elif active_id == "logs-tab":
            self.query_one(LogsScreen).update_data(self.snapshot)

    def action_toggle_theme(self) -> None:
        self.dark_mode = not self.dark_mode
        set_theme(dark=self.dark_mode)
        _ = self.set_class(self.dark_mode, "dark-theme")
        _ = self.set_class(not self.dark_mode, "light-theme")
        self.update_ui()
