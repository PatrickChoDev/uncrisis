import { useRef } from 'react'
import { useFrame } from '@react-three/fiber'
import { Sphere, MeshDistortMaterial, Stars } from '@react-three/drei'
import type { Mesh } from 'three'

interface GlobeProps {
  peaceScore?: number // 0-100 influences glow colour
}

/**
 * Animated 3-D globe used as the background visual for the game.
 * A higher peace score shifts the sphere colour from orange-red toward blue-green.
 */
export function Globe({ peaceScore = 50 }: GlobeProps) {
  const meshRef = useRef<Mesh>(null)

  // Lerp between crisis-red and peace-blue based on peaceScore
  const t = Math.min(Math.max(peaceScore / 100, 0), 1)
  const r = Math.round(255 * (1 - t))
  const g = Math.round(100 + 155 * t)
  const b = Math.round(100 + 155 * t)
  const color = `rgb(${r},${g},${b})`

  useFrame((_, delta) => {
    if (meshRef.current) {
      meshRef.current.rotation.y += delta * 0.08
    }
  })

  return (
    <>
      <Stars radius={200} depth={60} count={3000} factor={4} fade speed={0.5} />

      <ambientLight intensity={0.3} />
      <pointLight position={[10, 10, 10]} intensity={1.5} />
      <pointLight position={[-10, -10, -10]} intensity={0.5} color="#4488ff" />

      <Sphere ref={meshRef} args={[2.5, 64, 64]}>
        <MeshDistortMaterial
          color={color}
          distort={0.25}
          speed={1.5}
          roughness={0.4}
          metalness={0.2}
        />
      </Sphere>
    </>
  )
}
