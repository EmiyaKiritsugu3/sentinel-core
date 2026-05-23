/** Mirrors Go's graph.GraphEvent JSON shape. */
export interface GraphEvent {
  type: string;
  payload: Record<string, unknown>;
  timestamp: string;
}

/** Mirrors Go's TaskStatus JSON shape. */
export interface TaskStatus {
  id: string;
  description: string;
  status: string;
  tier?: string;
  verification?: string;
  created_at?: string;
}
