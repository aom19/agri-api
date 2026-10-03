package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"agri-api/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Tag-urile grupează datele pe tabele. Versiunea tag-ului intră în cheie și e
// incrementată la scriere, deci intrările vechi nu mai sunt găsite și expiră prin TTL.
const (
	TagMachines        = "machines"
	TagOperators       = "operators"
	TagImplements      = "implements"
	TagFields          = "fields"
	TagResources       = "resources"  // resources + resource_types
	TagStocks          = "stocks"     // stocks + stock_movements
	TagOperations      = "operations" // operation_types + operation_templates
	TagFieldOperations = "field_operations"
	TagCrops           = "crops"           // seasons + crops + field_crops
	TagUsers           = "users"           // users + user_profiles + roles + permissions
	TagCompatibilities = "compatibilities" // implement_compatibilities (fără rute de scriere)
	TagWeather         = "weather"         // weather_snapshots (scrise de monitorul meteo)
)

const (
	versionPrefix  = "cache:ver:"
	responsePrefix = "cache:resp:"

	// peste această limită nu se pune în cache
	maxEntrySize = 4 << 20
)

// VaryFunc adaugă la cheie o componentă per utilizator.
type VaryFunc func(c *gin.Context) string

// ResponseCache ține în Redis răspunsurile rutelor GET. Nil = cache dezactivat.
type ResponseCache struct {
	rdb        *redis.Client
	log        *logger.Logger
	defaultTTL time.Duration
}

func New(rdb *redis.Client, log *logger.Logger, defaultTTL time.Duration) *ResponseCache {
	return &ResponseCache{rdb: rdb, log: log, defaultTTL: defaultTTL}
}

func passThrough(c *gin.Context) { c.Next() }

// Cached pune în cache răspunsurile 200. Se montează după auth/permisiuni;
// ttl 0 = TTL implicit.
func (rc *ResponseCache) Cached(ttl time.Duration, vary VaryFunc, tags ...string) gin.HandlerFunc {
	if rc == nil {
		return passThrough
	}
	if ttl <= 0 {
		ttl = rc.defaultTTL
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		versions, err := rc.versions(ctx, tags)
		if err != nil {
			rc.log.Warnf("cache: citirea versiunilor a eșuat, servesc fără cache: %v", err)
			c.Next()
			return
		}

		varyKey := ""
		if vary != nil {
			varyKey = vary(c)
		}
		key := responseKey(c.Request, varyKey, versions)

		if raw, err := rc.rdb.Get(ctx, key).Bytes(); err == nil {
			if contentType, body, ok := decodeEntry(raw); ok {
				c.Header("X-Cache", "HIT")
				c.Data(http.StatusOK, contentType, body)
				c.Abort()
				return
			}
		} else if err != redis.Nil {
			rc.log.Warnf("cache: GET %s a eșuat: %v", key, err)
		}

		c.Header("X-Cache", "MISS")
		writer := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()
		c.Writer = writer.ResponseWriter

		if writer.Status() != http.StatusOK || writer.overflow {
			return
		}
		entry := encodeEntry(writer.Header().Get("Content-Type"), writer.body.Bytes())
		if err := rc.rdb.Set(ctx, key, entry, ttl).Err(); err != nil {
			rc.log.Warnf("cache: SET %s a eșuat: %v", key, err)
		}
	}
}

// Invalidates incrementează versiunea tag-urilor după un 2xx. Răspunsul pleacă
// abia după invalidare, ca clientul să nu primească date vechi.
func (rc *ResponseCache) Invalidates(tags ...string) gin.HandlerFunc {
	if rc == nil {
		return passThrough
	}

	return func(c *gin.Context) {
		writer := newBufferedWriter(c.Writer)
		c.Writer = writer
		c.Next()
		c.Writer = writer.ResponseWriter

		if writer.status >= 200 && writer.status < 300 {
			if err := rc.Invalidate(c.Request.Context(), tags...); err != nil {
				rc.log.Warnf("cache: invalidarea %v a eșuat: %v", tags, err)
			}
		}
		writer.flush()
	}
}

// Invalidate e varianta pentru scrieri din afara HTTP (ex. job-uri).
func (rc *ResponseCache) Invalidate(ctx context.Context, tags ...string) error {
	if rc == nil || len(tags) == 0 {
		return nil
	}
	pipe := rc.rdb.Pipeline()
	for _, tag := range tags {
		pipe.Incr(ctx, versionPrefix+tag)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (rc *ResponseCache) versions(ctx context.Context, tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	keys := make([]string, len(tags))
	for i, tag := range tags {
		keys[i] = versionPrefix + tag
	}
	values, err := rc.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	versions := make([]string, len(values))
	for i, value := range values {
		if s, ok := value.(string); ok {
			versions[i] = tags[i] + "=" + s
		} else {
			versions[i] = tags[i] + "=0"
		}
	}
	return versions, nil
}

// responseKey = hash(path, query sortat, vary, versiuni).
func responseKey(r *http.Request, vary string, versions []string) string {
	h := sha256.New()
	h.Write([]byte(r.URL.Path))
	h.Write([]byte{'?'})
	h.Write([]byte(r.URL.Query().Encode()))
	h.Write([]byte{'|'})
	h.Write([]byte(vary))
	h.Write([]byte{'|'})
	h.Write([]byte(strings.Join(versions, ",")))
	return responsePrefix + hex.EncodeToString(h.Sum(nil))
}

// Formatul unei intrări: "<content-type>\n<body>".
func encodeEntry(contentType string, body []byte) []byte {
	entry := make([]byte, 0, len(contentType)+1+len(body))
	entry = append(entry, contentType...)
	entry = append(entry, '\n')
	return append(entry, body...)
}

func decodeEntry(raw []byte) (string, []byte, bool) {
	i := bytes.IndexByte(raw, '\n')
	if i < 0 {
		return "", nil, false
	}
	return string(raw[:i]), raw[i+1:], true
}

// captureWriter trimite la client și păstrează o copie a body-ului.
type captureWriter struct {
	gin.ResponseWriter
	body     bytes.Buffer
	overflow bool
}

func (w *captureWriter) Write(b []byte) (int, error) {
	w.capture(b)
	return w.ResponseWriter.Write(b)
}

func (w *captureWriter) WriteString(s string) (int, error) {
	w.capture([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *captureWriter) capture(b []byte) {
	if w.overflow {
		return
	}
	if w.body.Len()+len(b) > maxEntrySize {
		w.overflow = true
		w.body.Reset()
		return
	}
	w.body.Write(b)
}

// bufferedWriter ține status-ul și body-ul în memorie până la flush().
type bufferedWriter struct {
	gin.ResponseWriter
	status  int
	body    bytes.Buffer
	written bool
}

func newBufferedWriter(w gin.ResponseWriter) *bufferedWriter {
	return &bufferedWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *bufferedWriter) WriteHeader(code int) {
	if code > 0 && !w.written {
		w.status = code
	}
}

func (w *bufferedWriter) WriteHeaderNow() { w.written = true }

func (w *bufferedWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.body.Write(b)
}

func (w *bufferedWriter) WriteString(s string) (int, error) {
	w.written = true
	return w.body.WriteString(s)
}

func (w *bufferedWriter) Status() int   { return w.status }
func (w *bufferedWriter) Written() bool { return w.written }
func (w *bufferedWriter) Flush()        {}

func (w *bufferedWriter) Size() int {
	if !w.written {
		return -1
	}
	return w.body.Len()
}

func (w *bufferedWriter) flush() {
	w.ResponseWriter.WriteHeader(w.status)
	if w.body.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.body.Bytes())
		return
	}
	w.ResponseWriter.WriteHeaderNow()
}
