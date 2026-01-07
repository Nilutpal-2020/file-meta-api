# Deploying File-Meta to Vercel

This guide walks you through deploying the file-meta application to Vercel using the web interface.

## Prerequisites

1. **Vercel Account**: Sign up at [vercel.com](https://vercel.com)
2. **GitHub Repository**: Push your code to GitHub
3. **Redis Database**: You'll need a Redis instance (Upstash, Redis Cloud, etc.)

---

## Step 1: Push Code to GitHub

If you haven't already, push your code to a GitHub repository:

```bash
# Initialize git (if not already done)
git init

# Add all files
git add .

# Commit
git commit -m "Prepare for Vercel deployment"

# Add remote and push
git remote add origin https://github.com/YOUR_USERNAME/file-meta.git
git push -u origin main
```

---

## Step 2: Import Project to Vercel

1. **Go to Vercel Dashboard**
   - Visit [vercel.com/new](https://vercel.com/new)
   - Sign in if you haven't already

2. **Import Git Repository**
   - Click "Import Git Repository"
   - Select your GitHub account
   - Choose the `file-meta` repository
   - Click "Import"

3. **Configure Project**
   - **Project Name**: `file-meta` (or your preferred name)
   - **Framework Preset**: Other
   - **Root Directory**: `./` (leave as is)
   - **Build Command**: Leave empty (Vercel auto-detects Go)
   - **Output Directory**: Leave empty

---

## Step 3: Configure Environment Variables

Before deploying, add these environment variables:

1. Click "Environment Variables" section
2. Add each variable:

### Required Variables

| Variable | Value | Description |
|----------|-------|-------------|
| `API_KEYS` | `your_key_1,your_key_2` | Comma-separated API keys |
| `REDIS_URL` | `redis://user:pass@host:port` | Your Redis connection URL |

### Optional Variables

| Variable | Value | Default | Description |
|----------|-------|---------|-------------|
| `MAX_FILE_SIZE_MB` | `4` | `20` | Max file size (Vercel limit: 4.5MB hobby) |
| `RATE_LIMIT_REQUESTS` | `10` | `10` | Requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | `1m` | Rate limit window |
| `LOG_LEVEL` | `info` | `info` | Logging level |
| `ENV` | `production` | `development` | Environment name |

**Important**: 
- Use `Environment: Production` for production variables
- You can add different values for Preview/Development if needed

---

## Step 4: Get Your Redis URL

### Option 1: Upstash (Recommended for Vercel)

1. Go to [upstash.com](https://upstash.com)
2. Create a free account
3. Click "Create Database"
4. Choose a region close to your Vercel deployment
5. Copy the **Redis URL** (starts with `redis://`)
6. Paste it as `REDIS_URL` in Vercel

### Option 2: Your Existing Redis

If you have a global Redis instance:

1. Get your Redis connection URL
2. Format: `redis://username:password@host:port/db`
3. Or use individual variables:
   - `REDIS_HOST`
   - `REDIS_PORT`
   - `REDIS_PASSWORD`
   - `REDIS_DB`

---

## Step 5: Deploy

1. Click **"Deploy"** button
2. Wait for the build to complete (1-2 minutes)
3. Vercel will show you the deployment URL

---

## Step 6: Test Your Deployment

### Test Health Endpoint

```bash
curl https://your-app.vercel.app/api/health
```

Expected response:
```json
{"status": "ok"}
```

### Test Metadata Endpoint

```bash
curl -X POST https://your-app.vercel.app/api/metadata \
  -H "X-API-Key: your_api_key" \
  -F "file=@test.txt"
```

Expected response:
```json
{
  "filename": "test.txt",
  "size_bytes": 123,
  "mime_type": "text/plain",
  "checksum_sha256": "abc123..."
}
```

---

## Vercel Dashboard Features

### View Logs

1. Go to your project in Vercel
2. Click on a deployment
3. Click "Functions" tab
4. Click on a function to see logs

### Monitor Performance

1. Click "Analytics" tab
2. View request counts, response times, errors

### Update Environment Variables

1. Click "Settings" tab
2. Click "Environment Variables"
3. Add/Edit/Delete variables
4. Redeploy to apply changes

---

## Important Vercel Limits

### Hobby Plan (Free)
- **Request Body**: 4.5 MB max
- **Function Timeout**: 10 seconds
- **Function Memory**: 1024 MB
- **Deployments**: Unlimited

### Pro Plan
- **Request Body**: 100 MB max
- **Function Timeout**: 60 seconds (configurable)
- **Function Memory**: 3008 MB
- **Priority support**

**Recommendation**: Start with Hobby plan, upgrade if you need larger files.

---

## Troubleshooting

### Build Fails

**Error**: `go.mod not found`
- Make sure `go.mod` is in the repository root
- Run `go mod tidy` locally and commit

**Error**: `module not found`
- Run `go mod download` locally
- Commit `go.sum` file

### Function Timeout

**Error**: `FUNCTION_INVOCATION_TIMEOUT`
- Large files take time to process
- Consider upgrading to Pro plan for 60s timeout
- Or reduce `MAX_FILE_SIZE_MB`

### Redis Connection Failed

**Error**: Rate limiting not working
- Check `REDIS_URL` is correct
- Test Redis connection from your local machine
- Ensure Redis allows connections from Vercel IPs

### Rate Limit Headers Missing

- Check Redis is connected (view function logs)
- Verify `REDIS_URL` environment variable is set
- Check API key is valid

---

## Custom Domain (Optional)

1. Go to project Settings
2. Click "Domains"
3. Add your domain
4. Update DNS records as shown
5. Wait for SSL certificate (automatic)

---

## Redeployment

Vercel automatically redeploys when you push to GitHub:

```bash
# Make changes
git add .
git commit -m "Update feature"
git push

# Vercel will automatically redeploy
```

Manual redeploy:
1. Go to Deployments tab
2. Click "..." on a deployment
3. Click "Redeploy"

---

## Next Steps

- ✅ Test all endpoints thoroughly
- ✅ Monitor logs in Vercel dashboard
- ✅ Set up custom domain (optional)
- ✅ Configure production API keys
- ✅ Monitor Redis usage
- ✅ Set up error tracking (Sentry, etc.)

---

## Support

- **Vercel Docs**: [vercel.com/docs](https://vercel.com/docs)
- **Vercel Discord**: [vercel.com/discord](https://vercel.com/discord)
- **Redis Docs**: Check your Redis provider's documentation

---

## Summary

Your deployment will be live at:
```
https://your-project-name.vercel.app
```

Endpoints:
- Health: `GET /api/health`
- Metadata: `POST /api/metadata`

All requests require `X-API-Key` header (except health check).
