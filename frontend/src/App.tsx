import { useState, useRef, useCallback, useEffect } from 'react'
import { GetDefaultOutputDir, PickFolder, SelectVideoFile, LoadVideo } from '../wailsjs/go/main/App'
import { OnFileDrop, BrowserOpenURL } from '../wailsjs/runtime'
import { domain } from '../wailsjs/go/models'

type Stage = 'upload' | 'config' | 'progress' | 'result'

interface FileInfo {
  name: string
  path: string
  rawSize: number
  duration: number
  type: string
  preview?: string
}

function fmt(bytes: number): string {
  if (bytes === 0) return '0 B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

export default function App() {
  const [stage, setStage] = useState<Stage>('upload')
  const [file, setFile] = useState<FileInfo | null>(null)
  const [targetMB, setTargetMB] = useState(1)
  const [targetInput, setTargetInput] = useState('1')
  const [progress, setProgress] = useState(15)
  const [passTitle, setPassTitle] = useState('Pass 1 of 2: Analyzing bitrate...')
  const [supportModal, setSupportModal] = useState(false)
  const [supportView, setSupportView] = useState<'main' | 'vodafone' | 'work'>('main')
  const [copied, setCopied] = useState(false)
  const [drag, setDrag] = useState(false)
  const [destination, setDestination] = useState('')
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const dropStageRef = useRef(stage)

  const originalSize = file ? file.rawSize : 320.4 * 1024 * 1024
  const originalMB = originalSize / (1024 * 1024)
  const rawTarget = parseInt(targetInput)
  const effectiveMB = !isNaN(rawTarget) && rawTarget >= 1 ? rawTarget : targetMB
  const reductionPct = effectiveMB >= originalMB
    ? '0.0'
    : Math.max(0, ((originalMB - effectiveMB) / originalMB) * 100).toFixed(1)
  const isOverOriginal = !isNaN(rawTarget) && rawTarget >= originalMB

  const applyVideo = useCallback((info: domain.Video) => {
    const f: FileInfo = {
      name: info.name,
      path: info.path,
      rawSize: info.size,
      duration: info.duration,
      type: 'video',
    }
    setFile(f)
    setStage('config')
    GetDefaultOutputDir().then(setDestination).catch(() => {})
    const mb = Math.max(1, Math.floor(info.size / (1024 * 1024) / 4))
    setTargetMB(mb)
    setTargetInput(String(mb))
  }, [])

  const loadFromPath = useCallback(async (path: string) => {
    applyVideo(await LoadVideo(path))
  }, [applyVideo])

  const openVideoPicker = useCallback(async () => {
    const info = await SelectVideoFile()
    if (info) applyVideo(info)
  }, [applyVideo])

  dropStageRef.current = stage

  useEffect(() => {
    const onDrop = (_x: number, _y: number, paths: string[]) => {
      if (dropStageRef.current !== 'upload') return
      const path = paths[0]
      if (!path) return
      setDrag(false)
      loadFromPath(path).catch(() => {})
    }
    OnFileDrop(onDrop, true)
  }, [loadFromPath])

  const openFolderPicker = async () => {
    let dir = destination
    try {
      dir = await GetDefaultOutputDir()
    } catch { /* fall back to current destination */ }
    const picked = await PickFolder(dir)
    if (picked) setDestination(picked)
  }

  const startCompression = () => {
    setStage('progress')
    setProgress(0)
    setPassTitle('Pass 1 of 2: Analyzing bitrate...')

    let pass = 1
    let p = 0
    timerRef.current = setInterval(() => {
      p += 2
      if (p >= 100) {
        if (pass === 1) {
          pass = 2
          p = 0
          setProgress(0)
          setPassTitle('Pass 2 of 2: Compressing video...')
        } else {
          clearInterval(timerRef.current!)
          setStage('result')
        }
        return
      }
      setProgress(p)
    }, 80)
  }

  const cancelCompression = () => {
    if (timerRef.current) clearInterval(timerRef.current)
    setStage('config')
  }

  const reset = () => {
    if (timerRef.current) clearInterval(timerRef.current)
    setStage('upload')
    setFile(null)
    setProgress(15)
  }

  const copyVodafoneNumber = () => {
    navigator.clipboard.writeText('+201006311537').then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1800)
    })
  }

  useEffect(() => () => { if (timerRef.current) clearInterval(timerRef.current) }, [])
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setSupportModal(false)
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  const circumference = 2 * Math.PI * 42

  return (
    <div className="min-h-screen flex flex-col bg-bg text-fg font-sans select-none overflow-x-hidden">
      {/* Top Window Bar */}
      <header className="h-12 border-b border-border/70 flex items-center justify-between px-5 shrink-0 bg-bg">
        <div className="flex items-center gap-4">
          <span className="font-medium text-sm tracking-tight text-fg">Byteless</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            className="p-1.5 rounded-md hover:bg-card border border-transparent hover:border-border text-muted hover:text-red-400 transition-all flex items-center justify-center"
            onClick={() => { setSupportView('main'); setSupportModal(true) }}
            title="Support Byto"
          >
            <svg className="w-[18px] h-[18px]" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" viewBox="0 0 24 24">
              <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z" />
            </svg>
          </button>
        </div>
      </header>

      {/* Main Workspace */}
      <main className="flex-1 flex items-center justify-center p-6 w-full max-w-xl mx-auto">

        {/* Stage 1: Upload */}
        {stage === 'upload' && (
          <div className="w-full flex-col items-center justify-center animate-fadeIn flex">
            <div
              className={`group w-full h-80 rounded-2xl border border-dashed border-border hover:border-zinc-500 bg-card/40 hover:bg-card/90 transition-all duration-200 cursor-pointer flex flex-col items-center justify-center p-8 text-center relative overflow-hidden ${drag ? 'border-zinc-500 bg-card/90' : ''}`}
              onDragOver={e => { e.preventDefault(); setDrag(true) }}
              onDragLeave={() => setDrag(false)}
              onClick={openVideoPicker}
              style={{ ['--wails-drop-target' as string]: 'drop' } as React.CSSProperties}
            >
              <div className={`w-14 h-14 rounded-2xl bg-subtle border border-border flex items-center justify-center mb-5 transition-all ${drag ? 'text-fg scale-105' : 'text-secondary group-hover:text-fg group-hover:scale-105'}`}>
                <svg className="w-6 h-6" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                  <path d="M3 16V6a3 3 0 0 1 3-3h7.5a3 3 0 0 1 2.1.86l3.4 3.4A3 3 0 0 1 19.5 8H21" strokeLinecap="round" strokeLinejoin="round" />
                  <path d="M3 16l3.5-3.5M6.5 12.5L10 16" strokeLinecap="round" strokeLinejoin="round" />
                  <path d="M3 16h14a3 3 0 0 0 3-3v-2" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </div>
              <h2 className="text-base font-medium text-fg mb-1.5">{drag ? 'Drop to compress' : 'Drop video here or click to browse'}</h2>
              <p className="text-xs text-muted max-w-xs mb-5">Supports MP4, MOV, MKV, and WebM</p>
            </div>
          </div>
        )}

        {/* Stage 2: Config */}
        {stage === 'config' && file && (
          <div className="w-full flex-col animate-fadeIn flex">
            <div className="w-full rounded-2xl border border-border bg-card p-6 shadow-xl space-y-6">
              {/* Video Details */}
              <div className="flex items-center justify-between pb-5 border-b border-border/80">
                <div className="flex items-center gap-3.5 min-w-0">
                  <div className="w-11 h-11 rounded-xl bg-subtle border border-border flex items-center justify-center text-muted shrink-0">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                      <path d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  </div>
                  <div className="min-w-0">
                    <div className="text-sm font-medium text-fg truncate">{file.name}</div>
                    <div className="text-xs text-muted font-mono mt-0.5 flex items-center gap-2">
                      <span>{fmt(file.rawSize)}</span>
                      <span>&#8226;</span>
                      <span>{formatDuration(file.duration)}</span>
                    </div>
                  </div>
                </div>
                <button
                  className="text-xs text-muted hover:text-fg px-2 py-1 rounded hover:bg-subtle border border-transparent hover:border-border transition-all"
                  onClick={reset}
                >
                  Change
                </button>
              </div>

              {/* Target Size */}
              <div className="space-y-3">
                <label className="text-xs font-medium text-secondary uppercase tracking-wider">Desired Target Size</label>
                <div className="relative flex items-center">
                  <input
                    className="w-full bg-subtle border border-border focus:border-zinc-400 focus:ring-0 text-fg text-2xl font-mono font-medium rounded-xl px-4 py-3 h-14 outline-none transition-colors [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                    type="text"
                    inputMode="numeric"
                    pattern="[0-9]*"
                    value={targetInput}
                    onChange={e => {
                      const raw = e.target.value.replace(/[^0-9]/g, '')
                      setTargetInput(raw)
                      const v = parseInt(raw)
                      if (!isNaN(v) && v >= 1) {
                        setTargetMB(v)
                      }
                    }}
                    onBlur={() => {
                      const v = parseInt(targetInput)
                      if (isNaN(v) || v < 1) {
                        setTargetMB(1)
                        setTargetInput('1')
                      } else {
                        setTargetMB(v)
                        setTargetInput(String(v))
                      }
                    }}
                  />
                  <div className="absolute right-4 flex items-center gap-3">
                    <span className="text-sm font-mono text-muted font-medium pointer-events-none">MB</span>
                    <div className="flex flex-col border-l border-border pl-3">
                      <button className="text-muted hover:text-fg text-xs leading-none p-0.5" onClick={() => setTargetMB(t => { const v = t + 1; setTargetInput(String(v)); return v })}>&#9650;</button>
                      <button className="text-muted hover:text-fg text-xs leading-none p-0.5 mt-1" onClick={() => setTargetMB(t => { const v = Math.max(t - 1, 1); setTargetInput(String(v)); return v })}>&#9660;</button>
                    </div>
                  </div>
                </div>
              </div>

              {/* Destination Folder */}
              <div className="space-y-3">
                <label className="text-xs font-medium text-secondary uppercase tracking-wider">Destination Folder</label>
                <div className="relative flex items-center">
                  <div className="w-full bg-subtle border border-border focus-within:border-zinc-400 text-fg text-sm font-mono rounded-xl pl-4 pr-4 h-14 outline-none transition-colors flex items-center gap-3 overflow-hidden">
                    <svg className="w-4 h-4 text-muted shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" viewBox="0 0 24 24">
                      <path d="M3 7.5A1.5 1.5 0 0 1 4.5 6h4.6a1.5 1.5 0 0 1 1.06.44L11.3 7.5h9.2A1.5 1.5 0 0 1 22 9v9a1.5 1.5 0 0 1-1.5 1.5h-16A1.5 1.5 0 0 1 3 18z" />
                    </svg>
                    <span className="flex-1 truncate">{destination || 'Choose a folder'}</span>
                  </div>
                  <button
                    className="absolute right-2.5 shrink-0 h-9 px-3 rounded-lg text-xs font-medium bg-fg text-bg hover:bg-zinc-200 active:scale-[0.97] transition-all flex items-center gap-1.5"
                    onClick={openFolderPicker}
                  >
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                    <span>Browse</span>
                  </button>
                </div>
              </div>

              {/* Warning */}
              {isOverOriginal && (
                <div className="flex items-center gap-3 px-3 py-2.5 rounded-xl bg-amber-500/10 border border-amber-500/20 animate-fadeIn">
                  <svg className="w-4 h-4 text-amber-400 shrink-0" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                    <path d="M12 9v4M12 17h.01M10.29 3.86l-8.6 14.86A2 2 0 003.4 22h17.2a2 2 0 001.71-3.28l-8.6-14.86a2 2 0 00-3.42 0z" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                  <span className="text-xs text-amber-300">Target size is not smaller than the original ({fmt(file.rawSize)}). Pick a smaller value to compress.</span>
                </div>
              )}

              {/* Compress Button */}
              <button
                className={`w-full h-12 rounded-xl font-medium text-sm transition-all flex items-center justify-center gap-2 ${isOverOriginal ? 'bg-subtle text-muted border border-border cursor-not-allowed' : 'bg-fg text-bg hover:bg-zinc-200 active:scale-[0.99]'}`}
                onClick={startCompression}
                disabled={isOverOriginal}
              >
                <span>Compress Video</span>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path d="M5 12h14M12 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </button>
            </div>
          </div>
        )}

        {/* Stage 3: Progress */}
        {stage === 'progress' && (
          <div className="w-full flex-col animate-fadeIn flex">
            <div className="w-full rounded-2xl border border-border bg-card p-8 shadow-xl flex flex-col items-center text-center space-y-6">
              {/* Circular Progress */}
              <div className="relative w-36 h-36 flex items-center justify-center my-2">
                <svg className="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
                  <circle className="text-subtle" cx="50" cy="50" fill="transparent" r="42" stroke="currentColor" strokeWidth="6" />
                  <circle
                    className="text-fg transition-all duration-300 ease-out"
                    cx="50" cy="50" fill="transparent" r="42"
                    stroke="currentColor"
                    strokeDasharray={circumference}
                    strokeDashoffset={circumference - (progress / 100) * circumference}
                    strokeLinecap="round" strokeWidth="6"
                  />
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <span className="text-2xl font-mono font-medium text-fg">{progress}%</span>
                  <span className="text-[11px] font-mono text-muted mt-0.5">{Math.max(0, Math.round((100 - progress) * 0.18))}s left</span>
                </div>
              </div>

              {/* Pass Status */}
              <div className="space-y-1">
                <h3 className="text-sm font-medium text-fg">{passTitle}</h3>
                <p className="text-xs text-muted font-mono truncate max-w-sm">
                  Targeting {effectiveMB} MB
                </p>
              </div>

              {/* Actions */}
              <div className="pt-2 w-full flex items-center justify-center">
                <button
                  className="px-4 py-1.5 rounded-lg bg-subtle hover:bg-zinc-800 border border-border text-xs text-secondary hover:text-fg transition-all"
                  onClick={cancelCompression}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Stage 4: Result */}
        {stage === 'result' && (
          <div className="w-full flex-col space-y-4 animate-fadeIn flex">
            <div className="w-full rounded-2xl border border-border bg-card p-6 shadow-xl space-y-6">
              {/* Header */}
              <div className="flex items-center justify-between pb-4 border-b border-border/80">
                <div className="flex items-center gap-2.5">
                  <div className="w-8 h-8 rounded-full bg-subtle border border-border flex items-center justify-center text-fg">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path d="M5 13l4 4L19 7" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  </div>
                  <div>
                    <div className="text-sm font-medium text-fg">Compression Finished</div>
                    <div className="text-xs text-muted font-mono">Elapsed time: {Math.round(progress * 0.38)}s</div>
                  </div>
                </div>
                <span className="text-xs font-mono px-2 py-0.5 rounded bg-subtle border border-border text-fg font-medium">-{reductionPct}%</span>
              </div>

              {/* Size Comparison */}
              <div className="bg-subtle rounded-xl p-4 border border-border flex items-center justify-between">
                <div className="space-y-1">
                  <div className="text-[11px] uppercase tracking-wider text-muted font-medium">Original</div>
                  <div className="text-base font-mono text-secondary">{fmt(file?.rawSize || 0)}</div>
                </div>
                <div className="text-muted text-xs">&rarr;</div>
                <div className="space-y-1 text-right">
                  <div className="text-[11px] uppercase tracking-wider text-secondary font-medium">New Size</div>
                  <div className="text-base font-mono text-fg font-medium">{fmt(effectiveMB * 1024 * 1024)}</div>
                </div>
              </div>

              {/* File Details */}
              <div className="text-xs font-mono text-muted truncate px-1 flex items-center justify-between">
                <span className="truncate">{file?.name || 'output.mp4'}</span>
                <span className="shrink-0 text-[11px]">AV1 &bull; 1080p</span>
              </div>

              {/* Actions */}
              <div className="space-y-2 pt-1">
                <button className="w-full h-11 rounded-xl bg-fg text-bg font-medium text-xs hover:bg-zinc-200 transition-all flex items-center justify-center gap-1.5">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                    <path d="M5 3l14 9-14 9V3z" fill="currentColor" />
                  </svg>
                  <span>Open File</span>
                </button>
                <div className="grid grid-cols-2 gap-2">
                  <button className="h-10 rounded-xl bg-subtle hover:bg-zinc-800 border border-border text-xs text-secondary hover:text-fg transition-all flex items-center justify-center gap-1.5">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                    <span>Show in Folder</span>
                  </button>
                  <button
                    className="h-10 rounded-xl bg-subtle hover:bg-zinc-800 border border-border text-xs text-secondary hover:text-fg transition-all flex items-center justify-center gap-1.5"
                    onClick={reset}
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path d="M1 4v6h6M23 20v-6h-6" strokeLinecap="round" strokeLinejoin="round" />
                      <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                    <span>Compress Another</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Support Modal */}
      {supportModal && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4 transition-opacity duration-200">
          <div className="bg-[#0f1013] border border-[#23252b] rounded-2xl shadow-2xl max-w-[440px] w-full p-6 text-white relative transition-all duration-200">
            {/* Header */}
            <div className="flex items-start justify-between">
              <div>
                <h2 className="text-lg font-semibold tracking-tight text-zinc-100">Support Byteless</h2>
                <p className="text-[13px] text-zinc-400 mt-1">If you find Byteless useful, consider supporting the project</p>
              </div>
              <button className="text-zinc-400 hover:text-zinc-200 transition-colors -mr-1 -mt-1 p-1" onClick={() => setSupportModal(false)}>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                  <line x1="18" x2="6" y1="6" y2="18" />
                  <line x1="6" x2="18" y1="6" y2="18" />
                </svg>
              </button>
            </div>

            {/* Main View */}
            {supportView === 'main' && (
              <div className="mt-5 space-y-2.5">
                {/* Looking for work */}
                <button className="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl bg-[#17181c] border border-[#262830] hover:bg-[#1f2127] text-left transition-colors group" onClick={() => setSupportView('work')}>
                  <svg className="w-4 h-4 text-zinc-300 shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                    <rect height="10" rx="2" ry="2" width="20" x="2" y="7" />
                    <path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2" />
                  </svg>
                  <span className="text-[13px] font-medium text-zinc-200 group-hover:text-white">I'm looking for work — Hire me</span>
                </button>

                {/* Vodafone Cash */}
                <button className="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl bg-[#17181c] border border-[#262830] hover:bg-[#1f2127] text-left transition-colors group" onClick={() => setSupportView('vodafone')}>
                  <svg className="w-4 h-4 text-zinc-300 shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                    <rect height="20" rx="2" ry="2" width="14" x="5" y="2" />
                    <line x1="12" x2="12.01" y1="18" y2="18" />
                  </svg>
                  <span className="text-[13px] font-medium text-zinc-200 group-hover:text-white">Support via Vodafone Cash</span>
                </button>

                {/* Ko-fi */}
                <button className="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl bg-[#17181c] border border-[#262830] hover:bg-[#1f2127] text-left transition-colors group" onClick={() => BrowserOpenURL('https://ko-fi.com')}>
                  <svg className="w-4 h-4 text-zinc-300 shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                    <path d="M17 8h1a4 4 0 1 1 0 8h-1" />
                    <path d="M3 8h14v9a4 4 0 0 1-4 4H7a4 4 0 0 1-4-4Z" />
                    <line x1="6" x2="6" y1="2" y2="4" />
                    <line x1="10" x2="10" y1="2" y2="4" />
                    <line x1="14" x2="14" y1="2" y2="4" />
                  </svg>
                  <span className="text-[13px] font-medium text-zinc-200 group-hover:text-white">Support me on Ko-fi</span>
                </button>

                {/* Star GitHub */}
                <button className="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl bg-[#17181c] border border-[#262830] hover:bg-[#1f2127] text-left transition-colors group" onClick={() => BrowserOpenURL('https://github.com/OmarNaru1110/byteless')}>
                  <svg className="w-4 h-4 text-zinc-300 shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                    <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                  </svg>
                  <span className="text-[13px] font-medium text-zinc-200 group-hover:text-white">Star/Contribute to byteless on GitHub</span>
                </button>

                <div className="pt-4 flex justify-end">
                  <button className="px-5 py-2 rounded-xl bg-[#0e56c8] hover:bg-[#1466ea] active:scale-[0.98] text-white text-xs font-semibold shadow-md transition-all" onClick={() => setSupportModal(false)}>
                    Close
                  </button>
                </div>
              </div>
            )}

            {/* Vodafone View */}
            {supportView === 'vodafone' && (
              <div className="mt-4 animate-fadeIn">
                <div className="rounded-xl bg-[#141519] border border-[#24262d] p-4 text-left shadow-lg">
                  <div className="flex items-center justify-between pb-3">
                    <div className="flex items-center gap-2">
                      <span className="w-5 h-6 rounded border border-red-500/80 bg-red-500/10 flex items-center justify-center">
                        <span className="w-1.5 h-1.5 rounded-full bg-red-500" />
                      </span>
                      <span className="text-sm font-semibold text-zinc-100">Vodafone Cash</span>
                    </div>
                    <button className="text-zinc-500 hover:text-zinc-300 p-0.5" onClick={() => setSupportView('main')}>
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                        <line x1="18" x2="6" y1="6" y2="18" />
                        <line x1="6" x2="18" y1="6" y2="18" />
                      </svg>
                    </button>
                  </div>
                  <p className="text-xs text-zinc-400 mb-3">You can send your support to the following number:</p>
                  <div className="w-full bg-[#1b1c22] rounded-xl py-3 px-4 border border-[#2c2f38] text-center mb-3">
                    <span className="font-mono font-semibold tracking-wider text-sm text-zinc-100">+20 100 631 1537</span>
                  </div>
                  <button
                    className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-[#191b20] border border-[#2c2f38] hover:bg-[#22242c] active:scale-[0.99] text-xs font-semibold text-zinc-200 hover:text-white transition-all shadow-sm"
                    onClick={copyVodafoneNumber}
                  >
                    <svg className="w-4 h-4 text-zinc-400" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                      <rect height="13" rx="2" ry="2" width="13" x="9" y="9" />
                      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                    </svg>
                    <span>{copied ? 'Copied!' : 'Copy Number'}</span>
                  </button>
                </div>
                <div className="pt-4 flex justify-end">
                  <button className="px-5 py-2 rounded-xl bg-[#0e56c8] hover:bg-[#1466ea] active:scale-[0.98] text-white text-xs font-semibold shadow-md transition-all" onClick={() => setSupportModal(false)}>
                    Close
                  </button>
                </div>
              </div>
            )}

            {/* Work View */}
            {supportView === 'work' && (
              <div className="mt-4 animate-fadeIn">
                <div className="rounded-xl bg-[#141519] border border-[#24262d] p-4 text-left shadow-lg">
                  <div className="flex items-center justify-between pb-3">
                    <div className="flex items-center gap-2">
                      <span className="w-5 h-6 rounded border border-green-500/80 bg-green-500/10 flex items-center justify-center">
                        <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                      </span>
                      <span className="text-sm font-semibold text-zinc-100">Looking for work</span>
                    </div>
                    <button className="text-zinc-500 hover:text-zinc-300 p-0.5" onClick={() => setSupportView('main')}>
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                        <line x1="18" x2="6" y1="6" y2="18" />
                        <line x1="6" x2="18" y1="6" y2="18" />
                      </svg>
                    </button>
                  </div>
                  <p className="text-xs text-zinc-400 mb-3">I'm a junior software engineer looking for work. If your team or company needs someone, I'd love to hear from you:</p>
                  <div className="space-y-2.5">
                    <button className="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl bg-[#1b1c22] border border-[#2c2f38] hover:bg-[#22242c] text-left transition-all group" onClick={() => BrowserOpenURL('https://omarnaru.online/')}>
                      <svg className="w-4 h-4 text-zinc-400 shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                        <circle cx="12" cy="12" r="10" />
                        <line x1="2" x2="22" y1="12" y2="12" />
                        <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                      </svg>
                      <span className="text-[13px] font-medium text-zinc-200 group-hover:text-white">omarnaru.online</span>
                    </button>
                    <button className="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl bg-[#1b1c22] border border-[#2c2f38] hover:bg-[#22242c] text-left transition-all group" onClick={() => BrowserOpenURL('mailto:omarnaru2002@gmail.com')}>
                      <svg className="w-4 h-4 text-zinc-400 shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                        <rect height="16" rx="2" ry="2" width="20" x="2" y="4" />
                        <path d="m22 7-10 5L2 7" />
                      </svg>
                      <span className="text-[13px] font-medium text-zinc-200 group-hover:text-white">omarnaru2002@gmail.com</span>
                    </button>
                    <button className="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl bg-[#1b1c22] border border-[#2c2f38] hover:bg-[#22242c] text-left transition-all group" onClick={() => BrowserOpenURL('https://www.linkedin.com/in/omarnaru/')}>
                      <svg className="w-4 h-4 text-zinc-400 shrink-0" fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
                        <path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6z" />
                        <rect height="12" width="4" x="2" y="9" />
                        <circle cx="4" cy="4" r="2" />
                      </svg>
                      <span className="text-[13px] font-medium text-zinc-200 group-hover:text-white">linkedin.com/in/omarnaru</span>
                    </button>
                  </div>
                </div>
                <div className="pt-4 flex justify-end">
                  <button className="px-5 py-2 rounded-xl bg-[#0e56c8] hover:bg-[#1466ea] active:scale-[0.98] text-white text-xs font-semibold shadow-md transition-all" onClick={() => setSupportModal(false)}>
                    Close
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
