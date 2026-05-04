import { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import { useSentinelData } from '../hooks/useSentinelData';

export function GraphCanvas() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [cyInstance, setCyInstance] = useState<cytoscape.Core | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    const cy = cytoscape({
      container: containerRef.current,
      style: [
        {
          selector: 'node',
          style: {
            'label': 'data(label)',
            'background-color': '#00ADB5',
            'color': '#fff',
            'text-valign': 'center',
            'text-halign': 'center',
            'font-size': '12px',
            'text-wrap': 'wrap',
            'text-max-width': '80px'
          }
        },
        {
          selector: 'edge',
          style: {
            'label': 'data(label)',
            'width': 1.5,
            'line-color': '#555',
            'target-arrow-color': '#555',
            'target-arrow-shape': 'triangle',
            'curve-style': 'bezier',
            'font-size': '10px',
            'color': '#999',
            'text-rotation': 'autorotate'
          }
        },
        {
          selector: 'node[type="unresolved_import"]',
          style: {
            'background-color': '#B53737'
          }
        }
      ],
      layout: { name: 'preset' } // Use preset initially, then trigger layout in hook
    });

    setCyInstance(cy);

    return () => {
      cy.destroy();
    };
  }, []);

  useSentinelData(cyInstance);

  return (
    <div style={{ position: 'relative', width: '100vw', height: '100vh', background: '#1e1e1e' }}>
      <div style={{ position: 'absolute', top: 16, left: 16, color: '#fff', fontFamily: 'monospace', zIndex: 10 }}>
        <h2>Sentinel Live View</h2>
      </div>
      <div ref={containerRef} style={{ width: '100%', height: '100%' }} />
    </div>
  );
}