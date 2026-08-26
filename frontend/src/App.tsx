import { useState, useRef, useCallback, useEffect } from 'react'

// ─── Types ────────────────────────────────────────────────────────────────────

type Screen = 'idle' | 'selected' | 'compressing' | 'done' | 'error'

interface FileInfo {
  name: string
  rawSize: number
  type: string
  preview?: string
}

interface Preset {
  id: string
  name: string
  sublabel: string
  bytes: number
  platform?: string
}

// ─── Constants ────────────────────────────────────────────────────────────────

const PRESETS: Preset[] = [
  { id: 'discord',     name: 'Discord',      sublabel: '8 MB',  bytes: 8  * 1024 * 1024, platform: 'Free' },
  { id: 'nitro',       name: 'Nitro',        sublabel: '25 MB', bytes: 25 * 1024 * 1024, platform: 'Discord' },
  { id: 'email',       name: 'Email',        sublabel: '10 MB', bytes: 10 * 1024 * 1024 },
  { id: 'custom',      name: 'Custom',       sublabel: '…',     bytes: 0 },
]

// ─── Helpers ──────────────────────────────────────────────────────────────────

function fmt(bytes: number): string {
  if (bytes === 0) return '0 B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function isImage(t: string) { return t.startsWith('image/') }
function isVideo(t: string) { return t.startsWith('video/') }

function compressionStage(p: number) {
  if (p < 25) return 'Reading file…'
  if (p < 55) return 'Finding the best quality…'
  if (p < 82) return 'Compressing…'
  return 'Wrapping up…'
}

// ─── Tokens ───────────────────────────────────────────────────────────────────

const C = {
  bg:           '#090910',
  panel:        '#111118',
  raised:       '#18181f',
  hover:        '#1f1f28',
  border:       'rgba(255,255,255,0.07)',
  borderMid:    'rgba(255,255,255,0.12)',
  borderStrong: 'rgba(255,255,255,0.18)',
  text:         '#e8e8f0',
  muted:        'rgba(232,232,240,0.42)',
  faint:        'rgba(232,232,240,0.22)',
  accent:       '#7c6af7',
  accentHover:  '#9283ff',
  accentDim:    'rgba(124,106,247,0.14)',
  accentGlow:   'rgba(124,106,247,0.30)',
  green:        '#34d399',
  greenDim:     'rgba(52,211,153,0.12)',
  amber:        '#fbbf24',
  amberDim:     'rgba(251,191,36,0.10)',
  red:          '#f87171',
  redDim:       'rgba(248,113,113,0.12)',
  mono:         'JetBrains Mono, ui-monospace, monospace',
  sans:         'Outfit, ui-sans-serif, system-ui, sans-serif',
}

// ─── Micro components ─────────────────────────────────────────────────────────

function MonoTag({ children, color = C.muted }: { children: React.ReactNode; color?: string }) {
  return (
    <span style={{ fontFamily: C.mono, fontSize: 11, letterSpacing: '0.06em', color, fontWeight: 400 }}>
      {children}
    </span>
  )
}

function Divider() {
  return <div style={{ height: 1, background: C.border }} />
}

// ─── Drop zone ────────────────────────────────────────────────────────────────

function IdleScreen({ onFile }: { onFile: (f: File) => void }) {
  const [drag, setDrag] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDrag(false)
    const f = e.dataTransfer.files[0]
    if (f) onFile(f)
  }

  return (
    <div className="anim-fadeup" style={{ padding: '12px' }}>
      <div
        onDragOver={e => { e.preventDefault(); setDrag(true) }}
        onDragLeave={() => setDrag(false)}
        onDrop={onDrop}
        onClick={() => inputRef.current?.click()}
        style={{
          position: 'relative',
          borderRadius: 14,
          border: `1.5px dashed ${drag ? C.accent : C.borderMid}`,
          background: drag ? C.accentDim : 'transparent',
          padding: '52px 24px 48px',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          cursor: 'pointer',
          transition: 'all 0.18s ease',
          overflow: 'hidden',
        }}
      >
        {/* Ambient glow on drag */}
        {drag && (
          <div style={{
            position: 'absolute', inset: 0, borderRadius: 14,
            background: `radial-gradient(ellipse at 50% 40%, ${C.accentGlow}, transparent 70%)`,
            pointerEvents: 'none',
          }} />
        )}

        {/* Upload icon */}
        <div style={{
          width: 56, height: 56, borderRadius: 14, marginBottom: 20,
          background: drag ? C.accentDim : C.raised,
          border: `1px solid ${drag ? C.accent : C.borderStrong}`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          transition: 'all 0.18s',
          position: 'relative', zIndex: 1,
        }}>
          <svg width="22" height="22" viewBox="0 0 22 22" fill="none">
            <path d="M11 14.5V5M11 5L7.5 8.5M11 5L14.5 8.5"
              stroke={drag ? C.accent : C.muted}
              strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"
            />
            <path d="M3.5 15.5C3.5 17.16 4.84 18.5 6.5 18.5h9c1.66 0 3-1.34 3-3"
              stroke={drag ? C.accent : C.faint}
              strokeWidth="1.5" strokeLinecap="round"
            />
          </svg>
        </div>

        <p style={{ margin: '0 0 6px', fontSize: 17, fontWeight: 600, color: drag ? '#fff' : C.text, position: 'relative', zIndex: 1, transition: 'color 0.15s' }}>
          {drag ? 'Drop to compress' : 'Drop your file here'}
        </p>

        <button
          onClick={e => { e.stopPropagation(); inputRef.current?.click() }}
          style={{
            padding: '9px 22px',
            borderRadius: 8,
            background: C.raised,
            border: `1px solid ${C.borderStrong}`,
            color: C.text,
            fontSize: 13.5,
            fontWeight: 500,
            cursor: 'pointer',
            fontFamily: C.sans,
            position: 'relative', zIndex: 1,
            transition: 'background 0.12s, border-color 0.12s',
          }}
          onMouseEnter={e => { e.currentTarget.style.background = C.hover; e.currentTarget.style.borderColor = C.borderStrong }}
          onMouseLeave={e => { e.currentTarget.style.background = C.raised; e.currentTarget.style.borderColor = C.borderStrong }}
        >
          Choose File
        </button>
        <input ref={inputRef} type="file" accept="image/*,video/*" style={{ display: 'none' }} onChange={e => { const f = e.target.files?.[0]; if (f) onFile(f) }} />
      </div>


    </div>
  )
}

// ─── File header ──────────────────────────────────────────────────────────────

function FileHeader({ file, onClear }: { file: FileInfo; onClear: () => void }) {
  return (
    <div style={{
      display: 'flex', alignItems: 'center', gap: 12,
      padding: '16px 20px',
      background: C.raised,
      borderRadius: 12,
    }}>
      {file.preview
        ? <img src={file.preview} alt="" style={{ width: 44, height: 44, borderRadius: 8, objectFit: 'cover', flexShrink: 0 }} />
        : (
          <div style={{
            width: 44, height: 44, borderRadius: 8, flexShrink: 0,
            background: C.panel, border: `1px solid ${C.border}`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              {isVideo(file.type)
                ? <path d="M4 5.5A1.5 1.5 0 015.5 4h5.586a1.5 1.5 0 011.06.44l3.415 3.414a1.5 1.5 0 01.439 1.06V14.5A1.5 1.5 0 0114.5 16h-9A1.5 1.5 0 014 14.5v-9z" stroke={C.muted} strokeWidth="1.3"/>
                : <><rect x="4" y="3" width="12" height="14" rx="2" stroke={C.muted} strokeWidth="1.3"/><path d="M7 7.5h6M7 10h4" stroke={C.muted} strokeWidth="1.3" strokeLinecap="round"/></>
              }
            </svg>
          </div>
        )
      }
      <div style={{ flex: 1, minWidth: 0 }}>
        <p style={{ margin: 0, fontSize: 13.5, fontWeight: 500, color: C.text, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {file.name}
        </p>
        <MonoTag>{fmt(file.rawSize)}</MonoTag>
      </div>
      <button onClick={onClear} title="Remove" style={{
        background: 'none', border: 'none', cursor: 'pointer',
        color: C.faint, padding: 4, borderRadius: 6, flexShrink: 0,
        transition: 'color 0.12s',
        display: 'flex', alignItems: 'center',
      }}
        onMouseEnter={e => e.currentTarget.style.color = C.muted}
        onMouseLeave={e => e.currentTarget.style.color = C.faint}
      >
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M3 3L11 11M11 3L3 11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
        </svg>
      </button>
    </div>
  )
}

// ─── Selected screen ──────────────────────────────────────────────────────────

function SelectedScreen({
  file, onCompress, onClear,
}: {
  file: FileInfo
  onCompress: (targetBytes: number) => void
  onClear: () => void
}) {
  const [preset, setPreset] = useState(0)
  const [customMB, setCustomMB] = useState('5')

  const targetBytes = preset === 3
    ? (parseFloat(customMB) || 5) * 1024 * 1024
    : PRESETS[preset].bytes

  const alreadyFits = file.rawSize <= targetBytes

  return (
    <div className="anim-fadeup" style={{ padding: '24px 28px 28px', display: 'flex', flexDirection: 'column', gap: 22 }}>
      <FileHeader file={file} onClear={onClear} />

      {/* Section: target size */}
      <div>
        <p style={{ margin: '0 0 11px', fontSize: 11.5, fontWeight: 600, letterSpacing: '0.09em', color: C.muted, textTransform: 'uppercase', fontFamily: C.mono }}>
          What size do you need?
        </p>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
          {PRESETS.map((p, i) => {
            const active = preset === i
            return (
              <button key={p.id} onClick={() => setPreset(i)} style={{
                padding: '13px 14px',
                borderRadius: 10,
                border: `1px solid ${active ? C.accent : C.border}`,
                background: active ? C.accentDim : C.raised,
                cursor: 'pointer',
                textAlign: 'left',
                fontFamily: C.sans,
                transition: 'all 0.14s',
                position: 'relative',
                overflow: 'hidden',
              }}
                onMouseEnter={e => { if (!active) e.currentTarget.style.background = C.hover }}
                onMouseLeave={e => { if (!active) e.currentTarget.style.background = C.raised }}
              >
                {active && (
                  <div style={{
                    position: 'absolute', inset: 0,
                    background: `radial-gradient(ellipse at 20% 50%, ${C.accentGlow}, transparent 70%)`,
                    pointerEvents: 'none',
                  }} />
                )}
                <div style={{ position: 'relative' }}>
                  <p style={{ margin: 0, fontSize: 14, fontWeight: 600, color: active ? '#fff' : C.text }}>{p.name}</p>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 3 }}>
                    <MonoTag color={active ? C.accent : C.muted}>{p.sublabel}</MonoTag>
                    {p.platform && <span style={{ fontSize: 11, color: C.faint }}>{p.platform}</span>}
                  </div>
                </div>
              </button>
            )
          })}
        </div>

        {/* Custom input */}
        {preset === 3 && (
          <div className="anim-fadein" style={{
            marginTop: 10,
            display: 'flex', alignItems: 'center', gap: 10,
            padding: '11px 14px',
            background: C.raised,
            border: `1px solid ${C.borderMid}`,
            borderRadius: 10,
          }}>
            <span style={{ fontSize: 13.5, color: C.muted, flex: 1 }}>Make it fit under</span>
            <input
              autoFocus
              type="number" min="1" max="2000"
              value={customMB}
              onChange={e => setCustomMB(e.target.value)}
              style={{
                background: 'transparent', border: 'none', outline: 'none',
                color: C.text, fontSize: 16, fontWeight: 600,
                fontFamily: C.mono, width: 52, textAlign: 'right',
                MozAppearance: 'textfield',
              } as React.CSSProperties}
            />
            <span style={{ fontSize: 13.5, color: C.muted }}>MB</span>
          </div>
        )}
      </div>

      {/* Warning: already fits */}
      {alreadyFits && (
        <div className="anim-fadein" style={{
          display: 'flex', gap: 10, padding: '12px 14px',
          background: C.amberDim,
          border: `1px solid rgba(251,191,36,0.22)`,
          borderRadius: 10,
        }}>
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" style={{ flexShrink: 0, marginTop: 1 }}>
            <path d="M8 2L14 13H2L8 2Z" stroke={C.amber} strokeWidth="1.3" strokeLinejoin="round"/>
            <path d="M8 6.5V9.5" stroke={C.amber} strokeWidth="1.3" strokeLinecap="round"/>
            <circle cx="8" cy="11.5" r="0.6" fill={C.amber}/>
          </svg>
          <p style={{ margin: 0, fontSize: 12.5, color: C.amber, lineHeight: 1.55 }}>
            Already under {fmt(targetBytes)} — you can still compress for a smaller result.
          </p>
        </div>
      )}

      <Divider />

      {/* CTA */}
      <button
        onClick={() => onCompress(targetBytes)}
        style={{
          width: '100%', padding: '14px',
          borderRadius: 11,
          background: C.accent,
          border: 'none',
          color: '#fff',
          fontSize: 15, fontWeight: 600,
          cursor: 'pointer', fontFamily: C.sans,
          transition: 'background 0.14s, box-shadow 0.14s',
          boxShadow: `0 0 0 0 ${C.accentGlow}`,
          letterSpacing: '0.01em',
        }}
        onMouseEnter={e => { e.currentTarget.style.background = C.accentHover; e.currentTarget.style.boxShadow = `0 4px 24px ${C.accentGlow}` }}
        onMouseLeave={e => { e.currentTarget.style.background = C.accent; e.currentTarget.style.boxShadow = `0 0 0 0 ${C.accentGlow}` }}
      >
        Compress
      </button>
    </div>
  )
}

// ─── Compressing screen ───────────────────────────────────────────────────────

function CompressingScreen({ progress, file }: { progress: number; file: FileInfo }) {
  const r = 42
  const circ = 2 * Math.PI * r
  const filledArc = (progress / 100) * circ

  return (
    <div className="anim-fadeup" style={{ padding: '32px', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 0 }}>
      {/* Ring */}
      <div style={{ position: 'relative', width: 108, height: 108, marginBottom: 28 }}>
        {/* Ambient pulse */}
        <div style={{
          position: 'absolute', inset: -12,
          borderRadius: '50%',
          background: C.accentGlow,
          animation: 'pulse-ring 2.4s ease-in-out infinite',
        }} />
        <svg width="108" height="108" viewBox="0 0 108 108" fill="none" style={{ position: 'relative', zIndex: 1, transform: 'rotate(-90deg)' }}>
          <circle cx="54" cy="54" r={r} stroke={C.border} strokeWidth="3" />
          <circle
            cx="54" cy="54" r={r}
            stroke={C.accent}
            strokeWidth="3"
            strokeLinecap="round"
            strokeDasharray={`${filledArc} ${circ - filledArc}`}
            style={{ transition: 'stroke-dasharray 0.12s linear', filter: `drop-shadow(0 0 6px ${C.accent})` }}
          />
        </svg>
        {/* Center pct */}
        <div style={{
          position: 'absolute', inset: 0, zIndex: 2,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <span style={{ fontFamily: C.mono, fontSize: 15, fontWeight: 500, color: C.text }}>
            {Math.round(progress)}%
          </span>
        </div>
      </div>

      <p style={{ margin: 0, fontSize: 13, color: C.muted }}>
        <MonoTag>{file.name.length > 32 ? file.name.slice(0, 29) + '…' : file.name}</MonoTag>
      </p>
    </div>
  )
}

// ─── Done screen ──────────────────────────────────────────────────────────────

function DoneScreen({ file, resultSize, onReset }: { file: FileInfo; resultSize: number; onReset: () => void }) {
  const saved = file.rawSize - resultSize
  const pct = Math.round((saved / file.rawSize) * 100)
  const barWidth = Math.round((resultSize / file.rawSize) * 100)

  return (
    <div className="anim-fadeup" style={{ padding: '24px 28px 28px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      {/* Success header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
        <div style={{
          width: 44, height: 44, borderRadius: '50%', flexShrink: 0,
          background: C.greenDim, border: `1px solid ${C.green}`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          boxShadow: `0 0 16px rgba(52,211,153,0.2)`,
        }}>
          <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
            <path d="M4 9L7.5 12.5L14 5.5" stroke={C.green} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </div>
        <div>
          <p style={{ margin: 0, fontSize: 17, fontWeight: 700, color: C.text }}>Ready to send</p>
          <p style={{ margin: '3px 0 0', fontSize: 13, color: C.muted }}>Your file fits the limit</p>
        </div>
      </div>

      <Divider />

      {/* Buttons */}
      <button style={{
        width: '100%', padding: '14px',
        borderRadius: 11,
        background: C.green,
        border: 'none', color: '#001f14',
        fontSize: 15, fontWeight: 700,
        cursor: 'pointer', fontFamily: C.sans,
        transition: 'background 0.14s, box-shadow 0.14s',
        boxShadow: `0 4px 20px rgba(52,211,153,0.25)`,
      }}
        onMouseEnter={e => { e.currentTarget.style.background = '#4ade80' }}
        onMouseLeave={e => { e.currentTarget.style.background = C.green }}
      >
        Save File
      </button>

      <button onClick={onReset} style={{
        width: '100%', padding: '12px',
        borderRadius: 11, background: 'transparent',
        border: `1px solid ${C.border}`,
        color: C.muted, fontSize: 14,
        cursor: 'pointer', fontFamily: C.sans,
        transition: 'border-color 0.12s, color 0.12s',
      }}
        onMouseEnter={e => { e.currentTarget.style.borderColor = C.borderMid; e.currentTarget.style.color = C.text }}
        onMouseLeave={e => { e.currentTarget.style.borderColor = C.border; e.currentTarget.style.color = C.muted }}
      >
        Compress another file
      </button>
    </div>
  )
}

// ─── Error screen ─────────────────────────────────────────────────────────────

function ErrorScreen({ message, onRetry, onReset }: { message: string; onRetry: () => void; onReset: () => void }) {
  return (
    <div className="anim-fadeup" style={{ padding: '32px', textAlign: 'center' }}>
      <div style={{
        width: 52, height: 52, borderRadius: '50%', margin: '0 auto 22px',
        background: C.redDim, border: `1px solid ${C.red}`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}>
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
          <path d="M10 5.5V11" stroke={C.red} strokeWidth="1.8" strokeLinecap="round"/>
          <circle cx="10" cy="14.5" r="1.1" fill={C.red}/>
        </svg>
      </div>
      <p style={{ margin: '0 0 10px', fontSize: 17, fontWeight: 700, color: C.text }}>Something went wrong</p>
      <p style={{ margin: '0 0 30px', fontSize: 13.5, color: C.muted, lineHeight: 1.65, maxWidth: 320, marginLeft: 'auto', marginRight: 'auto' }}>
        {message}
      </p>
      <div style={{ display: 'flex', gap: 10, justifyContent: 'center' }}>
        <button onClick={onRetry} style={{
          padding: '10px 22px', borderRadius: 9,
          background: C.raised, border: `1px solid ${C.borderStrong}`,
          color: C.text, fontSize: 13.5, fontWeight: 500,
          cursor: 'pointer', fontFamily: C.sans,
        }}>
          Try a larger limit
        </button>
        <button onClick={onReset} style={{
          padding: '10px 22px', borderRadius: 9,
          background: 'transparent', border: `1px solid ${C.border}`,
          color: C.muted, fontSize: 13.5,
          cursor: 'pointer', fontFamily: C.sans,
        }}>
          Start over
        </button>
      </div>
    </div>
  )
}

// ─── Root ─────────────────────────────────────────────────────────────────────

export default function App() {
  const [screen, setScreen] = useState<Screen>('idle')
  const [file, setFile] = useState<FileInfo | null>(null)
  const [progress, setProgress] = useState(0)
  const [resultSize, setResultSize] = useState(0)
  const [errorMsg, setErrorMsg] = useState('')
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const loadFile = useCallback((f: File) => {
    const info: FileInfo = { name: f.name, rawSize: f.size, type: f.type }
    if ((isImage(f.type) && f.size < 15 * 1024 * 1024) || (isVideo(f.type) && f.size < 8 * 1024 * 1024)) {
      info.preview = URL.createObjectURL(f)
    }
    setFile(info)
    setScreen('selected')
    setProgress(0)
    setResultSize(0)
    setErrorMsg('')
  }, [])

  const startCompress = useCallback((targetBytes: number) => {
    if (!file) return
    setScreen('compressing')
    setProgress(0)

    let p = 0
    timerRef.current = setInterval(() => {
      const step = Math.random() * 3.5 + 1.2
      p = Math.min(p + step, 100)
      setProgress(p)

      if (p >= 100) {
        clearInterval(timerRef.current!)
        const ratio = targetBytes / file.rawSize
        if (ratio < 0.04) {
          setTimeout(() => {
            setErrorMsg("That size limit is too small for this file — we can't compress it that far without making it unusable. Try a larger limit.")
            setScreen('error')
          }, 350)
        } else {
          const simulated = file.rawSize <= targetBytes
            ? file.rawSize * (0.82 + Math.random() * 0.12)
            : targetBytes * (0.86 + Math.random() * 0.1)
          setTimeout(() => {
            setResultSize(simulated)
            setScreen('done')
          }, 350)
        }
      }
    }, 75)
  }, [file])

  useEffect(() => () => { if (timerRef.current) clearInterval(timerRef.current) }, [])

  const reset = () => {
    setScreen('idle')
    setFile(null)
    setProgress(0)
    setResultSize(0)
    setErrorMsg('')
    if (inputRef.current) inputRef.current.value = ''
  }

  return (
    <div style={{
      width: '100vw', height: '100vh',
      background: C.panel,
      fontFamily: C.sans,
      overflow: 'hidden',
      display: 'flex', flexDirection: 'column',
    }}>
      {/* Screen content */}
      <div key={screen} style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', overflow: 'auto' }}>
        {screen === 'idle'        && <IdleScreen onFile={loadFile} />}
        {screen === 'selected'    && file && <SelectedScreen file={file} onCompress={startCompress} onClear={reset} />}
        {screen === 'compressing' && file && <CompressingScreen progress={progress} file={file} />}
        {screen === 'done'        && file && <DoneScreen file={file} resultSize={resultSize} onReset={reset} />}
        {screen === 'error'       && <ErrorScreen message={errorMsg} onRetry={() => setScreen('selected')} onReset={reset} />}
      </div>
    </div>
  )
}
