#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$BASE_DIR"

mkdir -p input output tmp

WHISPER_BIN="${WHISPER_BIN:-../whisper.cpp/build/bin/whisper-cli}"
WHISPER_MODEL="${WHISPER_MODEL:-../whisper.cpp/models/ggml-base.en.bin}"
PIPER_BIN="${PIPER_BIN:-../piper1-gpl/build/piper}"
PIPER_MODEL="${PIPER_MODEL:-../piper1-gpl/voices/en_US-lessac-medium.onnx}"
OLLAMA_URL="${OLLAMA_URL:-http://localhost:11434/api/chat}"
OLLAMA_MODEL="${OLLAMA_MODEL:-llama3.1}"
RECORD_SECONDS="${RECORD_SECONDS:-5}"

INPUT_WAV="input/input.wav"
TRANSCRIPT_TXT="output/transcript.txt"
LLM_TEXT="output/llm_response.txt"
OUTPUT_WAV="output/output.wav"
RAW_JSON="tmp/ollama.json"

echo "Recording ${RECORD_SECONDS}s..."
ffmpeg -y -f avfoundation -i ":0" -t "$RECORD_SECONDS" -ac 1 -ar 16000 "$INPUT_WAV" >/dev/null 2>&1

echo "Running Whisper..."
"$WHISPER_BIN" \
  -m "$WHISPER_MODEL" \
  -f "$INPUT_WAV" \
  -l en \
  -nt \
  -of output/transcript >/dev/null 2>&1

if [[ ! -f "$TRANSCRIPT_TXT" ]]; then
  echo "Whisper transcript not found: $TRANSCRIPT_TXT"
  exit 1
fi

USER_TEXT="$(cat "$TRANSCRIPT_TXT" | tr -d '\r')"
echo "You said: $USER_TEXT"

echo "Calling LLM..."
curl -sS "$OLLAMA_URL" \
  -H "Content-Type: application/json" \
  -d "$(printf '%s' "$USER_TEXT" | python3 - <<'PY'
import json, sys
text = sys.stdin.read()
payload = {
    "model": "llama3.1",
    "messages": [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": text}
    ],
    "stream": False
}
print(json.dumps(payload))
PY
)" > "$RAW_JSON"

ASSISTANT_TEXT="$(python3 - <<'PY'
import json
from pathlib import Path
data = json.loads(Path("tmp/ollama.json").read_text(encoding="utf-8"))
print(data["message"]["content"].strip())
PY
)"

printf '%s\n' "$ASSISTANT_TEXT" > "$LLM_TEXT"
echo "Assistant: $ASSISTANT_TEXT"

echo "Running Piper..."
printf '%s' "$ASSISTANT_TEXT" | "$PIPER_BIN" \
  --model "$PIPER_MODEL" \
  --output_file "$OUTPUT_WAV" >/dev/null 2>&1

echo "Playing..."
afplay "$OUTPUT_WAV"
