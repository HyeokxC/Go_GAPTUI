"""Semantic color tokens for Rich markup — canonical source for all UI colors.

Mirrors the TCSS theme files.  Widget code reads ``C.<TOKEN>``.
"""

from __future__ import annotations


class Dark:
    # ── Surface ──────────────────────────────────────────────
    PANEL_BG = "#161b22"
    BG_DEEP = "#0d1117"
    CARD_BG = "#1c2128"
    CARD_BG_ALT = "#161b22"
    WIDGET_BG = "#21262d"
    BORDER = "#30363d"
    STATUSBAR_BG = "#0f3460"
    HOVER_BG = "#30363d"

    # ── Text ─────────────────────────────────────────────────
    TEXT_PRIMARY = "#e6edf3"
    TEXT_SECONDARY = "#8b949e"
    TEXT_MUTED = "#484f58"

    # ── Action / Accent ──────────────────────────────────────
    ACCENT = "#58a6ff"
    ACCENT_ACTIVE = "#1f6feb"
    SELECTION_BG = "#1f6feb33"

    # ── Status ───────────────────────────────────────────────
    SUCCESS = "#3fb950"
    ERROR = "#f85149"
    WARNING = "#d29922"
    INFO = "#58a6ff"
    PROGRESS = "#58a6ff"

    RED = "#f85149"
    YELLOW = "#d29922"
    GREEN = "#3fb950"
    PURPLE = "#bc8cff"
    PILL_TEXT = "#000000"
    INACTIVE_BG = "#323232"
    INACTIVE_TEXT = "#646464"
    MID_PREMIUM = "#787878"

    # ── Kimchi-premium tiers (Section 4.2) ───────────────────
    # Doc convention: red = hot/high premium, blue = negative
    KIMP_5PLUS = "#f85149"  # bold red  (>=5%)
    KIMP_3TO5 = "#d29922"  # yellow    (>=3%)
    KIMP_1TO3 = "#3fb950"  # green     (>=1%)
    KIMP_NEUTRAL = "#8b949e"  # gray      (-1~+1%)
    KIMP_NEG1TO3 = "#58a6ff"  # blue      (-1~-3%)
    KIMP_NEG3TO5 = "#58a6ff"  # deep blue (-3~-5%)
    KIMP_NEG5PLUS = "#58a6ff"  # bold blue (<=-5%)

    # ── Domestic-gap tiers ───────────────────────────────────
    DOMGAP_HIGH = "#ff8c00"
    DOMGAP_MID = "#d2991a"
    DOMGAP_LOW = "#8b949e"

    # ── Futures-basis tiers (Section 4.2) ────────────────────
    FUTBASIS_HIGH = "#f85149"
    FUTBASIS_MID = "#d29922"
    FUTBASIS_LOW = "#3fb950"

    # ── Badge backgrounds ────────────────────────────────────
    BADGE_POSITIVE_BG = "#238636"
    BADGE_NEGATIVE_BG = "#da3633"
    BADGE_BT = "#30363d"
    BADGE_UP = "#58a6ff"

    # ── Scenario badges ──────────────────────────────────────
    BADGE_KIMP = "#d29922"
    BADGE_DOMGAP = "#58a6ff"
    BADGE_FUTBASIS = "#bc8cff"

    EXCHANGE_UPBIT = "#58a6ff"
    EXCHANGE_BITHUMB = "#bc8cff"
    EXCHANGE_BINANCE = "#d29922"
    EXCHANGE_BYBIT = "#f85149"
    EXCHANGE_OKX = "#3fb950"
    EXCHANGE_DEFAULT = "#8b949e"

    # ── DW status ────────────────────────────────────────────
    DW_OK = "green"
    DW_BLOCKED = "red"
    DW_CHAIN_BADGE_BG = "#da3633"


class Light:
    # ── Surface ──────────────────────────────────────────────
    PANEL_BG = "#ffffff"
    BG_DEEP = "#f6f8fa"
    CARD_BG = "#ffffff"
    CARD_BG_ALT = "#f6f8fa"
    WIDGET_BG = "#f6f8fa"
    BORDER = "#d0d7de"
    STATUSBAR_BG = "#0969da"
    HOVER_BG = "#d8dee4"

    # ── Text ─────────────────────────────────────────────────
    TEXT_PRIMARY = "#24292f"
    TEXT_SECONDARY = "#656d76"
    TEXT_MUTED = "#8c959f"

    # ── Action / Accent ──────────────────────────────────────
    ACCENT = "#0969da"
    ACCENT_ACTIVE = "#0550ae"
    SELECTION_BG = "#0550ae22"

    # ── Status ───────────────────────────────────────────────
    SUCCESS = "#1a7f37"
    ERROR = "#cf222e"
    WARNING = "#9a6700"
    INFO = "#0969da"
    PROGRESS = "#0969da"

    RED = "#cf222e"
    YELLOW = "#9a6700"
    GREEN = "#1a7f37"
    PURPLE = "#8250df"
    PILL_TEXT = "#ffffff"
    INACTIVE_BG = "#e1e4e8"
    INACTIVE_TEXT = "#8c959f"
    MID_PREMIUM = "#8c959f"

    # ── Kimchi-premium tiers (Section 4.2) ───────────────────
    # Doc convention: red = hot/high premium, blue = negative
    KIMP_5PLUS = "#cf222e"  # bold red  (>=5%)
    KIMP_3TO5 = "#9a6700"  # yellow    (>=3%)
    KIMP_1TO3 = "#1a7f37"  # green     (>=1%)
    KIMP_NEUTRAL = "#57606a"  # gray      (-1~+1%)
    KIMP_NEG1TO3 = "#0969da"  # blue      (-1~-3%)
    KIMP_NEG3TO5 = "#0969da"  # deep blue (-3~-5%)
    KIMP_NEG5PLUS = "#0969da"  # bold blue (<=-5%)

    # ── Domestic-gap tiers ───────────────────────────────────
    DOMGAP_HIGH = "#9a6700"
    DOMGAP_MID = "#9a6700"
    DOMGAP_LOW = "#57606a"

    # ── Futures-basis tiers ──────────────────────────────────
    FUTBASIS_HIGH = "#cf222e"
    FUTBASIS_MID = "#9a6700"
    FUTBASIS_LOW = "#1a7f37"

    # ── Badge backgrounds ────────────────────────────────────
    BADGE_POSITIVE_BG = "#1a7f37"
    BADGE_NEGATIVE_BG = "#cf222e"
    BADGE_BT = "#d0d7de"
    BADGE_UP = "#0969da"

    # ── Scenario badges ──────────────────────────────────────
    BADGE_KIMP = "#9a6700"
    BADGE_DOMGAP = "#0969da"
    BADGE_FUTBASIS = "#8250df"

    EXCHANGE_UPBIT = "#0969da"
    EXCHANGE_BITHUMB = "#8250df"
    EXCHANGE_BINANCE = "#9a6700"
    EXCHANGE_BYBIT = "#cf222e"
    EXCHANGE_OKX = "#1a7f37"
    EXCHANGE_DEFAULT = "#8c959f"

    # ── DW status ────────────────────────────────────────────
    DW_OK = "#1a7f37"
    DW_BLOCKED = "#cf222e"
    DW_CHAIN_BADGE_BG = "#cf222e"


def get_kimchi_color(kimp: float) -> str:
    if kimp >= 5.0:
        return C.RED
    elif kimp >= 3.0:
        return C.YELLOW
    elif kimp >= 0.0:
        return C.GREEN
    else:
        return C.ACCENT


def scenario_color(scenario_type: str) -> str:
    colors = get_scenario_badge_colors()
    return colors.get(scenario_type, C.TEXT_SECONDARY)


def exchange_tag_color(exchange: str) -> str:
    mapping = {
        "Upbit": C.EXCHANGE_UPBIT,
        "UP": C.EXCHANGE_UPBIT,
        "Bithumb": C.EXCHANGE_BITHUMB,
        "BT": C.EXCHANGE_BITHUMB,
        "Binance": C.EXCHANGE_BINANCE,
        "BN": C.EXCHANGE_BINANCE,
        "Bybit": C.EXCHANGE_BYBIT,
        "BB": C.EXCHANGE_BYBIT,
        "Okx": C.EXCHANGE_OKX,
        "OK": C.EXCHANGE_OKX,
        "OKX": C.EXCHANGE_OKX,
    }
    return mapping.get(exchange, C.EXCHANGE_DEFAULT)


def get_futures_basis_color(basis: float) -> str:
    abs_val = abs(basis)
    if abs_val >= 1.0:
        return C.RED
    elif abs_val >= 0.3:
        return C.YELLOW
    else:
        return C.GREEN


class _ThemeProxy:
    """Proxy that delegates attribute reads to the active theme class.

    All existing ``from .colors import C`` bindings keep working because
    ``C`` is a stable singleton — only its *backing* class changes.
    """

    _theme: type = Dark

    def __getattr__(self, name: str) -> str:
        try:
            return getattr(self._theme, name)  # type: ignore[return-value]
        except AttributeError:
            raise AttributeError(f"Theme has no color token '{name}'") from None


C = _ThemeProxy()


def set_theme(*, dark: bool) -> None:
    """Switch the global color proxy between Dark and Light."""
    C._theme = Dark if dark else Light  # noqa: SLF001


SCENARIO_BADGE_LABELS: dict[str, str] = {
    "GapThreshold": "KIMP",
    "DomesticGap": "DOM-GAP",
    "FutBasis": "FUT%",
}


def get_scenario_badge_colors() -> dict[str, str]:
    """Return scenario badge colors for the *active* theme."""
    return {
        "GapThreshold": C.BADGE_KIMP,
        "DomesticGap": C.BADGE_DOMGAP,
        "FutBasis": C.BADGE_FUTBASIS,
    }


# Backward-compat alias — static dict kept for imports that don't need live updates.
SCENARIO_BADGE_COLORS: dict[str, str] = {
    "GapThreshold": Dark.BADGE_KIMP,
    "DomesticGap": Dark.BADGE_DOMGAP,
    "FutBasis": Dark.BADGE_FUTBASIS,
}
