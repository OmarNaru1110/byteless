# byteless

> A modern, lightweight, and powerful desktop app for compressing videos to a target size without the eye-bleeding bitrate guessing.

![byteless tour](assets/byteless.jpg)

<p align="center">
  <img src="https://img.shields.io/github/downloads/OmarNaru1110/byteless/total?style=for-the-badge" alt="GitHub Downloads (all releases)">
</p>

**byteless** wraps the complexity of the ffmpeg command line into a clean, easy-to-use desktop application. Whether you're shrinking a clip to fit a file-size limit or squeezing a video down for sharing, byteless hits your target size every time with two-pass precision.

---

## Key Features

- **Modern UI**: Built with React and modern design principles, offering a clean, dark-themed interface.
- **Exact Target Size**: Enter the size you want in MB and byteless computes the right bitrate to land right on it.
- **Two-Pass Encoding**: A smart "pass 1 → pass 2" pipeline analyzes your video first, then encodes, so quality is preserved while hitting your target.
- **Smart Size Floor**: Automatically calculates the minimum achievable size for your clip, so you never over-compress.
- **Encoder Choice**: Pick between **H.264** (fast, widely compatible) and **H.265** (better compression, slower encodes).
- **Drag & Drop**: Drop a video anywhere on the upload screen, or browse for files — `mp4`, `mov`, `mkv`, and `webm` supported.
- **Real-time Progress**: Watch live, per-pass progress on the circular progress ring.
- **Live Size Comparison**: See the original vs. new size and your total reduction percentage before you finish.
- **Auto-Dependency Discovery**: byteless automatically finds [ffmpeg](https://www.ffmpeg.org/) and ffprobe from the app's config directory, next to the executable, or on your PATH. No manual setup required.
- **Cancel Anytime**: Stop an in-progress compression cleanly, no leftover processes.
- **One-Click Results**: Open the file or jump straight to it in your file manager when encoding finishes.

## Technology Stack

Byteless relies on a robust stack to deliver high performance and a small footprint:

- **Backend**: [Go](https://go.dev/) (powered by the [Wails](https://wails.io/) framework)
- **Frontend**: [React](https://react.dev/), [TypeScript](https://www.typescriptlang.org/), and [Vite](https://vitejs.dev/)
- **Styling**: [Tailwind CSS](https://tailwindcss.com/)
- **Core Engine**: [FFmpeg](https://www.ffmpeg.org/) & [ffprobe](https://ffmpeg.org/ffprobe.html)

## Getting Started

### Installation

Download the latest release from the [Releases](https://github.com/OmarNaru1110/byteless/releases) page.

#### Windows
1. Download `byteless.rar` from the release page
2. Extract it with [WinRAR](https://www.win-rar.com/) or [7-Zip](https://www.7-zip.org/). Place the folder anywhere you like
3. Run `byteless.exe` from the extracted folder
4. You may see a SmartScreen warning (the app is not code-signed). Click **"More info"** → **"Run anyway"**

### Building from Source

Byteless requires [Go](https://go.dev/dl/) and [Wails CLI](https://wails.io/docs/gettingstarted/installation), plus [Node.js](https://nodejs.org/) for the frontend.

```bash
# Install the Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build for production
wails build

# Live development with hot reload
wails dev
```

The production binary will be placed in `build/bin/`.

### FFmpeg / FFprobe Setup

Byteless looks for `ffmpeg` and `ffprobe` (in this order):

1. The `BYTELESS_FFMPEG_PATH` / `BYTELESS_FFPROBE_PATH` environment variables
2. The app's config directory (e.g. `AppData\Roaming\byteless` on Windows)
3. Next to the running executable (or its `bin/` subfolder)
4. The working directory or `build/bin`
5. Your system PATH

For the tool to work, make sure [ffmpeg](https://www.ffmpeg.org/download.html) is present in one of those locations.

## Usage

1. **Launch byteless**
2. **Pick a Video**: Drag and drop a video file onto the upload screen, or click to browse
3. **Set Target Size**: Enter your desired file size in MB. byteless shows the minimum size your clip can reach
4. **Choose an Encoder**: Pick **H.264** for fast, widely compatible output, or **H.265** for better compression at slower encode times
5. **Choose Destination**: Select the folder where the compressed file will be saved
6. **Compress**: Click **Compress Video** and follow pass 1 (bitrate analysis) and pass 2 (encoding) on the live progress ring
7. **Finish Up**: Open the file directly or reveal it in your file manager

> **Note**: Make sure ffmpeg and ffprobe are available before compressing. If an encode fails with a missing-tool error, check the dependency scan order in the section above.

## Contributing

Contributions are welcome! If you have ideas for new features or have found a bug, check out our [Contribution Guide](CONTRIBUTING.md) to get started.

## License

This project is released into the public domain under the [The Unlicense](LICENSE).