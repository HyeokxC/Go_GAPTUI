from __future__ import annotations

import asyncio
import json
from typing import Any, Awaitable, Callable, Optional

from websockets.exceptions import ConnectionClosed

try:
    from websockets.asyncio.client import connect
except ImportError:
    from websockets import connect  # type: ignore


SnapshotCallback = Callable[[dict[str, Any]], None]
ConnectionCallback = Callable[[], None]


class IpcClient:
    def __init__(self, url: str = "ws://127.0.0.1:9876") -> None:
        self.url = url
        self.ws: Any = None
        self.on_snapshot: Optional[SnapshotCallback] = None
        self.on_connect: Optional[ConnectionCallback] = None
        self.on_disconnect: Optional[ConnectionCallback] = None
        self._running = True
        self._write_lock = asyncio.Lock()

    async def connect(self) -> None:
        retry_delay = 2.0
        while self._running:
            try:
                async with connect(
                    self.url,
                    ping_interval=20,
                    ping_timeout=20,
                    close_timeout=2,
                    max_queue=128,
                ) as ws:
                    self.ws = ws
                    if self.on_connect is not None:
                        self.on_connect()

                    async for message in ws:
                        if not self._running:
                            return
                        try:
                            data = json.loads(message)
                        except json.JSONDecodeError:
                            continue
                        if isinstance(data, dict) and self.on_snapshot is not None:
                            self.on_snapshot(data)
            except (ConnectionClosed, ConnectionRefusedError, OSError):
                pass
            finally:
                self.ws = None
                if self.on_disconnect is not None:
                    self.on_disconnect()

            if self._running:
                await asyncio.sleep(retry_delay)

    async def send_command(self, cmd: str, **params: Any) -> None:
        if self.ws is None:
            return
        payload = json.dumps({"type": cmd, "params": params}, ensure_ascii=False)
        async with self._write_lock:
            try:
                await self.ws.send(payload)
            except ConnectionClosed:
                self.ws = None

    def stop(self) -> None:
        self._running = False

    async def send_shutdown(self) -> None:
        """Send shutdown command to Go backend."""
        await self.send_command("shutdown")
