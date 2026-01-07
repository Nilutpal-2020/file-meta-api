# Deploying File-Meta to Render

Complete guide for deploying the file-meta application to Render with Redis support.

## Why Render?

- ✅ Native Go support (no serverless limitations)
- ✅ Persistent connections
- ✅ No file size limits (unlike Vercel's 4.5MB)
- ✅ Full server runtime (not serverless)
- ✅ Free tier available
- ✅ Auto-deploy from GitHub
- ✅ Free SSL certificates

---

## Prerequisites

1. **Render Account**: Sign up at [render.com](https://render.com)
2. **GitHub Repository**: Your code pushed to GitHub
3. **Redis Database** (Optional): For distributed rate limiting
   - Use Render's free Redis instance, or
   - Your existing Redis database

---

## Deployment Steps

### Step 1: Push Code to GitHub

```bash
# Commit all changes
git add .
git commit -m "Ready for Render deployment"
git push origin main
```

### Step 2: Create Redis Instance (Optional but Recommended)

**If using Render's Redis:**

1. Go to [dashboard.render.com](https://dashboard.render.com)
2. Click "New" → "Redis"
3. Configure:
   - **Name**: `file-meta-redis`
   - **Plan**: Free (25MB, perfect for rate limiting)
   - **Region**: Choose closest to your users
4. Click "Create Redis"
5. Copy the **Internal Redis URL** (starts with `redis://`)

**If using your own Redis:**
- Use your existing Redis URL
- Ensure it's accessible from the internet
- Format: `redis://user:password@host:port/db`

### Step 3: Create Web Service

1. Go to [dashboard.render.com/create](https://dashboard.render.com/create)
2. Click "New Web Service"
3. Connect your GitHub repository
4. Select the `file-meta` repository

### Step 4: Configure Web Service

**Basic Settings:**
- **Name**: `file-meta` (or your preferred name)
- **Environment**: `Go`
- **Region**: Choose closest to your users
- **Branch**: `main`
- **Build Command**: `go build -o file-meta .`
- **Start Command**: `./file-meta`

**Instance Type:**
- **Free**: 512MB RAM, shared CPU (great for testing)
- **Starter ($7/mo)**: 512MB RAM, dedicated CPU
- **Standard**: More resources for production

### Step 5: Environment Variables

Click "Advanced" → "Add Environment Variable" for each:

#### Required Variables

| Variable | Value | Description |
|----------|-------|-------------|
| `API_KEYS` | `your_key_1,your_key_2` | Comma-separated API keys |

#### Optional Variables (with defaults)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port (auto-set by Render) |
| `REDIS_URL` | none | Redis connection URL (for distributed rate limiting) |
| `MAX_FILE_SIZE_MB` | `20` | Maximum upload size in MB |
| `RATE_LIMIT_REQUESTS` | `10` | Requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | Rate limit window (e.g., `1m`, `60s`) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `ENV` | `development` | `production` recommended |

**Example Configuration:**
```
API_KEYS=sk_prod_abc123,sk_prod_def456
REDIS_URL=redis://red-xxxxx:6379
MAX_FILE_SIZE_MB=50
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
LOG_LEVEL=info
ENV=production
```

### Step 6: Deploy

1. Click "Create Web Service"
2. Render will:
   - Clone your repository
   - Run `go build`
   - Start your service
   - Assign a URL: `https://file-meta-xxxx.onrender.com`
3. Wait 2-3 minutes for first deployment

---

## Testing Your Deployment

### Health Check

```bash
curl https://file-meta-xxxx.onrender.com/health
```

Expected response:
```json
{"status":"ok"}
```

### Upload File

```bash
curl -X POST https://file-meta-xxxx.onrender.com/v1/metadata \
  -H "X-API-Key: your_api_key" \
  -F "file=@document.pdf"
```

Expected response:
```json
{
  "filename": "document.pdf",
  "size_bytes": 1048576,
  "mime_type": "application/pdf",
  "checksum_sha256": "abc123..."
}
```

### Test Rate Limiting

```bash
# Make 11 rapid requests (if limit is 10)
for i in {1..11}; do
  echo "Request $i:"
  curl -X POST https://file-meta-xxxx.onrender.com/v1/metadata \
    -H "X-API-Key: your_api_key" \
    -F "file=@test.txt"
  echo "\n"
done
```

Expected: First 10 succeed, 11th returns `429 Too Many Requests`

---

## Using render.yaml (Alternative)

Instead of manual configuration, you can use the included `render.yaml`:

1. Push `render.yaml` to your repository
2. Go to Render dashboard
3. Click "New" → "Blueprint"
4. Connect repository
5. Render will auto-detect `render.yaml`
6. Add secret environment variables (API_KEYS, REDIS_URL)
7. Click "Apply"

This automatically creates all services with correct configuration.

---

## Render Features

### Auto-Deploy

Render automatically redeploys when you push to GitHub:

```bash
git add .
git commit -m "Update feature"
git push
# Render auto-deploys in 1-2 minutes
```

### Custom Domain

1. Go to your service settings
2. Click "Custom Domains"
3. Add your domain (e.g., `api.yourdomain.com`)
4. Update your DNS:
   ```
   CNAME api -> file-meta-xxxx.onrender.com
   ```
5. SSL certificate is automatic and free

### View Logs

1. Go to your service in Render dashboard
2. Click "Logs" tab
3. See real-time logs with:
   - Request IDs
   - Processing times
   - Errors and warnings
   - Redis connection status

### Metrics

Render provides built-in metrics:
- CPU usage
- Memory usage
- Request count
- Response times

---

## Redis Setup on Render

### Create Free Redis Instance

1. Click "New" → "Redis"
2. **Name**: `file-meta-redis`
3. **Plan**: Free (25MB)
4. **Region**: Same as web service
5. Click "Create Redis"

### Connect to Web Service

1. Copy the **Internal Redis URL**:
   ```
   redis://red-xxxxx:6379
   ```
2. Go to web service settings
3. Add environment variable:
   ```
   REDIS_URL=redis://red-xxxxx:6379
   ```
4. Save and redeploy

**Benefits:**
- Free 25MB storage (sufficient for millions of rate limit entries)
- Same region as your app (low latency)
- Automatic backups
- Automatic failover

---

## Scaling

### Vertical Scaling

Upgrade instance type in settings:
- **Free**: 512MB RAM
- **Starter ($7/mo)**: 512MB RAM, dedicated CPU
- **Standard ($25/mo)**: 2GB RAM
- **Pro**: Up to 16GB RAM

### Horizontal Scaling

1. Go to service settings
2. Increase "Instances" count
3. Load is automatically balanced
4. **Requires Redis** for rate limiting to work correctly

---

## Monitoring & Debugging

### View Logs

```bash
# Install Render CLI (optional)
brew tap render-oss/render
brew install render

# View live logs
render logs -s file-meta
```

Or use the web dashboard → Logs tab.

### Health Checks

Render automatically checks `/health` endpoint:
- If unhealthy, service is restarted
- Configure in service settings

### Alerts

Set up alerts in Settings → Notifications:
- Deploy failures
- Service crashes
- High CPU/memory

---

## Troubleshooting

### Build Fails

**Error**: `go: cannot find module`
```bash
# Solution: Ensure go.mod exists
go mod tidy
git add go.mod go.sum
git commit -m "Fix modules"
git push
```

### Service Won't Start

**Check Logs**:
1. Go to Logs tab
2. Look for initialization errors
3. Common issues:
   - Missing `API_KEYS` env var
   - Invalid `REDIS_URL`
   - Port binding issues (use Render's `$PORT`)

### Redis Connection Failed

**Error in logs**: `Redis connection failed`

**Solutions**:
1. Check `REDIS_URL` is correct
2. Ensure Redis instance is in same region
3. Use internal URL (not external/public)
4. Verify Redis instance is running

**App will still work** - falls back to in-memory rate limiting

### High Memory Usage

**Solutions**:
- Upgrade to larger instance
- Review `MAX_FILE_SIZE_MB` setting
- Check for memory leaks in logs
- Enable Redis for distributed rate limiting

---

## Cost Breakdown

### Free Tier
- **Web Service**: 750 hours/month free
- **Redis**: Free 25MB instance
- **Bandwidth**: 100GB/month
- **SSL**: Free
- **Total**: $0/month

### Production Setup
- **Web Service**: Starter ($7/mo)
- **Redis**: Starter 100MB ($7/mo)
- **Total**: $14/month

**Much more economical than Vercel Pro ($20/mo) with better flexibility!**

---

## Environment-Specific Deployments

### Production

```
ENV=production
LOG_LEVEL=warn
API_KEYS=prod_key_1,prod_key_2
REDIS_URL=redis://prod-redis:6379
```

### Staging

Create a separate web service:
- Branch: `staging`
- Environment variables with staging values
- Different API keys

---

## Backup & Recovery

### Database Backups

Redis on Render:
- Free tier: Daily backups (7 days retention)
- Paid tier: Hourly backups (30 days retention)

### Disaster Recovery

1. Code is in GitHub (version controlled)
2. Redis data is backed up
3. Can recreate service in minutes
4. Environment variables are saved in Render

---

## Security Best Practices

- ✅ Use strong API keys (32+ characters)
- ✅ Enable HTTPS only (automatic on Render)
- ✅ Set `ENV=production`
- ✅ Use `LOG_LEVEL=warn` or `error` in production
- ✅ Rotate API keys periodically
- ✅ Monitor access logs
- ✅ Use Redis password if using external Redis
- ✅ Rate limiting enabled

---

## Next Steps

1. ✅ Deploy to Render
2. ✅ Test all endpoints
3. ✅ Set up custom domain
4. ✅ Configure production API keys
5. ✅ Set up Redis for distributed rate limiting
6. ✅ Monitor logs and metrics
7. ✅ Set up alerts
8. ✅ Create staging environment
9. ✅ Document your API for users

---

## Support

- **Render Docs**: [render.com/docs](https://render.com/docs)
- **Render Community**: [community.render.com](https://community.render.com)
- **Status Page**: [status.render.com](https://status.render.com)

---

## Summary

Your API will be live at:
```
https://file-meta-xxxx.onrender.com
```

Endpoints:
- `GET /health` - Health check
- `POST /v1/metadata` - File metadata extraction (requires `X-API-Key`)

The application runs as a full HTTP server with:
- No file size limits (unlike Vercel's 4.5MB)
- Persistent connections
- Redis support for distributed rate limiting
- Auto-deploy from GitHub
- Free SSL certificates
- Built-in monitoring
