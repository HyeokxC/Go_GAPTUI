"""Shared numeric formatters — single source of truth for display precision."""

from __future__ import annotations

from typing import Optional

from rich.text import Text

from .colors import C


# ── Price formatters ─────────────────────────────────────────


# KRW: >=1000 → 0dp, >=1 → 2dp, <1 → 4dp
def fmt_krw_price(value: Optional[float]) -> str:
    if value is None:
        return "-"
    if value >= 1000:
        return f"{value:,.0f}"
    if value >= 1:
        return f"{value:,.2f}"
    return f"{value:,.4f}"


# USD: >=1 → 4dp, <1 → 6dp
def fmt_usd_price(value: Optional[float]) -> str:
    if value is None:
        return "-"
    if value >= 1.0:
        return f"{value:,.4f}"
    return f"{value:.6f}"


def fmt_usd_equiv(krw_price: Optional[float], usdt_krw: Optional[float]) -> str:
    if krw_price is None or usdt_krw is None or usdt_krw <= 0:
        return "$-"
    usd = krw_price / usdt_krw
    if usd >= 1.0:
        return f"${usd:,.2f}"
    return f"${usd:.6f}"


# ── Amount / rate formatters ─────────────────────────────────


def fmt_amount(value: float) -> str:
    return f"{value:.8f}".rstrip("0").rstrip(".") or "0"


def fmt_rate(value: Optional[float]) -> str:
    if value is None:
        return "--"
    return f"{value:,.0f}"


def fmt_rate_detailed(value: Optional[float]) -> str:
    if value is None:
        return "-"
    return f"{value:,.1f}"


# ── Percent formatters ───────────────────────────────────────


def fmt_signed_pct(value: Optional[float]) -> str:
    if value is None:
        return "-"
    sign = "+" if value > 0 else ""
    return f"{sign}{value:.2f}%"


def fmt_pct_rich(value: Optional[float]) -> Text:
    if value is None:
        return Text("-")
    if value >= 5.0:
        style = f"bold {C.KIMP_5PLUS}"
    elif value >= 3.0:
        style = f"bold {C.KIMP_3TO5}"
    elif value >= 1.0:
        style = C.KIMP_1TO3
    elif value > -1.0:
        style = C.KIMP_NEUTRAL
    elif value > -3.0:
        style = C.KIMP_NEG1TO3
    elif value > -5.0:
        style = f"bold {C.KIMP_NEG3TO5}"
    else:
        style = f"bold {C.KIMP_NEG5PLUS}"
    sign = "+" if value > 0 else ""
    return Text(f"{sign}{value:.2f}", style=style)


def fmt_domgap_rich(value: Optional[float]) -> Text:
    if value is None:
        return Text("-")
    abs_value = abs(value)
    if abs_value >= 3.0:
        style = f"bold {C.DOMGAP_HIGH}"
    elif abs_value >= 1.0:
        style = C.DOMGAP_MID
    else:
        style = C.DOMGAP_LOW
    sign = "+" if value > 0 else ""
    return Text(f"{sign}{value:.2f}", style=style)


def fmt_kimp_markup(value: Optional[float]) -> str:
    if value is None:
        return f"[{C.TEXT_MUTED}]-[/]"
    rendered = fmt_pct_rich(value)
    style = rendered.style or ""
    return f"[{style}]{rendered.plain}%[/]"


def fmt_domgap_markup(value: Optional[float]) -> str:
    if value is None:
        return f"[{C.TEXT_MUTED}]-[/]"
    rendered = fmt_domgap_rich(value)
    style = rendered.style or ""
    return f"[{style}]{rendered.plain}%[/]"


# ── Time formatters ──────────────────────────────────────────


def fmt_elapsed(secs: int) -> str:
    if secs >= 3600:
        return f"{secs // 3600}:{(secs % 3600) // 60:02d}:{secs % 60:02d}"
    return f"{secs // 60}:{secs % 60:02d}"
