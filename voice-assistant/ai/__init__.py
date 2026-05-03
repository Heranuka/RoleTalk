import os
import sys

# Help Python find the generated modules in this package
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

# Explicitly expose the classes for IDE type hinting (Optional but helpful)
try:
    from .ai_pb2 import ProcessVoiceTurnRequest, ProcessVoiceTurnResponse
    from .ai_pb2_grpc import AIServiceStub, AIServiceServicer, add_AIServiceServicer_to_server
except ImportError:
    pass