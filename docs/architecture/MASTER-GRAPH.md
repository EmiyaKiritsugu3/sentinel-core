# Project Master Architecture [PID-SENTINEL]

> [!IMPORTANT]
> This is an auto-generated live map of the codebase.

```mermaid
graph TD
    file_cmd_sentinel_main_go["main.go (file)"]
    func_cmd_sentinel_main_go_main["main (function)"]:::func
    file_internal_graph_scanner_go_go["scanner_go.go (file)"]
    struct_internal_graph_scanner_go_go_GoScanner["GoScanner (struct)"]:::struct
    func_internal_graph_scanner_go_go_NewGoScanner["NewGoScanner (function)"]:::func
    func_internal_graph_scanner_go_go_ScanProject["ScanProject (function)"]:::func
    func_internal_graph_scanner_go_go_scanFile["scanFile (function)"]:::func
    func_internal_graph_scanner_go_go_upsertNode["upsertNode (function)"]:::func
    func_internal_graph_scanner_go_go_createEdge["createEdge (function)"]:::func
    func_internal_graph_scanner_go_go_isIgnored["isIgnored (function)"]:::func
    func_internal_graph_scanner_go_go_calculateHash["calculateHash (function)"]:::func
    file_internal_graph_schema_go["schema.go (file)"]
    func_internal_graph_schema_go_Migrate["Migrate (function)"]:::func
    file_pkg_sqlite_db_go["db.go (file)"]
    struct_pkg_sqlite_db_go_DB["DB (struct)"]:::struct
    func_pkg_sqlite_db_go_Init["Init (function)"]:::func
    func_pkg_sqlite_db_go_Close["Close (function)"]:::func
    file_internal_graph_visualizer_go["visualizer.go (file)"]
    struct_internal_graph_visualizer_go_Visualizer["Visualizer (struct)"]:::struct
    func_internal_graph_visualizer_go_NewVisualizer["NewVisualizer (function)"]:::func
    func_internal_graph_visualizer_go_GenerateMasterDiagram["GenerateMasterDiagram (function)"]:::func
    func_internal_graph_visualizer_go_GenerateTaskSnapshot["GenerateTaskSnapshot (function)"]:::func
    struct_internal_graph_visualizer_go_Node["Node (struct)"]:::struct
    struct_internal_graph_visualizer_go_Edge["Edge (struct)"]:::struct
    func_internal_graph_visualizer_go_getNodes["getNodes (function)"]:::func
    func_internal_graph_visualizer_go_getEdges["getEdges (function)"]:::func
    func_internal_graph_visualizer_go_formatMermaid["formatMermaid (function)"]:::func
    file_cmd_sentinel_main_go -->|contains| func_cmd_sentinel_main_go_main
    file_internal_graph_scanner_go_go -->|contains| struct_internal_graph_scanner_go_go_GoScanner
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_NewGoScanner
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_ScanProject
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_scanFile
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_upsertNode
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_createEdge
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_isIgnored
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_calculateHash
    file_internal_graph_schema_go -->|contains| func_internal_graph_schema_go_Migrate
    file_pkg_sqlite_db_go -->|contains| struct_pkg_sqlite_db_go_DB
    file_pkg_sqlite_db_go -->|contains| func_pkg_sqlite_db_go_Init
    file_pkg_sqlite_db_go -->|contains| func_pkg_sqlite_db_go_Close
    file_internal_graph_visualizer_go -->|contains| struct_internal_graph_visualizer_go_Visualizer
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_NewVisualizer
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_GenerateMasterDiagram
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_GenerateTaskSnapshot
    file_internal_graph_visualizer_go -->|contains| struct_internal_graph_visualizer_go_Node
    file_internal_graph_visualizer_go -->|contains| struct_internal_graph_visualizer_go_Edge
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_getNodes
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_getEdges
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_formatMermaid

    classDef struct fill:#f9f,stroke:#333,stroke-width:2px;
    classDef func fill:#bbf,stroke:#333,stroke-width:1px;
```
