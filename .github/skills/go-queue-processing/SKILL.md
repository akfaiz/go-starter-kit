---
name: go-queue-processing
description: "Asynq queue client, worker modules, payload DTOs, task registration, retries, and trace propagation. Use when enqueueing jobs, adding task handlers, or changing background processing."
---

# Queue Processing

This project uses `github.com/hibiken/asynq` for background jobs with Redis-backed queues.

## Core Pieces
- `internal/delivery/queue/client.go`: enqueue tasks from application code.
- `internal/delivery/queue/task.go`: create tasks with injected trace headers.
- `internal/delivery/queue/server/server.go`: run the worker server.
- `internal/delivery/queue/module.go`: FX modules for queue client and workers.
- `internal/delivery/queue/middleware/`: worker logging and tracing middleware.
- `internal/delivery/queue/handler/`: task handlers.
- `internal/delivery/queue/handler/payload/`: JSON payloads shared between producers and workers.

## Adding a New Task
1. Define a payload struct under `internal/delivery/queue/handler/payload/`.
2. Add a task type constant in `internal/delivery/queue/task.go`.
3. Implement a handler in `internal/delivery/queue/handler/`.
4. Register the handler in `internal/delivery/queue/module.go`.
5. Enqueue the task with `queue.NewTask(ctx, taskType, payload)` and `queue.Client.EnqueueContext`.
6. Add handler tests for decode failure, successful processing, and retryable dependency failure.

## Existing Pattern
The mail flow is the current reference:
- `internal/service/auth/auth_service.go` builds a `payload.MailPayload`.
- `queue.NewTask` injects trace headers before enqueueing.
- `internal/delivery/queue/handler/mail_handler.go` unmarshals the payload and calls `domain.Mailer`.
- `internal/delivery/queue/handler/payload/payload.go` maps JSON into `domain.Mail`.

## Handler Rules
- Treat malformed JSON as non-retryable. Return `asynq.SkipRetry` for decode failures.
- Return real errors for transient failures so the worker can retry them.
- Keep handlers thin: decode payload, log minimal task context, call the domain interface.
- Do not import HTTP handlers or HTTP DTOs from queue handlers.

## Tracing and Logging
- The client and worker both use OTel spans.
- `internal/delivery/queue/middleware/otel.go` extracts propagated headers and starts a task span.
- `internal/delivery/queue/middleware/logger.go` logs execution timing and errors.

## Practical Rules
- Prefer explicit payload structs over generic `map[string]any`.
- Keep task names stable because they are routing keys in the worker mux.
- Preserve request context when enqueueing so trace headers are propagated.
- Use `ClientModule` for producers and `WorkerModule` for worker processes.
- Keep payload JSON backward-compatible if queued tasks may already exist in Redis.
