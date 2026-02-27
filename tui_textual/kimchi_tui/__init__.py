from __future__ import annotations

from pathlib import Path

_SRC_PACKAGE = Path(__file__).resolve().parent.parent / "src" / "kimchi_tui"
if _SRC_PACKAGE.exists():
    __path__.append(str(_SRC_PACKAGE))

from kimchi_tui.app import KimchiApp  # type: ignore  # noqa: E402


def main() -> None:
    KimchiApp().run()
