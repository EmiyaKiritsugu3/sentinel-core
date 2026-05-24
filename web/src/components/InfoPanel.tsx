import { useEffect, useState } from 'react';
import './InfoPanel.css';

/* ── Types ─────────────────────────────────────── */

interface NodeInfo {
  id: string;
  label: string;
  type: string;
  file_path?: string;
  start_line?: number;
  end_line?: number;
}

interface InfoPanelProps {
  node: NodeInfo;
  baseUrl: string;
  onClose: () => void;
}

interface ADRInfo {
  id: string;
  title: string;
  filename: string;
}

interface ADRContent {
  id: string;
  title: string;
  content: string;
  filename: string;
}

/* ── Component ─────────────────────────────────── */

type FetchState<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'ok'; data: T }
  | { status: 'error'; message: string };

export function InfoPanel({ node, baseUrl, onClose }: InfoPanelProps) {
  const [codeState, setCodeState] = useState<FetchState<string[]>>({ status: 'idle' });
  const [adrListState, setAdrListState] = useState<FetchState<ADRInfo[]>>({ status: 'idle' });
  const [selectedAdr, setSelectedAdr] = useState<ADRContent | null>(null);

  const canFetchCode = node.file_path && node.start_line != null && node.end_line != null;

  /* Fetch code snippet */
  useEffect(() => {
    if (!canFetchCode) return;
    setCodeState({ status: 'loading' });

    const params = new URLSearchParams({
      path: node.file_path!,
      start: String(node.start_line!),
      end: String(node.end_line!),
    });

    fetch(`${baseUrl}/api/code?${params}`)
      .then(async (res) => {
        if (!res.ok) {
          const text = await res.text().catch(() => '');
          throw new Error(`HTTP ${res.status}${text ? ': ' + text : ''}`);
        }
        return res.json();
      })
      .then((data: { lines: string[] }) => {
        setCodeState({ status: 'ok', data: data.lines });
      })
      .catch((err: Error) => {
        setCodeState({ status: 'error', message: err.message });
      });
  }, [node.file_path, node.start_line, node.end_line, baseUrl, canFetchCode]);

  /* Fetch ADR list */
  useEffect(() => {
    setAdrListState({ status: 'loading' });

    fetch(`${baseUrl}/api/adr`)
      .then(async (res) => {
        if (!res.ok) {
          const text = await res.text().catch(() => '');
          throw new Error(`HTTP ${res.status}${text ? ': ' + text : ''}`);
        }
        return res.json();
      })
      .then((data: { adrs: ADRInfo[] }) => {
        setAdrListState({ status: 'ok', data: data.adrs });
      })
      .catch((err: Error) => {
        setAdrListState({ status: 'error', message: err.message });
      });
  }, [baseUrl]);

  /* ADR filtering */
  const relatedAdrs: ADRInfo[] =
    adrListState.status === 'ok'
      ? filterAdrs(adrListState.data, node)
      : [];

  /* ADR content fetch */
  const handleAdrClick = (adr: ADRInfo) => {
    setSelectedAdr(null);
    fetch(`${baseUrl}/api/adr/${encodeURIComponent(adr.filename)}`)
      .then(async (res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data: ADRContent = await res.json();
        setSelectedAdr(data);
      })
      .catch((err: Error) => {
        setSelectedAdr({
          id: adr.id,
          title: adr.title,
          filename: adr.filename,
          content: `Error loading ADR: ${err.message}`,
        });
      });
  };

  return (
    <div className="info-panel">
      {/* Header */}
      <div className="info-panel__header">
        <span className="info-panel__node-id">{node.id}</span>
        <button className="info-panel__close" onClick={onClose} title="Close">
          X
        </button>
      </div>

      {/* Meta */}
      <div className="info-panel__meta">
        <span className="info-panel__tag">{node.type}</span>
        {node.file_path && (
          <span className="info-panel__file" title={node.file_path}>
            {node.file_path}
          </span>
        )}
        {node.start_line != null && node.end_line != null && (
          <span className="info-panel__lines">
            L{node.start_line}–{node.end_line}
          </span>
        )}
      </div>

      {/* Code section */}
      <div className="info-panel__section">
        <div className="info-panel__section-title">Code</div>
        {renderCodeSection()}
      </div>

      {/* ADR section */}
      <div className="info-panel__section">
        <div className="info-panel__section-title">Related ADRs</div>
        {renderAdrSection()}
      </div>
    </div>
  );

  /* ── Render helpers ────────────────────────── */

  function renderCodeSection() {
    if (!node.file_path) {
      return <div className="info-panel__empty">No code data available for this node.</div>;
    }
    if (!canFetchCode) {
      return (
        <div className="info-panel__code-block">
          <pre className="info-panel__code">
            <code>{node.file_path}</code>
          </pre>
          <div className="info-panel__empty">Line range not available.</div>
        </div>
      );
    }

    switch (codeState.status) {
      case 'idle':
      case 'loading':
        return <div className="info-panel__loading">Loading code...</div>;
      case 'error':
        return <div className="info-panel__error">{codeState.message}</div>;
      case 'ok': {
        const lineDigits = String(node.end_line).length;
        return (
          <div className="info-panel__code-block">
            <pre className="info-panel__code">
              {codeState.data.map((line, i) => (
                <span key={i} className="info-panel__code-line">
                  <span className="info-panel__line-num">
                    {String(node.start_line! + i).padStart(lineDigits, ' ')}
                  </span>
                  <span className="info-panel__line-text">{line}</span>
                </span>
              ))}
            </pre>
          </div>
        );
      }
    }
  }

  function renderAdrSection() {
    switch (adrListState.status) {
      case 'idle':
      case 'loading':
        return <div className="info-panel__loading">Loading ADRs...</div>;
      case 'error':
        return <div className="info-panel__error">{adrListState.message}</div>;
      case 'ok':
        if (relatedAdrs.length === 0) {
          return <div className="info-panel__empty">No related ADRs found.</div>;
        }
        return (
          <div>
            <ul className="info-panel__adr-list">
              {relatedAdrs.map((adr) => (
                <li key={adr.id}>
                  <button
                    className="info-panel__adr-link"
                    onClick={() => handleAdrClick(adr)}
                  >
                    {adr.title || adr.filename}
                  </button>
                </li>
              ))}
            </ul>
            {selectedAdr && (
              <div className="info-panel__adr-content">
                <div className="info-panel__adr-content-title">{selectedAdr.title}</div>
                <pre className="info-panel__adr-content-text">{selectedAdr.content}</pre>
              </div>
            )}
          </div>
        );
    }
  }
}

/* ── Helpers ───────────────────────────────────── */

function filterAdrs(adrs: ADRInfo[], node: NodeInfo): ADRInfo[] {
  const nodeName = node.label.toLowerCase();
  const pathParts = (node.file_path ?? '')
    .toLowerCase()
    .split('/')
    .filter(Boolean);

  return adrs.filter((adr) => {
    const adrFilename = adr.filename.toLowerCase();
    const adrTitle = adr.title.toLowerCase();

    if (adrFilename.includes(nodeName) || adrTitle.includes(nodeName)) {
      return true;
    }

    for (const part of pathParts) {
      if (adrFilename.includes(part) || adrTitle.includes(part)) {
        return true;
      }
    }

    return false;
  });
}
