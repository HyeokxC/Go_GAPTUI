from __future__ import annotations

from textual.widgets import Static

from ..colors import (
    C,
    exchange_tag_color,
    get_futures_basis_color,
    get_kimchi_color,
)
from ..formatters import fmt_krw_price, fmt_signed_pct, fmt_usd_price
from ..indicators import dw_dots
from ..models import ExchangeWalletStatus, Snapshot


class CoinDetailPanel(Static):
    def __init__(self) -> None:
        super().__init__(id="coin-detail-panel")

    def update_detail(self, symbol: str, snapshot: Snapshot) -> None:
        coin = snapshot.coin_states.get(symbol)
        if coin is None:
            self.update(f"[{C.TEXT_MUTED}]No coin data available[/]")
            return

        orderbook = snapshot.orderbook_info.get(symbol)
        wallet = snapshot.wallet_status.get(symbol)

        lines = [
            f"[bold {C.ACCENT}]━━━ {symbol} Detail ━━━[/]",
            "",
            f"[bold {C.ACCENT}]Exchange Grid[/]",
            self._exchange_row(
                code="UP",
                krw_price=coin.upbit_price,
                usd_price=self._usd_from_krw(coin.upbit_price, snapshot.usdt_krw),
                kimp=coin.upbit_kimchi,
                dw=wallet.upbit if wallet else None,
                krw_first=True,
            ),
            self._exchange_row(
                code="BT",
                krw_price=coin.bithumb_price,
                usd_price=self._usd_from_krw(coin.bithumb_price, snapshot.usdt_krw),
                kimp=coin.bithumb_kimchi,
                dw=wallet.bithumb if wallet else None,
                krw_first=True,
            ),
            self._exchange_row(
                code="BN",
                krw_price=coin.binance_krw,
                usd_price=coin.binance_price,
                kimp=None,
                dw=wallet.binance if wallet else None,
                krw_first=False,
            ),
            self._exchange_row(
                code="BB",
                krw_price=coin.bybit_krw,
                usd_price=coin.bybit_price,
                kimp=coin.bybit_kimchi_up,
                dw=wallet.bybit if wallet else None,
                krw_first=False,
            ),
            self._exchange_row(
                code="OK",
                krw_price=coin.okx_krw,
                usd_price=coin.okx_price,
                kimp=coin.okx_kimchi_up,
                dw=wallet.okx if wallet else None,
                krw_first=False,
            ),
            "",
            f"[bold {C.ACCENT}]Domestic Gap[/]  {self._domestic_gap_text(coin.domestic_gap)}",
            "",
            f"[bold {C.ACCENT}]Futures Basis[/]  {self._futures_pill('BNF', self._basis_pct(coin.binance_price, coin.binance_futures_price))}  {self._futures_pill('BBF', self._basis_pct(coin.bybit_price, coin.bybit_futures_price))}",
            "",
            f"[bold {C.ACCENT}]Orderbook[/]  BT buy {fmt_signed_pct(orderbook.bt_buy_slippage if orderbook else None)}  |  UP sell {fmt_signed_pct(orderbook.up_sell_slippage if orderbook else None)}  |  Spread {fmt_signed_pct(orderbook.real_gap_bt_up if orderbook else None)}",
        ]
        self.update("\n".join(lines))

    def _exchange_row(
        self,
        *,
        code: str,
        krw_price: float | None,
        usd_price: float | None,
        kimp: float | None,
        dw: ExchangeWalletStatus | None,
        krw_first: bool,
    ) -> str:
        tag = f"[bold {C.PILL_TEXT} on {exchange_tag_color(code)}] {code} [/]"
        krw_text = f"₩{fmt_krw_price(krw_price)}" if krw_price is not None else "₩-"
        usd_text = self._fmt_usd(usd_price, compact=krw_first)
        if krw_first:
            prices = f"{krw_text:<14}  ({usd_text})"
        else:
            prices = f"{usd_text:<13}  {krw_text}"

        if kimp is None:
            kimp_text = f"[{C.TEXT_MUTED}]KIMP: --[/]"
        else:
            kimp_color = get_kimchi_color(kimp)
            sign = "+" if kimp > 0 else ""
            kimp_text = f"[{kimp_color}]KIMP: {sign}{kimp:.2f}%[/]"

        dots = dw_dots(dw.deposit if dw else None, dw.withdraw if dw else None)
        return f"  {tag}  {prices}  {kimp_text}  {dots}"

    def _futures_pill(self, label: str, basis: float | None) -> str:
        if basis is None:
            return f"[bold {C.TEXT_MUTED}] {label} -- [/]"
        color = get_futures_basis_color(basis)
        sign = "+" if basis > 0 else ""
        return f"[bold {C.PILL_TEXT} on {color}] {label} {sign}{basis:.2f}% [/]"

    def _domestic_gap_text(self, value: float | None) -> str:
        if value is None:
            return f"[{C.TEXT_MUTED}]--[/]"
        abs_value = abs(value)
        if abs_value >= 3.0:
            color = C.DOMGAP_HIGH
        elif abs_value >= 1.0:
            color = C.DOMGAP_MID
        else:
            color = C.DOMGAP_LOW
        sign = "+" if value > 0 else ""
        return f"[{color}]{sign}{value:.2f}%[/]"

    def _fmt_usd(self, value: float | None, *, compact: bool) -> str:
        if value is None:
            return "$-"
        if compact:
            if value >= 1.0:
                return f"${value:,.2f}"
            return f"${value:.6f}"
        return f"${fmt_usd_price(value)}"

    def _usd_from_krw(self, krw: float | None, usdt_krw: float | None) -> float | None:
        if krw is None or usdt_krw is None or usdt_krw <= 0:
            return None
        return krw / usdt_krw

    def _basis_pct(self, spot: float | None, futures: float | None) -> float | None:
        if spot is None or futures is None or futures == 0:
            return None
        return ((spot - futures) / futures) * 100
