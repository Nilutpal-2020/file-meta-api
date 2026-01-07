# Quick Start - Deploy to Render

Deploy file-meta to Render in 5 minutes.

## Prerequisites
- [ ] Render account (free at [render.com](https://render.com))
- [ ] GitHub repository with your code

## Steps

### 1. Push to GitHub
```bash
git add .
git commit -m "Deploy to Render"
git push origin main
```

### 2. Create Web Service
1. Go to [dashboard.render.com/create](https://dashboard.render.com/create)
2. Click "New Web Service"
3. Connect your GitHub `file-meta` repository

### 3. Configure
- **Build Command**: `go build -o file-meta .`
- **Start Command**: `./file-meta`
- **Instance Type**: Free (for testing)

### 4. Add Environment Variables
**Required:**
- `API_KEYS` = `your_api_key_1,your_api_key_2`

**Optional:**
- `REDIS_URL` = `redis://...` (for distributed rate limiting)
- `MAX_FILE_SIZE_MB` = `20`
- `ENV` = `production`

### 5. Deploy
Click "Create Web Service" and wait 2-3 minutes.

## Test Your API

```bash
# Health check
curl https://your-app.onrender.com/health

# Upload file
curl -X POST https://your-app.onrender.com/v1/metadata \
  -H "X-API-Key: your_api_key" \
  -F "file=@test.txt"
```

## Optional: Add Redis (Free)

1. In Render dashboard: "New" → "Redis"
2. Plan: Free (25MB)
3. Copy Internal Redis URL
4. Add to web service: `REDIS_URL=redis://...`

**Done!** Your API is live at `https://your-app.onrender.com`

For detailed instructions, see [docs/RENDER_DEPLOYMENT.md](docs/RENDER_DEPLOYMENT.md)
