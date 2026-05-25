---
task_id: "ec002a7"
title: "Observer Pattern for Graph Events"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a7: Observer Pattern for Graph Events

## Contexto

O motor de escaneamento do grafo (`graph.Engine`) precisa notificar consumidores externos sobre eventos do ciclo de vida: início e fim de scan, inserção de nós e criação de arestas. Os consumidores incluem o servidor Live View (que transmite para clientes WebSocket), futuros plugins de notificação (Slack, webhook), e o próprio `EventBuffer` para telemetria. O acoplamento deve ser mínimo — o engine não deve conhecer os consumidores concretos.

## Decisão

Implementamos o padrão **Observer** com interface Go nativa:

```go
type Observer interface {
    Notify(event GraphEvent)
}
```

`graph.Engine` mantém um slice `observers []Observer` protegido por `sync.RWMutex`. O método `RegisterObserver(o Observer)` adiciona observers com write lock. O método `notifyObservers(event GraphEvent)` itera sob read lock e dispara cada observer em uma goroutine separada, com semáforo de backpressure (`observeSem chan struct{}` com capacidade 16) para limitar goroutines simultâneas e evitar sobrecarga.

Tipos de eventos (`EventType`) são constantes tipadas: `SCAN_STARTED`, `NODE_UPSERTED`, `EDGE_CREATED`, `SCAN_COMPLETED`. Cada `GraphEvent` carrega `Type`, `Payload interface{}` (Node ou Edge) e `Time`.

O servidor Live View (`liveview.Server`) implementa `graph.Observer` e injeta eventos recebidos no canal `broadcast` do hub WebSocket, fechando o ciclo sem acoplamento direto entre engine e rede.

## Consequências

- **Positivo**: Desacoplamento total — engine não importa pacotes de rede ou UI. Novos observers são registrados sem modificar o engine.
- **Positivo**: Semáforo de backpressure (16 goroutines simultâneas) previne explosão de goroutines em cenários de alto throughput.
- **Positivo**: `Notify` do Live View usa `select` com `default` — nunca bloqueia o engine, descartando eventos se o canal estiver cheio.
- **Negativo**: Ordem de notificação não é garantida entre observers diferentes. Cada observer deve ser idempotente.

## Referências

- Task ID: [ec002a7]
- Implementação: `internal/graph/events.go:18-29`, `internal/graph/engine.go:40-61`, `internal/liveview/server.go:118-124`
