# Project Master Architecture [PID-SENTINEL]

> [!IMPORTANT]
> This is an auto-generated live map of the codebase.

```mermaid
graph TD
    file_cmd_sentinel_commands_scan_go["scan.go (file)"]
    import_cmd_sentinel_commands_scan_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_scan_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_scan_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_scan_go_NewScanCmd["NewScanCmd (function)"]:::func
    file_cmd_sentinel_commands_visualize_go["visualize.go (file)"]
    import_cmd_sentinel_commands_visualize_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_visualize_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_visualize_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_visualize_go_NewVisualizeCmd["NewVisualizeCmd (function)"]:::func
    file_cmd_sentinel_commands_root_go["root.go (file)"]
    import_cmd_sentinel_commands_root_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_root_go_os["os (unresolved_import)"]
    import_cmd_sentinel_commands_root_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_root_go_NewRootCmd["NewRootCmd (function)"]:::func
    func_cmd_sentinel_commands_root_go_Execute["Execute (function)"]:::func
    file_cmd_sentinel_commands_start_go["start.go (file)"]
    import_cmd_sentinel_commands_start_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_start_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_start_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_start_go_NewStartCmd["NewStartCmd (function)"]:::func
    file_cmd_sentinel_commands_plan_go["plan.go (file)"]
    import_cmd_sentinel_commands_plan_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_plan_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_plan_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_plan_go_NewPlanCmd["NewPlanCmd (function)"]:::func
    file_cmd_sentinel_commands_status_go["status.go (file)"]
    import_cmd_sentinel_commands_status_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_status_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_status_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_status_go_NewStatusCmd["NewStatusCmd (function)"]:::func
    file_cmd_sentinel_commands_report_go["report.go (file)"]
    import_cmd_sentinel_commands_report_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_report_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_report_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_report_go_NewReportCmd["NewReportCmd (function)"]:::func
    file_cmd_sentinel_commands_audit_go["audit.go (file)"]
    import_cmd_sentinel_commands_audit_go_errors["errors (unresolved_import)"]
    import_cmd_sentinel_commands_audit_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_audit_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_audit_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_audit_go_NewAuditCmd["NewAuditCmd (function)"]:::func
    file_cmd_sentinel_commands_instruct_go["instruct.go (file)"]
    import_cmd_sentinel_commands_instruct_go_bufio["bufio (unresolved_import)"]
    import_cmd_sentinel_commands_instruct_go_context["context (unresolved_import)"]
    import_cmd_sentinel_commands_instruct_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_commands_instruct_go_os["os (unresolved_import)"]
    import_cmd_sentinel_commands_instruct_go_strings["strings (unresolved_import)"]
    import_cmd_sentinel_commands_instruct_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_cmd_sentinel_commands_instruct_go_init["init (function)"]:::func
    func_cmd_sentinel_commands_instruct_go_NewInstructCmd["NewInstructCmd (function)"]:::func
    func_cmd_sentinel_commands_instruct_go_performDiagnostic["performDiagnostic (function)"]:::func
    func_cmd_sentinel_commands_instruct_go_runSocraticInterview["runSocraticInterview (function)"]:::func
    file_cmd_sentinel_main_go["main.go (file)"]
    import_cmd_sentinel_main_go_fmt["fmt (unresolved_import)"]
    import_cmd_sentinel_main_go_os["os (unresolved_import)"]
    func_cmd_sentinel_main_go_main["main (function)"]:::func
    file_internal_agents_auth_provider_go["auth_provider.go (file)"]
    import_internal_agents_auth_provider_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_auth_provider_go_os["os (unresolved_import)"]
    struct_internal_agents_auth_provider_go_SovereignAuthProvider["SovereignAuthProvider (struct)"]:::struct
    func_internal_agents_auth_provider_go_SovereignAuthProvider_GetAPIKey["GetAPIKey (function)"]:::func
    file_internal_agents_auth_provider_test_go["auth_provider_test.go (file)"]
    import_internal_agents_auth_provider_test_go_os["os (unresolved_import)"]
    import_internal_agents_auth_provider_test_go_testing["testing (unresolved_import)"]
    func_internal_agents_auth_provider_test_go_TestSovereignAuthProvider_GetAPIKey["TestSovereignAuthProvider_GetAPIKey (function)"]:::func
    file_internal_agents_decompose_test_go["decompose_test.go (file)"]
    import_internal_agents_decompose_test_go_context["context (unresolved_import)"]
    import_internal_agents_decompose_test_go_testing["testing (unresolved_import)"]
    func_internal_agents_decompose_test_go_TestDecomposeTool["TestDecomposeTool (function)"]:::func
    file_internal_agents_dispatcher_go["dispatcher.go (file)"]
    import_internal_agents_dispatcher_go_bufio["bufio (unresolved_import)"]
    import_internal_agents_dispatcher_go_context["context (unresolved_import)"]
    import_internal_agents_dispatcher_go_encoding_json["encoding/json (unresolved_import)"]
    import_internal_agents_dispatcher_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_dispatcher_go_io["io (unresolved_import)"]
    import_internal_agents_dispatcher_go_os["os (unresolved_import)"]
    import_internal_agents_dispatcher_go_path_filepath["path/filepath (unresolved_import)"]
    struct_internal_agents_dispatcher_go_Dispatcher["Dispatcher (struct)"]:::struct
    func_internal_agents_dispatcher_go_NewDispatcher["NewDispatcher (function)"]:::func
    func_internal_agents_dispatcher_go_Dispatcher_Dispatch["Dispatch (function)"]:::func
    func_internal_agents_dispatcher_go_Dispatcher_ReconcileEvents["ReconcileEvents (function)"]:::func
    file_internal_agents_dispatcher_test_go["dispatcher_test.go (file)"]
    import_internal_agents_dispatcher_test_go_context["context (unresolved_import)"]
    import_internal_agents_dispatcher_test_go_encoding_json["encoding/json (unresolved_import)"]
    import_internal_agents_dispatcher_test_go_os["os (unresolved_import)"]
    import_internal_agents_dispatcher_test_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_agents_dispatcher_test_go_testing["testing (unresolved_import)"]
    func_internal_agents_dispatcher_test_go_TestDispatcher_ReconcileEvents["TestDispatcher_ReconcileEvents (function)"]:::func
    file_internal_agents_engine_test_go["engine_test.go (file)"]
    import_internal_agents_engine_test_go_testing["testing (unresolved_import)"]
    struct_internal_agents_engine_test_go_mockAuthProvider["mockAuthProvider (struct)"]:::struct
    func_internal_agents_engine_test_go_mockAuthProvider_GetAPIKey["GetAPIKey (function)"]:::func
    func_internal_agents_engine_test_go_TestNewEngine["TestNewEngine (function)"]:::func
    file_internal_agents_engine_go["engine.go (file)"]
    import_internal_agents_engine_go_context["context (unresolved_import)"]
    import_internal_agents_engine_go_encoding_json["encoding/json (unresolved_import)"]
    import_internal_agents_engine_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_engine_go_log["log (unresolved_import)"]
    import_internal_agents_engine_go_strings["strings (unresolved_import)"]
    import_internal_agents_engine_go_sync["sync (unresolved_import)"]
    import_internal_agents_engine_go_github_com_google_generative_ai_go_genai["github.com/google/generative-ai-go/genai (unresolved_import)"]
    import_internal_agents_engine_go_golang_org_x_sync_errgroup["golang.org/x/sync/errgroup (unresolved_import)"]
    import_internal_agents_engine_go_google_golang_org_api_option["google.golang.org/api/option (unresolved_import)"]
    struct_internal_agents_engine_go_Registry["Registry (struct)"]:::struct
    func_internal_agents_engine_go_NewRegistry["NewRegistry (function)"]:::func
    struct_internal_agents_engine_go_Engine["Engine (struct)"]:::struct
    func_internal_agents_engine_go_NewEngine["NewEngine (function)"]:::func
    func_internal_agents_engine_go_Engine_Close["Close (function)"]:::func
    func_internal_agents_engine_go_Engine_getGenaiTools["getGenaiTools (function)"]:::func
    func_internal_agents_engine_go_Engine_Execute["Execute (function)"]:::func
    func_internal_agents_engine_go_Engine_processSubTasks["processSubTasks (function)"]:::func
    func_internal_agents_engine_go_Engine_executeToolsWithResults["executeToolsWithResults (function)"]:::func
    func_internal_agents_engine_go_Engine_runPACDeliberation["runPACDeliberation (function)"]:::func
    func_internal_agents_engine_go_Engine_shouldEscalate["shouldEscalate (function)"]:::func
    func_internal_agents_engine_go_Engine_escalate["escalate (function)"]:::func
    file_internal_agents_git_shield_go["git_shield.go (file)"]
    import_internal_agents_git_shield_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_git_shield_go_os["os (unresolved_import)"]
    import_internal_agents_git_shield_go_os_exec["os/exec (unresolved_import)"]
    import_internal_agents_git_shield_go_strings["strings (unresolved_import)"]
    struct_internal_agents_git_shield_go_GitShield["GitShield (struct)"]:::struct
    func_internal_agents_git_shield_go_NewGitShield["NewGitShield (function)"]:::func
    func_internal_agents_git_shield_go_GitShield_run["run (function)"]:::func
    func_internal_agents_git_shield_go_GitShield_CreateTaskBranch["CreateTaskBranch (function)"]:::func
    func_internal_agents_git_shield_go_GitShield_CreateWorktree["CreateWorktree (function)"]:::func
    func_internal_agents_git_shield_go_GitShield_RemoveWorktree["RemoveWorktree (function)"]:::func
    func_internal_agents_git_shield_go_GitShield_CleanupWorktrees["CleanupWorktrees (function)"]:::func
    func_internal_agents_git_shield_go_GitShield_AtomicCommit["AtomicCommit (function)"]:::func
    file_internal_agents_git_shield_test_go["git_shield_test.go (file)"]
    import_internal_agents_git_shield_test_go_os["os (unresolved_import)"]
    import_internal_agents_git_shield_test_go_os_exec["os/exec (unresolved_import)"]
    import_internal_agents_git_shield_test_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_agents_git_shield_test_go_testing["testing (unresolved_import)"]
    struct_internal_agents_git_shield_test_go_mockValidator["mockValidator (struct)"]:::struct
    func_internal_agents_git_shield_test_go_mockValidator_ValidatePath["ValidatePath (function)"]:::func
    func_internal_agents_git_shield_test_go_mockValidator_ValidateCommand["ValidateCommand (function)"]:::func
    func_internal_agents_git_shield_test_go_TestGitShield_CreateWorktree["TestGitShield_CreateWorktree (function)"]:::func
    file_internal_agents_loader_go["loader.go (file)"]
    import_internal_agents_loader_go_bufio["bufio (unresolved_import)"]
    import_internal_agents_loader_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_loader_go_os["os (unresolved_import)"]
    import_internal_agents_loader_go_strings["strings (unresolved_import)"]
    import_internal_agents_loader_go_github_com_go_playground_validator_v10["github.com/go-playground/validator/v10 (unresolved_import)"]
    import_internal_agents_loader_go_gopkg_in_yaml_v3["gopkg.in/yaml.v3 (unresolved_import)"]
    struct_internal_agents_loader_go_Loader["Loader (struct)"]:::struct
    func_internal_agents_loader_go_NewLoader["NewLoader (function)"]:::func
    func_internal_agents_loader_go_Loader_LoadAgent["LoadAgent (function)"]:::func
    file_internal_agents_mutation_go["mutation.go (file)"]
    import_internal_agents_mutation_go_bufio["bufio (unresolved_import)"]
    import_internal_agents_mutation_go_context["context (unresolved_import)"]
    import_internal_agents_mutation_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_mutation_go_io["io (unresolved_import)"]
    import_internal_agents_mutation_go_os["os (unresolved_import)"]
    import_internal_agents_mutation_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_agents_mutation_go_regexp["regexp (unresolved_import)"]
    struct_internal_agents_mutation_go_MutationEngine["MutationEngine (struct)"]:::struct
    func_internal_agents_mutation_go_NewMutationEngine["NewMutationEngine (function)"]:::func
    func_internal_agents_mutation_go_MutationEngine_Mutate["Mutate (function)"]:::func
    func_internal_agents_mutation_go_MutationEngine_Rollback["Rollback (function)"]:::func
    file_internal_agents_registry_go["registry.go (file)"]
    import_internal_agents_registry_go_context["context (unresolved_import)"]
    import_internal_agents_registry_go_encoding_json["encoding/json (unresolved_import)"]
    import_internal_agents_registry_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_registry_go_strings["strings (unresolved_import)"]
    struct_internal_agents_registry_go_RegistryManager["RegistryManager (struct)"]:::struct
    func_internal_agents_registry_go_NewRegistryManager["NewRegistryManager (function)"]:::func
    func_internal_agents_registry_go_RegistryManager_SelectBest["SelectBest (function)"]:::func
    func_internal_agents_registry_go_RegistryManager_matchesAll["matchesAll (function)"]:::func
    file_internal_agents_registry_test_go["registry_test.go (file)"]
    import_internal_agents_registry_test_go_context["context (unresolved_import)"]
    import_internal_agents_registry_test_go_database_sql["database/sql (unresolved_import)"]
    import_internal_agents_registry_test_go_encoding_json["encoding/json (unresolved_import)"]
    import_internal_agents_registry_test_go_os["os (unresolved_import)"]
    import_internal_agents_registry_test_go_testing["testing (unresolved_import)"]
    import_internal_agents_registry_test_go_modernc_org_sqlite["modernc.org/sqlite (unresolved_import)"]
    func_internal_agents_registry_test_go_TestRegistryManager_SelectBest["TestRegistryManager_SelectBest (function)"]:::func
    file_internal_agents_types_go["types.go (file)"]
    import_internal_agents_types_go_context["context (unresolved_import)"]
    import_internal_agents_types_go_sync["sync (unresolved_import)"]
    struct_internal_agents_types_go_Specialist["Specialist (struct)"]:::struct
    struct_internal_agents_types_go_TokenBudget["TokenBudget (struct)"]:::struct
    func_internal_agents_types_go_TokenBudget_AddTokens["AddTokens (function)"]:::func
    func_internal_agents_types_go_TokenBudget_IncSteps["IncSteps (function)"]:::func
    struct_internal_agents_types_go_AgentDefinition["AgentDefinition (struct)"]:::struct
    struct_internal_agents_types_go_Message["Message (struct)"]:::struct
    struct_internal_agents_types_go_AgentContext["AgentContext (struct)"]:::struct
    struct_internal_agents_types_go_SubTask["SubTask (struct)"]:::struct
    func_internal_agents_types_go_NewAgentContext["NewAgentContext (function)"]:::func
    file_internal_agents_tools_go["tools.go (file)"]
    import_internal_agents_tools_go_bufio["bufio (unresolved_import)"]
    import_internal_agents_tools_go_bytes["bytes (unresolved_import)"]
    import_internal_agents_tools_go_context["context (unresolved_import)"]
    import_internal_agents_tools_go_encoding_json["encoding/json (unresolved_import)"]
    import_internal_agents_tools_go_fmt["fmt (unresolved_import)"]
    import_internal_agents_tools_go_os["os (unresolved_import)"]
    import_internal_agents_tools_go_os_exec["os/exec (unresolved_import)"]
    import_internal_agents_tools_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_agents_tools_go_regexp["regexp (unresolved_import)"]
    import_internal_agents_tools_go_strings["strings (unresolved_import)"]
    import_internal_agents_tools_go_github_com_google_generative_ai_go_genai["github.com/google/generative-ai-go/genai (unresolved_import)"]
    import_internal_agents_tools_go_github_com_google_shlex["github.com/google/shlex (unresolved_import)"]
    import_internal_agents_tools_go_github_com_google_uuid["github.com/google/uuid (unresolved_import)"]
    struct_internal_agents_tools_go_ReadFileTool["ReadFileTool (struct)"]:::struct
    func_internal_agents_tools_go_ReadFileTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_ReadFileTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_ReadFileTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_ReadFileTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_ReadFileTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_WriteFileTool["WriteFileTool (struct)"]:::struct
    func_internal_agents_tools_go_WriteFileTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_WriteFileTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_WriteFileTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_WriteFileTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_WriteFileTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_ReplaceTool["ReplaceTool (struct)"]:::struct
    func_internal_agents_tools_go_ReplaceTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_ReplaceTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_ReplaceTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_ReplaceTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_ReplaceTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_GrepSearchTool["GrepSearchTool (struct)"]:::struct
    func_internal_agents_tools_go_GrepSearchTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_GrepSearchTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_GrepSearchTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_GrepSearchTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_GrepSearchTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_AuditTool["AuditTool (struct)"]:::struct
    func_internal_agents_tools_go_AuditTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_AuditTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_AuditTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_AuditTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_AuditTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_RunTool["RunTool (struct)"]:::struct
    func_internal_agents_tools_go_RunTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_RunTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_RunTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_RunTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_RunTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_ADRTool["ADRTool (struct)"]:::struct
    func_internal_agents_tools_go_ADRTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_ADRTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_ADRTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_ADRTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_ADRTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_ScanTool["ScanTool (struct)"]:::struct
    func_internal_agents_tools_go_ScanTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_ScanTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_ScanTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_ScanTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_ScanTool_Execute["Execute (function)"]:::func
    struct_internal_agents_tools_go_DecomposeTool["DecomposeTool (struct)"]:::struct
    func_internal_agents_tools_go_DecomposeTool_Name["Name (function)"]:::func
    func_internal_agents_tools_go_DecomposeTool_Description["Description (function)"]:::func
    func_internal_agents_tools_go_DecomposeTool_Definition["Definition (function)"]:::func
    func_internal_agents_tools_go_DecomposeTool_ValidateArguments["ValidateArguments (function)"]:::func
    func_internal_agents_tools_go_DecomposeTool_Execute["Execute (function)"]:::func
    func_internal_agents_tools_go_RegisterCoreTools["RegisterCoreTools (function)"]:::func
    file_internal_audit_runner_go["runner.go (file)"]
    import_internal_audit_runner_go_bytes["bytes (unresolved_import)"]
    import_internal_audit_runner_go_context["context (unresolved_import)"]
    import_internal_audit_runner_go_errors["errors (unresolved_import)"]
    import_internal_audit_runner_go_fmt["fmt (unresolved_import)"]
    import_internal_audit_runner_go_os_exec["os/exec (unresolved_import)"]
    import_internal_audit_runner_go_time["time (unresolved_import)"]
    import_internal_audit_runner_go_github_com_google_shlex["github.com/google/shlex (unresolved_import)"]
    struct_internal_audit_runner_go_Runner["Runner (struct)"]:::struct
    func_internal_audit_runner_go_NewRunner["NewRunner (function)"]:::func
    func_internal_audit_runner_go_Runner_ExecuteAudit["ExecuteAudit (function)"]:::func
    file_internal_bridge_prompt_factory_go["prompt_factory.go (file)"]
    import_internal_bridge_prompt_factory_go_bufio["bufio (unresolved_import)"]
    import_internal_bridge_prompt_factory_go_fmt["fmt (unresolved_import)"]
    import_internal_bridge_prompt_factory_go_os["os (unresolved_import)"]
    import_internal_bridge_prompt_factory_go_strings["strings (unresolved_import)"]
    import_internal_bridge_prompt_factory_go_text_template["text/template (unresolved_import)"]
    struct_internal_bridge_prompt_factory_go_ADR["ADR (struct)"]:::struct
    struct_internal_bridge_prompt_factory_go_ContextNode["ContextNode (struct)"]:::struct
    struct_internal_bridge_prompt_factory_go_ContextPayload["ContextPayload (struct)"]:::struct
    struct_internal_bridge_prompt_factory_go_Factory["Factory (struct)"]:::struct
    func_internal_bridge_prompt_factory_go_NewFactory["NewFactory (function)"]:::func
    func_internal_bridge_prompt_factory_go_Factory_GeneratePayload["GeneratePayload (function)"]:::func
    struct_internal_bridge_prompt_factory_go_SystemData["SystemData (struct)"]:::struct
    func_internal_bridge_prompt_factory_go_Factory_loadADRs["loadADRs (function)"]:::func
    func_internal_bridge_prompt_factory_go_Factory_loadSurgicalContext["loadSurgicalContext (function)"]:::func
    func_internal_bridge_prompt_factory_go_extractLines["extractLines (function)"]:::func
    file_internal_graph_adr_generator_go["adr_generator.go (file)"]
    import_internal_graph_adr_generator_go_fmt["fmt (unresolved_import)"]
    import_internal_graph_adr_generator_go_os["os (unresolved_import)"]
    import_internal_graph_adr_generator_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_graph_adr_generator_go_time["time (unresolved_import)"]
    struct_internal_graph_adr_generator_go_ADRGenerator["ADRGenerator (struct)"]:::struct
    func_internal_graph_adr_generator_go_NewADRGenerator["NewADRGenerator (function)"]:::func
    struct_internal_graph_adr_generator_go_ADRData["ADRData (struct)"]:::struct
    func_internal_graph_adr_generator_go_ADRGenerator_Generate["Generate (function)"]:::func
    file_internal_graph_adr_generator_test_go["adr_generator_test.go (file)"]
    import_internal_graph_adr_generator_test_go_io["io (unresolved_import)"]
    import_internal_graph_adr_generator_test_go_os["os (unresolved_import)"]
    import_internal_graph_adr_generator_test_go_strings["strings (unresolved_import)"]
    import_internal_graph_adr_generator_test_go_testing["testing (unresolved_import)"]
    func_internal_graph_adr_generator_test_go_TestADRGenerator_Generate["TestADRGenerator_Generate (function)"]:::func
    file_internal_graph_engine_go["engine.go (file)"]
    import_internal_graph_engine_go_database_sql["database/sql (unresolved_import)"]
    import_internal_graph_engine_go_fmt["fmt (unresolved_import)"]
    import_internal_graph_engine_go_os["os (unresolved_import)"]
    import_internal_graph_engine_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_graph_engine_go_sync["sync (unresolved_import)"]
    struct_internal_graph_engine_go_Engine["Engine (struct)"]:::struct
    func_internal_graph_engine_go_NewEngine["NewEngine (function)"]:::func
    func_internal_graph_engine_go_Engine_RegisterScanner["RegisterScanner (function)"]:::func
    func_internal_graph_engine_go_Engine_ScanProject["ScanProject (function)"]:::func
    func_internal_graph_engine_go_Engine_scanFileWithIncrementalCheck["scanFileWithIncrementalCheck (function)"]:::func
    func_internal_graph_engine_go_Engine_persistResult["persistResult (function)"]:::func
    func_internal_graph_engine_go_Engine_upsertNodeTx["upsertNodeTx (function)"]:::func
    func_internal_graph_engine_go_Engine_createEdgeTx["createEdgeTx (function)"]:::func
    file_internal_graph_linker_go["linker.go (file)"]
    import_internal_graph_linker_go_fmt["fmt (unresolved_import)"]
    import_internal_graph_linker_go_os["os (unresolved_import)"]
    import_internal_graph_linker_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_graph_linker_go_strings["strings (unresolved_import)"]
    func_internal_graph_linker_go_Engine_LinkDependencies["LinkDependencies (function)"]:::func
    struct_internal_graph_linker_go_pendingLink["pendingLink (struct)"]:::struct
    func_internal_graph_linker_go_Engine_resolveImport["resolveImport (function)"]:::func
    func_internal_graph_linker_go_Engine_createRealEdge["createRealEdge (function)"]:::func
    file_internal_graph_scanner_go_go["scanner_go.go (file)"]
    import_internal_graph_scanner_go_go_fmt["fmt (unresolved_import)"]
    import_internal_graph_scanner_go_go_go_ast["go/ast (unresolved_import)"]
    import_internal_graph_scanner_go_go_go_parser["go/parser (unresolved_import)"]
    import_internal_graph_scanner_go_go_go_token["go/token (unresolved_import)"]
    import_internal_graph_scanner_go_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_graph_scanner_go_go_strings["strings (unresolved_import)"]
    struct_internal_graph_scanner_go_go_GoScanner["GoScanner (struct)"]:::struct
    func_internal_graph_scanner_go_go_NewGoScanner["NewGoScanner (function)"]:::func
    func_internal_graph_scanner_go_go_GoScanner_SupportedExtensions["SupportedExtensions (function)"]:::func
    func_internal_graph_scanner_go_go_GoScanner_Scan["Scan (function)"]:::func
    file_internal_graph_scanner_treesitter_go["scanner_treesitter.go (file)"]
    import_internal_graph_scanner_treesitter_go_context["context (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_fmt["fmt (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_io["io (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_os["os (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_strings["strings (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_sync["sync (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_github_com_smacker_go_tree_sitter["github.com/smacker/go-tree-sitter (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_github_com_smacker_go_tree_sitter_typescript_tsx["github.com/smacker/go-tree-sitter/typescript/tsx (unresolved_import)"]
    import_internal_graph_scanner_treesitter_go_github_com_smacker_go_tree_sitter_typescript_typescript["github.com/smacker/go-tree-sitter/typescript/typescript (unresolved_import)"]
    struct_internal_graph_scanner_treesitter_go_TreeSitterScanner["TreeSitterScanner (struct)"]:::struct
    func_internal_graph_scanner_treesitter_go_NewTreeSitterScanner["NewTreeSitterScanner (function)"]:::func
    func_internal_graph_scanner_treesitter_go_TreeSitterScanner_SupportedExtensions["SupportedExtensions (function)"]:::func
    func_internal_graph_scanner_treesitter_go_TreeSitterScanner_Scan["Scan (function)"]:::func
    func_internal_graph_scanner_treesitter_go_TreeSitterScanner_processSymbol["processSymbol (function)"]:::func
    file_internal_graph_schema_go["schema.go (file)"]
    import_internal_graph_schema_go_fmt["fmt (unresolved_import)"]
    func_internal_graph_schema_go_Migrate["Migrate (function)"]:::func
    file_internal_graph_schema_test_go["schema_test.go (file)"]
    import_internal_graph_schema_test_go_database_sql["database/sql (unresolved_import)"]
    import_internal_graph_schema_test_go_os["os (unresolved_import)"]
    import_internal_graph_schema_test_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_graph_schema_test_go_testing["testing (unresolved_import)"]
    import_internal_graph_schema_test_go_modernc_org_sqlite["modernc.org/sqlite (unresolved_import)"]
    func_internal_graph_schema_test_go_TestMigrate["TestMigrate (function)"]:::func
    file_internal_graph_types_go["types.go (file)"]
    import_internal_graph_types_go_time["time (unresolved_import)"]
    struct_internal_graph_types_go_Node["Node (struct)"]:::struct
    struct_internal_graph_types_go_Edge["Edge (struct)"]:::struct
    struct_internal_graph_types_go_ScanResult["ScanResult (struct)"]:::struct
    file_internal_graph_visualizer_go["visualizer.go (file)"]
    import_internal_graph_visualizer_go_fmt["fmt (unresolved_import)"]
    import_internal_graph_visualizer_go_os["os (unresolved_import)"]
    import_internal_graph_visualizer_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_graph_visualizer_go_strings["strings (unresolved_import)"]
    struct_internal_graph_visualizer_go_Visualizer["Visualizer (struct)"]:::struct
    func_internal_graph_visualizer_go_NewVisualizer["NewVisualizer (function)"]:::func
    func_internal_graph_visualizer_go_Visualizer_GenerateMasterDiagram["GenerateMasterDiagram (function)"]:::func
    func_internal_graph_visualizer_go_Visualizer_GenerateTaskSnapshot["GenerateTaskSnapshot (function)"]:::func
    func_internal_graph_visualizer_go_Visualizer_getNodes["getNodes (function)"]:::func
    func_internal_graph_visualizer_go_Visualizer_GenerateC4ContainerDiagram["GenerateC4ContainerDiagram (function)"]:::func
    func_internal_graph_visualizer_go_Visualizer_formatC4Mermaid["formatC4Mermaid (function)"]:::func
    struct_internal_graph_visualizer_go_container["container (struct)"]:::struct
    struct_internal_graph_visualizer_go_relKey["relKey (struct)"]:::struct
    func_internal_graph_visualizer_go_Visualizer_formatMermaid["formatMermaid (function)"]:::func
    func_internal_graph_visualizer_go_Visualizer_getEdges["getEdges (function)"]:::func
    file_internal_reflect_validator_go["validator.go (file)"]
    import_internal_reflect_validator_go_bufio["bufio (unresolved_import)"]
    import_internal_reflect_validator_go_fmt["fmt (unresolved_import)"]
    import_internal_reflect_validator_go_os["os (unresolved_import)"]
    import_internal_reflect_validator_go_path_filepath["path/filepath (unresolved_import)"]
    import_internal_reflect_validator_go_strings["strings (unresolved_import)"]
    struct_internal_reflect_validator_go_Violation["Violation (struct)"]:::struct
    struct_internal_reflect_validator_go_Validator["Validator (struct)"]:::struct
    func_internal_reflect_validator_go_NewValidator["NewValidator (function)"]:::func
    func_internal_reflect_validator_go_Validator_ValidateProject["ValidateProject (function)"]:::func
    func_internal_reflect_validator_go_Validator_ValidatePath["ValidatePath (function)"]:::func
    func_internal_reflect_validator_go_Validator_ValidateCommand["ValidateCommand (function)"]:::func
    func_internal_reflect_validator_go_Validator_checkFile["checkFile (function)"]:::func
    func_internal_reflect_validator_go_isIgnored["isIgnored (function)"]:::func
    file_internal_registry_commands_go["commands.go (file)"]
    import_internal_registry_commands_go_github_com_spf13_cobra["github.com/spf13/cobra (unresolved_import)"]
    func_internal_registry_commands_go_Register["Register (function)"]:::func
    func_internal_registry_commands_go_GetCommands["GetCommands (function)"]:::func
    file_internal_report_aggregator_go["aggregator.go (file)"]
    import_internal_report_aggregator_go_fmt["fmt (unresolved_import)"]
    import_internal_report_aggregator_go_os["os (unresolved_import)"]
    import_internal_report_aggregator_go_path_filepath["path/filepath (unresolved_import)"]
    struct_internal_report_aggregator_go_TaskInfo["TaskInfo (struct)"]:::struct
    struct_internal_report_aggregator_go_ProjectStats["ProjectStats (struct)"]:::struct
    struct_internal_report_aggregator_go_Aggregator["Aggregator (struct)"]:::struct
    func_internal_report_aggregator_go_NewAggregator["NewAggregator (function)"]:::func
    func_internal_report_aggregator_go_Aggregator_FetchStats["FetchStats (function)"]:::func
    func_internal_report_aggregator_go_Aggregator_GenerateMarkdown["GenerateMarkdown (function)"]:::func
    file_internal_state_manager_go["manager.go (file)"]
    import_internal_state_manager_go_fmt["fmt (unresolved_import)"]
    import_internal_state_manager_go_time["time (unresolved_import)"]
    import_internal_state_manager_go_github_com_google_uuid["github.com/google/uuid (unresolved_import)"]
    struct_internal_state_manager_go_Task["Task (struct)"]:::struct
    struct_internal_state_manager_go_Manager["Manager (struct)"]:::struct
    func_internal_state_manager_go_NewManager["NewManager (function)"]:::func
    func_internal_state_manager_go_Manager_CreateTask["CreateTask (function)"]:::func
    func_internal_state_manager_go_Manager_StartTask["StartTask (function)"]:::func
    func_internal_state_manager_go_Manager_GetTaskByID["GetTaskByID (function)"]:::func
    func_internal_state_manager_go_Manager_UpdateStatus["UpdateStatus (function)"]:::func
    func_internal_state_manager_go_Manager_GetActiveTask["GetActiveTask (function)"]:::func
    func_internal_state_manager_go_Manager_ListTasks["ListTasks (function)"]:::func
    file_legacy_ts_src_brainstorm_ts["brainstorm.ts (file)"]
    import_legacy_ts_src_brainstorm_ts_fs["fs (unresolved_import)"]
    import_legacy_ts_src_brainstorm_ts_path["path (unresolved_import)"]
    import_legacy_ts_src_brainstorm_ts_inquirer["inquirer (unresolved_import)"]
    function_legacy_ts_src_brainstorm_ts_runBrainstorm["runBrainstorm (function)"]:::func
    function_legacy_ts_src_brainstorm_ts_config["config (function)"]:::func
    component_legacy_ts_src_brainstorm_ts_INSIGHTS_DIR["INSIGHTS_DIR (component)"]
    function_legacy_ts_src_brainstorm_ts_theme["theme (function)"]:::func
    function_legacy_ts_src_brainstorm_ts_sanitizedTheme["sanitizedTheme (function)"]:::func
    function_legacy_ts_src_brainstorm_ts_reportFile["reportFile (function)"]:::func
    function_legacy_ts_src_brainstorm_ts_reportPath["reportPath (function)"]:::func
    function_legacy_ts_src_brainstorm_ts_skeleton["skeleton (function)"]:::func
    file_legacy_ts_src_config_loader_ts["config-loader.ts (file)"]
    import_legacy_ts_src_config_loader_ts_fs["fs (unresolved_import)"]
    import_legacy_ts_src_config_loader_ts_path["path (unresolved_import)"]
    interface_legacy_ts_src_config_loader_ts_SentinelConfig["SentinelConfig (interface)"]
    component_legacy_ts_src_config_loader_ts_CONFIG_FILENAME["CONFIG_FILENAME (component)"]
    function_legacy_ts_src_config_loader_ts_loadConfig["loadConfig (function)"]:::func
    function_legacy_ts_src_config_loader_ts_configPath["configPath (function)"]:::func
    function_legacy_ts_src_config_loader_ts_raw["raw (function)"]:::func
    file_legacy_ts_src_cli_ts["cli.ts (file)"]
    import_legacy_ts_src_cli_ts_child_process["child_process (unresolved_import)"]
    import_legacy_ts_src_cli_ts_fs["fs (unresolved_import)"]
    import_legacy_ts_src_cli_ts_path["path (unresolved_import)"]
    import_legacy_ts_src_cli_ts_inquirer["inquirer (unresolved_import)"]
    function_legacy_ts_src_cli_ts_plan["plan (function)"]:::func
    function_legacy_ts_src_cli_ts_config["config (function)"]:::func
    component_legacy_ts_src_cli_ts_PLAN_PATH["PLAN_PATH (component)"]
    function_legacy_ts_src_cli_ts_skeleton["skeleton (function)"]:::func
    function_legacy_ts_src_cli_ts_forge["forge (function)"]:::func
    function_legacy_ts_src_cli_ts_insightsDir["insightsDir (function)"]:::func
    function_legacy_ts_src_cli_ts_files["files (function)"]:::func
    function_legacy_ts_src_cli_ts_audit["audit (function)"]:::func
    function_legacy_ts_src_cli_ts_verifyPlan["verifyPlan (function)"]:::func
    function_legacy_ts_src_cli_ts_content["content (function)"]:::func
    function_legacy_ts_src_cli_ts_markers["markers (function)"]:::func
    function_legacy_ts_src_cli_ts_command["command (function)"]:::func
    function_legacy_ts_src_cli_ts_main["main (function)"]:::func
    function_legacy_ts_src_cli_ts_c["c (function)"]:::func
    file_legacy_ts_src_sentinel_wrapper_ts["sentinel-wrapper.ts (file)"]
    file_legacy_ts_src_forge_engine_ts["forge-engine.ts (file)"]
    import_legacy_ts_src_forge_engine_ts_fs["fs (unresolved_import)"]
    import_legacy_ts_src_forge_engine_ts_path["path (unresolved_import)"]
    function_legacy_ts_src_forge_engine_ts_getAllFiles["getAllFiles (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_files["files (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_forgePlan["forgePlan (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_config["config (function)"]:::func
    component_legacy_ts_src_forge_engine_ts_INSIGHTS_DIR["INSIGHTS_DIR (component)"]
    component_legacy_ts_src_forge_engine_ts_PLAN_PATH["PLAN_PATH (component)"]
    component_legacy_ts_src_forge_engine_ts_FPA_RULES_PATH["FPA_RULES_PATH (component)"]
    function_legacy_ts_src_forge_engine_ts_insightPath["insightPath (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_content["content (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_pathDivider["pathDivider (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_nextPathDivider["nextPathDivider (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_splitParts["splitParts (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_splitStart["splitStart (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_chunk["chunk (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_lines["lines (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_pathName["pathName (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_visionLine["visionLine (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_vision["vision (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_impactFiles["impactFiles (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_keywords["keywords (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_allFilesToScan["allFilesToScan (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_fileContent["fileContent (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_totalFP["totalFP (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_criticalPathEntries["criticalPathEntries (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_relativePath["relativePath (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_criticalMatch["criticalMatch (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_tier["tier (function)"]:::func
    function_legacy_ts_src_forge_engine_ts_plan["plan (function)"]:::func
    file_pkg_sqlite_db_go["db.go (file)"]
    import_pkg_sqlite_db_go_database_sql["database/sql (unresolved_import)"]
    import_pkg_sqlite_db_go_fmt["fmt (unresolved_import)"]
    import_pkg_sqlite_db_go_os["os (unresolved_import)"]
    import_pkg_sqlite_db_go_path_filepath["path/filepath (unresolved_import)"]
    import_pkg_sqlite_db_go_modernc_org_sqlite["modernc.org/sqlite (unresolved_import)"]
    struct_pkg_sqlite_db_go_DB["DB (struct)"]:::struct
    func_pkg_sqlite_db_go_Init["Init (function)"]:::func
    func_pkg_sqlite_db_go_DB_Close["Close (function)"]:::func
    file_pkg_utils_hash_go["hash.go (file)"]
    import_pkg_utils_hash_go_crypto_sha256["crypto/sha256 (unresolved_import)"]
    import_pkg_utils_hash_go_fmt["fmt (unresolved_import)"]
    import_pkg_utils_hash_go_io["io (unresolved_import)"]
    import_pkg_utils_hash_go_os["os (unresolved_import)"]
    func_pkg_utils_hash_go_CalculateHash["CalculateHash (function)"]:::func
    file_pkg_utils_text_go["text.go (file)"]
    import_pkg_utils_text_go_regexp["regexp (unresolved_import)"]
    import_pkg_utils_text_go_strings["strings (unresolved_import)"]
    func_pkg_utils_text_go_SanitizeID["SanitizeID (function)"]:::func
    func_pkg_utils_text_go_Slugify["Slugify (function)"]:::func
    func_pkg_utils_text_go_EscapeYAML["EscapeYAML (function)"]:::func
    file_pkg_utils_filter_go["filter.go (file)"]
    import_pkg_utils_filter_go_bufio["bufio (unresolved_import)"]
    import_pkg_utils_filter_go_os["os (unresolved_import)"]
    import_pkg_utils_filter_go_path_filepath["path/filepath (unresolved_import)"]
    import_pkg_utils_filter_go_strings["strings (unresolved_import)"]
    struct_pkg_utils_filter_go_IgnoreFilter["IgnoreFilter (struct)"]:::struct
    func_pkg_utils_filter_go_NewIgnoreFilter["NewIgnoreFilter (function)"]:::func
    func_pkg_utils_filter_go_IgnoreFilter_loadGitignore["loadGitignore (function)"]:::func
    func_pkg_utils_filter_go_IgnoreFilter_IsIgnored["IsIgnored (function)"]:::func
    file_cmd_sentinel_commands_scan_go -->|imports| import_cmd_sentinel_commands_scan_go_fmt
    file_cmd_sentinel_commands_scan_go -->|imports| import_cmd_sentinel_commands_scan_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_scan_go -->|contains| func_cmd_sentinel_commands_scan_go_init
    file_cmd_sentinel_commands_scan_go -->|contains| func_cmd_sentinel_commands_scan_go_NewScanCmd
    file_cmd_sentinel_commands_visualize_go -->|imports| import_cmd_sentinel_commands_visualize_go_fmt
    file_cmd_sentinel_commands_visualize_go -->|imports| import_cmd_sentinel_commands_visualize_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_visualize_go -->|contains| func_cmd_sentinel_commands_visualize_go_init
    file_cmd_sentinel_commands_visualize_go -->|contains| func_cmd_sentinel_commands_visualize_go_NewVisualizeCmd
    file_cmd_sentinel_commands_root_go -->|imports| import_cmd_sentinel_commands_root_go_fmt
    file_cmd_sentinel_commands_root_go -->|imports| import_cmd_sentinel_commands_root_go_os
    file_cmd_sentinel_commands_root_go -->|imports| import_cmd_sentinel_commands_root_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_root_go -->|contains| func_cmd_sentinel_commands_root_go_NewRootCmd
    file_cmd_sentinel_commands_root_go -->|contains| func_cmd_sentinel_commands_root_go_Execute
    file_cmd_sentinel_commands_start_go -->|imports| import_cmd_sentinel_commands_start_go_fmt
    file_cmd_sentinel_commands_start_go -->|imports| import_cmd_sentinel_commands_start_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_start_go -->|contains| func_cmd_sentinel_commands_start_go_init
    file_cmd_sentinel_commands_start_go -->|contains| func_cmd_sentinel_commands_start_go_NewStartCmd
    file_cmd_sentinel_commands_plan_go -->|imports| import_cmd_sentinel_commands_plan_go_fmt
    file_cmd_sentinel_commands_plan_go -->|imports| import_cmd_sentinel_commands_plan_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_plan_go -->|contains| func_cmd_sentinel_commands_plan_go_init
    file_cmd_sentinel_commands_plan_go -->|contains| func_cmd_sentinel_commands_plan_go_NewPlanCmd
    file_cmd_sentinel_commands_status_go -->|imports| import_cmd_sentinel_commands_status_go_fmt
    file_cmd_sentinel_commands_status_go -->|imports| import_cmd_sentinel_commands_status_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_status_go -->|contains| func_cmd_sentinel_commands_status_go_init
    file_cmd_sentinel_commands_status_go -->|contains| func_cmd_sentinel_commands_status_go_NewStatusCmd
    file_cmd_sentinel_commands_report_go -->|imports| import_cmd_sentinel_commands_report_go_fmt
    file_cmd_sentinel_commands_report_go -->|imports| import_cmd_sentinel_commands_report_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_report_go -->|contains| func_cmd_sentinel_commands_report_go_init
    file_cmd_sentinel_commands_report_go -->|contains| func_cmd_sentinel_commands_report_go_NewReportCmd
    file_cmd_sentinel_commands_audit_go -->|imports| import_cmd_sentinel_commands_audit_go_errors
    file_cmd_sentinel_commands_audit_go -->|imports| import_cmd_sentinel_commands_audit_go_fmt
    file_cmd_sentinel_commands_audit_go -->|imports| import_cmd_sentinel_commands_audit_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_audit_go -->|contains| func_cmd_sentinel_commands_audit_go_init
    file_cmd_sentinel_commands_audit_go -->|contains| func_cmd_sentinel_commands_audit_go_NewAuditCmd
    file_cmd_sentinel_commands_instruct_go -->|imports| import_cmd_sentinel_commands_instruct_go_bufio
    file_cmd_sentinel_commands_instruct_go -->|imports| import_cmd_sentinel_commands_instruct_go_context
    file_cmd_sentinel_commands_instruct_go -->|imports| import_cmd_sentinel_commands_instruct_go_fmt
    file_cmd_sentinel_commands_instruct_go -->|imports| import_cmd_sentinel_commands_instruct_go_os
    file_cmd_sentinel_commands_instruct_go -->|imports| import_cmd_sentinel_commands_instruct_go_strings
    file_cmd_sentinel_commands_instruct_go -->|imports| import_cmd_sentinel_commands_instruct_go_github_com_spf13_cobra
    file_cmd_sentinel_commands_instruct_go -->|contains| func_cmd_sentinel_commands_instruct_go_init
    file_cmd_sentinel_commands_instruct_go -->|contains| func_cmd_sentinel_commands_instruct_go_NewInstructCmd
    file_cmd_sentinel_commands_instruct_go -->|contains| func_cmd_sentinel_commands_instruct_go_performDiagnostic
    file_cmd_sentinel_commands_instruct_go -->|contains| func_cmd_sentinel_commands_instruct_go_runSocraticInterview
    file_cmd_sentinel_main_go -->|imports| import_cmd_sentinel_main_go_fmt
    file_cmd_sentinel_main_go -->|imports| import_cmd_sentinel_main_go_os
    file_cmd_sentinel_main_go -->|contains| func_cmd_sentinel_main_go_main
    file_internal_agents_auth_provider_go -->|imports| import_internal_agents_auth_provider_go_fmt
    file_internal_agents_auth_provider_go -->|imports| import_internal_agents_auth_provider_go_os
    file_internal_agents_auth_provider_go -->|contains| struct_internal_agents_auth_provider_go_SovereignAuthProvider
    file_internal_agents_auth_provider_go -->|contains| func_internal_agents_auth_provider_go_SovereignAuthProvider_GetAPIKey
    file_internal_agents_auth_provider_test_go -->|imports| import_internal_agents_auth_provider_test_go_os
    file_internal_agents_auth_provider_test_go -->|imports| import_internal_agents_auth_provider_test_go_testing
    file_internal_agents_auth_provider_test_go -->|contains| func_internal_agents_auth_provider_test_go_TestSovereignAuthProvider_GetAPIKey
    file_internal_agents_decompose_test_go -->|imports| import_internal_agents_decompose_test_go_context
    file_internal_agents_decompose_test_go -->|imports| import_internal_agents_decompose_test_go_testing
    file_internal_agents_decompose_test_go -->|contains| func_internal_agents_decompose_test_go_TestDecomposeTool
    file_internal_agents_dispatcher_go -->|imports| import_internal_agents_dispatcher_go_bufio
    file_internal_agents_dispatcher_go -->|imports| import_internal_agents_dispatcher_go_context
    file_internal_agents_dispatcher_go -->|imports| import_internal_agents_dispatcher_go_encoding_json
    file_internal_agents_dispatcher_go -->|imports| import_internal_agents_dispatcher_go_fmt
    file_internal_agents_dispatcher_go -->|imports| import_internal_agents_dispatcher_go_io
    file_internal_agents_dispatcher_go -->|imports| import_internal_agents_dispatcher_go_os
    file_internal_agents_dispatcher_go -->|imports| import_internal_agents_dispatcher_go_path_filepath
    file_internal_agents_dispatcher_go -->|contains| struct_internal_agents_dispatcher_go_Dispatcher
    file_internal_agents_dispatcher_go -->|contains| func_internal_agents_dispatcher_go_NewDispatcher
    file_internal_agents_dispatcher_go -->|contains| func_internal_agents_dispatcher_go_Dispatcher_Dispatch
    file_internal_agents_dispatcher_go -->|contains| func_internal_agents_dispatcher_go_Dispatcher_ReconcileEvents
    file_internal_agents_dispatcher_test_go -->|imports| import_internal_agents_dispatcher_test_go_context
    file_internal_agents_dispatcher_test_go -->|imports| import_internal_agents_dispatcher_test_go_encoding_json
    file_internal_agents_dispatcher_test_go -->|imports| import_internal_agents_dispatcher_test_go_os
    file_internal_agents_dispatcher_test_go -->|imports| import_internal_agents_dispatcher_test_go_path_filepath
    file_internal_agents_dispatcher_test_go -->|imports| import_internal_agents_dispatcher_test_go_testing
    file_internal_agents_dispatcher_test_go -->|contains| func_internal_agents_dispatcher_test_go_TestDispatcher_ReconcileEvents
    file_internal_agents_engine_test_go -->|imports| import_internal_agents_engine_test_go_testing
    file_internal_agents_engine_test_go -->|contains| struct_internal_agents_engine_test_go_mockAuthProvider
    file_internal_agents_engine_test_go -->|contains| func_internal_agents_engine_test_go_mockAuthProvider_GetAPIKey
    file_internal_agents_engine_test_go -->|contains| func_internal_agents_engine_test_go_TestNewEngine
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_context
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_encoding_json
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_fmt
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_log
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_strings
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_sync
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_github_com_google_generative_ai_go_genai
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_golang_org_x_sync_errgroup
    file_internal_agents_engine_go -->|imports| import_internal_agents_engine_go_google_golang_org_api_option
    file_internal_agents_engine_go -->|contains| struct_internal_agents_engine_go_Registry
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_NewRegistry
    file_internal_agents_engine_go -->|contains| struct_internal_agents_engine_go_Engine
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_NewEngine
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_Close
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_getGenaiTools
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_Execute
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_processSubTasks
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_executeToolsWithResults
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_runPACDeliberation
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_shouldEscalate
    file_internal_agents_engine_go -->|contains| func_internal_agents_engine_go_Engine_escalate
    file_internal_agents_git_shield_go -->|imports| import_internal_agents_git_shield_go_fmt
    file_internal_agents_git_shield_go -->|imports| import_internal_agents_git_shield_go_os
    file_internal_agents_git_shield_go -->|imports| import_internal_agents_git_shield_go_os_exec
    file_internal_agents_git_shield_go -->|imports| import_internal_agents_git_shield_go_strings
    file_internal_agents_git_shield_go -->|contains| struct_internal_agents_git_shield_go_GitShield
    file_internal_agents_git_shield_go -->|contains| func_internal_agents_git_shield_go_NewGitShield
    file_internal_agents_git_shield_go -->|contains| func_internal_agents_git_shield_go_GitShield_run
    file_internal_agents_git_shield_go -->|contains| func_internal_agents_git_shield_go_GitShield_CreateTaskBranch
    file_internal_agents_git_shield_go -->|contains| func_internal_agents_git_shield_go_GitShield_CreateWorktree
    file_internal_agents_git_shield_go -->|contains| func_internal_agents_git_shield_go_GitShield_RemoveWorktree
    file_internal_agents_git_shield_go -->|contains| func_internal_agents_git_shield_go_GitShield_CleanupWorktrees
    file_internal_agents_git_shield_go -->|contains| func_internal_agents_git_shield_go_GitShield_AtomicCommit
    file_internal_agents_git_shield_test_go -->|imports| import_internal_agents_git_shield_test_go_os
    file_internal_agents_git_shield_test_go -->|imports| import_internal_agents_git_shield_test_go_os_exec
    file_internal_agents_git_shield_test_go -->|imports| import_internal_agents_git_shield_test_go_path_filepath
    file_internal_agents_git_shield_test_go -->|imports| import_internal_agents_git_shield_test_go_testing
    file_internal_agents_git_shield_test_go -->|contains| struct_internal_agents_git_shield_test_go_mockValidator
    file_internal_agents_git_shield_test_go -->|contains| func_internal_agents_git_shield_test_go_mockValidator_ValidatePath
    file_internal_agents_git_shield_test_go -->|contains| func_internal_agents_git_shield_test_go_mockValidator_ValidateCommand
    file_internal_agents_git_shield_test_go -->|contains| func_internal_agents_git_shield_test_go_TestGitShield_CreateWorktree
    file_internal_agents_loader_go -->|imports| import_internal_agents_loader_go_bufio
    file_internal_agents_loader_go -->|imports| import_internal_agents_loader_go_fmt
    file_internal_agents_loader_go -->|imports| import_internal_agents_loader_go_os
    file_internal_agents_loader_go -->|imports| import_internal_agents_loader_go_strings
    file_internal_agents_loader_go -->|imports| import_internal_agents_loader_go_github_com_go_playground_validator_v10
    file_internal_agents_loader_go -->|imports| import_internal_agents_loader_go_gopkg_in_yaml_v3
    file_internal_agents_loader_go -->|contains| struct_internal_agents_loader_go_Loader
    file_internal_agents_loader_go -->|contains| func_internal_agents_loader_go_NewLoader
    file_internal_agents_loader_go -->|contains| func_internal_agents_loader_go_Loader_LoadAgent
    file_internal_agents_mutation_go -->|imports| import_internal_agents_mutation_go_bufio
    file_internal_agents_mutation_go -->|imports| import_internal_agents_mutation_go_context
    file_internal_agents_mutation_go -->|imports| import_internal_agents_mutation_go_fmt
    file_internal_agents_mutation_go -->|imports| import_internal_agents_mutation_go_io
    file_internal_agents_mutation_go -->|imports| import_internal_agents_mutation_go_os
    file_internal_agents_mutation_go -->|imports| import_internal_agents_mutation_go_path_filepath
    file_internal_agents_mutation_go -->|imports| import_internal_agents_mutation_go_regexp
    file_internal_agents_mutation_go -->|contains| struct_internal_agents_mutation_go_MutationEngine
    file_internal_agents_mutation_go -->|contains| func_internal_agents_mutation_go_NewMutationEngine
    file_internal_agents_mutation_go -->|contains| func_internal_agents_mutation_go_MutationEngine_Mutate
    file_internal_agents_mutation_go -->|contains| func_internal_agents_mutation_go_MutationEngine_Rollback
    file_internal_agents_registry_go -->|imports| import_internal_agents_registry_go_context
    file_internal_agents_registry_go -->|imports| import_internal_agents_registry_go_encoding_json
    file_internal_agents_registry_go -->|imports| import_internal_agents_registry_go_fmt
    file_internal_agents_registry_go -->|imports| import_internal_agents_registry_go_strings
    file_internal_agents_registry_go -->|contains| struct_internal_agents_registry_go_RegistryManager
    file_internal_agents_registry_go -->|contains| func_internal_agents_registry_go_NewRegistryManager
    file_internal_agents_registry_go -->|contains| func_internal_agents_registry_go_RegistryManager_SelectBest
    file_internal_agents_registry_go -->|contains| func_internal_agents_registry_go_RegistryManager_matchesAll
    file_internal_agents_registry_test_go -->|imports| import_internal_agents_registry_test_go_context
    file_internal_agents_registry_test_go -->|imports| import_internal_agents_registry_test_go_database_sql
    file_internal_agents_registry_test_go -->|imports| import_internal_agents_registry_test_go_encoding_json
    file_internal_agents_registry_test_go -->|imports| import_internal_agents_registry_test_go_os
    file_internal_agents_registry_test_go -->|imports| import_internal_agents_registry_test_go_testing
    file_internal_agents_registry_test_go -->|imports| import_internal_agents_registry_test_go_modernc_org_sqlite
    file_internal_agents_registry_test_go -->|contains| func_internal_agents_registry_test_go_TestRegistryManager_SelectBest
    file_internal_agents_types_go -->|imports| import_internal_agents_types_go_context
    file_internal_agents_types_go -->|imports| import_internal_agents_types_go_sync
    file_internal_agents_types_go -->|contains| struct_internal_agents_types_go_Specialist
    file_internal_agents_types_go -->|contains| struct_internal_agents_types_go_TokenBudget
    file_internal_agents_types_go -->|contains| func_internal_agents_types_go_TokenBudget_AddTokens
    file_internal_agents_types_go -->|contains| func_internal_agents_types_go_TokenBudget_IncSteps
    file_internal_agents_types_go -->|contains| struct_internal_agents_types_go_AgentDefinition
    file_internal_agents_types_go -->|contains| struct_internal_agents_types_go_Message
    file_internal_agents_types_go -->|contains| struct_internal_agents_types_go_AgentContext
    file_internal_agents_types_go -->|contains| struct_internal_agents_types_go_SubTask
    file_internal_agents_types_go -->|contains| func_internal_agents_types_go_NewAgentContext
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_bufio
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_bytes
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_context
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_encoding_json
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_fmt
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_os
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_os_exec
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_path_filepath
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_regexp
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_strings
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_github_com_google_generative_ai_go_genai
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_github_com_google_shlex
    file_internal_agents_tools_go -->|imports| import_internal_agents_tools_go_github_com_google_uuid
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_ReadFileTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReadFileTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReadFileTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReadFileTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReadFileTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReadFileTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_WriteFileTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_WriteFileTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_WriteFileTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_WriteFileTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_WriteFileTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_WriteFileTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_ReplaceTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReplaceTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReplaceTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReplaceTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReplaceTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ReplaceTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_GrepSearchTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_GrepSearchTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_GrepSearchTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_GrepSearchTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_GrepSearchTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_GrepSearchTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_AuditTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_AuditTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_AuditTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_AuditTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_AuditTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_AuditTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_RunTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_RunTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_RunTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_RunTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_RunTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_RunTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_ADRTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ADRTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ADRTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ADRTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ADRTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ADRTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_ScanTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ScanTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ScanTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ScanTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ScanTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_ScanTool_Execute
    file_internal_agents_tools_go -->|contains| struct_internal_agents_tools_go_DecomposeTool
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_DecomposeTool_Name
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_DecomposeTool_Description
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_DecomposeTool_Definition
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_DecomposeTool_ValidateArguments
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_DecomposeTool_Execute
    file_internal_agents_tools_go -->|contains| func_internal_agents_tools_go_RegisterCoreTools
    file_internal_audit_runner_go -->|imports| import_internal_audit_runner_go_bytes
    file_internal_audit_runner_go -->|imports| import_internal_audit_runner_go_context
    file_internal_audit_runner_go -->|imports| import_internal_audit_runner_go_errors
    file_internal_audit_runner_go -->|imports| import_internal_audit_runner_go_fmt
    file_internal_audit_runner_go -->|imports| import_internal_audit_runner_go_os_exec
    file_internal_audit_runner_go -->|imports| import_internal_audit_runner_go_time
    file_internal_audit_runner_go -->|imports| import_internal_audit_runner_go_github_com_google_shlex
    file_internal_audit_runner_go -->|contains| struct_internal_audit_runner_go_Runner
    file_internal_audit_runner_go -->|contains| func_internal_audit_runner_go_NewRunner
    file_internal_audit_runner_go -->|contains| func_internal_audit_runner_go_Runner_ExecuteAudit
    file_internal_bridge_prompt_factory_go -->|imports| import_internal_bridge_prompt_factory_go_bufio
    file_internal_bridge_prompt_factory_go -->|imports| import_internal_bridge_prompt_factory_go_fmt
    file_internal_bridge_prompt_factory_go -->|imports| import_internal_bridge_prompt_factory_go_os
    file_internal_bridge_prompt_factory_go -->|imports| import_internal_bridge_prompt_factory_go_strings
    file_internal_bridge_prompt_factory_go -->|imports| import_internal_bridge_prompt_factory_go_text_template
    file_internal_bridge_prompt_factory_go -->|contains| struct_internal_bridge_prompt_factory_go_ADR
    file_internal_bridge_prompt_factory_go -->|contains| struct_internal_bridge_prompt_factory_go_ContextNode
    file_internal_bridge_prompt_factory_go -->|contains| struct_internal_bridge_prompt_factory_go_ContextPayload
    file_internal_bridge_prompt_factory_go -->|contains| struct_internal_bridge_prompt_factory_go_Factory
    file_internal_bridge_prompt_factory_go -->|contains| func_internal_bridge_prompt_factory_go_NewFactory
    file_internal_bridge_prompt_factory_go -->|contains| func_internal_bridge_prompt_factory_go_Factory_GeneratePayload
    file_internal_bridge_prompt_factory_go -->|contains| struct_internal_bridge_prompt_factory_go_SystemData
    file_internal_bridge_prompt_factory_go -->|contains| func_internal_bridge_prompt_factory_go_Factory_loadADRs
    file_internal_bridge_prompt_factory_go -->|contains| func_internal_bridge_prompt_factory_go_Factory_loadSurgicalContext
    file_internal_bridge_prompt_factory_go -->|contains| func_internal_bridge_prompt_factory_go_extractLines
    file_internal_graph_adr_generator_go -->|imports| import_internal_graph_adr_generator_go_fmt
    file_internal_graph_adr_generator_go -->|imports| import_internal_graph_adr_generator_go_os
    file_internal_graph_adr_generator_go -->|imports| import_internal_graph_adr_generator_go_path_filepath
    file_internal_graph_adr_generator_go -->|imports| import_internal_graph_adr_generator_go_time
    file_internal_graph_adr_generator_go -->|contains| struct_internal_graph_adr_generator_go_ADRGenerator
    file_internal_graph_adr_generator_go -->|contains| func_internal_graph_adr_generator_go_NewADRGenerator
    file_internal_graph_adr_generator_go -->|contains| struct_internal_graph_adr_generator_go_ADRData
    file_internal_graph_adr_generator_go -->|contains| func_internal_graph_adr_generator_go_ADRGenerator_Generate
    file_internal_graph_adr_generator_test_go -->|imports| import_internal_graph_adr_generator_test_go_io
    file_internal_graph_adr_generator_test_go -->|imports| import_internal_graph_adr_generator_test_go_os
    file_internal_graph_adr_generator_test_go -->|imports| import_internal_graph_adr_generator_test_go_strings
    file_internal_graph_adr_generator_test_go -->|imports| import_internal_graph_adr_generator_test_go_testing
    file_internal_graph_adr_generator_test_go -->|contains| func_internal_graph_adr_generator_test_go_TestADRGenerator_Generate
    file_internal_graph_engine_go -->|imports| import_internal_graph_engine_go_database_sql
    file_internal_graph_engine_go -->|imports| import_internal_graph_engine_go_fmt
    file_internal_graph_engine_go -->|imports| import_internal_graph_engine_go_os
    file_internal_graph_engine_go -->|imports| import_internal_graph_engine_go_path_filepath
    file_internal_graph_engine_go -->|imports| import_internal_graph_engine_go_sync
    file_internal_graph_engine_go -->|contains| struct_internal_graph_engine_go_Engine
    file_internal_graph_engine_go -->|contains| func_internal_graph_engine_go_NewEngine
    file_internal_graph_engine_go -->|contains| func_internal_graph_engine_go_Engine_RegisterScanner
    file_internal_graph_engine_go -->|contains| func_internal_graph_engine_go_Engine_ScanProject
    file_internal_graph_engine_go -->|contains| func_internal_graph_engine_go_Engine_scanFileWithIncrementalCheck
    file_internal_graph_engine_go -->|contains| func_internal_graph_engine_go_Engine_persistResult
    file_internal_graph_engine_go -->|contains| func_internal_graph_engine_go_Engine_upsertNodeTx
    file_internal_graph_engine_go -->|contains| func_internal_graph_engine_go_Engine_createEdgeTx
    file_internal_graph_linker_go -->|imports| import_internal_graph_linker_go_fmt
    file_internal_graph_linker_go -->|imports| import_internal_graph_linker_go_os
    file_internal_graph_linker_go -->|imports| import_internal_graph_linker_go_path_filepath
    file_internal_graph_linker_go -->|imports| import_internal_graph_linker_go_strings
    file_internal_graph_linker_go -->|contains| func_internal_graph_linker_go_Engine_LinkDependencies
    file_internal_graph_linker_go -->|contains| struct_internal_graph_linker_go_pendingLink
    file_internal_graph_linker_go -->|contains| func_internal_graph_linker_go_Engine_resolveImport
    file_internal_graph_linker_go -->|contains| func_internal_graph_linker_go_Engine_createRealEdge
    file_internal_graph_scanner_go_go -->|imports| import_internal_graph_scanner_go_go_fmt
    file_internal_graph_scanner_go_go -->|imports| import_internal_graph_scanner_go_go_go_ast
    file_internal_graph_scanner_go_go -->|imports| import_internal_graph_scanner_go_go_go_parser
    file_internal_graph_scanner_go_go -->|imports| import_internal_graph_scanner_go_go_go_token
    file_internal_graph_scanner_go_go -->|imports| import_internal_graph_scanner_go_go_path_filepath
    file_internal_graph_scanner_go_go -->|imports| import_internal_graph_scanner_go_go_strings
    file_internal_graph_scanner_go_go -->|contains| struct_internal_graph_scanner_go_go_GoScanner
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_NewGoScanner
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_GoScanner_SupportedExtensions
    file_internal_graph_scanner_go_go -->|contains| func_internal_graph_scanner_go_go_GoScanner_Scan
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_context
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_fmt
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_io
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_os
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_path_filepath
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_strings
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_sync
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_github_com_smacker_go_tree_sitter
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_github_com_smacker_go_tree_sitter_typescript_tsx
    file_internal_graph_scanner_treesitter_go -->|imports| import_internal_graph_scanner_treesitter_go_github_com_smacker_go_tree_sitter_typescript_typescript
    file_internal_graph_scanner_treesitter_go -->|contains| struct_internal_graph_scanner_treesitter_go_TreeSitterScanner
    file_internal_graph_scanner_treesitter_go -->|contains| func_internal_graph_scanner_treesitter_go_NewTreeSitterScanner
    file_internal_graph_scanner_treesitter_go -->|contains| func_internal_graph_scanner_treesitter_go_TreeSitterScanner_SupportedExtensions
    file_internal_graph_scanner_treesitter_go -->|contains| func_internal_graph_scanner_treesitter_go_TreeSitterScanner_Scan
    file_internal_graph_scanner_treesitter_go -->|contains| func_internal_graph_scanner_treesitter_go_TreeSitterScanner_processSymbol
    file_internal_graph_schema_go -->|imports| import_internal_graph_schema_go_fmt
    file_internal_graph_schema_go -->|contains| func_internal_graph_schema_go_Migrate
    file_internal_graph_schema_test_go -->|imports| import_internal_graph_schema_test_go_database_sql
    file_internal_graph_schema_test_go -->|imports| import_internal_graph_schema_test_go_os
    file_internal_graph_schema_test_go -->|imports| import_internal_graph_schema_test_go_path_filepath
    file_internal_graph_schema_test_go -->|imports| import_internal_graph_schema_test_go_testing
    file_internal_graph_schema_test_go -->|imports| import_internal_graph_schema_test_go_modernc_org_sqlite
    file_internal_graph_schema_test_go -->|contains| func_internal_graph_schema_test_go_TestMigrate
    file_internal_graph_types_go -->|imports| import_internal_graph_types_go_time
    file_internal_graph_types_go -->|contains| struct_internal_graph_types_go_Node
    file_internal_graph_types_go -->|contains| struct_internal_graph_types_go_Edge
    file_internal_graph_types_go -->|contains| struct_internal_graph_types_go_ScanResult
    file_internal_graph_visualizer_go -->|imports| import_internal_graph_visualizer_go_fmt
    file_internal_graph_visualizer_go -->|imports| import_internal_graph_visualizer_go_os
    file_internal_graph_visualizer_go -->|imports| import_internal_graph_visualizer_go_path_filepath
    file_internal_graph_visualizer_go -->|imports| import_internal_graph_visualizer_go_strings
    file_internal_graph_visualizer_go -->|contains| struct_internal_graph_visualizer_go_Visualizer
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_NewVisualizer
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_Visualizer_GenerateMasterDiagram
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_Visualizer_GenerateTaskSnapshot
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_Visualizer_getNodes
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_Visualizer_GenerateC4ContainerDiagram
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_Visualizer_formatC4Mermaid
    file_internal_graph_visualizer_go -->|contains| struct_internal_graph_visualizer_go_container
    file_internal_graph_visualizer_go -->|contains| struct_internal_graph_visualizer_go_relKey
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_Visualizer_formatMermaid
    file_internal_graph_visualizer_go -->|contains| func_internal_graph_visualizer_go_Visualizer_getEdges
    file_internal_reflect_validator_go -->|imports| import_internal_reflect_validator_go_bufio
    file_internal_reflect_validator_go -->|imports| import_internal_reflect_validator_go_fmt
    file_internal_reflect_validator_go -->|imports| import_internal_reflect_validator_go_os
    file_internal_reflect_validator_go -->|imports| import_internal_reflect_validator_go_path_filepath
    file_internal_reflect_validator_go -->|imports| import_internal_reflect_validator_go_strings
    file_internal_reflect_validator_go -->|contains| struct_internal_reflect_validator_go_Violation
    file_internal_reflect_validator_go -->|contains| struct_internal_reflect_validator_go_Validator
    file_internal_reflect_validator_go -->|contains| func_internal_reflect_validator_go_NewValidator
    file_internal_reflect_validator_go -->|contains| func_internal_reflect_validator_go_Validator_ValidateProject
    file_internal_reflect_validator_go -->|contains| func_internal_reflect_validator_go_Validator_ValidatePath
    file_internal_reflect_validator_go -->|contains| func_internal_reflect_validator_go_Validator_ValidateCommand
    file_internal_reflect_validator_go -->|contains| func_internal_reflect_validator_go_Validator_checkFile
    file_internal_reflect_validator_go -->|contains| func_internal_reflect_validator_go_isIgnored
    file_internal_registry_commands_go -->|imports| import_internal_registry_commands_go_github_com_spf13_cobra
    file_internal_registry_commands_go -->|contains| func_internal_registry_commands_go_Register
    file_internal_registry_commands_go -->|contains| func_internal_registry_commands_go_GetCommands
    file_internal_report_aggregator_go -->|imports| import_internal_report_aggregator_go_fmt
    file_internal_report_aggregator_go -->|imports| import_internal_report_aggregator_go_os
    file_internal_report_aggregator_go -->|imports| import_internal_report_aggregator_go_path_filepath
    file_internal_report_aggregator_go -->|contains| struct_internal_report_aggregator_go_TaskInfo
    file_internal_report_aggregator_go -->|contains| struct_internal_report_aggregator_go_ProjectStats
    file_internal_report_aggregator_go -->|contains| struct_internal_report_aggregator_go_Aggregator
    file_internal_report_aggregator_go -->|contains| func_internal_report_aggregator_go_NewAggregator
    file_internal_report_aggregator_go -->|contains| func_internal_report_aggregator_go_Aggregator_FetchStats
    file_internal_report_aggregator_go -->|contains| func_internal_report_aggregator_go_Aggregator_GenerateMarkdown
    file_internal_state_manager_go -->|imports| import_internal_state_manager_go_fmt
    file_internal_state_manager_go -->|imports| import_internal_state_manager_go_time
    file_internal_state_manager_go -->|imports| import_internal_state_manager_go_github_com_google_uuid
    file_internal_state_manager_go -->|contains| struct_internal_state_manager_go_Task
    file_internal_state_manager_go -->|contains| struct_internal_state_manager_go_Manager
    file_internal_state_manager_go -->|contains| func_internal_state_manager_go_NewManager
    file_internal_state_manager_go -->|contains| func_internal_state_manager_go_Manager_CreateTask
    file_internal_state_manager_go -->|contains| func_internal_state_manager_go_Manager_StartTask
    file_internal_state_manager_go -->|contains| func_internal_state_manager_go_Manager_GetTaskByID
    file_internal_state_manager_go -->|contains| func_internal_state_manager_go_Manager_UpdateStatus
    file_internal_state_manager_go -->|contains| func_internal_state_manager_go_Manager_GetActiveTask
    file_internal_state_manager_go -->|contains| func_internal_state_manager_go_Manager_ListTasks
    file_legacy_ts_src_brainstorm_ts -->|imports| import_legacy_ts_src_brainstorm_ts_fs
    file_legacy_ts_src_brainstorm_ts -->|imports| import_legacy_ts_src_brainstorm_ts_path
    file_legacy_ts_src_brainstorm_ts -->|imports| import_legacy_ts_src_brainstorm_ts_inquirer
    file_legacy_ts_src_brainstorm_ts -->|contains| function_legacy_ts_src_brainstorm_ts_runBrainstorm
    file_legacy_ts_src_brainstorm_ts -->|contains| function_legacy_ts_src_brainstorm_ts_config
    file_legacy_ts_src_brainstorm_ts -->|contains| component_legacy_ts_src_brainstorm_ts_INSIGHTS_DIR
    file_legacy_ts_src_brainstorm_ts -->|contains| function_legacy_ts_src_brainstorm_ts_theme
    file_legacy_ts_src_brainstorm_ts -->|contains| function_legacy_ts_src_brainstorm_ts_sanitizedTheme
    file_legacy_ts_src_brainstorm_ts -->|contains| function_legacy_ts_src_brainstorm_ts_reportFile
    file_legacy_ts_src_brainstorm_ts -->|contains| function_legacy_ts_src_brainstorm_ts_reportPath
    file_legacy_ts_src_brainstorm_ts -->|contains| function_legacy_ts_src_brainstorm_ts_skeleton
    file_legacy_ts_src_config_loader_ts -->|imports| import_legacy_ts_src_config_loader_ts_fs
    file_legacy_ts_src_config_loader_ts -->|imports| import_legacy_ts_src_config_loader_ts_path
    file_legacy_ts_src_config_loader_ts -->|contains| interface_legacy_ts_src_config_loader_ts_SentinelConfig
    file_legacy_ts_src_config_loader_ts -->|contains| component_legacy_ts_src_config_loader_ts_CONFIG_FILENAME
    file_legacy_ts_src_config_loader_ts -->|contains| function_legacy_ts_src_config_loader_ts_loadConfig
    file_legacy_ts_src_config_loader_ts -->|contains| function_legacy_ts_src_config_loader_ts_configPath
    file_legacy_ts_src_config_loader_ts -->|contains| function_legacy_ts_src_config_loader_ts_raw
    file_legacy_ts_src_cli_ts -->|imports| import_legacy_ts_src_cli_ts_child_process
    file_legacy_ts_src_cli_ts -->|imports| import_legacy_ts_src_cli_ts_fs
    file_legacy_ts_src_cli_ts -->|imports| import_legacy_ts_src_cli_ts_path
    file_legacy_ts_src_cli_ts -->|imports| import_legacy_ts_src_cli_ts_inquirer
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_plan
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_config
    file_legacy_ts_src_cli_ts -->|contains| component_legacy_ts_src_cli_ts_PLAN_PATH
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_skeleton
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_forge
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_insightsDir
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_files
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_audit
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_verifyPlan
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_content
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_markers
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_command
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_main
    file_legacy_ts_src_cli_ts -->|contains| function_legacy_ts_src_cli_ts_c
    file_legacy_ts_src_forge_engine_ts -->|imports| import_legacy_ts_src_forge_engine_ts_fs
    file_legacy_ts_src_forge_engine_ts -->|imports| import_legacy_ts_src_forge_engine_ts_path
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_getAllFiles
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_files
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_forgePlan
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_config
    file_legacy_ts_src_forge_engine_ts -->|contains| component_legacy_ts_src_forge_engine_ts_INSIGHTS_DIR
    file_legacy_ts_src_forge_engine_ts -->|contains| component_legacy_ts_src_forge_engine_ts_PLAN_PATH
    file_legacy_ts_src_forge_engine_ts -->|contains| component_legacy_ts_src_forge_engine_ts_FPA_RULES_PATH
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_insightPath
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_content
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_pathDivider
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_nextPathDivider
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_splitParts
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_splitStart
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_chunk
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_lines
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_pathName
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_visionLine
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_vision
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_impactFiles
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_keywords
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_allFilesToScan
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_fileContent
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_totalFP
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_criticalPathEntries
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_relativePath
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_criticalMatch
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_tier
    file_legacy_ts_src_forge_engine_ts -->|contains| function_legacy_ts_src_forge_engine_ts_plan
    file_pkg_sqlite_db_go -->|imports| import_pkg_sqlite_db_go_database_sql
    file_pkg_sqlite_db_go -->|imports| import_pkg_sqlite_db_go_fmt
    file_pkg_sqlite_db_go -->|imports| import_pkg_sqlite_db_go_os
    file_pkg_sqlite_db_go -->|imports| import_pkg_sqlite_db_go_path_filepath
    file_pkg_sqlite_db_go -->|imports| import_pkg_sqlite_db_go_modernc_org_sqlite
    file_pkg_sqlite_db_go -->|contains| struct_pkg_sqlite_db_go_DB
    file_pkg_sqlite_db_go -->|contains| func_pkg_sqlite_db_go_Init
    file_pkg_sqlite_db_go -->|contains| func_pkg_sqlite_db_go_DB_Close
    file_pkg_utils_hash_go -->|imports| import_pkg_utils_hash_go_crypto_sha256
    file_pkg_utils_hash_go -->|imports| import_pkg_utils_hash_go_fmt
    file_pkg_utils_hash_go -->|imports| import_pkg_utils_hash_go_io
    file_pkg_utils_hash_go -->|imports| import_pkg_utils_hash_go_os
    file_pkg_utils_hash_go -->|contains| func_pkg_utils_hash_go_CalculateHash
    file_pkg_utils_text_go -->|imports| import_pkg_utils_text_go_regexp
    file_pkg_utils_text_go -->|imports| import_pkg_utils_text_go_strings
    file_pkg_utils_text_go -->|contains| func_pkg_utils_text_go_SanitizeID
    file_pkg_utils_text_go -->|contains| func_pkg_utils_text_go_Slugify
    file_pkg_utils_text_go -->|contains| func_pkg_utils_text_go_EscapeYAML
    file_pkg_utils_filter_go -->|imports| import_pkg_utils_filter_go_bufio
    file_pkg_utils_filter_go -->|imports| import_pkg_utils_filter_go_os
    file_pkg_utils_filter_go -->|imports| import_pkg_utils_filter_go_path_filepath
    file_pkg_utils_filter_go -->|imports| import_pkg_utils_filter_go_strings
    file_pkg_utils_filter_go -->|contains| struct_pkg_utils_filter_go_IgnoreFilter
    file_pkg_utils_filter_go -->|contains| func_pkg_utils_filter_go_NewIgnoreFilter
    file_pkg_utils_filter_go -->|contains| func_pkg_utils_filter_go_IgnoreFilter_loadGitignore
    file_pkg_utils_filter_go -->|contains| func_pkg_utils_filter_go_IgnoreFilter_IsIgnored
    file_legacy_ts_src_brainstorm_ts -->|imports| file_legacy_ts_src_config_loader_ts
    file_legacy_ts_src_cli_ts -->|imports| file_legacy_ts_src_brainstorm_ts
    file_legacy_ts_src_cli_ts -->|imports| file_legacy_ts_src_forge_engine_ts
    file_legacy_ts_src_cli_ts -->|imports| file_legacy_ts_src_config_loader_ts
    file_legacy_ts_src_forge_engine_ts -->|imports| file_legacy_ts_src_config_loader_ts

    classDef struct fill:#f9f,stroke:#333,stroke-width:2px;
    classDef func fill:#bbf,stroke:#333,stroke-width:1px;
```
