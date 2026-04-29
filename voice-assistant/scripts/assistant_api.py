import os
import shutil
import subprocess
import requests
import logging
import time
import uuid
from pathlib import Path
from typing import Optional

from fastapi import FastAPI, UploadFile, File, HTTPException, Header, BackgroundTasks, Request
from fastapi.responses import FileResponse
from faster_whisper import WhisperModel

# --- CONFIGURATION & LOGGING ---
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - [%(request_id)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
# Force logger to include request_id even if empty
old_factory = logging.getLogRecordFactory()
def record_factory(*args, **kwargs):
    record = old_factory(*args, **kwargs)
    if not hasattr(record, 'request_id'):
        record.request_id = 'GLOBAL'
    return record
logging.setLogRecordFactory(record_factory)

logger = logging.getLogger("RoleTalk-AI")

app = FastAPI(title="RoleTalk AI Service")

BASE_DIR = Path("/app")
UPLOAD_DIR = BASE_DIR / "input"
RESULT_DIR = BASE_DIR / "output"
UPLOAD_DIR.mkdir(exist_ok=True)
RESULT_DIR.mkdir(exist_ok=True)

PIPER_MODEL = "/app/models/piper/en_US-libritts_r-medium.onnx"
OLLAMA_URL = os.getenv("OLLAMA_URL", "http://host.docker.internal:11434")

# --- MODEL INITIALIZATION ---
logger.info("Loading Faster-Whisper model (base)...")
try:
    stt_model = WhisperModel(
        "base",
        device="cpu",
        compute_type="int8",
        download_root="/app/models/whisper"
    )
    logger.info("Whisper model loaded successfully.")
except Exception as e:
    logger.error(f"Failed to load Whisper model: {e}")
    os._exit(1)

# --- HELPERS ---

def cleanup_files(*files: str):
    """Background task to remove temporary files after response is sent."""
    for f in files:
        try:
            if os.path.exists(f):
                os.remove(f)
        except Exception as e:
            logger.error(f"Cleanup error for {f}: {e}")

def ask_ollama(user_text: str, lang: str, rid: str):
    """Communicates with Ollama LLM."""
    logger.info(f"Querying Ollama (Lang: {lang})")
    try:
        start_t = time.time()
        payload = {
            "model": "qwen2.5:latest",
            "messages": [
                {
                    "role": "system",
                    "content": f"You are a roleplay partner. Speak only in {lang}. Keep responses natural and very short (1-2 sentences)."
                },
                {"role": "user", "content": user_text}
            ],
            "stream": False
        }
        r = requests.post(f"{OLLAMA_URL}/api/chat", json=payload, timeout=30)
        r.raise_for_status()
        text = r.json()["message"]["content"].strip()
        logger.info(f"Ollama responded in {time.time() - start_t:.2f}s")
        return text
    except Exception as e:
        logger.error(f"Ollama error: {e}", extra={"request_id": rid})
        return "I'm sorry, I'm having trouble thinking. Can you repeat?"

# --- API ENDPOINTS ---

@app.post("/voice")
async def process_voice(
        background_tasks: BackgroundTasks,
        request: Request,
        file: UploadFile = File(...),
        x_request_id: Optional[str] = Header(None),
        x_practice_language: Optional[str] = Header("English")
):
    # Tracing & Identification
    request_id = x_request_id or str(uuid.uuid4())
    # Add request_id to logger context via extra
    log_extra = {"request_id": request_id}

    request_start = time.time()
    logger.info(f"New request: {file.filename} | Target Lang: {x_practice_language}", extra=log_extra)

    # Unique file paths for this specific request
    temp_m4a = f"/tmp/{request_id}.m4a"
    input_wav = UPLOAD_DIR / f"{request_id}.wav"
    output_wav = RESULT_DIR / f"{request_id}_res.wav"

    try:
        # 1. Save inbound stream
        with open(temp_m4a, "wb") as buffer:
            shutil.copyfileobj(file.file, buffer)

        # 2. Transcode to WAV for Whisper
        ffmpeg_start = time.time()
        subprocess.run([
            "ffmpeg", "-y", "-i", temp_m4a,
            "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", str(input_wav)
        ], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        logger.info(f"FFMPEG conversion done in {time.time() - ffmpeg_start:.2f}s", extra=log_extra)

        # 3. Speech-to-Text
        stt_start = time.time()
        # Map human language name to Whisper ISO code (simple logic)
        whisper_lang = "en" if "en" in x_practice_language.lower() else None

        segments, _ = stt_model.transcribe(str(input_wav), language=whisper_lang, beam_size=5)
        user_text = " ".join([s.text for s in segments]).strip()

        if not user_text:
            logger.warning("No speech detected", extra=log_extra)
            user_text = "..."
        logger.info(f"STT: '{user_text}' in {time.time() - stt_start:.2f}s", extra=log_extra)

        # 4. LLM Response
        ai_text = ask_ollama(user_text, x_practice_language, request_id)

        # 5. Text-to-Speech (Piper)
        if not os.path.exists(PIPER_MODEL):
            raise FileNotFoundError(f"Piper model missing: {PIPER_MODEL}")

        tts_start = time.time()
        subprocess.run(
            ["piper", "--model", PIPER_MODEL, "--output_file", str(output_wav)],
            input=ai_text.encode('utf-8'),
            check=True,
            capture_output=True
        )
        logger.info(f"TTS synthesis done in {time.time() - tts_start:.2f}s", extra=log_extra)

        # Total processing stats
        total_duration = time.time() - request_start
        logger.info(f"Request {request_id} completed in {total_duration:.2f}s", extra=log_extra)

        # Prepare response with metadata in headers
        response = FileResponse(
            output_wav,
            media_type="audio/wav",
            filename="response.wav",
            headers={
                "X-STT-Transcription": user_text.encode('utf-8').decode('latin-1'),
                "X-LLM-Response": ai_text.encode('utf-8').decode('latin-1'),
                "X-Request-ID": request_id
            }
        )

        # 6. Background cleanup
        background_tasks.add_task(cleanup_files, temp_m4a, str(input_wav), str(output_wav))

        return response

    except Exception as e:
        logger.error(f"Request failed: {e}", exc_info=True, extra=log_extra)
        # Cleanup on failure
        background_tasks.add_task(cleanup_files, temp_m4a, str(input_wav), str(output_wav))
        raise HTTPException(status_code=500, detail="Internal AI Processing Error")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)