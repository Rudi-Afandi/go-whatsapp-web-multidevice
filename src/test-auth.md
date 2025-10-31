# 🔐 Authentication Testing Guide

## Cara Test Authentication

### 1. Setup Environment
```bash
cd src
echo "APP_BASIC_AUTH=fanfanra:K4gur@aa" > .env
```

### 2. Run Application
```bash
go run . rest
```

### 3. Test URLs

#### ✅ Web Interface (via /web)
- **URL**: http://localhost:3000/web
- **Auth**: Browser akan prompt untuk username/password
- **Credentials**: fanfanra / K4gur@aa

#### ✅ API Endpoints (direct)
```bash
# Test dengan curl - harus return 200
curl -u "fanfanra:K4gur@aa" http://localhost:3000/chats

# Test tanpa auth - harus return 401
curl http://localhost:3000/chats
```

#### ✅ Next.js API Routes (via /web proxy)
```bash
# Test Next.js API routes via Go backend proxy
curl -u "fanfanra:K4gur@aa" http://localhost:3000/web/api/chats

# Ini akan di-proxy ke: http://localhost:3000/chats
```

### 4. Expected Behavior

1. **Web Interface**:
   - http://localhost:3000/ → Dashboard utama (link ke /web)
   - http://localhost:3000/web → Next.js app dengan shared auth

2. **API Authentication**:
   - ✅ Valid auth → 200 OK
   - ❌ Invalid auth → 401 Unauthorized
   - ❌ No auth → 401 Unauthorized

3. **Next.js Integration**:
   - Next.js API routes work via /web/* prefix
   - Shared APP_BASIC_AUTH credentials
   - Auto-detect backend URL

### 5. Debug Logs

Lihat console output untuk:
```
🚀 WhatsApp API Server starting...
📱 Web Interface: http://localhost:3000/web
🔗 API Documentation: http://localhost:3000
Next.js app configured to serve via /web route
```

### 6. Troubleshooting

#### 401 Unauthorized?
- Check .env file ada APP_BASIC_AUTH
- Verify format: username:password
- Restart aplikasi

#### /web not found?
- Check whatsapp-web directory exists
- Run `npm run build` di whatsapp-web folder
- Restart aplikasi

#### Next.js API not working?
- Check proxyToNextJSAPI function
- Verify CORS headers include Authorization
- Check network tab in browser dev tools