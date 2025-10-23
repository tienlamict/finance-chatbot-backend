# CORS Configuration Guide

## Overview

This document explains the CORS (Cross-Origin Resource Sharing) configuration implemented in the finance-chatbot-backend to resolve frontend integration issues.

## Problem

The backend was experiencing CORS errors when integrating with frontend applications because:
1. No CORS middleware was configured in the Gin server
2. Browser security policies blocked cross-origin requests
3. Preflight OPTIONS requests were not handled properly

## Solution

### 1. CORS Middleware Implementation

Created `middleware/cors.go` with comprehensive CORS configuration:

- **Default Origins**: Supports common frontend development servers (React, Vue, Angular, Vite)
- **Environment-based Configuration**: Allows production origins via `CORS_ALLOWED_ORIGINS` environment variable
- **Security**: Removes localhost origins in production mode
- **Credentials Support**: Enables authentication headers and cookies

### 2. Gin Server Integration

Updated `cmd/root.go` to include CORS middleware:

```go
router.Use(middleware.CORSMiddleware())
```

The middleware is applied globally to all routes.

### 3. Docker Configuration

Updated `docker-compose.yaml` with CORS environment variables:

```yaml
CORS_ALLOWED_ORIGINS: "http://localhost:3000,http://localhost:3001,http://localhost:8080,http://localhost:4200,http://localhost:5173"
```

## Configuration Details

### Allowed Origins (Development)
- `http://localhost:3000` - React default dev server
- `http://localhost:3001` - Alternative React port
- `http://localhost:8080` - Vue.js default dev server
- `http://localhost:4200` - Angular default dev server
- `http://localhost:5173` - Vite default dev server
- `http://127.0.0.1:*` - Alternative localhost addresses

### Allowed Methods
- GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS

### Allowed Headers
- Origin, Content-Length, Content-Type, Authorization
- X-Requested-With, Accept, Accept-Encoding, Accept-Language
- Cache-Control, Connection, DNT, Host, Pragma, Referer, User-Agent

### Security Features
- **Credentials**: Enabled for authentication
- **Max Age**: 12 hours for preflight cache
- **Production Safety**: Removes localhost origins in release mode

## Environment Variables

### Development
```bash
# No additional configuration needed - uses defaults
```

### Production
```bash
CORS_ALLOWED_ORIGINS="https://yourdomain.com,https://app.yourdomain.com"
```

## Testing CORS

### 1. Browser Developer Tools
Check Network tab for:
- Preflight OPTIONS requests
- CORS headers in responses
- No CORS error messages

### 2. curl Testing
```bash
# Test preflight request
curl -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type,Authorization" \
  http://localhost:3001/v1/authenticate

# Test actual request
curl -X POST \
  -H "Origin: http://localhost:3000" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  http://localhost:3001/v1/profile
```

### 3. Frontend Integration Test
```javascript
// Test from frontend console
fetch('http://localhost:3001/v1/ping', {
  method: 'GET',
  headers: {
    'Content-Type': 'application/json',
  },
  credentials: 'include'
})
.then(response => response.json())
.then(data => console.log('CORS working:', data))
.catch(error => console.error('CORS error:', error));
```

## Troubleshooting

### Common Issues

1. **Still getting CORS errors**
   - Check if frontend origin is in allowed list
   - Verify CORS middleware is loaded
   - Check browser console for specific error messages

2. **Credentials not working**
   - Ensure `credentials: 'include'` in frontend requests
   - Verify `Access-Control-Allow-Credentials: true` in response

3. **Preflight requests failing**
   - Check OPTIONS method is allowed
   - Verify all required headers are in allowed list

### Debug Steps

1. **Check middleware loading**:
   ```bash
   # Look for CORS middleware in startup logs
   docker-compose logs app | grep -i cors
   ```

2. **Verify headers**:
   ```bash
   curl -I -H "Origin: http://localhost:3000" http://localhost:3001/v1/ping
   ```

3. **Test specific endpoints**:
   ```bash
   # Test authentication endpoint
   curl -X OPTIONS -H "Origin: http://localhost:3000" http://localhost:3001/v1/authenticate
   ```

## Security Considerations

1. **Production Origins**: Always specify exact production domains
2. **No Wildcards**: Avoid using `*` for origins in production
3. **HTTPS**: Use HTTPS origins in production
4. **Regular Updates**: Review and update allowed origins regularly

## Dependencies

- `github.com/gin-contrib/cors v1.7.6` - Official Gin CORS middleware

## Files Modified

- `middleware/cors.go` - New CORS middleware implementation
- `cmd/root.go` - Added CORS middleware to Gin router
- `docker-compose.yaml` - Added CORS environment variables
- `go.mod` - Added CORS dependency

## Next Steps

1. Test with your frontend application
2. Configure production origins in deployment
3. Monitor CORS-related logs in production
4. Consider adding CORS metrics/monitoring if needed
