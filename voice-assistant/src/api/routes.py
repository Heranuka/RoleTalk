import shutil
import os
from typing import Optional
from fastapi import APIRouter, UploadFile, File, Header, BackgroundTasks, HTTPException
from fastapi.responses import FileResponse
import uuid

router = APIRouter()
orchestrator = None # Будет внедрен в main.py

def cleanup(files: list):
    for f in files:
        if os.path.exists(f):
            os.remove(f)

@router.post("/voice")
async def process_voice(
        background_tasks: BackgroundTasks,
        file: UploadFile = File(...),
        x_request_id: Optional[str] = Header(None),
        x_practice_language: str = Header("English")
):
    rid = x_request_id or str(uuid.uuid4())
    temp_m4a = f"/tmp/{rid}.m4a"

    # Save upload
    with open(temp_m4a, "wb") as buffer:
        shutil.copyfileobj(file.file, buffer)

    try:
        user_text, ai_text, out_path, in_wav = await orchestrator.execute_turn(temp_m4a, x_practice_language, rid)

        # Регистрация удаления временных файлов
        background_tasks.add_task(cleanup, [temp_m4a, in_wav, out_path])

        return FileResponse(
            out_path,
            media_type="audio/wav",
            headers={
                "X-STT-Transcription": user_text.encode('utf-8').decode('latin-1'),
                "X-LLM-Response": ai_text.encode('utf-8').decode('latin-1'),
                "X-Request-ID": rid
            }
        )
    except Exception as e:
        background_tasks.add_task(cleanup, [temp_m4a])
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/health")
async def health():
    return {"status": "ok"}