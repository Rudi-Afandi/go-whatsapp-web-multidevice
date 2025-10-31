# 🚀 WhatsApp Web Multi-Device dengan Next.js Auto-Connect

## 🔥 Fitur Utama

- ✅ **Web via /web**: Next.js diakses melalui http://localhost:3000/web
- ✅ **Single Port**: Semua service berjalan di port 3000
- ✅ **Shared APP_BASIC_AUTH**: Authentication otomatis terintegrasi
- ✅ **One-Command Startup**: Menjalankan aplikasi dengan satu perintah
- ✅ **API Proxy**: Next.js API routes di-proxy ke backend Go

## 🛠 Cara Penggunaan

### Single Command Setup
```bash
cd src
# Set credentials
echo "APP_BASIC_AUTH=fanfanra:K4gur@aa" > .env

# Run aplikasi
go run . rest
```

### Akses Aplikasi
- **📱 Web Interface**: http://localhost:3000/web (Next.js app)
- **🔗 API Documentation**: http://localhost:3000 (Go backend)
- **📊 Dashboard**: http://localhost:3000/ (main dashboard)

### Build & Run
```bash
cd src
go build -o whatsapp
./whatsapp rest
```

## 📁 Struktur Project

```
src/
├── .env                    # Backend credentials (shared)
├── .env.example           # Template environment variables
├── cmd/rest.go            # Modified to auto-start Next.js
├── config/settings.go     # Go backend configuration
├── whatsapp-web/          # Next.js frontend
│   ├── .env.local         # Minimal frontend config
│   ├── lib/api-config.ts  # Smart API configuration
│   └── app/api/           # API routes with shared auth
└── ...
```

## 🔗 Konfigurasi Auto-Connect

### Backend (Go)
- **Port**: 3000 (sesuai config `APP_PORT`)
- **Authentication**: `APP_BASIC_AUTH` (shared dengan Next.js)
- **Auto-exports**: `APP_BASIC_AUTH` otomatis dibagi ke Next.js

### Frontend (Next.js)
- **Port**: 3001 (default Next.js dev server)
- **Auto-detect**: Mencoba connect ke `http://localhost:3000`
- **Authentication**: Menggunakan `APP_BASIC_AUTH` dari backend
- **Fallback**: Gunakan `NEXT_PUBLIC_WHATSAPP_API_URL` jika lokal gagal

## 🔧 Environment Variables

### Backend (.env)
```bash
# Application Settings
APP_PORT=3000
APP_DEBUG=true
APP_OS=AldinoKemal
APP_BASIC_AUTH=fanfanra:K4gur@aa
APP_BASE_PATH=

# Database Settings
DB_URI="file:storages/whatsapp.db?_foreign_keys=on"
DB_KEYS_URI="file::memory:?cache=shared&_foreign_keys=on"
```

### Frontend (.env.local)
```bash
# Auto-connect to local Go backend
NEXT_PUBLIC_WHATSAPP_API_URL=http://localhost:3000
```

## 🔐 Authentication System

### Single Source of Truth
- ✅ **Backend**: `APP_BASIC_AUTH=fanfanra:K4gur@aa`
- ✅ **Frontend**: Otomatis menerima dari backend
- ✅ **Format**: `username:password` (support multiple: `user1:pass1,user2:pass2`)
- ✅ **Auto-export**: Go backend export ke Next.js environment

## 🎯 Cara Kerja Auto-Connect

1. **Go REST API** starts di port 3000
2. **Next.js** otomatis detect dan connect ke backend
3. **Credentials** otomatis dibagi melalui environment variables
4. **Smart fallback** jika backend lokal tidak tersedia

### Console Logs Examples:
```
✅ Connected to local Go backend: http://localhost:3000
❌ Local backend not found, falling back to environment URL
📡 Using backend URL: https://your-remote-api.com
```

## 🔄 API Routes

### Frontend API Routes (auto-authenticated dengan APP_BASIC_AUTH)
- `GET /api/chats` - Mendapatkan daftar chat
- `GET /api/messages` - Mendapatkan pesan
- `POST /api/send-message` - Mengirim pesan

### Backend API Routes
- `GET /chats` - Backend chats endpoint
- `GET /chat/{jid}/messages` - Backend messages endpoint
- `POST /send/message` - Backend send message endpoint
- `GET /app/auth-info` - Auth info endpoint

## 🛡 Security Features

- ✅ **Single source of truth** - `APP_BASIC_AUTH` di backend
- ✅ **No hardcoded credentials** di frontend
- ✅ **Environment-based authentication**
- ✅ **Basic Auth** protection di backend
- ✅ **No credential exposure** di client-side code
- ✅ **Auto-export credentials** dari backend ke frontend

## 🚀 Deployment Notes

### Development
```bash
cd src
go run . rest
```

### Production
```bash
cd src
go build -o whatsapp
./whatsapp rest
```

Both applications will start automatically:
- Go API: http://localhost:3000
- Next.js: http://localhost:3001

## 🔍 Troubleshooting

### Next.js tidak bisa connect ke backend?
1. Pastikan Go API berjalan di port 3000
2. Check firewall settings
3. Verify environment variables

### Credentials tidak bekerja?
1. Check `.env` file di backend
2. Pastikan `APP_BASIC_AUTH` ter-set dengan format `username:password`
3. Restart aplikasi

### Port conflict?
1. Ubah `APP_PORT` di backend
2. Ubah Next.js port dengan: `npm run dev -- -p 3002`