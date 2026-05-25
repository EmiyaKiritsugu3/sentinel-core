# internal/bridge

AI client abstraction layer connecting the agent engine to external providers (Gemini SDK).

## Overview

The bridge package provides interfaces that decouple the agent engine from the concrete Gemini SDK. It includes an SDK client wrapper, intent classification with AI fallback, context routing strategies, and a prompt factory that assembles surgical context payloads.

## Key Types

### `GenaiClient` (interface)
Abstracts `*genai.Client`. Methods: `GenerativeModel(model)`, `Close()`. Implemented by `sdkClient` wrapping the real SDK. Enables mocking in tests.

### `GenaiModel` (interface)
Abstracts `*genai.GenerativeModel`. Methods: `SetTemperature`, `SetSystemInstruction`, `SetTools`, `StartChat`, `GenerateContent`.

### `MessageSender` (interface)
Abstracts `*genai.ChatSession`. Single method: `SendMessage(ctx, parts...)`.

### `NewSDKClient(client *genai.Client) (GenaiClient, error)`
Wraps a concrete Gemini client. Returns `ErrNilClient` if nil.

## Intent Classification

### `IntentClassifier`
Tiered classification: heuristic keyword matching first, AI fallback when confidence is below threshold (0.60). Results cached per taskID via `sync.Map`.

Intents: `diagnose`, `implement`, `refactor`, `review`, `unknown`. Keywords support English and Portuguese.

### `AIClassifier` (interface)
AI-powered fallback. Implemented by `GeminiClassifier` and `NilClassifier` (test null object).

## Context Routing

### `ContextStrategy`
Defines what context to inject per intent type:
- **Diagnose**: high-coupling nodes, recent changes, limit 15
- **Implement**: test files, ADRs, limit 10
- **Refactor**: high-coupling nodes, technical debt markers, limit 12
- **Review**: ADRs, limit 8
- **Unknown**: default behavior (no routing)

Strategy resolved by `StrategyFor(intent)`.

## Prompt Factory

### `Factory`
Assembles `ContextPayload` with system instruction (persona + ADRs + standards + rules of engagement), surgical context (code nodes with snippets), and task description.
- `GeneratePayload(ctx, taskID, personaPrompt)` — full payload assembly
- `loadContextByStrategy(ctx, taskID, strategy)` — strategy-driven node selection from DB

## Dependencies

- `internal/state` — task manager
- `pkg/sqlite` — DB for context queries
- `github.com/google/generative-ai-go/genai` — Gemini SDK types

## Usage

```go
sdk, _ := bridge.NewSDKClient(genaiClient)
classifier, _ := bridge.NewGeminiClassifier(sdk)
intentCls := bridge.NewIntentClassifier(classifier, 0.60)
factory, _ := bridge.NewFactory(db, intentCls)

payload, _ := factory.GeneratePayload(ctx, taskID, persona)
// payload.SystemInstruction, payload.SurgicalContext, payload.TaskDescription
```
