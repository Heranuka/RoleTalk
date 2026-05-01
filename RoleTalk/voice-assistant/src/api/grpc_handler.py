import grpc
from src.api.ai.v1 import ai_pb2, ai_pb2_grpc
from src.core.logger import logger

class AIServiceHandler(ai_pb2_grpc.AIServiceServicer):
    """
    AIServiceHandler implements the gRPC servicer defined in the proto file.
    It orchestrates the flow: Audio Bytes -> Text -> LLM Response -> Voice Bytes.
    """
    def __init__(self, orchestrator):
        self.orchestrator = orchestrator

    async def ProcessVoiceTurn(self, request, context):
        """
        Main gRPC method. Note the 'async' keyword – we are using
        Python's gRPC AsyncIO for maximum throughput.
        """
        try:
            logger.info(f"Received gRPC request. Audio size: {len(request.audio_data)} bytes")

            # Вызываем асинхронный процесс оркестратора
            # Порядок возврата: user_text, ai_text, ai_audio_bytes
            user_text, ai_text, ai_audio_bytes = await self.orchestrator.process(
                audio_bytes=request.audio_data,
                lang=request.language,
                system_context=request.system_prompt
            )

            return ai_pb2.ProcessVoiceTurnResponse(
                user_transcription=user_text,
                ai_response_text=ai_text,
                ai_audio_data=ai_audio_bytes
            )

        except Exception as e:
            logger.error(f"Critical error in gRPC handler: {str(e)}", exc_info=True)
            context.set_details(f"Internal AI Error: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return ai_pb2.ProcessVoiceTurnResponse()