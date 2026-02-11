import * as THREE from 'three';
import type { SceneGraph, Entity } from '../api';
import { getMaterial, getBuildingMaterial } from './materials';

export interface SceneState {
  entityMap: Map<string, THREE.Object3D>;
  entityMetadata: Map<string, Record<string, unknown>>;
  groups: SceneGraph['groups'];
  cityBounds: SceneGraph['metadata']['city_bounds'];
}

/**
 * Loads a scene graph into a Three.js scene.
 * Returns SceneState for visibility toggling, route tracing, and camera fitting.
 */
export function loadSceneGraph(
  sceneGraph: SceneGraph,
  scene: THREE.Scene,
): SceneState {
  const entityMap = new Map<string, THREE.Object3D>();
  const entityMetadata = new Map<string, Record<string, unknown>>();

  // Find max building height for color gradient
  let maxHeight = 60;
  for (const e of sceneGraph.entities) {
    if (e.type === 'building' && e.dimensions.y > maxHeight) {
      maxHeight = e.dimensions.y;
    }
  }

  for (const entity of sceneGraph.entities) {
    const mesh = createMesh(entity, maxHeight);
    if (mesh) {
      entityMap.set(entity.id, mesh);
      scene.add(mesh);
    }
    if (entity.metadata) {
      entityMetadata.set(entity.id, entity.metadata);
    }
  }

  return {
    entityMap,
    entityMetadata,
    groups: sceneGraph.groups,
    cityBounds: sceneGraph.metadata.city_bounds,
  };
}

function createMesh(entity: Entity, maxHeight: number): THREE.Mesh | null {
  const { type, position, dimensions, rotation, material } = entity;
  let geometry: THREE.BufferGeometry;

  switch (type) {
    case 'building':
    case 'path':
    case 'pipe':
    case 'lane':
    case 'park':
    case 'pedway':
    case 'bike_tunnel':
    case 'battery':
    case 'bike_path':
    case 'shuttle_route':
    case 'station':
    case 'sports_field':
    case 'plaza':
    case 'tree':
      geometry = new THREE.BoxGeometry(dimensions.x, dimensions.y, dimensions.z);
      break;
    default:
      return null;
  }

  // Buildings get height-graded materials; others use the standard palette
  let mat: THREE.MeshStandardMaterial;
  if (type === 'building' && material === 'concrete') {
    mat = getBuildingMaterial(dimensions.y / maxHeight);
  } else {
    mat = getMaterial(material);
  }

  const mesh = new THREE.Mesh(geometry, mat);

  // Position: solver sets Y = base of entity.
  // BoxGeometry is centered at origin, so offset Y by half height.
  mesh.position.set(
    position.x,
    position.y + dimensions.y / 2,
    position.z,
  );

  // Apply quaternion rotation [x, y, z, w]
  if (rotation) {
    mesh.quaternion.set(rotation[0], rotation[1], rotation[2], rotation[3]);
  }

  mesh.name = entity.id;
  return mesh;
}
