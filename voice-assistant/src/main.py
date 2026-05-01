import asyncio
import grpc
import signal
import sys
from src.core.stt import STTManager
from src.core.llm import LLMManager
from src.core.tts import TTSManager
from src.services.orchestrator import VoiceOrchestrator
from src.api.ai.v1 import ai_pb2_grpc
from src.api.grpc_handler import AIServiceHandler
from src.core.logger import logger

async def serve():
    """
    Production entry point for the Voice Assistant gRPC server.
    Loads AI models and starts the async gRPC engine.
    """
    logger.info("Initializing AI Engines (Whisper, Ollama, Piper)...")

    # Инициализация менеджеров (Heavy objects)
    stt = STTManager(model_path="/app/models/whisper")
    llm = LLMManager()
    tts = TTSManager(model_path="/app/models/piper/en_US-libritts_r-medium.onnx")

    orchestrator = VoiceOrchestrator(stt, llm, tts)

    # Создаем асинхронный gRPC сервер
    server = grpc.aio.server()

    # Регистрируем обработчик
    ai_pb2_grpc.add_AIServiceServicer_to_server(
        AIServiceHandler(orchestrator), server
    )

    listen_addr = "[::]:50051"
    server.add_insecure_port(listen_addr)
    logger.info(f"gRPC server is listening on {listen_addr}")

    await server.start()

    # Graceful shutdown
    async def shutdown():
        logger.info("Shutting down gRPC server...")
        await server.stop(5)
        logger.info("Server stopped.")

    loop = asyncio.get_event_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, lambda: asyncio.create_task(shutdown()))

    await server.wait_for_termination()

if __name__ == "__main__":
    try:
        asyncio.run(serve())
    except KeyboardInterrupt:
        pass
    except Exception as e:
        logger.error(f"Fatal error: {e}")
        sys.exit(1)