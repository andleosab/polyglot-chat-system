FROM python:3.9-slim AS build

WORKDIR /app

RUN pip install uv

COPY pyproject.toml uv.lock* ./
RUN uv sync --frozen --no-dev

COPY . .

# Generate gRPC stubs after source copy so they are never overwritten
RUN mkdir -p app/proto && uv run python -m grpc_tools.protoc \
      -I proto \
      --python_out=app/proto \
      --grpc_python_out=app/proto \
      proto/presence.proto


FROM python:3.9-slim

WORKDIR /app

COPY --from=build /app /app

EXPOSE 8000 50051

CMD ["/app/.venv/bin/uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
