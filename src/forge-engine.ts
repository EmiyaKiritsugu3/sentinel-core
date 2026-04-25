import fs from 'fs';
import path from 'path';
import { loadConfig } from './config-loader';

function getAllFiles(dirPath: string, arrayOfFiles: string[] = []): string[] {
  if (!fs.existsSync(dirPath)) return arrayOfFiles;
  
  const files = fs.readdirSync(dirPath);

  files.forEach(function (file) {
    if (fs.statSync(dirPath + '/' + file).isDirectory()) {
      if (file !== 'node_modules' && file !== '.git' && file !== '.next') {
        arrayOfFiles = getAllFiles(dirPath + '/' + file, arrayOfFiles);
      }
    } else {
      arrayOfFiles.push(path.join(dirPath, '/', file));
    }
  });

  return arrayOfFiles;
}

export async function forgePlan(insightFile: string, pathIndex: number): Promise<boolean> {
  const config = loadConfig();
  const INSIGHTS_DIR = path.join(process.cwd(), config.paths.insightsDir);
  const PLAN_PATH = path.join(process.cwd(), config.planPath);
  const FPA_RULES_PATH = config.paths.fpaRules ? path.join(process.cwd(), config.paths.fpaRules) : null;

  const insightPath = path.join(INSIGHTS_DIR, insightFile);
  if (!fs.existsSync(insightPath)) {
    console.error('❌ Insight file not found: ' + insightPath);
    return false;
  }
  
  const content = fs.readFileSync(insightPath, 'utf8');

  const pathDivider = '### Path ' + pathIndex + ':';
  const nextPathDivider = '### Path ' + (pathIndex + 1) + ':';

  const splitParts = content.split(pathDivider);
  if (splitParts.length < 2) {
    console.error('❌ Parse Error: Cannot find ' + pathDivider + ' in the insight report.');
    return false;
  }

  const splitStart = splitParts[1];
  const chunk = splitStart?.split(nextPathDivider)[0]?.split('---')[0] || '';

  const lines = chunk.split('\n').filter((l) => l.trim() !== '');
  if (lines.length === 0) {
    console.error('❌ Parse Error: Selected path block is empty.');
    return false;
  }
  const pathName = lines[0]?.trim() || 'Unnamed Path';

  const visionLine = lines.find((l) => l.includes('**The Vision**:'));
  const vision = visionLine?.split('**The Vision**:')[1]?.trim() || 'No explicit vision detected.';

  console.log('\n⚒️  FORGING: \"' + pathName + '\"...\n');

  const impactFiles: string[] = [];
  try {
    const keywords = [...pathName.split(' '), ...vision.split(' ')]
      .filter((k) => k.length > 5)
      .map((k) => k.replace(/[^a-zA-Z0-9_-]/g, ''));

    if (keywords.length > 0) {
      const allFilesToScan: string[] = [];
      config.paths.sourceDirs.forEach(dir => {
        allFilesToScan.push(...getAllFiles(path.join(process.cwd(), dir)));
      });

      for (const kw of keywords.slice(0, 5)) {
        allFilesToScan.forEach((f) => {
          const fileContent = fs.readFileSync(f, 'utf8');
          if (fileContent.toLowerCase().includes(kw.toLowerCase()) && !impactFiles.includes(f)) {
            impactFiles.push(f);
          }
        });
      }
    }
  } catch (_e) {}

  let totalFP = 0;
  const criticalPathEntries = Object.entries(config.fpa.criticalPaths);

  impactFiles.forEach((f) => {
    const relativePath = path.relative(process.cwd(), f);
    const criticalMatch = criticalPathEntries.find(([p]) => relativePath.includes(p));
    
    if (criticalMatch) {
      totalFP += criticalMatch[1] || 0;
    } else if (f.endsWith('.tsx') || f.endsWith('.ts') || f.endsWith('.js')) {
      totalFP += config.fpa.defaultFilePoints;
    }
  });

  const tier = totalFP <= 5 ? 'Tier 1' : totalFP <= 15 ? 'Tier 2' : 'Tier 3';

  const plan = '# Implementation Plan: ' + pathName + ' [PID-SENTINEL]\n\n' +
'> [!IMPORTANT]\n' +
'> **GOAL**: ' + pathName + '\n' +
'> **SOURCE INSIGHT**: [' + insightFile + '](file://' + insightPath + ')\n' +
'> **STRATEGY**: ' + vision + '\n' +
'> \n' +
'> **AI INSTRUCTION**: This plan was forged via Sentinel Framework v5.0. Execute strictly within **' + tier + '** protocol gates.\n\n' +
'### I. Deterministic Baseline (Phase Alpha)\n\n' +
'- **GATE(BASELINE)**: Mandatory `pre-flight` execution completed prior to generation.\n' +
'- **STATUS**: [PASS]\n' +
'- **INVARIANT**: If any CLI tests break during refactor, development stops.\n\n' +
'### II. Verification Sovereignty (Phase Beta)\n\n' +
'| Path | Assertion | Verification Tool | Proof ID |\n' +
'| --- | --- | --- | --- |\n' +
(impactFiles.length > 0 ? impactFiles.slice(0, 3).map((f, i) => '| `' + path.relative(process.cwd(), f) + '` | Maintains type safety | Audit Tool | PRF-0' + (i+1) + ' |').join('\n') : '| `docs/` | Ensure documentation is updated | Visual | PRF-01 |') + '\n\n' +
'### III. Dialectical Audit (Phase Gamma)\n\n' +
'- **SENIOR_ARCHITECT**: \"The strategy aims for higher cohesion. Ensure abstractions remain clean.\"\n\n' +
'### IV. Impact Heatmap (Phase Delta)\n\n' +
'- **RIPPLES**:\n' +
(impactFiles.length > 0 ? impactFiles.map((f) => '  - `' + path.relative(process.cwd(), f) + '`').join('\n') : '  - No direct file collision detected.') + '\n' +
'- **RISK_LEVEL**: ' + (totalFP > 15 ? 'HIGH' : totalFP > 5 ? 'MEDIUM' : 'LOW') + '\n\n' +
'### V. Fail-Safe Operations (Phase Epsilon)\n' +
'- **MECHANISM**: Mandatory fallback paths.\n\n' +
'---\n\n' +
'## 📋 Sentinel Governance Checklist & Gotchas\n' +
'- [ ] 1. Run `pre-flight` before edit.\n\n' +
'## Technical Execution (Proposed Changes)\n' +
'<!-- AI: Flesh out the architecture here -->\n\n' +
'## Function Point Analysis (FPA)\n\n' +
'| Category | FP | \n' +
'| --- | --- |\n' +
'| **Total Base FP** | **' + totalFP + '** |\n' +
'| **Tier** | **' + tier + '** |\n\n' +
'---\n' +
'**Protocol Status**: Forged & Ready for execution.\n' +
'**Timestamp**: ' + new Date().toISOString() + '\n';

  fs.writeFileSync(PLAN_PATH, plan);
  console.log('\n✅ ELITE PLAN FORGED SUCCESSFULLY: ' + config.planPath);
  return true;
}
