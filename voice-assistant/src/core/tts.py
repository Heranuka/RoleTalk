import subprocess
import os

class TTSManager:
    def __init__(self, model_path: str):
        self.model_path = model_path

    def synthesize(self, text: str, output_path: str):
        if not os.path.exists(self.model_path):
            raise FileNotFoundError(f"Piper model missing at {self.model_path}")

        subprocess.run(
            ["piper", "--model", self.model_path, "--output_file", output_path],
            input=text.encode('utf-8'),
            check=True,
            capture_output=True
        )