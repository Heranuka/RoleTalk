import os
import requests
import sounddevice as sd
import numpy as np
from pynput import keyboard
from scipy.io.wavfile import write
import tempfile

# Настройки
API_URL = "http://localhost:8080/voice"
FS = 16000  # Частота дискретизации (Whisper требует 16к)
CHANNELS = 1

class VoiceClient:
    def __init__(self):
        self.recording = False
        self.audio_data = []

    def start_recording(self):
        self.audio_data = []
        self.recording = True
        print("\n🎤 Запись пошла... (говорите)")

        # Начинаем захват звука
        def callback(indata, frames, time, status):
            if self.recording:
                self.audio_data.append(indata.copy())

        self.stream = sd.InputStream(samplerate=FS, channels=CHANNELS, callback=callback)
        self.stream.start()

    def stop_and_send(self):
        self.recording = False
        self.stream.stop()
        self.stream.close()
        print("🛑 Запись остановлена. Отправка на сервер...")

        # Сохраняем во временный файл
        if not self.audio_data:
            print("Пустая запись")
            return

        full_audio = np.concatenate(self.audio_data, axis=0)

        with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp_file:
            write(tmp_file.name, FS, (full_audio * 32767).astype(np.int16))

            # Отправляем на сервер
            try:
                with open(tmp_file.name, 'rb') as f:
                    files = {'file': ('audio.wav', f, 'audio/wav')}
                    response = requests.post(API_URL, files=files)

                if response.status_code == 200:
                    # Сохраняем ответный голос ИИ и проигрываем его
                    with open("ai_response.wav", "wb") as res_file:
                        res_file.write(response.content)
                    print("✅ Получен ответ. Проигрываю...")
                    # Проигрывание (на Маке используется afplay)
                    os.system("afplay ai_response.wav")
                else:
                    print(f"Ошибка сервера: {response.text}")
            except Exception as e:
                print(f"Ошибка соединения: {e}")
            finally:
                os.unlink(tmp_file.name)

# Логика клавиш
client = VoiceClient()

def on_press(key):
    # Используем клавишу Пробел (или любую другую)
    if key == keyboard.Key.space and not client.recording:
        client.start_recording()

def on_release(key):
    if key == keyboard.Key.space:
        client.stop_and_send()
        print("\nНажми и держи ПРОБЕЛ, чтобы говорить (ESC для выхода)")

print("Нажми и держи ПРОБЕЛ, чтобы говорить (ESC для выхода)")

with keyboard.Listener(on_press=on_press, on_release=on_release) as listener:
    listener.join()