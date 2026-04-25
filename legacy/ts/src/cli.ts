import { execSync, spawn } from 'child_process';
import fs from 'fs';
import path from 'path';
import inquirer from 'inquirer';
import { runBrainstorm } from './brainstorm';
import { forgePlan } from './forge-engine';
import { loadConfig } from './config-loader';

async function plan() {
  const config = loadConfig();
  const PLAN_PATH = path.join(process.cwd(), config.planPath);
  const { goal } = await inquirer.prompt([{ type: 'input', name: 'goal', message: 'Primary Objective:' }]);
  const skeleton = '# Implementation Plan: ' + goal + ' [PID-SENTINEL]\n\n' +
'## 🔍 Pre-Planning Tool Log & Deliberation (Step 0)\n' +
'<!-- AI: MUST RUN SEARCH TOOLS (Tool Audit) BEFORE PROCEEDING -->\n' +
'- [ ] `sequential-thinking` executed for strategy mapping.\n' +
'- [ ] `wiki-index.md` consulted for historical trace.\n\n' +
'## 📋 Sentinel Governance Checklist\n' +
'- [ ] 1. Run `pre-flight` before edit.\n' +
'- [ ] 2. Cite origin Spec in the plan.\n\n' +
'### I. Deterministic Baseline (Phase Alpha)\n' +
'### II. Verification Sovereignty (Phase Beta)\n' +
'### III. Dialectical Audit (Phase Gamma)\n' +
'### IV. Impact Heatmap (Phase Delta)\n' +
'### V. Fail-Safe Operations (Phase Epsilon)\n\n' +
'## 🎓 Epiphany Protocol (Closing Mandate)\n' +
'- [ ] Did a library bug occur? -> Log in `TECHNICAL-DEBT.md`.\n' +
'- [ ] Was a new project rule defined? -> Log in `sentinel-log.md`.\n' +
'- [ ] Did the AI behavior fail? -> Update global `GEMINI.md` via `save_memory`.\n\n' +
'## Technical Execution\n\n' +
'## Function Point Analysis\n';
  fs.writeFileSync(PLAN_PATH, skeleton);
  console.log('✅ SKELETON GENERATED WITH EPIPHANY PROTOCOL: ' + PLAN_PATH);
}

async function forge() {
  const config = loadConfig();
  const insightsDir = path.join(process.cwd(), config.paths.insightsDir);
  if (!fs.existsSync(insightsDir)) return console.error('❌ No insights found.');
  const files = fs.readdirSync(insightsDir).filter((f) => f.endsWith('.md'));
  const { insightFile } = await inquirer.prompt([{ type: 'list', name: 'insightFile', message: 'Select Insight:', choices: files }]);
  const { pathIndex } = await inquirer.prompt([{ type: 'input', name: 'pathIndex', message: 'Path index:', default: '1' }]);
  await forgePlan(insightFile, parseInt(pathIndex, 10));
}

async function audit() {
  const config = loadConfig();
  try {
    execSync(config.preFlightCommand, { stdio: 'inherit' });
    console.log('\n✅ BASELINE SECURE.');
  } catch {
    process.exit(1);
  }
}

async function verifyPlan() {
  const config = loadConfig();
  const PLAN_PATH = path.join(process.cwd(), config.planPath);
  if (!fs.existsSync(PLAN_PATH)) { console.error('❌ Plan not found.'); process.exit(1); }
  const content = fs.readFileSync(PLAN_PATH, 'utf8');
  const markers = ['[PID-SENTINEL]', 'Phase Alpha', 'Phase Beta', 'Phase Gamma', 'Phase Delta', 'Phase Epsilon', 'Audit', 'Analysis', 'Checklist', 'Epiphany Protocol'];
  for (const m of markers) { 
    if (!content.includes(m)) {
      console.error(`❌ Plan failed verification. Missing required marker: ${m}`);
      process.exit(1); 
    }
  }
  console.log('✅ PLAN VERIFIED & COMPLIANT WITH ADF v1.2.');
}

const command = process.argv[2];
async function main() {
  switch (command) {
    case 'plan': await plan(); break;
    case 'verify-plan': case 'v': await verifyPlan(); break;
    case 'brainstorm': case 'shout': await runBrainstorm(); break;
    case 'forge': await forge(); break;
    case 'audit': await audit(); break;
    case 'status': 
      const c = loadConfig();
      console.log('🛡️ SENTINEL v5.1.0-standalone\nConfig: ' + c.paths.baseline + '\nActive Protocols: ADF v1.2, Epiphany, Traceability'); 
      break;
    default: 
      console.log('Available commands: plan, verify-plan (v), brainstorm (shout), forge, audit, status');
      process.exit(1);
  }
}
main().catch(console.error);
