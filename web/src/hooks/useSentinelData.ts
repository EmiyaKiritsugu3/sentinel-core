import { useEffect, useRef } from 'react';
import cytoscape from 'cytoscape';
import { useEventLogStore } from '../stores';

type GraphEvent = {
  type: string;
  payload: any;
  timestamp: string;
};

export function useSentinelData(cy: cytoscape.Core | null) {
  const queue = useRef<GraphEvent[]>([]);
  const isLoaded = useRef(false);

  useEffect(() => {
    if (!cy) return;

    // Use current host for WebSocket and API, fallback to localhost:8080 if not available (e.g. running dev server locally)
    const port = window.location.port === '5173' ? '8080' : window.location.port;
    const host = `${window.location.hostname}:${port}`;
    
    const ws = new WebSocket(`ws://${host}/ws`);
    
    ws.onmessage = (msg) => {
      try {
        const event: GraphEvent = JSON.parse(msg.data);
        useEventLogStore.getState().addEvent(event);
        if (!isLoaded.current) {
          queue.current.push(event);
        } else {
          applyEvent(cy, event);
        }
      } catch (e) {
        console.error('Failed to parse WS message:', e);
      }
    };

    fetch(`http://${host}/api/graph`)
      .then(res => res.json())
      .then(data => {
        const cyNodes = (data.nodes || []).map((n: any) => ({
          data: { id: n.ID, label: n.Name, type: n.Type, file_path: n.FilePath }
        }));
        const cyEdges = (data.edges || []).map((e: any) => ({
          data: { id: `${e.From}-${e.To}-${e.Rel}`, source: e.From, target: e.To, label: e.Rel }
        }));

        cy.add([...cyNodes, ...cyEdges]);
        
        // Concentric or cose layout usually works well for generic architecture graphs
        cy.layout({ 
          name: 'cose',
          animate: false,
          nodeRepulsion: 400000,
          idealEdgeLength: 100,
        }).run();
        
        isLoaded.current = true;
        
        // Flush queue
        while (queue.current.length > 0) {
          const event = queue.current.shift();
          if (event) applyEvent(cy, event);
        }
      })
      .catch(err => console.error('Failed to fetch graph data:', err));

    return () => {
      ws.close();
    };
  }, [cy]);

  function applyEvent(cy: cytoscape.Core, event: GraphEvent) {
    if (event.type === 'NODE_UPSERTED') {
      const n = event.payload;
      const el = cy.getElementById(n.ID);
      if (el.length > 0) {
        el.data('label', n.Name);
        el.data('type', n.Type);
      } else {
        cy.add({ data: { id: n.ID, label: n.Name, type: n.Type, file_path: n.FilePath } });
      }
    } else if (event.type === 'EDGE_CREATED') {
      const e = event.payload;
      const edgeId = `${e.From}-${e.To}-${e.Rel}`;
      if (cy.getElementById(edgeId).length === 0) {
        cy.add({ data: { id: edgeId, source: e.From, target: e.To, label: e.Rel } });
      }
    }
  }
}