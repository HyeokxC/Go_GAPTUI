from __future__ import annotations

from typing import Any, cast

from textual import events
from textual.containers import Horizontal, Vertical
from textual.widgets import Button, Checkbox, Input, RichLog, Static

from ..colors import C, exchange_tag_color
from ..formatters import fmt_amount, fmt_elapsed
from ..indicators import job_status_badge, step_marker, transfer_log_badge
from ..models import Snapshot, TransferJob, TransferState


class TransferScreen(Horizontal):
    ALL_EXCHANGES = [
        ("Upbit", "UP"),
        ("Bithumb", "BT"),
        ("Binance", "BN"),
        ("Bybit", "BB"),
        ("Bitget", "BG"),
        ("Okx", "OK"),
        ("Gate", "GT"),
    ]
    DOMESTIC_EXCHANGES = [("Upbit", "UP"), ("Bithumb", "BT")]
    GLOBAL_EXCHANGES = [
        ("Binance", "BN"),
        ("Bybit", "BB"),
        ("Bitget", "BG"),
        ("Okx", "OK"),
        ("Gate", "GT"),
    ]

    def __init__(self) -> None:
        super().__init__(id="transfer-screen")
        self._last_snapshot_transfer = TransferState()
        self._last_transfer_log_lines: list[str] = []
        self._last_network_key = ""
        self._history_scroll_offset = 0

    def compose(self):
        with Vertical(id="transfer-left"):
            yield Static(f"[bold {C.TEXT_PRIMARY}]Quick Transfer[/]", classes="panel-title")

            with Horizontal(classes="row-wrap"):
                yield Static("Coin", classes="field-label")
                yield Input(placeholder="Coin", id="coin-input")
                yield Button("Refresh", id="coin-refresh", classes="chip-btn")
            with Horizontal(id="from-to-row"):
                with Vertical(id="from-section"):
                    yield Static("From:", classes="field-label")
                    with Horizontal(classes="row-wrap"):
                        yield Static("Domestic", classes="sub-label")
                        for ex, short in self.DOMESTIC_EXCHANGES:
                            yield Button(short, id=f"from-{ex}", classes="chip-btn")
                    with Horizontal(classes="row-wrap"):
                        yield Static("Global", classes="sub-label")
                        for ex, short in self.GLOBAL_EXCHANGES:
                            yield Button(short, id=f"from-{ex}", classes="chip-btn")

                with Vertical(id="to-section"):
                    with Horizontal(classes="row-wrap"):
                        yield Static("To:", classes="field-label")
                        yield Static("", id="addr-status")
                    with Horizontal(classes="row-wrap"):
                        yield Static("Domestic", classes="sub-label")
                        for ex, short in self.DOMESTIC_EXCHANGES:
                            yield Button(short, id=f"to-{ex}", classes="chip-btn")
                    with Horizontal(classes="row-wrap"):
                        yield Static("Global", classes="sub-label")
                        for ex, short in self.GLOBAL_EXCHANGES:
                            yield Button(short, id=f"to-{ex}", classes="chip-btn")
                        yield Button("Wallet", id="to-wallet", classes="chip-btn")
                        yield Button("My", id="to-my", classes="chip-btn")

            with Vertical(id="chain-selector", classes="hidden"):
                yield Static("Chains", classes="field-label")
                with Horizontal(id="chain-buttons", classes="row-wrap"):
                    yield Static("No networks fetched", id="no-network")

            with Horizontal(classes="row-wrap"):
                yield Static("Address", classes="field-label")
                yield Input(placeholder="Destination address", id="transfer-address")
                yield Button("Copy", id="copy-address", classes="chip-btn")

            with Horizontal(classes="row-wrap"):
                yield Static("Tag/Memo", classes="field-label")
                yield Input(placeholder="Tag or memo if needed", id="transfer-tag")
                yield Button("Copy", id="copy-tag", classes="chip-btn")

            with Horizontal(classes="row-wrap"):
                yield Static("Amount", classes="field-label")
                yield Input(placeholder="0.0", id="transfer-amount")
                yield Button("25%", id="amt-25", classes="amount-btn")
                yield Button("50%", id="amt-50", classes="amount-btn")
                yield Button("75%", id="amt-75", classes="amount-btn")
                yield Button("100%", id="amt-100", classes="amount-btn")
                yield Button("↻", id="amt-refresh", classes="amount-btn")

            with Horizontal(classes="row-wrap"):
                yield Checkbox("Auto-buy before transfer", id="auto-buy")
                yield Checkbox("Auto-sell on arrival", value=True, id="auto-sell")

            with Horizontal(classes="row-wrap"):
                yield Button("Market Buy", id="market-buy", classes="action-btn")
                yield Button("Market Sell", id="market-sell", classes="action-btn")
                yield Button(
                    "Execute Transfer",
                    id="execute-transfer",
                    classes="action-btn primary",
                )

            yield Static(f"[bold {C.TEXT_PRIMARY}]Transfer Log[/]", classes="panel-title")
            yield RichLog(id="transfer-log", highlight=True, wrap=True, markup=True)

        with Vertical(id="transfer-right"):
            yield Static(f"[bold {C.TEXT_PRIMARY}]Progress[/]", classes="panel-title")
            yield Static(id="transfer-progress")
            yield Static(f"[bold {C.TEXT_PRIMARY}]History[/]", classes="panel-title")
            yield Static(id="transfer-history")

    def update_data(self, snapshot: Snapshot) -> None:
        transfer = snapshot.transfer
        self._last_snapshot_transfer = transfer

        self._sync_selection_buttons(transfer)
        self._sync_form_values(transfer)
        self._sync_network_buttons(transfer)

        log = self.query_one("#transfer-log", RichLog)
        transfer_log_lines: list[str] = []
        for entry in transfer.logs[-120:]:
            badge = transfer_log_badge(entry.is_error)
            color = C.ERROR if entry.is_error else C.INFO
            transfer_log_lines.append(
                f"[{C.TEXT_SECONDARY}]{entry.timestamp}[/] {badge} [{color}]{entry.message}[/]"
            )
        if (
            transfer_log_lines[: len(self._last_transfer_log_lines)]
            == self._last_transfer_log_lines
        ):
            for line in transfer_log_lines[len(self._last_transfer_log_lines) :]:
                log.write(line)
        else:
            log.clear()
            for line in transfer_log_lines:
                log.write(line)
        self._last_transfer_log_lines = transfer_log_lines

        self._render_active_transfer_progress(snapshot.transfer_jobs)
        self._render_transfer_history(snapshot.transfer_jobs)

    def on_key(self, event: events.Key) -> None:
        focused = self.app.focused
        if isinstance(focused, Input):
            return

        key = event.key
        char = (event.character or "").lower()

        if char == "s":
            self._cycle_from_exchange()
            event.stop()
            return
        if char == "d":
            self._cycle_to_exchange()
            event.stop()
            return
        if char == "c":
            self.query_one("#coin-input", Input).focus()
            event.stop()
            return
        if char == "a":
            self.query_one("#transfer-amount", Input).focus()
            event.stop()
            return
        if char == "p":
            self.query_one("#transfer-address", Input).focus()
            event.stop()
            return
        if char == "m":
            self.query_one("#transfer-tag", Input).focus()
            event.stop()
            return
        if char == "n":
            self._cycle_network()
            event.stop()
            return
        if char == "r":
            self._refresh_transfer_data()
            event.stop()
            return
        if char == "b":
            auto_buy = self.query_one("#auto-buy", Checkbox)
            auto_buy.value = not auto_buy.value
            app = cast(Any, self.app)
            self.run_worker(app.ipc.send_command("SetAutoBuy", enabled=auto_buy.value))
            event.stop()
            return
        if char == "v":
            auto_sell = self.query_one("#auto-sell", Checkbox)
            auto_sell.value = not auto_sell.value
            app = cast(Any, self.app)
            self.run_worker(
                app.ipc.send_command("SetAutoSell", enabled=auto_sell.value)
            )
            event.stop()
            return
        if char == "w":
            self._last_snapshot_transfer.to_is_personal_wallet = (
                not self._last_snapshot_transfer.to_is_personal_wallet
            )
            self._sync_wallet_button()
            event.stop()
            return
        if char == "1":
            self._set_amount_ratio(0.25)
            event.stop()
            return
        if char == "2":
            self._set_amount_ratio(0.50)
            event.stop()
            return
        if char == "3":
            self._set_amount_ratio(0.75)
            event.stop()
            return
        if char == "4":
            self._set_amount_ratio(1.00)
            event.stop()
            return
        if char == "j":
            self._history_scroll_offset += 1
            self._render_transfer_history(cast(Any, self.app).snapshot.transfer_jobs)
            event.stop()
            return
        if char == "k":
            self._history_scroll_offset = max(0, self._history_scroll_offset - 1)
            self._render_transfer_history(cast(Any, self.app).snapshot.transfer_jobs)
            event.stop()
            return
        if key == "enter":
            self._execute_transfer()
            event.stop()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        button_id = event.button.id or ""
        app = cast(Any, self.app)

        if button_id == "coin-refresh":
            self._apply_coin_from_input()
            return

        if button_id.startswith("from-"):
            exchange = button_id.split("-", 1)[1]
            self.run_worker(app.ipc.send_command("SetTransferFrom", exchange=exchange))
            self.run_worker(
                app.ipc.send_command(
                    "FetchNetworks",
                    coin=self._last_snapshot_transfer.selected_coin,
                    from_exchange=exchange,
                    to_exchange=self._last_snapshot_transfer.to_exchange,
                )
            )
            self.run_worker(
                app.ipc.send_command(
                    "FetchBalance",
                    coin=self._last_snapshot_transfer.selected_coin,
                    exchange=exchange,
                )
            )
            return

        if button_id.startswith("to-") and button_id not in ("to-wallet", "to-my"):
            exchange = button_id.split("-", 1)[1]
            self.run_worker(app.ipc.send_command("SetTransferTo", exchange=exchange))
            self.run_worker(
                app.ipc.send_command(
                    "FetchNetworks",
                    coin=self._last_snapshot_transfer.selected_coin,
                    from_exchange=self._last_snapshot_transfer.from_exchange,
                    to_exchange=exchange,
                )
            )
            return

        if button_id == "to-wallet":
            self._last_snapshot_transfer.to_is_personal_wallet = (
                not self._last_snapshot_transfer.to_is_personal_wallet
            )
            self._sync_wallet_button()
            return

        if button_id == "to-my":
            self._last_snapshot_transfer.to_is_personal_wallet = True
            self._sync_wallet_button()
            return

        if button_id == "copy-address":
            self._copy_to_clipboard(
                self.query_one("#transfer-address", Input).value, "address"
            )
            return

        if button_id == "copy-tag":
            self._copy_to_clipboard(
                self.query_one("#transfer-tag", Input).value, "tag/memo"
            )
            return

        if button_id.startswith("chain-"):
            idx = int(button_id.split("-", 1)[1])
            self.run_worker(app.ipc.send_command("SelectNetwork", index=idx))
            if idx < len(self._last_snapshot_transfer.available_networks):
                network = self._last_snapshot_transfer.available_networks[idx]
                self.run_worker(
                    app.ipc.send_command(
                        "FetchDepositAddress",
                        coin=self._last_snapshot_transfer.selected_coin,
                        exchange=self._last_snapshot_transfer.to_exchange,
                        network=network.network,
                    ),
                )
            return

        if button_id.startswith("amt-"):
            ratio_map = {
                "amt-25": 0.25,
                "amt-50": 0.50,
                "amt-75": 0.75,
                "amt-100": 1.00,
            }
            if button_id in ratio_map:
                self.run_worker(
                    app.ipc.send_command("SetAmountRatio", ratio=ratio_map[button_id])
                )
            elif button_id == "amt-refresh":
                self.run_worker(
                    app.ipc.send_command(
                        "FetchBalance",
                        coin=self._last_snapshot_transfer.selected_coin,
                        exchange=self._last_snapshot_transfer.from_exchange,
                    ),
                )
            return

        if button_id == "market-buy":
            self.run_worker(
                app.ipc.send_command(
                    "SubmitMarketOrder",
                    exchange=self._last_snapshot_transfer.from_exchange,
                    coin=self._last_snapshot_transfer.selected_coin,
                    side="Buy",
                    amount=self._last_snapshot_transfer.amount,
                ),
            )
            return

        if button_id == "market-sell":
            self.run_worker(
                app.ipc.send_command(
                    "SubmitMarketOrder",
                    exchange=self._last_snapshot_transfer.from_exchange,
                    coin=self._last_snapshot_transfer.selected_coin,
                    side="Sell",
                    amount=self._last_snapshot_transfer.amount,
                ),
            )
            return

        if button_id == "execute-transfer":
            self._execute_transfer()

    def on_input_submitted(self, event: Input.Submitted) -> None:
        app = cast(Any, self.app)
        if event.input.id == "coin-input":
            self._apply_coin_from_input()
        elif event.input.id == "transfer-address":
            self.run_worker(
                app.ipc.send_command("SetTransferAddress", address=event.value)
            )
        elif event.input.id == "transfer-tag":
            self.run_worker(app.ipc.send_command("SetTransferTag", tag=event.value))
        elif event.input.id == "transfer-amount":
            self.run_worker(
                app.ipc.send_command("SetTransferAmount", amount=event.value)
            )

    def on_checkbox_changed(self, event: Checkbox.Changed) -> None:
        app = cast(Any, self.app)
        if event.checkbox.id == "auto-buy":
            self.run_worker(app.ipc.send_command("SetAutoBuy", enabled=event.value))
        elif event.checkbox.id == "auto-sell":
            self.run_worker(app.ipc.send_command("SetAutoSell", enabled=event.value))

    def _apply_coin_from_input(self) -> None:
        app = cast(Any, self.app)
        coin = self.query_one("#coin-input", Input).value.strip().upper()
        if not coin:
            return
        self.run_worker(app.ipc.send_command("SetTransferCoin", coin=coin))
        self.run_worker(
            app.ipc.send_command(
                "FetchNetworks",
                coin=coin,
                from_exchange=self._last_snapshot_transfer.from_exchange,
                to_exchange=self._last_snapshot_transfer.to_exchange,
            )
        )
        self.run_worker(
            app.ipc.send_command(
                "FetchBalance",
                coin=coin,
                exchange=self._last_snapshot_transfer.from_exchange,
            )
        )

    def _copy_to_clipboard(self, text: str, label: str) -> None:
        if not text:
            return
        try:
            clip = __import__("pyperclip")
            clip.copy(text)
        except Exception:
            pass

    def _sync_selection_buttons(self, transfer: TransferState) -> None:
        coin_input = self.query_one("#coin-input", Input)
        if coin_input.value != transfer.selected_coin and not coin_input.has_focus:
            coin_input.value = transfer.selected_coin

        for ex, short in self.ALL_EXCHANGES:
            try:
                btn = self.query_one(f"#from-{ex}", Button)
                is_selected = ex == transfer.from_exchange
                btn.variant = "success" if is_selected else "default"
                color = exchange_tag_color(short)
                btn.label = f"[bold {color}]{short}[/]" if is_selected else short
            except Exception:
                pass

        for ex, short in self.ALL_EXCHANGES:
            try:
                btn = self.query_one(f"#to-{ex}", Button)
                is_selected = ex == transfer.to_exchange
                btn.variant = "success" if is_selected else "default"
                color = exchange_tag_color(short)
                btn.label = f"[bold {color}]{short}[/]" if is_selected else short
            except Exception:
                pass

        self._sync_wallet_button()

    def _sync_wallet_button(self) -> None:
        try:
            wallet_btn = self.query_one("#to-wallet", Button)
            wallet_btn.variant = (
                "success"
                if self._last_snapshot_transfer.to_is_personal_wallet
                else "default"
            )
        except Exception:
            pass

    def _sync_form_values(self, transfer: TransferState) -> None:
        amount_input = self.query_one("#transfer-amount", Input)
        address_input = self.query_one("#transfer-address", Input)
        tag_input = self.query_one("#transfer-tag", Input)
        auto_buy = self.query_one("#auto-buy", Checkbox)
        auto_sell = self.query_one("#auto-sell", Checkbox)

        if amount_input.value != transfer.amount:
            amount_input.value = transfer.amount
        if address_input.value != transfer.deposit_address:
            address_input.value = transfer.deposit_address
        if tag_input.value != transfer.deposit_tag:
            tag_input.value = transfer.deposit_tag

        auto_buy.value = transfer.auto_buy_before_transfer
        auto_sell.value = transfer.auto_sell_on_arrival

        addr_status = self.query_one("#addr-status", Static)
        if transfer.deposit_address:
            addr_status.update(f"[{C.SUCCESS}]✓ Addr loaded[/]")
        else:
            addr_status.update(f"[{C.TEXT_MUTED}]No addr[/]")

    def _sync_network_buttons(self, transfer: TransferState) -> None:
        chain_selector = self.query_one("#chain-selector", Vertical)
        if not transfer.available_networks:
            chain_selector.add_class("hidden")
        else:
            chain_selector.remove_class("hidden")

        net_key = "|".join(
            f"{network.network}:{network.display_name}:{network.withdraw_fee}"
            for network in transfer.available_networks
        )
        holder = self.query_one("#chain-buttons", Horizontal)

        if net_key == self._last_network_key:
            for idx in range(len(transfer.available_networks)):
                try:
                    button = self.query_one(f"#chain-{idx}", Button)
                    button.variant = (
                        "success" if transfer.selected_network_idx == idx else "default"
                    )
                except Exception:
                    continue
            return

        self._last_network_key = net_key
        holder.remove_children()

        if not transfer.available_networks:
            holder.mount(Static("No networks fetched", id="no-network"))
            return

        for idx, network in enumerate(transfer.available_networks):
            fee_txt = (
                "-" if network.withdraw_fee is None else f"{network.withdraw_fee:g}"
            )
            label = f"{network.display_name or network.network} ({fee_txt})"
            btn = Button(label, id=f"chain-{idx}", classes="chain-btn")
            btn.variant = (
                "success" if transfer.selected_network_idx == idx else "default"
            )
            holder.mount(btn)

    def _execute_transfer(self) -> None:
        app = cast(Any, self.app)
        self.run_worker(
            app.ipc.send_command(
                "ExecuteTransfer",
                coin=self._last_snapshot_transfer.selected_coin,
                from_exchange=self._last_snapshot_transfer.from_exchange,
                to_exchange=self._last_snapshot_transfer.to_exchange,
                amount=self.query_one("#transfer-amount", Input).value,
                address=self.query_one("#transfer-address", Input).value,
                tag=self.query_one("#transfer-tag", Input).value,
                auto_buy=self.query_one("#auto-buy", Checkbox).value,
                auto_sell=self.query_one("#auto-sell", Checkbox).value,
            ),
        )

    def _cycle_from_exchange(self) -> None:
        app = cast(Any, self.app)
        exchanges = [name for name, _ in self.ALL_EXCHANGES]
        current = self._last_snapshot_transfer.from_exchange
        next_idx = (
            (exchanges.index(current) + 1) % len(exchanges)
            if current in exchanges
            else 0
        )
        exchange = exchanges[next_idx]

        self.run_worker(app.ipc.send_command("SetTransferFrom", exchange=exchange))
        self.run_worker(
            app.ipc.send_command(
                "FetchNetworks",
                coin=self._last_snapshot_transfer.selected_coin,
                from_exchange=exchange,
                to_exchange=self._last_snapshot_transfer.to_exchange,
            )
        )
        self.run_worker(
            app.ipc.send_command(
                "FetchBalance",
                coin=self._last_snapshot_transfer.selected_coin,
                exchange=exchange,
            )
        )

    def _cycle_to_exchange(self) -> None:
        app = cast(Any, self.app)
        exchanges = [name for name, _ in self.ALL_EXCHANGES]
        current = self._last_snapshot_transfer.to_exchange
        next_idx = (
            (exchanges.index(current) + 1) % len(exchanges)
            if current in exchanges
            else 0
        )
        exchange = exchanges[next_idx]

        self.run_worker(app.ipc.send_command("SetTransferTo", exchange=exchange))
        self.run_worker(
            app.ipc.send_command(
                "FetchNetworks",
                coin=self._last_snapshot_transfer.selected_coin,
                from_exchange=self._last_snapshot_transfer.from_exchange,
                to_exchange=exchange,
            )
        )

    def _cycle_network(self) -> None:
        networks = self._last_snapshot_transfer.available_networks
        if not networks:
            return

        current = self._last_snapshot_transfer.selected_network_idx
        if current is None or current < 0 or current >= len(networks):
            next_idx = 0
        else:
            next_idx = (current + 1) % len(networks)

        app = cast(Any, self.app)
        self.run_worker(app.ipc.send_command("SelectNetwork", index=next_idx))
        network = networks[next_idx]
        self.run_worker(
            app.ipc.send_command(
                "FetchDepositAddress",
                coin=self._last_snapshot_transfer.selected_coin,
                exchange=self._last_snapshot_transfer.to_exchange,
                network=network.network,
            ),
        )

    def _refresh_transfer_data(self) -> None:
        app = cast(Any, self.app)
        self.run_worker(
            app.ipc.send_command(
                "FetchBalance",
                coin=self._last_snapshot_transfer.selected_coin,
                exchange=self._last_snapshot_transfer.from_exchange,
            )
        )
        self.run_worker(
            app.ipc.send_command(
                "FetchNetworks",
                coin=self._last_snapshot_transfer.selected_coin,
                from_exchange=self._last_snapshot_transfer.from_exchange,
                to_exchange=self._last_snapshot_transfer.to_exchange,
            )
        )

    def _set_amount_ratio(self, ratio: float) -> None:
        app = cast(Any, self.app)
        self.run_worker(app.ipc.send_command("SetAmountRatio", ratio=ratio))

    def _render_active_transfer_progress(self, jobs: list[TransferJob]) -> None:
        widget = self.query_one("#transfer-progress", Static)
        active = next((job for job in jobs if job.is_executing), None)
        if active is None:
            widget.update(
                f"[bold {C.TEXT_PRIMARY}]Active Transfer[/]\n[{C.TEXT_MUTED}]No active transfer[/]"
            )
            return

        header = (
            f"[bold {C.TEXT_PRIMARY}]Active Transfer[/]  "
            f"Job #{active.id}  {active.coin}  "
            f"{active.from_exchange} → {active.to_exchange}  "
            f"{fmt_amount(active.amount)}"
        )
        lines: list[str] = [header, ""]

        for step in active.steps:
            status = step.status
            if status == "Completed":
                border_color = C.SUCCESS
                icon = "✓"
            elif status == "InProgress":
                border_color = C.PROGRESS
                icon = "● In progress..."
            elif status == "Failed":
                border_color = C.ERROR
                icon = "✗ Failed"
            else:
                border_color = C.TEXT_MUTED
                icon = ""

            lines.append(f"  [{border_color}]┌─ [bold]{step.title}[/bold] ─┐[/]")
            if icon:
                lines.append(f"  [{border_color}]│  {icon}[/]")
            if status == "Failed" and step.message:
                lines.append(f"  [{C.ERROR}]│  {step.message}[/]")
            lines.append(f"  [{border_color}]└──────────────────┘[/]")
            lines.append(f"  [{C.TEXT_MUTED}]│[/]")

        if active.error_message:
            lines.append(f"[{C.ERROR}]✗ Error: {active.error_message}[/]")
        else:
            failed_step = next(
                (step for step in active.steps if step.status == "Failed"), None
            )
            if failed_step is not None and failed_step.message:
                lines.append(f"[{C.ERROR}]✗ Error: {failed_step.message}[/]")

        widget.update("\n".join(lines))

    def _render_transfer_history(self, jobs: list[TransferJob]) -> None:
        widget = self.query_one("#transfer-history", Static)
        total_jobs = len(jobs)
        if total_jobs == 0:
            widget.update(f"[{C.TEXT_MUTED}]No transfer history[/]")
            self._history_scroll_offset = 0
            return

        available_height = widget.size.height
        visible_rows = max(5, (available_height - 4) if available_height > 0 else 8)
        max_offset = max(0, total_jobs - visible_rows)
        self._history_scroll_offset = min(self._history_scroll_offset, max_offset)

        start = self._history_scroll_offset
        end = min(total_jobs, start + visible_rows)
        rows = jobs[start:end]

        header = f"[bold {C.TEXT_PRIMARY}]{'ID':<4} {'Coin':<8} {'Route':<24} {'Amount':>14} {'Status':<10} {'Time':>10}[/]"
        divider = f"[{C.BORDER}]" + ("─" * 74) + "[/]"
        table_lines = [header, divider]

        for job in rows:
            status_txt = job_status_badge(job.is_executing, job.error_message)

            from_color = exchange_tag_color(job.from_exchange)
            to_color = exchange_tag_color(job.to_exchange)
            route_plain = f"{job.from_exchange}->{job.to_exchange}"

            table_lines.append(
                f"{job.id:<4} {job.coin:<8} "
                f"[{from_color}]{job.from_exchange}[/]→[{to_color}]{job.to_exchange}[/]"
                f"{' ' * max(0, 22 - len(route_plain))}"
                f" {fmt_amount(job.amount):>14} {status_txt}  {fmt_elapsed(job.started_at_secs):>10}"
            )

        table_lines.append(
            f"[{C.TEXT_MUTED}]Rows {start + 1}-{end} / {total_jobs} (j/k scroll)[/]"
        )
        widget.update("\n".join(table_lines))
