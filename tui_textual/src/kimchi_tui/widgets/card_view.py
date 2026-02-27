from __future__ import annotations

from typing import Optional

from textual.widgets import Static

from ..colors import C
from ..formatters import fmt_kimp_markup, fmt_krw_price, fmt_signed_pct, fmt_usd_equiv, fmt_usd_price
from ..indicators import dw_checkmarks, dw_dots
from ..models import ExchangeWalletStatus, OrderbookInfo, Snapshot
from .coin_table import CoinTable


class CardView(Static):
    def __init__(self) -> None:
        super().__init__(id="card-view")

    def update_cards(
        self,
        snapshot: Snapshot,
        query: str,
        dw_only: bool,
        exchange_filter: str,
        sort_column: str,
        sort_desc: bool,
        favorites: set[str],
    ) -> None:
        coins = CoinTable.select_coins(
            snapshot,
            query=query,
            dw_only=dw_only,
            exchange_filter=exchange_filter,
            sort_column=sort_column,
            sort_desc=sort_desc,
        )

        if not coins:
            self.update(f"[{C.TEXT_MUTED}]No coins match current filters[/]")
            return

        usdt_krw = snapshot.usdt_krw
        blocks: list[str] = []

        for coin in coins:
            symbol = coin.symbol
            name = snapshot.korean_names.get(symbol, "-")
            kimchi = CoinTable._primary_kimchi(coin)
            wallet = snapshot.wallet_status.get(symbol)
            ob = snapshot.orderbook_info.get(symbol)
            star = "★" if symbol in favorites else " "

            kimp_str = fmt_kimp_markup(kimchi)
            line1 = f"{star} [bold]{symbol}[/]  {name}  {kimp_str}"

            bt_wallet = getattr(wallet, "bithumb", None) if wallet else None
            up_wallet = getattr(wallet, "upbit", None) if wallet else None
            bt_dw = self._extract_dw_dots(bt_wallet)
            up_dw = self._extract_dw_dots(up_wallet)

            bt_price_str = fmt_krw_price(coin.bithumb_price)
            bt_usd_str = fmt_usd_equiv(coin.bithumb_price, usdt_krw)
            up_price_str = fmt_krw_price(coin.upbit_price)
            up_usd_str = fmt_usd_equiv(coin.upbit_price, usdt_krw)

            line2 = (
                f"  [bold white on {C.BADGE_BT}] BT [/] ₩{bt_price_str} ({bt_usd_str})  {bt_dw}"
                f"    [bold white on {C.BADGE_UP}] UP [/] ₩{up_price_str} ({up_usd_str})  {up_dw}"
            )

            line3 = self._fmt_slip_spread(ob)
            line4 = self._fmt_exchange_badges(coin)

            card_lines = [line1, line2, line3]
            if line4:
                card_lines.append(line4)
            binance_usd = getattr(coin, "binance_price", None)
            if binance_usd is not None:
                line5 = f"  [{C.TEXT_MUTED}]${fmt_usd_price(binance_usd)}[/]"
                card_lines.append(line5)
            card_lines.append(
                f"[{C.TEXT_MUTED}]────────────────────────────────────────────────[/]"
            )

            blocks.append("\n".join(card_lines))

        self.update("\n".join(blocks))

    @staticmethod
    def _extract_dw_dots(exchange_wallet: Optional[ExchangeWalletStatus]) -> str:
        if exchange_wallet is None:
            return dw_dots(None, None)
        return dw_dots(exchange_wallet.deposit, exchange_wallet.withdraw)

    @staticmethod
    def _fmt_slip_spread(ob: Optional[OrderbookInfo]) -> str:
        if ob is None:
            return f"  [{C.TEXT_MUTED}]slip --  |  sprd --[/]"
        bt_slip = fmt_signed_pct(ob.bt_buy_slippage if ob.bt_buy_slippage else None)
        up_slip = fmt_signed_pct(ob.up_sell_slippage if ob.up_sell_slippage else None)
        real_gap = fmt_signed_pct(ob.real_gap_bt_up if ob.real_gap_bt_up else None)
        return f"  slip BT:{bt_slip} UP:{up_slip}  |  real_gap:{real_gap}"

    @staticmethod
    def _fmt_exchange_badges(coin: object) -> str:
        badges: list[str] = []
        fut_basis = getattr(coin, "futures_basis", None)
        if fut_basis is not None:
            bg = C.BADGE_POSITIVE_BG if fut_basis >= 0 else C.BADGE_NEGATIVE_BG
            sign = "+" if fut_basis >= 0 else ""
            badges.append(f"[bold white on {bg}] BNF {sign}{fut_basis:.2f}% [/]")
        bb_kimp = getattr(coin, "bybit_kimchi_up", None)
        if bb_kimp is not None:
            bg = C.BADGE_POSITIVE_BG if bb_kimp >= 0 else C.BADGE_NEGATIVE_BG
            sign = "+" if bb_kimp >= 0 else ""
            badges.append(f"[bold white on {bg}] BB {sign}{bb_kimp:.2f}% [/]")
        ok_kimp = getattr(coin, "okx_kimchi_up", None)
        if ok_kimp is not None:
            bg = C.BADGE_POSITIVE_BG if ok_kimp >= 0 else C.BADGE_NEGATIVE_BG
            sign = "+" if ok_kimp >= 0 else ""
            badges.append(f"[bold white on {bg}] OK {sign}{ok_kimp:.2f}% [/]")
        if not badges:
            return ""
        return "  " + "  ".join(badges)
