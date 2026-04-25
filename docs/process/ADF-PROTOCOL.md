# ADF Protocol — Antigravity Debugging Framework (v1.0.0)

Este protocolo é obrigatório para o diagnóstico de erros monumentais de infraestrutura ou build que persistam por mais de 3 tentativas de correção manual.

## 1. Deliberação e Auditoria (Passo Zero)
Antes de qualquer ação ou diagnóstico, o agente DEVE:
- **Tool Audit**: Listar todas as ferramentas disponíveis e selecionar as mais adequadas (ex: `context7` para documentação, `graphify` para arquitetura).
- **Sequential Thinking**: Iniciar uma sessão de pensamento sequencial para mapear hipóteses, listar incertezas e planejar a árvore de decisão.

## 2. Arqueologia Obrigatória (History-First)
Com base na deliberação inicial, correlacionar o sintoma com o histórico recente.
- **Ação**: `git log -n 5 --name-only` e leitura de `.nanostack/journal/`.
- **Objetivo**: Identificar mudanças de "Paradigma de Versão" ou "Estrutura de Raiz".

## 3. Hierarquia de Sanidade (The Stack)
A depuração deve seguir esta ordem rigorosa de baixo para cima, validando cada degrau com pensamento crítico:
1.  **INFRA**: A raiz segue o padrão da versão atual do framework?
2.  **ESTÁTICA**: O `tsc --noEmit` está verde? (Não depure renderização com TSC quebrado).
3.  **ENGINE**: O build nativo completa com um layout minimalista?
4.  **UI**: O erro ocorre no componente específico ou no seu contexto (Providers)?

## 3. Reducionismo de Isolação
Se o build falhar na fase de renderização, o agente DEVE:
- Criar um `layout.tsx` minimalista: `<html><body>{children}</body></html>`.
- Se o erro persistir: O problema é **CONFIG/DEPENDÊNCIA**.
- Se o erro sumir: O problema é **LÓGICA/UI**.

## 4. Matriz de Rastreabilidade (Traceability Chain)
Para evitar regressões e perda de propósito, cada alteração deve seguir a cadeia de custódia:
1.  **REFERÊNCIA**: Todo Plano de Implementação DEVE iniciar citando a Spec de origem.
2.  **IDENTIDADE**: Commits e arquivos de plano DEVEM conter a tag `[PID-SENTINEL]`.
3.  **SÍNTESE**: Após a conclusão, o aprendizado (positivo ou negativo) DEVE ser registrado no `sentinel-log.md`.
4.  **INDEXAÇÃO**: Novos documentos críticos devem ser linkados no `wiki-index.md`.

## 5. Protocolo de Epifania (The Learning Loop)
No encerramento de tarefas complexas ou resoluções de bugs críticos, o agente DEVE processar o aprendizado:
1.  **Filtro A (Contextual/Temporário)**: É um bug de biblioteca? -> Registrar em `TECHNICAL-DEBT.md`.
2.  **Filtro B (Projeto/Arquitetura)**: É uma nova regra de projeto? -> Registrar em `sentinel-log.md`.
3.  **Filtro C (Comportamento/Global)**: É uma falha no método de raciocínio da IA? -> Invocar `save_memory` e adicionar a `GEMINI.md`.

## 6. Governança de Raiz (Next.js 15 Special)
- Proibido arquivos de instrumentação customizados (ex: `instrumentation-client.ts`).
- Sentry deve usar obrigatoriamente os 3 arquivos de config + `withSentryConfig`.
- Tags `<html>` e `<body>` são reservadas para arquivos de layout raiz ou erro global.

---
*Assinado: Antigravity Sentinel Core*
