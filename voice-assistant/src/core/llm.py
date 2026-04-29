import requests
import os

class LLMManager:
    def __init__(self):
        self.url = os.getenv("OLLAMA_URL", "http://host.docker.internal:11434")
        self.model = "qwen2.5:latest"

    def ask(self, text: str, lang: str, rid: str):
        payload = {
            "model": self.model,
            "messages": [
                {"role": "system", "content": f"You are a roleplay partner. Speak only in {lang}. Short natural responses."},
                {"role": "user", "content": text}
            ],
            "stream": False
        }
        r = requests.post(f"{self.url}/api/chat", json=payload, timeout=30)
        r.raise_for_status()
        return r.json()["message"]["content"].strip()