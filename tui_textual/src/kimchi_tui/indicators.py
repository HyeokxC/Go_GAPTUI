"""Shared status indicators — glyph + text for accessibility (P2)."""

from __future__ import annotations

from typing import Optional

from rich.text import Text

from .colors import C


# ── Transfer step markers ────────────────────────────────────


def step_marker(status: str) -> str:
    if status == "Completed":
        return f"[{C.SUCCESS}][✓][/]"
    if status == "InProgress":
        return f"[{C.PROGRESS}][⟳][/]"
    if status == "Failed":
        return f"[{C.ERROR}][✗][/]"
    return f"[{C.TEXT_MUTED}][ ][/]"


# ── Transfer job status ──────────────────────────────────────


def job_status_badge(is_executing: bool, error_message: Optional[str]) -> str:
    if is_executing:
        return f"[{C.PROGRESS}]⟳ RUNNING[/]"
    if error_message is not None:
        return f"[{C.ERROR}]✗ FAILED[/]"
    return f"[{C.SUCCESS}]✓ DONE[/]"


# ── D/W status indicators ───────────────────────────────────
# Uppercase=OK, lowercase=blocked (non-color cue)


def dw_dots(deposit_ok: Optional[bool], withdraw_ok: Optional[bool]) -> str:
    if deposit_ok is None and withdraw_ok is None:
        return f"[{C.TEXT_MUTED}]D- W-[/]"
    d_dot = f"[{C.DW_OK}]D●[/]" if deposit_ok else f"[{C.DW_BLOCKED}]d●[/]"
    w_dot = f"[{C.DW_OK}]W●[/]" if withdraw_ok else f"[{C.DW_BLOCKED}]w●[/]"
    return f"{d_dot}{w_dot}"


def dw_checkmarks(deposit_ok: Optional[bool], withdraw_ok: Optional[bool]) -> str:
    if deposit_ok is None and withdraw_ok is None:
        return f"D[{C.TEXT_MUTED}]-[/]W[{C.TEXT_MUTED}]-[/]"
    d_mark = f"[{C.DW_OK}]✓[/]" if deposit_ok else f"[{C.DW_BLOCKED}]✗[/]"
    w_mark = f"[{C.DW_OK}]✓[/]" if withdraw_ok else f"[{C.DW_BLOCKED}]✗[/]"
    return f"D{d_mark}W{w_mark}"


def dw_cell(deposit_ok: Optional[bool], withdraw_ok: Optional[bool]) -> Text:
    if deposit_ok is None and withdraw_ok is None:
        return Text("-", style="dim")
    dep_char = "D" if deposit_ok else "d"
    wd_char = "W" if withdraw_ok else "w"
    style = C.DW_OK if (deposit_ok and withdraw_ok) else C.DW_BLOCKED
    return Text(f"{dep_char}/{wd_char}", style=style)


def dw_cell_with_chains(
    deposit_ok: Optional[bool],
    withdraw_ok: Optional[bool],
    blocked_chains: list[str],
) -> Text:
    text = dw_cell(deposit_ok, withdraw_ok)
    for chain in blocked_chains[:2]:
        text.append(" ")
        text.append(chain, style=f"bold white on {C.DW_CHAIN_BADGE_BG}")
    return text


# ── Log level badges ─────────────────────────────────────────
# Keyword matching: error/fail/blocked → ERR, warn → WRN, info/success → INF


def log_level_badge(level: str) -> str:
    lower = level.lower()
    if "error" in lower or "fail" in lower or "blocked" in lower:
        return f"[bold white on {C.ERROR}] ERR [/]"
    if "warn" in lower:
        return f"[bold white on {C.WARNING}] WRN [/]"
    if "info" in lower or "success" in lower:
        return f"[bold white on {C.INFO}] INF [/]"
    return f"[bold white on {C.BORDER}] DBG [/]"


def transfer_log_badge(is_error: bool) -> str:
    if is_error:
        return f"[bold white on {C.ERROR}] ERR [/]"
    return f"[bold white on {C.INFO}] INF [/]"


# ── Scenario badges ──────────────────────────────────────────


def scenario_badge(scenario: str) -> str:
    from .colors import get_scenario_badge_colors, SCENARIO_BADGE_LABELS

    label = SCENARIO_BADGE_LABELS.get(scenario, scenario)
    color = get_scenario_badge_colors().get(scenario, C.TEXT_SECONDARY)
    return f"[bold {color}][{label}][/]"


def scenario_state_text(is_active: bool) -> str:
    if is_active:
        return f"[{C.SUCCESS}]ACTIVE[/]"
    return f"[{C.TEXT_MUTED}]closed[/]"


# ── Connection / freshness ───────────────────────────────────


def freshness_indicator(age_ms: Optional[int]) -> str:
    if age_ms is None:
        return f"[{C.WARNING}]● Connecting...[/]"
    age_sec = age_ms / 1000
    if age_ms < 5000:
        return f"[{C.SUCCESS}]● Live {age_sec:.1f}s[/]"
    return f"[{C.ERROR}]● Stale {age_sec:.1f}s[/]"


def alert_indicator(active_count: int) -> str:
    if active_count > 0:
        return f"[{C.WARNING}]⚠ {active_count} alerts[/]"
    return f"[{C.TEXT_MUTED}]✓ No alerts[/]"
