from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Optional


@dataclass
class CoinState:
    symbol: str
    upbit_price: Optional[float] = None
    bithumb_price: Optional[float] = None
    binance_price: Optional[float] = None
    binance_krw: Optional[float] = None
    bybit_price: Optional[float] = None
    bybit_krw: Optional[float] = None
    okx_price: Optional[float] = None
    okx_krw: Optional[float] = None
    upbit_kimchi: Optional[float] = None
    bithumb_kimchi: Optional[float] = None
    bybit_kimchi_up: Optional[float] = None
    bybit_kimchi_bt: Optional[float] = None
    okx_kimchi_up: Optional[float] = None
    okx_kimchi_bt: Optional[float] = None
    domestic_gap: Optional[float] = None
    binance_futures_price: Optional[float] = None
    bybit_futures_price: Optional[float] = None
    futures_basis: Optional[float] = None
    timestamp: Optional[str] = None


@dataclass
class ExchangeWalletStatus:
    deposit: bool = False
    withdraw: bool = False
    deposit_blocked_chains: list[str] = field(default_factory=list)
    withdraw_blocked_chains: list[str] = field(default_factory=list)


@dataclass
class CoinWalletStatus:
    upbit: Optional[ExchangeWalletStatus] = None
    bithumb: Optional[ExchangeWalletStatus] = None
    binance: Optional[ExchangeWalletStatus] = None
    bybit: Optional[ExchangeWalletStatus] = None
    okx: Optional[ExchangeWalletStatus] = None


@dataclass
class OrderbookInfo:
    bt_buy_slippage: float = 0.0
    up_sell_slippage: float = 0.0
    real_gap_bt_up: float = 0.0
    real_kimp_up: dict[str, float] = field(default_factory=dict)
    real_kimp_bt: dict[str, float] = field(default_factory=dict)


@dataclass
class ThreadEntry:
    timestamp: str = ""
    message: str = ""


@dataclass
class LogThread:
    id: int = 0
    symbol: str = ""
    scenario: str = ""
    key: str = ""
    main_message: str = ""
    main_timestamp: str = ""
    sub_entries: list[ThreadEntry] = field(default_factory=list)
    is_active: bool = False
    initial_value: float = 0.0
    last_logged_value: float = 0.0


@dataclass
class LogEntry:
    timestamp: str = ""
    symbol: str = ""
    message: str = ""
    log_type: str = ""


@dataclass
class NetworkInfo:
    network: str = ""
    to_network: Optional[str] = None
    display_name: str = ""
    deposit_enabled: bool = False
    withdraw_enabled: bool = False
    withdraw_fee: Optional[float] = None
    withdraw_min: Optional[float] = None
    needs_memo: bool = False


@dataclass
class TransferLogEntry:
    timestamp: str = ""
    message: str = ""
    is_error: bool = False


@dataclass
class TransferState:
    selected_coin: str = ""
    from_exchange: str = ""
    to_exchange: str = ""
    available_networks: list[NetworkInfo] = field(default_factory=list)
    selected_network_idx: Optional[int] = None
    amount: str = ""
    balance: Optional[float] = None
    deposit_address: str = ""
    deposit_tag: str = ""
    to_is_personal_wallet: bool = False
    personal_wallet_address: str = ""
    personal_wallet_tag: str = ""
    auto_buy_before_transfer: bool = False
    auto_sell_on_arrival: bool = False
    market_order_pending: bool = False
    market_order_result: Optional[str] = None
    logs: list[TransferLogEntry] = field(default_factory=list)


@dataclass
class ScenarioConfig:
    gap_threshold_percent: float = 5.0
    domestic_gap_threshold: float = 1.5
    fut_basis_threshold: float = 0.5


@dataclass
class StepInfo:
    step: str = ""
    title: str = ""
    status: str = "Pending"
    message: Optional[str] = None


@dataclass
class TransferJob:
    id: int = 0
    coin: str = ""
    amount: float = 0.0
    from_exchange: str = ""
    to_exchange: str = ""
    network_display: str = ""
    current_step: str = "Idle"
    steps: list[StepInfo] = field(default_factory=list)
    is_executing: bool = False
    error_message: Optional[str] = None
    started_at_secs: int = 0


@dataclass
class Snapshot:
    coin_states: dict[str, CoinState] = field(default_factory=dict)
    wallet_status: dict[str, CoinWalletStatus] = field(default_factory=dict)
    korean_names: dict[str, str] = field(default_factory=dict)
    logs: list[LogEntry] = field(default_factory=list)
    orderbook_info: dict[str, OrderbookInfo] = field(default_factory=dict)
    scenario_threads: list[LogThread] = field(default_factory=list)
    usdt_krw: Optional[float] = None
    usd_krw_forex: Optional[float] = None
    last_ticker_age_ms: Optional[int] = None
    transfer: TransferState = field(default_factory=TransferState)
    scenario_config: ScenarioConfig = field(default_factory=ScenarioConfig)
    transfer_jobs: list[TransferJob] = field(default_factory=list)


def _as_float(value: Any) -> Optional[float]:
    if value is None:
        return None
    if isinstance(value, (int, float)):
        return float(value)
    if isinstance(value, str):
        try:
            return float(value)
        except ValueError:
            return None
    return None


def _as_int(value: Any) -> Optional[int]:
    if value is None:
        return None
    if isinstance(value, int):
        return value
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return None


def _as_bool(value: Any, default: bool = False) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        lowered = value.lower()
        if lowered in {"true", "1", "yes", "on"}:
            return True
        if lowered in {"false", "0", "no", "off"}:
            return False
    return default


def _parse_exchange_wallet_status(data: Any) -> Optional[ExchangeWalletStatus]:
    if not isinstance(data, dict):
        return None
    return ExchangeWalletStatus(
        deposit=_as_bool(data.get("deposit"), False),
        withdraw=_as_bool(data.get("withdraw"), False),
        deposit_blocked_chains=[
            str(x) for x in data.get("deposit_blocked_chains", []) if x is not None
        ],
        withdraw_blocked_chains=[
            str(x) for x in data.get("withdraw_blocked_chains", []) if x is not None
        ],
    )


def _parse_coin_wallet_status(data: Any) -> CoinWalletStatus:
    if not isinstance(data, dict):
        return CoinWalletStatus()
    return CoinWalletStatus(
        upbit=_parse_exchange_wallet_status(data.get("upbit")),
        bithumb=_parse_exchange_wallet_status(data.get("bithumb")),
        binance=_parse_exchange_wallet_status(data.get("binance")),
        bybit=_parse_exchange_wallet_status(data.get("bybit")),
        okx=_parse_exchange_wallet_status(data.get("okx")),
    )


def _parse_coin_state(symbol: str, data: Any) -> CoinState:
    if not isinstance(data, dict):
        return CoinState(symbol=symbol)
    return CoinState(
        symbol=str(data.get("symbol") or symbol),
        upbit_price=_as_float(data.get("upbit_price")),
        bithumb_price=_as_float(data.get("bithumb_price")),
        binance_price=_as_float(data.get("binance_price")),
        binance_krw=_as_float(data.get("binance_krw")),
        bybit_price=_as_float(data.get("bybit_price")),
        bybit_krw=_as_float(data.get("bybit_krw")),
        okx_price=_as_float(data.get("okx_price")),
        okx_krw=_as_float(data.get("okx_krw")),
        upbit_kimchi=_as_float(data.get("upbit_kimchi")),
        bithumb_kimchi=_as_float(data.get("bithumb_kimchi")),
        bybit_kimchi_up=_as_float(data.get("bybit_kimchi_up")),
        bybit_kimchi_bt=_as_float(data.get("bybit_kimchi_bt")),
        okx_kimchi_up=_as_float(data.get("okx_kimchi_up")),
        okx_kimchi_bt=_as_float(data.get("okx_kimchi_bt")),
        domestic_gap=_as_float(data.get("domestic_gap")),
        binance_futures_price=_as_float(data.get("binance_futures_price")),
        bybit_futures_price=_as_float(data.get("bybit_futures_price")),
        futures_basis=_as_float(data.get("futures_basis")),
        timestamp=str(data.get("timestamp"))
        if data.get("timestamp") is not None
        else None,
    )


def _parse_orderbook_info(data: Any) -> OrderbookInfo:
    if not isinstance(data, dict):
        return OrderbookInfo()
    raw_up = data.get("real_kimp_up", {})
    raw_bt = data.get("real_kimp_bt", {})
    return OrderbookInfo(
        bt_buy_slippage=_as_float(data.get("bt_buy_slippage")) or 0.0,
        up_sell_slippage=_as_float(data.get("up_sell_slippage")) or 0.0,
        real_gap_bt_up=_as_float(data.get("real_gap_bt_up")) or 0.0,
        real_kimp_up={str(k): _as_float(v) or 0.0 for k, v in raw_up.items()}
        if isinstance(raw_up, dict)
        else {},
        real_kimp_bt={str(k): _as_float(v) or 0.0 for k, v in raw_bt.items()}
        if isinstance(raw_bt, dict)
        else {},
    )


def _parse_thread_entry(data: Any) -> ThreadEntry:
    if not isinstance(data, dict):
        return ThreadEntry()
    return ThreadEntry(
        timestamp=str(data.get("timestamp") or ""),
        message=str(data.get("message") or ""),
    )


def _parse_log_thread(data: Any) -> LogThread:
    if not isinstance(data, dict):
        return LogThread()
    return LogThread(
        id=_as_int(data.get("id")) or 0,
        symbol=str(data.get("symbol") or ""),
        scenario=str(data.get("scenario") or ""),
        key=str(data.get("key") or ""),
        main_message=str(data.get("main_message") or ""),
        main_timestamp=str(data.get("main_timestamp") or ""),
        sub_entries=[_parse_thread_entry(x) for x in data.get("sub_entries", [])],
        is_active=_as_bool(data.get("is_active"), False),
        initial_value=_as_float(data.get("initial_value")) or 0.0,
        last_logged_value=_as_float(data.get("last_logged_value")) or 0.0,
    )


def _parse_log_entry(data: Any) -> LogEntry:
    if not isinstance(data, dict):
        return LogEntry()
    return LogEntry(
        timestamp=str(data.get("timestamp") or ""),
        symbol=str(data.get("symbol") or ""),
        message=str(data.get("message") or ""),
        log_type=str(data.get("log_type") or ""),
    )


def _parse_network_info(data: Any) -> NetworkInfo:
    if not isinstance(data, dict):
        return NetworkInfo()
    return NetworkInfo(
        network=str(data.get("network") or ""),
        to_network=str(data.get("to_network"))
        if data.get("to_network") is not None
        else None,
        display_name=str(data.get("display_name") or ""),
        deposit_enabled=_as_bool(data.get("deposit_enabled"), False),
        withdraw_enabled=_as_bool(data.get("withdraw_enabled"), False),
        withdraw_fee=_as_float(data.get("withdraw_fee")),
        withdraw_min=_as_float(data.get("withdraw_min")),
        needs_memo=_as_bool(data.get("needs_memo"), False),
    )


def _parse_transfer_log_entry(data: Any) -> TransferLogEntry:
    if not isinstance(data, dict):
        return TransferLogEntry()
    return TransferLogEntry(
        timestamp=str(data.get("timestamp") or ""),
        message=str(data.get("message") or ""),
        is_error=_as_bool(data.get("is_error"), False),
    )


def _parse_transfer_state(data: Any) -> TransferState:
    if not isinstance(data, dict):
        return TransferState()

    balance: Optional[float]
    raw_balance = data.get("balance")
    if isinstance(raw_balance, dict):
        balance = _as_float(raw_balance.get("available"))
    else:
        balance = _as_float(raw_balance)

    return TransferState(
        selected_coin=str(data.get("selected_coin") or ""),
        from_exchange=str(data.get("from_exchange") or ""),
        to_exchange=str(data.get("to_exchange") or ""),
        available_networks=[
            _parse_network_info(x) for x in data.get("available_networks", [])
        ],
        selected_network_idx=_as_int(data.get("selected_network_idx")),
        amount=str(data.get("amount") or ""),
        balance=balance,
        deposit_address=str(data.get("deposit_address") or ""),
        deposit_tag=str(data.get("deposit_tag") or ""),
        to_is_personal_wallet=_as_bool(data.get("to_is_personal_wallet"), False),
        personal_wallet_address=str(data.get("personal_wallet_address") or ""),
        personal_wallet_tag=str(data.get("personal_wallet_tag") or ""),
        auto_buy_before_transfer=_as_bool(data.get("auto_buy_before_transfer"), False),
        auto_sell_on_arrival=_as_bool(data.get("auto_sell_on_arrival"), False),
        market_order_pending=_as_bool(data.get("market_order_pending"), False),
        market_order_result=(
            str(data.get("market_order_result"))
            if data.get("market_order_result") is not None
            else None
        ),
        logs=[_parse_transfer_log_entry(x) for x in data.get("logs", [])],
    )


def _parse_scenario_config(data: Any) -> ScenarioConfig:
    if not isinstance(data, dict):
        return ScenarioConfig()
    return ScenarioConfig(
        gap_threshold_percent=_as_float(data.get("gap_threshold_percent")) or 5.0,
        domestic_gap_threshold=_as_float(data.get("domestic_gap_threshold")) or 1.5,
        fut_basis_threshold=_as_float(data.get("fut_basis_threshold")) or 0.5,
    )


def _parse_step_info(data: Any) -> StepInfo:
    if not isinstance(data, dict):
        return StepInfo()
    return StepInfo(
        step=str(data.get("step") or ""),
        title=str(data.get("title") or ""),
        status=str(data.get("status") or "Pending"),
        message=(
            str(data.get("message"))
            if data.get("message") is not None
            else None
        ),
    )


def _parse_transfer_job(data: Any) -> TransferJob:
    if not isinstance(data, dict):
        return TransferJob()
    return TransferJob(
        id=_as_int(data.get("id")) or 0,
        coin=str(data.get("coin") or ""),
        amount=_as_float(data.get("amount")) or 0.0,
        from_exchange=str(data.get("from_exchange") or ""),
        to_exchange=str(data.get("to_exchange") or ""),
        network_display=str(data.get("network_display") or ""),
        current_step=str(data.get("current_step") or "Idle"),
        steps=[_parse_step_info(x) for x in data.get("steps", [])],
        is_executing=_as_bool(data.get("is_executing"), False),
        error_message=(
            str(data.get("error_message"))
            if data.get("error_message") is not None
            else None
        ),
        started_at_secs=_as_int(data.get("started_at_secs")) or 0,
    )

def parse_snapshot(data: dict[str, Any]) -> Snapshot:
    coin_states_raw = data.get("coin_states", {})
    wallet_status_raw = data.get("wallet_status", {})
    korean_names_raw = data.get("korean_names", {})
    orderbook_info_raw = data.get("orderbook_info", {})

    coin_states = (
        {
            str(symbol): _parse_coin_state(str(symbol), coin_data)
            for symbol, coin_data in coin_states_raw.items()
        }
        if isinstance(coin_states_raw, dict)
        else {}
    )
    wallet_status = (
        {
            str(symbol): _parse_coin_wallet_status(ws_data)
            for symbol, ws_data in wallet_status_raw.items()
        }
        if isinstance(wallet_status_raw, dict)
        else {}
    )
    korean_names = (
        {str(k): str(v) for k, v in korean_names_raw.items()}
        if isinstance(korean_names_raw, dict)
        else {}
    )
    orderbook_info = (
        {
            str(symbol): _parse_orderbook_info(ob_data)
            for symbol, ob_data in orderbook_info_raw.items()
        }
        if isinstance(orderbook_info_raw, dict)
        else {}
    )

    return Snapshot(
        coin_states=coin_states,
        wallet_status=wallet_status,
        korean_names=korean_names,
        logs=[_parse_log_entry(x) for x in data.get("logs", [])],
        orderbook_info=orderbook_info,
        scenario_threads=[
            _parse_log_thread(x) for x in data.get("scenario_threads", [])
        ],
        usdt_krw=_as_float(data.get("usdt_krw")),
        usd_krw_forex=_as_float(data.get("usd_krw_forex")),
        last_ticker_age_ms=_as_int(data.get("last_ticker_age_ms")),
        transfer=_parse_transfer_state(data.get("transfer", {})),
        scenario_config=_parse_scenario_config(data.get("scenario_config", {})),
        transfer_jobs=[
            _parse_transfer_job(x) for x in data.get("transfer_jobs", [])
        ],
    )
