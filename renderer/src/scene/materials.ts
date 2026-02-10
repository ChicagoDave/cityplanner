import * as THREE from 'three';

const MATERIAL_DEFS: Record<string, { color: number; metalness?: number; roughness?: number }> = {
  concrete:  { color: 0xb0b0b0, roughness: 0.9 },
  glass:     { color: 0x88ccee, metalness: 0.3, roughness: 0.1 },
  brick:     { color: 0xc45a3c, roughness: 0.85 },
  steel:     { color: 0x8899aa, metalness: 0.6, roughness: 0.3 },
  copper:    { color: 0xcc7733, metalness: 0.7, roughness: 0.4 },
  fiber:     { color: 0xffcc00, roughness: 0.6 },
  asphalt:   { color: 0x444444, roughness: 0.95 },
  paver:     { color: 0x999988, roughness: 0.8 },
  grass:     { color: 0x3a7a3a, roughness: 0.95 },
};

const cache = new Map<string, THREE.MeshStandardMaterial>();

export function getMaterial(name: string): THREE.MeshStandardMaterial {
  let mat = cache.get(name);
  if (mat) return mat;

  const def = MATERIAL_DEFS[name] ?? { color: 0xff00ff };
  mat = new THREE.MeshStandardMaterial({
    color: def.color,
    metalness: def.metalness ?? 0.0,
    roughness: def.roughness ?? 0.5,
  });
  cache.set(name, mat);
  return mat;
}

const BUILDING_STEPS = 8;
const buildingCache = new Map<number, THREE.MeshStandardMaterial>();

/**
 * Returns a height-graded material for residential buildings.
 * t is 0..1 where 0 = shortest, 1 = tallest.
 * Gradient: warm sand (low-rise) → cool slate (high-rise).
 */
export function getBuildingMaterial(t: number): THREE.MeshStandardMaterial {
  const step = Math.min(Math.floor(t * BUILDING_STEPS), BUILDING_STEPS - 1);
  let mat = buildingCache.get(step);
  if (mat) return mat;

  const low = new THREE.Color(0xd4c4a0);  // warm sand
  const high = new THREE.Color(0x6688aa); // cool slate
  const color = low.clone().lerp(high, step / (BUILDING_STEPS - 1));

  mat = new THREE.MeshStandardMaterial({
    color,
    roughness: 0.85 - step * 0.05,
    metalness: step * 0.03,
  });
  buildingCache.set(step, mat);
  return mat;
}

export function getHighlightMaterial(): THREE.MeshStandardMaterial {
  let mat = cache.get('__highlight');
  if (mat) return mat;
  mat = new THREE.MeshStandardMaterial({
    color: 0x00ffaa,
    emissive: 0x00ff88,
    emissiveIntensity: 0.6,
    roughness: 0.3,
  });
  cache.set('__highlight', mat);
  return mat;
}

export function disposeMaterials(): void {
  for (const mat of cache.values()) mat.dispose();
  cache.clear();
  for (const mat of buildingCache.values()) mat.dispose();
  buildingCache.clear();
}
