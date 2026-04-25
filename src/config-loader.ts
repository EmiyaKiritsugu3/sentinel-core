import fs from 'fs';
import path from 'path';

export interface SentinelConfig {
  planPath: string;
  preFlightCommand: string;
  paths: {
    insightsDir: string;
    sourceDirs: string[];
    fpaRules?: string;
    baseline: string;
    currentState: string;
  };
  fpa: {
    criticalPaths: Record<string, number>;
    defaultFilePoints: number;
  };
}

const CONFIG_FILENAME = 'sentinel.config.json';

export function loadConfig(): SentinelConfig {
  const configPath = path.join(process.cwd(), CONFIG_FILENAME);
  
  if (!fs.existsSync(configPath)) {
    console.error('❌ Configuration file not found: ' + CONFIG_FILENAME);
    console.log('💡 Please create a ' + CONFIG_FILENAME + ' in your project root.');
    process.exit(1);
  }

  try {
    const raw = fs.readFileSync(configPath, 'utf8');
    return JSON.parse(raw) as SentinelConfig;
  } catch (error) {
    console.error('❌ Failed to parse ' + CONFIG_FILENAME + ': ' + (error instanceof Error ? error.message : 'Unknown error'));
    process.exit(1);
  }
}
