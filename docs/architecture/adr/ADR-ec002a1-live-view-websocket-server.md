---
task_id: "ec002a1"
title: "Live View WebSocket Server Architecture"
date: "2026-05-24"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-ec002a1: Live View WebSocket Server Architecture

## Contexto

O Sentinel necessita de uma interface de visualização em tempo real do grafo de dependências e eventos do motor de escaneamento. A comunicação deve ser bidirecional, de baixa latência e compatível com navegadores web modernos. O servidor atua como hub central que transmite eventos `GraphEvent` para múltiplos clientes conectados simultaneamente.

## Decisão

Implementamos o servidor Live View sobre **WebSocket** via `gorilla/websocket`, operando como um hub com canais de `register`, `unregister` e `broadcast`. Cada cliente recebe um `wsClient` com canal `send` dedicado (buffer 256), garantindo que apenas a goroutine `writePump` escreva na conexão — conformidade com o requisito do Gorilla de single-writer por conexão.

Formato de mensagem: JSON serializado de `graph.GraphEvent` (`{type, payload, timestamp}`), transmitido como `websocket.TextMessage`. O `broadcast` usa canal bufferizado (256) para evitar bloqueio do engine durante picos de eventos; em caso de saturação, eventos são descartados com log `WARN`.

Reconexão: o cliente é responsável pela lógica de retry. O servidor implementa ping/pong keep-alive com `pongWait=60s` e `pingPeriod=54s`. Clientes inativos têm a conexão fechada automaticamente. O `CheckOrigin` do upgrader restringe conexões a `localhost` e `127.0.0.1`, prevenindo CSRF via WebSocket em ambiente de desenvolvimento.

## Consequências

- **Positivo**: Push em tempo real de eventos do grafo, latência mínima, arquitetura hub/spoke testada em produção com Gorilla.
- **Positivo**: Isolamento de escrita por cliente (`writePump` dedicado) evita race conditions na conexão WebSocket.
- **Negativo**: Escalabilidade limitada a algumas dezenas de clientes — o hub usa broadcast síncrono sobre todos os clientes. Para dezenas de milhares, seria necessário sharding ou pub/sub externo.

## Referências

- Task ID: [ec002a1]
- Implementação: `internal/liveview/server.go`
- Gorilla WebSocket: https://github.com/gorilla/websocket
