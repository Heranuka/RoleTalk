from pathlib import Path


class VoiceOrchestrator:
    def __init__(self, stt, llm, tts):
        self.stt = stt
        self.llm = llm
        self.tts = tts
        self.upload_dir = Path("/app/input")
        self.result_dir = Path("/app/output")

    async def execute_turn(self, temp_m4a: str, lang: str, rid: str, system_context: str):
        # 1. Paths
        input_wav = str(self.upload_dir / f"{rid}.wav")
        output_wav = str(self.result_dir / f"{rid}_res.wav")

        # 2. STT Process
        self.stt.convert_to_wav(temp_m4a, input_wav)

        whisper_lang = "en" if "en" in lang.lower() else None
        user_text = self.stt.transcribe(input_wav, whisper_lang) or "..."

        # 3. LLM Process
        ai_text = self.llm.ask(user_text, lang, rid, system_context)

        # 4. TTS Process
        self.tts.synthesize(ai_text, output_wav)

        return user_text, ai_text, output_wav, input_wav