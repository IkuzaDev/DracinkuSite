# DracinkuToolKit

## About

DracinkuToolKit adalah aplikasi desktop untuk streaming dan download drama dari berbagai platform. Dibangun dengan Wails framework menggunakan Go backend dan Svelte frontend.

## Features

### 🎬 Platform Support
- **20+ Platform Streaming**: Dramabox, DramaMax, Anime, Movie, TV, ReelBuzz, RadReels, HiShort, DramaWave, JAV, Drakor, AnimeV2, dll
- **Multi-Language Support**: Indonesia, English, Mandarin, Korea, Jepang
- **Smart Pagination**: Handle berbagai format response API dengan robust

### 🎥 Video Player
- **Advanced Subtitle System**:
  - SRT to WebVTT conversion otomatis
  - Customizable subtitle settings (warna, ukuran, posisi, ketebalan)
  - Real-time subtitle positioning dengan slider
  - Support untuk subtitle lokal dan remote
- **Professional Controls**: Play/pause, seek, volume, fullscreen
- **Episode Navigation**: Easy switching antar episode

### 📥 Download Manager
- **Batch Download**: Download multiple episodes
- **Quality Selection**: Pilih kualitas video
- **Progress Tracking**: Monitor download progress
- **Resume Support**: Lanjutkan download yang terhenti

### 🎨 User Interface
- **Modern Design**: Dark theme dengan smooth animations
- **Responsive Layout**: Adaptif untuk berbagai ukuran layar
- **Search & Filter**: Cepat cari drama favorit
- **Library Management**: Organize koleksi drama

## Tech Stack

- **Backend**: Go dengan Wails framework
- **Frontend**: Svelte + Vite
- **Styling**: TailwindCSS
- **Video**: HTML5 Video API
- **Storage**: Local file system

## Getting Started

### Prerequisites
- Go 1.19+
- Node.js 16+
- Wails CLI

### Installation
```bash
# Clone repository
git clone <repository-url>
cd DracinkuToolKit

# Install dependencies
go mod tidy
npm install

# Run development
wails dev
```

### Development
```bash
# Live development dengan hot reload
wails dev

# Build untuk production
wails build

# Test backend
go test ./...
```

## API Endpoints

### Platform Management
- `GET /platforms` - Daftar semua platform
- `GET /platforms/{id}/tabs` - Tab list untuk platform
- `GET /platforms/{id}/feed` - Feed drama dengan pagination

### Video & Subtitle
- `GET /localfile/{path}` - Serve file lokal
- `POST /search` - Search drama
- `GET /video/{id}/detail` - Detail drama

## Configuration

### Environment Variables
```env
API_BASE_URL=https://api.dracinku.site
LICENSE_KEY=dk-demo-key-123456789
DEFAULT_LANGUAGE=id
```

### Platform Settings
Setiap platform dapat dikonfigurasi:
- Base URL
- Default language
- Custom headers
- Rate limiting

## Architecture

### Backend Structure
```
├── main.go              # Entry point
├── app.go               # Application logic
├── platforms/            # Platform implementations
│   ├── wrapper.go       # Generic API client
│   ├── dramabox.go      # Dramabox platform
│   ├── netshort.go      # NetShort platform
│   └── ...             # Other platforms
├── models/              # Data structures
├── downloader.go        # Download manager
└── config.go           # Configuration
```

### Frontend Structure
```
├── src/
│   ├── components/       # Svelte components
│   │   ├── VideoPlayer.svelte
│   │   ├── DramaList.svelte
│   │   └── ...
│   ├── routes/          # Page routing
│   └── stores/          # State management
```

## Features Detail

### Subtitle System
- **Format Support**: SRT, VTT
- **Styling Options**:
  - Font size: 0.8em - 2.0em
  - Colors: White, Yellow, Green, Cyan, Magenta, Orange
  - Background: Dark, Semi, None
  - Position: Top (0%) to Bottom (100%)
  - Weight: Normal, Bold
  - Shadow: On/Off

### Platform Integration
- **Wrapper Client**: Generic API client untuk platform serupa
- **Custom Logic**: Platform-specific implementations (NetShort, dll)
- **Error Handling**: Robust error management
- **Pagination**: Multi-format response handling

## Troubleshooting

### Common Issues
1. **Subtitle tidak tampil**: Check console untuk SRT conversion errors
2. **Platform loading gagal**: Verify API key dan network
3. **Download lambat**: Check disk space dan network speed

### Debug Mode
Enable debug logging:
```bash
DEBUG=true wails dev
```

## Contributing

1. Fork repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## License

MIT License - see LICENSE file untuk details

## Support

- 📧 Email: support@dracinku.com
- 🐛 Issues: GitHub Issues
- 📖 Documentation: Wiki

## Changelog

### v2.0.0 (Latest)
- ✅ Added 20+ platform support
- ✅ Advanced subtitle system with customization
- ✅ Improved pagination handling
- ✅ Enhanced video player controls
- ✅ Better error handling and logging

### v1.0.0
- ✅ Initial release
- ✅ Basic streaming functionality
- ✅ Download manager
