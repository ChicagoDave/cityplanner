import * as THREE from 'three';
import { CameraModeManager } from './camera/modes';

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
  1,
  10000,
);
const cameraManager = new CameraModeManager(camera);

// Lighting
const ambient = new THREE.AmbientLight(0xffffff, 0.4);
scene.add(ambient);

const directional = new THREE.DirectionalLight(0xffffff, 0.8);
directional.position.set(500, 1000, 500);
scene.add(directional);

// Ground plane placeholder
const ground = new THREE.Mesh(
  new THREE.CircleGeometry(900, 64),
  new THREE.MeshStandardMaterial({ color: 0x1a3a1a, side: THREE.DoubleSide }),
);
ground.rotation.x = -Math.PI / 2;
scene.add(ground);

// Ring indicators (center, middle, edge boundaries)
const ringRadii = [300, 600, 900];
for (const radius of ringRadii) {
  const ringGeo = new THREE.RingGeometry(radius - 1, radius + 1, 128);
  const ringMat = new THREE.MeshBasicMaterial({
    color: 0x333333,
    side: THREE.DoubleSide,
  });
  const ring = new THREE.Mesh(ringGeo, ringMat);
  ring.rotation.x = -Math.PI / 2;
  ring.position.y = 0.1;
  scene.add(ring);
}

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

// Hide overlay once scene has content
// For now, always show it since no scene is loaded
void overlay;
