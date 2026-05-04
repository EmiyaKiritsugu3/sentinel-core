---
task_id: "fe2bb6f9"
title: "Decisão Crítica --- com : caracteres @ perigosos \"aspas\""
date: "2026-04-28"
status: "PROPOSED"
author: "Sentinel Auto-ADR"
---

# ADR-fe2bb6f9: Decisão Crítica --- com : caracteres @ perigosos "aspas"

## Contexto

Esta decisão foi capturada proativamente pelo Sentinel via comando 'instruct'.
Intenção original: Decisão Crítica --- com : caracteres @ perigosos "aspas"

## Decisão

Estabeleceremos um protocolo rigoroso de sanitização e escaping para títulos e descrições que contenham caracteres especiais (`:`, `@`, `"`, `-`). O gerador automático de ADRs e o comando `instruct` devem aplicar neutralização de strings para evitar quebra de YAML frontmatter e garantir que caracteres perigosos não sejam interpretados erroneamente como comandos shell ou delimitadores de sistema durante a automação.

## Consequências

- **Positivo**: Robustez do sistema contra injeção de comandos e erros de parsing em metadados, permitindo flexibilidade total na nomenclatura de tarefas.
- **Negativo**: Adiciona uma pequena sobrecarga de processamento de strings em todas as operações de entrada do usuário.

## Referências

- Task ID: [fe2bb6f9]
