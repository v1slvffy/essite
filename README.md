# essite

Tool screenshot website yang simpel, ringan, dan mandiri berbasis Go.

Dibuat tanpa Puppeteer atau Chromium lokal, jadi gak bakal makan RAM ratusan megabyte di server. Essite mengemas web UI dan API ke dalam satu single binary (~5MB) yang siap jalan di VPS kecil (512MB RAM) atau langsung di-deploy ke Vercel Serverless.

## Kenapa Essite?

Kebanyakan tool screenshot biasanya butuh install Headless Chrome atau Puppeteer. Masalahnya, Chrome lokal gampang makan RAM 300MB sampai 1GB, sering berat di VPS kecil, dan ribet diatur di environment serverless atau container.

Essite mengambil pendekatan yang lebih praktis:
- Ringan: Penggunaan RAM cuma ~10MB saat runtime.
- Tanpa browser lokal: Render ditangani lewat pipeline multi-tier fallback (Microlink -> Thum.io -> S-Shot). Kalau provider utama lagi down atau kena rate limit, request otomatis dialihkan ke provider cadangan dalam hitungan milidetik.
- Single Binary & Vercel Ready: Backend dan frontend (HTML, CSS, JS) sudah di-embed langsung ke binary Go (`embed.FS`). Sekaligus sudah siap deploy ke Vercel via Serverless Function.
- White-label Image Stream: Dukung parameter `raw=true` buat streaming biner PNG langsung, jadi domain pihak ketiga gak bakal kelihatan sama client.

## Cara Menjalankan

### 1. Jalankan di VPS / Lokal

Syarat: Go 1.22+ terinstall.

```bash
git clone https://github.com/v1slvffy/essite.git
cd essite
go build -ldflags="-s -w" -o essite main.go
./essite
```

Buka `http://localhost:3300` di browser.

Kalau mau ubah port default:
```bash
PORT=8080 ./essite
```

### 2. Deploy ke Vercel

Repositori ini sudah include `vercel.json` dan handler `api/capture.go`.

- Lewat Vercel Dashboard: Cukup import repo GitHub ini lalu klik Deploy.
- Lewat Vercel CLI:
```bash
npm i -g vercel
vercel
```

## Penggunaan API

Endpoint utama: `/api/capture`

Method: `GET`, `POST`, `HEAD`

### Parameter

| Parameter | Tipe | Default | Keterangan |
|---|---|---|---|
| `url` | string | wajib | URL target yang mau di-screenshot |
| `device` | string | `desktop` | Pilihan viewport: `desktop` (1920x1080) atau `mobile` (414x896) |
| `fullPage` | bool | `true` | Ambil screenshot satu halaman penuh dari atas ke bawah |
| `raw` | bool | `false` | Stream file biner `image/png` langsung daripada respon JSON |

### Contoh Pemanggilan

#### 1. Dapatkan JSON Metadata
```bash
curl "http://localhost:3300/api/capture?url=https://github.com"
```

Respon JSON:
```json
{
  "status": true,
  "code": 200,
  "result": {
    "targetUrl": "https://github.com",
    "device": "desktop",
    "fullPage": true,
    "dimensions": { "width": 1920, "height": 2400 },
    "fileSize": "~150 KB",
    "imageUrl": "http://localhost:3300/api/capture?url=https%3A%2F%2Fgithub.com&device=desktop&fullPage=true&raw=true"
  }
}
```

#### 2. Download / Stream Gambar PNG Langsung (Raw Mode)
```bash
curl "http://localhost:3300/api/capture?url=https://github.com&raw=true" -o screenshot.png
```

Bisa juga langsung dipasang di tag HTML:
```html
<img src="http://localhost:3300/api/capture?url=https://github.com&raw=true" alt="Preview" />
```

## Struktur Project

```
essite/
├── api/
│   └── capture.go     # Vercel serverless function & core engine
├── public/
│   ├── index.html     # Web UI
│   ├── style.css      # Styling flat (#5A2132 & #EFE9E9)
│   └── app.js         # Client script
├── main.go            # Standalone Go HTTP server & embedded assets
├── vercel.json        # Routing Vercel
└── go.mod
```

## Lisensi

MIT
