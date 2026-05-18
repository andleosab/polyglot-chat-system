import asyncio
import logging
from concurrent import futures

import grpc

from app import presence as presence_svc
from app.deps import get_redis, get_pool

log = logging.getLogger(__name__)

# Generated stubs are placed in app/proto/ during Docker build.
# Import lazily so the module is importable even before generation.
try:
    from app.proto import presence_pb2, presence_pb2_grpc  # type: ignore
    _stubs_available = True
except ImportError:
    _stubs_available = False
    log.warning("gRPC stubs not found — gRPC server will not start")


class _PresenceServicer:
    async def Connected(self, request, context):
        await presence_svc.set_online(request.user_uuid, get_redis())
        log.debug("connected: %s", request.user_uuid)
        return presence_pb2.Ack(ok=True)

    async def Disconnected(self, request, context):
        await presence_svc.set_offline(request.user_uuid, get_redis(), get_pool())
        log.debug("disconnected: %s", request.user_uuid)
        return presence_pb2.Ack(ok=True)

    async def Heartbeat(self, request, context):
        await presence_svc.refresh_ttl(request.user_uuid, get_redis())
        return presence_pb2.Ack(ok=True)


async def serve(port: int = 50051) -> None:
    if not _stubs_available:
        log.error("gRPC stubs unavailable — skipping gRPC server startup")
        return

    server = grpc.aio.server(
        futures.ThreadPoolExecutor(max_workers=4),
        options=[("grpc.max_connection_idle_ms", 30_000)],
    )
    presence_pb2_grpc.add_PresenceServiceServicer_to_server(_PresenceServicer(), server)
    server.add_insecure_port(f"0.0.0.0:{port}")
    await server.start()
    log.info("gRPC server listening on :%d", port)
    await server.wait_for_termination()
