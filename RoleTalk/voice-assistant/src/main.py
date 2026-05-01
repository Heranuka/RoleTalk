import uvicorn
from fastapi import FastAPI
from src.core.stt import STTManager
from src.core.llm import LLMManager
from src.core.tts import TTSManager
from src.services.orchestrator import VoiceOrchestrator
from src.api import routes

app = FastAPI(title="RoleTalk AI Service")

# Инициализация (Dependency Injection)
stt = STTManager(model_path="/app/models/whisper")
llm = LLMManager()
tts = TTSManager(model_path="/app/models/piper/en_US-libritts_r-medium.onnx")

# Собираем оркестратор
orchestrator_service = VoiceOrchestrator(stt, llm, tts)

# Внедряем сервис в роутер
routes.orchestrator = orchestrator_service
app.include_router(routes.router)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8080)