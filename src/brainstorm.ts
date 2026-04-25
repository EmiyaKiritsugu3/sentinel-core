import fs from 'fs';
import path from 'path';
import inquirer from 'inquirer';
import { loadConfig } from './config-loader';

export async function runBrainstorm() {
  const config = loadConfig();
  const INSIGHTS_DIR = path.join(process.cwd(), config.paths.insightsDir);

  console.log('\n🎨 SENTINEL: ACTIVATING INSIGHT ENGINE (SHOUT)...\n');

  if (!fs.existsSync(INSIGHTS_DIR)) {
    fs.mkdirSync(INSIGHTS_DIR, { recursive: true });
  }

  const { themeInput } = await inquirer.prompt([
    {
      type: 'input',
      name: 'themeInput',
      message: 'What is the Epicenter of the brainstorm? (Leave blank for Autopilot 360º Scan):',
    },
  ]);

  const theme = themeInput.trim() || 'Autopilot 360 Scan';
  const sanitizedTheme = theme
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');

  const reportFile = 'INSIGHT-' + new Date().toISOString().split('T')[0] + '-' + (sanitizedTheme || 'unnamed') + '.md';
  const reportPath = path.join(INSIGHTS_DIR, reportFile);

  const skeleton = '# Insight Report: ' + theme + '\n\n' +
'> [!TIP]\n' +
'> **Epicenter**: ' + theme + '\n' +
'> **Status**: [PENDING AI DIALECTIC]\n' +
'> \n' +
'> **AI INSTRUCTION**: Act as **The Muse** and **The Artisan**. \n' +
'> 1. Read `' + config.paths.currentState + '`.\n' +
'> 2. Propose 3 state-of-the-art directions.\n' +
'> 3. Assign an **Innovation Score (0-100)** to each.\n\n' +
'## 🌌 Context Radiance (360º Scan)\n' +
'<!-- AI: Analyze how this theme impacts security, data, and user experience -->\n\n' +
'## 🛠️ The 3 Paths (Brainstorm)\n\n' +
'### Path 1: [Name]\n' +
'- **Innovation Score**: \n' +
'- **The Vision**: \n\n' +
'### Path 2: [Name]\n' +
'- ...\n\n' +
'---\n' +
'**Timestamp**: ' + new Date().toISOString() + '\n';

  fs.writeFileSync(reportPath, skeleton);
  console.log('\n✅ BRAINSTORM SKELETON GENERATED: ' + path.relative(process.cwd(), reportPath));
  console.log('🤖 AI is now primed to start the ideation loop here.\n');
}
