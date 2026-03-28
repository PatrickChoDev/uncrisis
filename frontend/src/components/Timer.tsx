interface TimerProps {
  secondsLeft: number
}

export function Timer({ secondsLeft }: TimerProps) {
  const isUrgent = secondsLeft <= 10

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '0.25rem',
      }}
    >
      <span style={{ fontSize: '0.75rem', color: '#aaa', textTransform: 'uppercase' }}>
        Time remaining
      </span>
      <span
        style={{
          fontSize: '2.5rem',
          fontWeight: 700,
          fontVariantNumeric: 'tabular-nums',
          color: isUrgent ? '#f44336' : '#ffffff',
          textShadow: isUrgent ? '0 0 12px #f44336' : 'none',
          transition: 'color 0.3s, text-shadow 0.3s',
        }}
      >
        {String(secondsLeft).padStart(2, '0')}s
      </span>

      {/* Progress bar */}
      <div
        style={{
          width: 200,
          height: 4,
          background: '#333',
          borderRadius: 2,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            height: '100%',
            width: `${(secondsLeft / 30) * 100}%`,
            background: isUrgent ? '#f44336' : '#4caf50',
            transition: 'width 1s linear, background 0.3s',
          }}
        />
      </div>
    </div>
  )
}
