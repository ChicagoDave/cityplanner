import * as THREE from 'three';
import { CameraModeManager } from './camera/modes';
import { loadSceneGraph } from './scene/loader';
import { fetchScene } from './api';
import { initControls } from './ui/controls';
import { RouteTracer } from './ui/route-tracer';

const app = document.getElementById('app')!;
const overlay = document.getElementById('overlay')!;

// Renderer
const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.setPixelRatio(window.devicePixelRatio);
renderer.setClearColor(0x111111);
app.appendChild(renderer.domElement);

// Scene
const scene = new THREE.Scene();

// Camera
const camera = new THREE.PerspectiveCamera(
  60,
  window.innerWidth / window.innerHeight,
  0.1,
  20000,
);
const cameraManager = new CameraModeManager(camera, renderer.domElement);

// Lighting
const ambient = new THREE.AmbientLight(0xffffff, 0.4);
scene.add(ambient);

const directional = new THREE.DirectionalLight(0xffffff, 0.8);
directional.position.set(500, 1000, 500);
scene.add(directional);

// Ground plane — sized after scene loads to cover full extent
const ground = new THREE.Mesh(
  new THREE.CircleGeometry(1500, 64),
  new THREE.MeshStandardMaterial({ color: 0x1a3a1a, side: THREE.DoubleSide }),
);
ground.rotation.x = -Math.PI / 2;
ground.position.y = -0.1; // slightly below surface entities
scene.add(ground);

// Resize handler
window.addEventListener('resize', () => {
  camera.aspect = window.innerWidth / window.innerHeight;
  camera.updateProjectionMatrix();
  renderer.setSize(window.innerWidth, window.innerHeight);
});

// Render loop
const clock = new THREE.Clock();

function animate(): void {
  requestAnimationFrame(animate);
  const delta = clock.getDelta();
  cameraManager.update(delta);
  renderer.render(scene, camera);
}

animate();

// Load scene from solver
loadCity();

async function loadCity(): Promise<void> {
  try {
    overlay.querySelector('p')!.textContent = 'Loading scene...';
    const sceneGraph = await fetchScene();
    const state = loadSceneGraph(sceneGraph, scene);

    if (state.cityBounds) {
      cameraManager.fitToBounds(state.cityBounds);
    }

    initControls(app, state, { ground });
    new RouteTracer(camera, state, renderer.domElement);
    overlay.style.display = 'none';
  } catch (err) {
    overlay.querySelector('p')!.textContent =
      `Failed to load: ${err instanceof Error ? err.message : 'unknown error'}. Is the solver running on :3000?`;
  }
}
