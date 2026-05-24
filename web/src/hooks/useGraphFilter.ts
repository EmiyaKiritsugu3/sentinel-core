import { useEffect, useState } from 'react';
import cytoscape from 'cytoscape';
import { useFilterStore } from '../stores';

/**
 * Update Cytoscape node visibility according to type filters, a text search, and an optional package selection.
 *
 * Matches `searchText` case-insensitively against node `label` and `file_path` (partial match). When `enabledTypes` is non-empty, only nodes whose `type` is in the set are shown. When `selectedPackage` is provided, a node is shown only if the top-level prefix of its `file_path` (first segment before `/`, or `'(root)'` if no `/`) equals `selectedPackage`.
 *
 * @param cy - The Cytoscape core instance whose nodes will be updated
 * @param enabledTypes - Set of node `type` values to allow; empty set disables type filtering
 * @param searchText - Text to match against node `label` and `file_path` (case-insensitive, partial)
 * @param selectedPackage - Optional top-level package name to filter by, or `null` to disable package filtering
 */
function applyFilter(
  cy: cytoscape.Core,
  enabledTypes: Set<string>,
  searchText: string,
  selectedPackage: string | null,
) {
  const search = searchText.toLowerCase();
  cy.batch(() => {
    cy.nodes().forEach(node => {
      let visible = true;

      // Type filter: non-empty set = only show those types
      if (enabledTypes.size > 0 && !enabledTypes.has((node.data('type') as string) || '')) {
        visible = false;
      }

      // Text search: match label or file_path (case-insensitive, partial)
      if (search) {
        const label = ((node.data('label') as string) || '').toLowerCase();
        const path  = ((node.data('file_path') as string) || '').toLowerCase();
        if (!label.includes(search) && !path.includes(search)) {
          visible = false;
        }
      }

      // Package filter: match top-level directory prefix
      if (selectedPackage) {
        const p = (node.data('file_path') as string) || '';
        const root = p.includes('/') ? p.split('/')[0] : '(root)';
        if (root !== selectedPackage) {
          visible = false;
        }
      }

      node.style('display', visible ? 'element' : 'none');
    });
  });
}

/**
 * Extracts a sorted list of unique top-level package identifiers from nodes' `file_path` values.
 *
 * For each node, the root is the substring before the first `/`, or `'(root)'` when no `/` is present.
 *
 * @param cy - Cytoscape instance whose nodes' `file_path` data will be inspected
 * @returns Sorted array of unique package roots (first segment of `file_path`, or `'(root)'` when absent)
 */
function extractPackages(cy: cytoscape.Core): string[] {
  return [...new Set(
    cy.nodes()
      .map(n => n.data('file_path') as string)
      .filter(p => p && p.length > 0)
      .map(p => p.includes('/') ? p.split('/')[0] : '(root)')
  )].sort();
}

/**
 * Synchronizes Cytoscape node visibility with filter state and exposes available packages.
 *
 * @param cy - Cytoscape `Core` instance or `null`
 * @returns An object with `packages`: a sorted array of unique top-level package identifiers derived from nodes' `file_path` (the first segment before `/`, or `"(root)"` when no slash)
 */
export function useGraphFilter(cy: cytoscape.Core | null) {
  const enabledTypes    = useFilterStore(s => s.enabledTypes);
  const searchText      = useFilterStore(s => s.searchText);
  const selectedPackage = useFilterStore(s => s.selectedPackage);

  const [packages, setPackages] = useState<string[]>([]);

  // Extract packages when cy instance is first available
  useEffect(() => {
    if (!cy) return;
    setPackages(extractPackages(cy));
  }, [cy]);

  // Apply filter on every store change
  useEffect(() => {
    if (!cy) return;
    applyFilter(cy, enabledTypes, searchText, selectedPackage);
  }, [cy, enabledTypes, searchText, selectedPackage]);

  // Re-apply filter when new nodes arrive via WebSocket
  useEffect(() => {
    if (!cy) return;
    const handler = () => {
      const s = useFilterStore.getState();
      applyFilter(cy, s.enabledTypes, s.searchText, s.selectedPackage);
      setPackages(extractPackages(cy));
    };
    cy.on('add', 'node', handler);
    return () => {
      if (!cy.destroyed()) {
        cy.off('add', 'node', handler);
      }
    };
  }, [cy]);

  return { packages };
}
