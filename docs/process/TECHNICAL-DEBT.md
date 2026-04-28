# Technical Debt Log [PID-SENTINEL]

## [2026-04-28] Dashboard & Discovery (Filtro A)
- **Scalability of Glob**: O uso de `filepath.Glob` para descobrir ADRs no Aggregator é $O(N)$ sobre o sistema de arquivos. Para projetos massivos, isso se tornará um gargalo de I/O.
- **Dashboard Growth**: `COMPLIANCE-DASHBOARD.md` crescerá linearmente. Falta suporte para arquivamento ou paginação.
- **Missing CLI Metadata**: O relatório CLI não exibe o `created_at`, dificultando a análise cronológica.

---
*Assinado: Security Auditor & Senior Architect*
