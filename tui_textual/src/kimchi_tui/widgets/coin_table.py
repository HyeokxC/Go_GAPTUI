from __future__ import annotations

from rich.markup import escape
from textual.containers import VerticalScroll
from textual.events import Click
from textual.widget import Widget
from textual.widgets import Static

from ..colors import C, exchange_tag_color, get_futures_basis_color, get_kimchi_color
from ..formatters import fmt_krw_price, fmt_usd_price
from ..indicators import dw_dots
from ..models import (
    CoinState,
    CoinWalletStatus,
    ExchangeWalletStatus,
    OrderbookInfo,
    Snapshot,
)


class CoinRow(Static):
    symbol: str

    def __init__(self, symbol: str, content: str) -> None:
        super().__init__(content, id=f"coin-{symbol}")
        self.symbol = symbol

    def on_click(self, _: Click) -> None:
        parent = self.parent
        if isinstance(parent, CoinTable):
            parent.select_symbol(self.symbol)


class CoinTable(VerticalScroll):
    SORT_COLUMNS: set[str] = {
        "Symbol",
        "Upbit",
        "Bithumb",
        "Binance",
        "Bybit",
        "Kimp",
    }

    def __init__(
        self,
        *children: Widget,
        name: str | None = None,
        id: str | None = None,
        classes: str | None = None,
        disabled: bool = False,
    ) -> None:
        super().__init__(
            *children,
            name=name,
            id=id,
            classes=classes,
            disabled=disabled,
        )
        self._row_keys: set[str] = set()
        self._card_cache: dict[str, str] = {}
        self._card_widgets: dict[str, CoinRow] = {}
        self._displayed_symbols: list[str] = []
        self._cursor_row: int | None = None

    @property
    def row_count(self) -> int:
        return len(self._displayed_symbols)

    @property
    def cursor_row(self) -> int | None:
        return self._cursor_row

    def update_rows(
        self,
        snapshot: Snapshot,
        query: str,
        dw_only: bool,
        exchange_filter: str,
        sort_column: str,
        sort_desc: bool,
        favorites: set[str],
        exchange_a: str = "BT",
        exchange_b: str = "UP",
    ) -> None:
        exchange_a = exchange_a.upper()
        exchange_b = exchange_b.upper()
        selected_symbol = self.selected_symbol()
        rows = self.select_coins(
            snapshot,
            query=query,
            dw_only=dw_only,
            exchange_filter=exchange_filter,
            sort_column=sort_column,
            sort_desc=sort_desc,
        )

        symbols = [coin.symbol for coin in rows]
        new_keys = set(symbols)
        stale_keys = self._row_keys - new_keys
        added_keys = new_keys - self._row_keys

        # Remove stale widgets
        for key in stale_keys:
            widget = self._card_widgets.pop(key, None)
            if widget is not None:
                _ = widget.remove()
            _ = self._card_cache.pop(key, None)

        self._row_keys = new_keys
        prev_symbols = self._displayed_symbols
        self._displayed_symbols = symbols

        self._sync_cursor(selected_symbol)

        # Build order map for sort_children
        order_map: dict[str, int] = {s: i for i, s in enumerate(symbols)}

        for idx, coin in enumerate(rows):
            symbol = coin.symbol
            base_content = self._render_coin_card(
                coin,
                snapshot,
                favorites,
                exchange_a=exchange_a,
                exchange_b=exchange_b,
                _row_index=idx,
            )
            content = self._apply_row_background(base_content, self._row_bg(idx))
            widget = self._card_widgets.get(symbol)
            if widget is None:
                widget = CoinRow(symbol, content)
                self._card_widgets[symbol] = widget
                self._card_cache[symbol] = base_content
                _ = self.mount(widget)
            elif self._card_cache.get(symbol) != base_content:
                widget.update(content)
                self._card_cache[symbol] = base_content

        # Reorder children to match the desired symbol order
        if prev_symbols != symbols:
            self.sort_children(
                key=lambda w: order_map.get(
                    w.symbol if isinstance(w, CoinRow) else '', len(symbols)
                )
            )

        self._scroll_cursor_into_view()

    def move_cursor(self, row: int, column: int = 0) -> None:
        del column
        if not self._displayed_symbols:
            self._cursor_row = None
            return
        bounded = max(0, min(row, len(self._displayed_symbols) - 1))
        if self._cursor_row == bounded:
            return
        prev_row = self._cursor_row
        self._cursor_row = bounded
        self._refresh_row(prev_row)
        self._refresh_row(self._cursor_row)
        self._scroll_cursor_into_view()

    def action_cursor_down(self) -> None:
        if not self._displayed_symbols:
            self._cursor_row = None
            return
        next_row = 0 if self._cursor_row is None else self._cursor_row + 1
        self.move_cursor(next_row)

    def action_cursor_up(self) -> None:
        if not self._displayed_symbols:
            self._cursor_row = None
            return
        next_row = (
            len(self._displayed_symbols) - 1
            if self._cursor_row is None
            else self._cursor_row - 1
        )
        self.move_cursor(next_row)

    def select_symbol(self, symbol: str) -> None:
        if symbol not in self._displayed_symbols:
            return
        self.move_cursor(self._displayed_symbols.index(symbol))

    def _sync_cursor(self, selected_symbol: str | None) -> None:
        if not self._displayed_symbols:
            self._cursor_row = None
            return
        if selected_symbol is not None and selected_symbol in self._row_keys:
            self._cursor_row = self._displayed_symbols.index(selected_symbol)
            return
        if self._cursor_row is None:
            self._cursor_row = 0
            return
        self._cursor_row = max(
            0, min(self._cursor_row, len(self._displayed_symbols) - 1)
        )

    def _refresh_row(self, row_index: int | None) -> None:
        if (
            row_index is None
            or row_index < 0
            or row_index >= len(self._displayed_symbols)
        ):
            return
        symbol = self._displayed_symbols[row_index]
        base = self._card_cache.get(symbol)
        widget = self._card_widgets.get(symbol)
        if base is None or widget is None:
            return
        widget.update(self._apply_row_background(base, self._row_bg(row_index)))

    def _scroll_cursor_into_view(self) -> None:
        symbol = self.selected_symbol()
        if symbol is None:
            return
        widget = self._card_widgets.get(symbol)
        if widget is None:
            return
        _ = self.scroll_to_widget(widget, animate=False)

    def _render_coin_card(
        self,
        coin: CoinState,
        snapshot: Snapshot,
        favorites: set[str],
        *,
        exchange_a: str,
        exchange_b: str,
        _row_index: int,
    ) -> str:
        line1 = self._line_one(coin, snapshot, favorites, exchange_a, exchange_b)
        line2 = self._line_two(coin, snapshot, exchange_a, exchange_b)
        line3 = self._line_three(snapshot.orderbook_info.get(coin.symbol))
        return "\n".join([line1, line2, line3] if line3 else [line1, line2])

    def _apply_row_background(
        self,
        content: str,
        bg_color: str,
    ) -> str:
        lines = content.splitlines()
        wrapped = [f"[on {bg_color}]{line}[/]" for line in lines]
        return "\n".join(wrapped)

    def _row_bg(self, row_index: int) -> str:
        if self._cursor_row == row_index:
            return C.HOVER_BG
        return C.CARD_BG if row_index % 2 == 0 else C.CARD_BG_ALT

    def _line_one(
        self,
        coin: CoinState,
        snapshot: Snapshot,
        favorites: set[str],
        exchange_a: str,
        exchange_b: str,
    ) -> str:
        symbol_markup = (
            f"[bold {C.ACCENT}]{escape(coin.symbol)}[/]"
            if coin.symbol in favorites
            else f"[bold {C.TEXT_PRIMARY}]{escape(coin.symbol)}[/]"
        )
        name_plain = snapshot.korean_names.get(coin.symbol, "-")
        name_markup = f"[{C.TEXT_SECONDARY}]{escape(name_plain)}[/]"
        left_markup = f"{symbol_markup} {name_markup}"
        left_plain = f"{coin.symbol} {name_plain}"

        premium = self._premium_value(
            coin, snapshot.orderbook_info.get(coin.symbol), exchange_a, exchange_b
        )
        premium_markup = self._premium_pill(premium)
        premium_plain = "-" if premium is None else f"{premium:+.2f}%"

        bnf_basis = self._futures_basis(
            self._spot_usd(coin, "BN", snapshot.usdt_krw), coin.binance_futures_price
        )
        bbf_basis = self._futures_basis(
            self._spot_usd(coin, "BB", snapshot.usdt_krw), coin.bybit_futures_price
        )
        bnf_markup = self._futures_pill("BNF", bnf_basis)
        bbf_markup = self._futures_pill("BBF", bbf_basis)
        bnf_plain = "BNF -" if bnf_basis is None else f"BNF {bnf_basis:+.2f}%"
        bbf_plain = "BBF -" if bbf_basis is None else f"BBF {bbf_basis:+.2f}%"

        right_markup = f"{premium_markup} {bnf_markup} {bbf_markup}"
        right_plain = f"{premium_plain} {bnf_plain} {bbf_plain}"
        return self._align_right(left_markup, left_plain, right_markup, right_plain)

    def _line_two(
        self, coin: CoinState, snapshot: Snapshot, exchange_a: str, exchange_b: str
    ) -> str:
        left = self._exchange_segment(coin, snapshot, exchange_a)
        right = self._exchange_segment(coin, snapshot, exchange_b)
        return f"{left} [{C.TEXT_MUTED}]|[/] {right}"

    def _line_three(self, ob: OrderbookInfo | None) -> str:
        if ob is None:
            return ""
        bt_slip = self._fmt_heat_pct(ob.bt_buy_slippage)
        up_slip = self._fmt_heat_pct(ob.up_sell_slippage)
        spread = self._fmt_heat_pct(ob.real_gap_bt_up)
        return (
            f"[{C.TEXT_MUTED}]slip[/] BT:{bt_slip} UP:{up_slip} "
            f"[{C.TEXT_MUTED}]| sprd[/] BT:{spread} UP:{spread}"
        )

    def _exchange_segment(
        self, coin: CoinState, snapshot: Snapshot, exchange: str
    ) -> str:
        tag_bg = exchange_tag_color(exchange)
        tag = f"[bold {C.PILL_TEXT} on {tag_bg}] {exchange} [/]"
        price = self._exchange_krw_price(coin, exchange)
        price_text = f"₩{fmt_krw_price(price)}" if price is not None else "-"
        usd_equiv = self._usd_equiv(price, snapshot.usdt_krw)
        usd_text = fmt_usd_price(usd_equiv) if usd_equiv is not None else "-"
        wallet = self._wallet_for_exchange(
            snapshot.wallet_status.get(coin.symbol), exchange
        )
        dots = dw_dots(
            wallet.deposit if wallet is not None else None,
            wallet.withdraw if wallet is not None else None,
        )
        return (
            f"{tag} [{C.TEXT_PRIMARY}]{price_text}[/] "
            f"[{C.TEXT_SECONDARY}](${usd_text})[/] {dots}"
        )

    def _align_right(
        self,
        left_markup: str,
        left_plain: str,
        right_markup: str,
        right_plain: str,
    ) -> str:
        width = self.size.width if self.size.width > 0 else 108
        space_count = max(2, width - len(left_plain) - len(right_plain) - 2)
        return f"{left_markup}{' ' * space_count}{right_markup}"

    def _premium_pill(self, value: float | None) -> str:
        if value is None:
            return f"[{C.TEXT_MUTED}]-[/]"
        bg = get_kimchi_color(value)
        return f"[bold {C.PILL_TEXT} on {bg}] {value:+.2f}% [/]"

    def _futures_pill(self, label: str, value: float | None) -> str:
        if value is None:
            return f"[{C.TEXT_MUTED}]{label} -[/]"
        bg = get_futures_basis_color(value)
        return f"[bold {C.PILL_TEXT} on {bg}] {label} {value:+.2f}% [/]"

    def _fmt_heat_pct(self, value: float) -> str:
        abs_val = abs(value)
        color = C.TEXT_MUTED
        if abs_val >= 1.0:
            color = C.RED
        elif abs_val >= 0.5:
            color = C.YELLOW
        return f"[{color}]{value:+.2f}%[/]"

    @staticmethod
    def _exchange_krw_price(coin: CoinState, exchange: str) -> float | None:
        mapping = {
            "UP": coin.upbit_price,
            "BT": coin.bithumb_price,
            "BN": coin.binance_krw,
            "BB": coin.bybit_krw,
            "OK": coin.okx_krw,
        }
        return mapping.get(exchange)

    @staticmethod
    def _wallet_for_exchange(
        status: CoinWalletStatus | None, exchange: str
    ) -> ExchangeWalletStatus | None:
        if status is None:
            return None
        mapping = {
            "UP": status.upbit,
            "BT": status.bithumb,
            "BN": status.binance,
            "BB": status.bybit,
            "OK": status.okx,
        }
        return mapping.get(exchange)

    @staticmethod
    def _usd_equiv(price_krw: float | None, usdt_krw: float | None) -> float | None:
        if price_krw is None or usdt_krw is None or usdt_krw <= 0:
            return None
        return price_krw / usdt_krw

    @staticmethod
    def _spot_usd(
        coin: CoinState, exchange: str, usdt_krw: float | None
    ) -> float | None:
        if exchange == "BN":
            return coin.binance_price or CoinTable._usd_equiv(
                coin.binance_krw, usdt_krw
            )
        if exchange == "BB":
            return coin.bybit_price or CoinTable._usd_equiv(coin.bybit_krw, usdt_krw)
        if exchange == "OK":
            return coin.okx_price or CoinTable._usd_equiv(coin.okx_krw, usdt_krw)
        if exchange == "UP":
            return CoinTable._usd_equiv(coin.upbit_price, usdt_krw)
        if exchange == "BT":
            return CoinTable._usd_equiv(coin.bithumb_price, usdt_krw)
        return None

    @staticmethod
    def _futures_basis(
        spot_usd: float | None, futures_usd: float | None
    ) -> float | None:
        if spot_usd is None or futures_usd is None or spot_usd <= 0 or futures_usd <= 0:
            return None
        return ((spot_usd - futures_usd) / futures_usd) * 100.0

    @staticmethod
    def _kimchi_for_pair(coin: CoinState, domestic: str, foreign: str) -> float | None:
        if domestic == "UP" and foreign == "BN":
            return coin.upbit_kimchi
        if domestic == "BT" and foreign == "BN":
            return coin.bithumb_kimchi
        if domestic == "UP" and foreign == "BB":
            return coin.bybit_kimchi_up
        if domestic == "BT" and foreign == "BB":
            return coin.bybit_kimchi_bt
        if domestic == "UP" and foreign == "OK":
            return coin.okx_kimchi_up
        if domestic == "BT" and foreign == "OK":
            return coin.okx_kimchi_bt
        return None

    @staticmethod
    def _mid_premium(coin: CoinState, exchange_a: str, exchange_b: str) -> float | None:
        values: list[float] = []
        for domestic, foreign in [(exchange_a, exchange_b), (exchange_b, exchange_a)]:
            value = CoinTable._kimchi_for_pair(coin, domestic, foreign)
            if value is not None:
                values.append(value)
        if not values:
            return None
        return sum(values) / len(values)

    @staticmethod
    def _depth_premium(
        ob: OrderbookInfo | None, exchange_a: str, exchange_b: str
    ) -> float | None:
        if ob is None:
            return None
        if exchange_a == "UP":
            return ob.real_kimp_up.get(exchange_b)
        if exchange_a == "BT":
            return ob.real_kimp_bt.get(exchange_b)
        if exchange_b == "UP":
            return ob.real_kimp_up.get(exchange_a)
        if exchange_b == "BT":
            return ob.real_kimp_bt.get(exchange_a)
        return None

    @staticmethod
    def _premium_value(
        coin: CoinState,
        ob: OrderbookInfo | None,
        exchange_a: str,
        exchange_b: str,
    ) -> float | None:
        depth = CoinTable._depth_premium(ob, exchange_a, exchange_b)
        if depth is not None:
            return depth
        return CoinTable._mid_premium(coin, exchange_a, exchange_b)

    def selected_symbol(self) -> str | None:
        row_idx = self._cursor_row
        if row_idx is None:
            return None
        if row_idx < 0 or row_idx >= len(self._displayed_symbols):
            return None
        return self._displayed_symbols[row_idx]

    def has_symbol(self, symbol: str) -> bool:
        return symbol in self._row_keys

    @staticmethod
    def select_coins(
        snapshot: Snapshot,
        query: str,
        dw_only: bool,
        exchange_filter: str,
        sort_column: str,
        sort_desc: bool,
    ) -> list[CoinState]:
        rows: list[CoinState] = []
        query_upper = query.strip().upper()

        for symbol, coin in snapshot.coin_states.items():
            if query_upper:
                name = snapshot.korean_names.get(symbol, "")
                if (
                    query_upper not in symbol.upper()
                    and query_upper not in name.upper()
                ):
                    continue
            if exchange_filter != "ALL" and not CoinTable._match_exchange_filter(
                coin, exchange_filter
            ):
                continue
            if dw_only and not CoinTable._is_wallet_blocked(
                snapshot.wallet_status.get(symbol)
            ):
                continue
            rows.append(coin)

        rows.sort(
            key=lambda coin: CoinTable._sort_key(coin, sort_column),
            reverse=sort_desc,
        )
        return rows

    @staticmethod
    def _sort_key(coin: CoinState, sort_column: str) -> float | str:
        column = sort_column if sort_column in CoinTable.SORT_COLUMNS else "Kimp"
        if column == "Symbol":
            return coin.symbol.upper()
        if column == "Upbit":
            return coin.upbit_price if coin.upbit_price is not None else float("-inf")
        if column == "Bithumb":
            return (
                coin.bithumb_price if coin.bithumb_price is not None else float("-inf")
            )
        if column == "Binance":
            return coin.binance_krw if coin.binance_krw is not None else float("-inf")
        if column == "Bybit":
            return coin.bybit_krw if coin.bybit_krw is not None else float("-inf")
        value = CoinTable._primary_kimchi(coin)
        return value if value is not None else float("-inf")

    @staticmethod
    def _primary_kimchi(coin: CoinState) -> float | None:
        return (
            coin.upbit_kimchi if coin.upbit_kimchi is not None else coin.bithumb_kimchi
        )

    @staticmethod
    def _is_wallet_blocked(status: CoinWalletStatus | None) -> bool:
        if status is None:
            return False
        for ex_status in [
            status.upbit,
            status.bithumb,
            status.binance,
            status.bybit,
            status.okx,
        ]:
            if ex_status and (not ex_status.deposit or not ex_status.withdraw):
                return True
        return False

    @staticmethod
    def _match_exchange_filter(coin: CoinState, exchange: str) -> bool:
        mapping = {
            "UP": coin.upbit_price,
            "BT": coin.bithumb_price,
            "BN": coin.binance_krw,
            "BB": coin.bybit_krw,
            "OK": coin.okx_krw,
        }
        return mapping.get(exchange) is not None
