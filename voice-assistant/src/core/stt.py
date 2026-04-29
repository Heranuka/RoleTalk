import subprocess
from faster_whisper import WhisperModel
from src.core.logger import logger

class STTManager:
    def __init__(self, model_path: str):
        logger.info(f"Loading Whisper model from {model_path}...")
        self.model = WhisperModel(
            "base", device="cpu", compute_type="int8", download_root=model_path
        )

    def convert_to_wav(self, input_path: str, output_path: str):
        subprocess.run([
            "ffmpeg", "-y", "-i", input_path,
            "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", output_path
        ], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    def transcribe(self, audio_path: str, lang_code: str):
        segments, _ = self.model.transcribe(audio_path, language=lang_code, beam_size=5)
        return " ".join([s.text for s in segments]).strip()