import { Canvas } from '@react-three/fiber'
import { Globe } from './Globe'

interface GlobeCanvasProps {
  peaceScore?: number
}

/**
 * Full-screen Three.js canvas rendered behind the UI.
 */
export function GlobeCanvas({ peaceScore }: GlobeCanvasProps) {
  return (
    <Canvas
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100vw',
        height: '100vh',
        zIndex: 0,
        pointerEvents: 'none',
      }}
      camera={{ position: [0, 0, 7], fov: 60 }}
    >
      <Globe peaceScore={peaceScore} />
    </Canvas>
  )
}
