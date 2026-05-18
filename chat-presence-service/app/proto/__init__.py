import sys
import os

# grpc_tools generates `import presence_pb2` (bare) in presence_pb2_grpc.py.
# Add this directory to sys.path so that bare import resolves correctly.
sys.path.insert(0, os.path.dirname(__file__))
