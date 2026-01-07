# Quick Start Guide for Vercel Deployment

Follow these steps to deploy your file-meta application to Vercel:

## Prerequisites
- [ ] GitHub account
- [ ] Vercel account (free at vercel.com)
- [ ] Redis database URL (Upstash recommended)

## Deployment Steps

### 1. Push to GitHub
```bash
git add .
git commit -m "Ready for Vercel deployment"
git push origin main
```

### 2. Import to Vercel
1. Go to [vercel.com/new]( https://vercel.com/new)
2. Click "Import Git Repository"
3. Select your `file-meta` repository
4. Click "Import"

### 3. Configure Environment Variables

Add these in Vercel dashboard before deployment:

**Required:**
- `API_KEYS` = `your_api_key_1,your_api_key_2`
- `REDIS_URL` = `redis://default:password@host:port`

**Optional:**
- `MAX_FILE_SIZE_MB` = `4` (Vercel limit: 4.5MB on free plan)
- `RATE_LIMIT_REQUESTS` = `10`
- `RATE_LIMIT_WINDOW` = `1m`
- `LOG_LEVEL` = `info`
- `ENV` = `production`

### 4. Deploy
Click "Deploy" and wait 1-2 minutes.

### 5. Test
```bash
# Health check
curl https://your-app.vercel.app/api/health

# Upload file
curl -X POST https://your-app.vercel.app/api/metadata \
  -H "X-API-Key: your_api_key" \
  -F "file=@test.txt"
```

## Get Redis (Free)

### Upstash (Recommended)
1. Go to [upstash.com](https://upstash.com)
2. Create free account
3. Create database
4. Copy Redis URL
5. Paste as `REDIS_URL` in Vercel

**That's it!** Your API is live at `https://your-project.vercel.app`

For detailed instructions, see [docs/VERCEL_DEPLOYMENT.md](docs/VERCEL_DEPLOYMENT.md)
