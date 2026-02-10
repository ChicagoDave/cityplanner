const API_BASE = '/api';

export interface SceneGraph {
  metadata: {
    spec_version: string;
    generated_at: string;
    city_bounds?: {
      min: { x: number; y: number; z: number };
      max: { x: number; y: number; z: number };
    };
  };
  entities: Entity[];
  groups: {
    pods: Record<string, string[]>;
    systems: Record<string, string[]>;
    layers: Record<string, string[]>;
    entity_types: Record<string, string[]>;
  };
}

export interface Entity {
  id: string;
  type: string;
  position: { x: number; y: number; z: number };
  dimensions: { x: number; y: number; z: number };
  rotation: [number, number, number, number];
  material: string;
  system?: string;
  pod?: string;
  layer: string;
  metadata?: Record<string, unknown>;
  children?: string[];
}

export async function fetchScene(): Promise<SceneGraph> {
  const res = await fetch(`${API_BASE}/scene`);
  if (!res.ok) throw new Error(`Failed to fetch scene: ${res.status}`);
  return res.json();
}

export async function fetchCost(): Promise<unknown> {
  const res = await fetch(`${API_BASE}/cost`);
  if (!res.ok) throw new Error(`Failed to fetch cost: ${res.status}`);
  return res.json();
}

export async function fetchValidation(): Promise<unknown> {
  const res = await fetch(`${API_BASE}/validation`);
  if (!res.ok) throw new Error(`Failed to fetch validation: ${res.status}`);
  return res.json();
}

export async function triggerSolve(): Promise<unknown> {
  const res = await fetch(`${API_BASE}/solve`, { method: 'POST' });
  if (!res.ok) throw new Error(`Failed to trigger solve: ${res.status}`);
  return res.json();
}
