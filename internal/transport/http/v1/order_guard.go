package v1

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

const (
	guestOrderRateWindow      = 15 * time.Minute
	guestOrderMaxPerIP        = 5
	guestOrderDuplicateWindow = 10 * time.Minute
)

type guestOrderGuard struct {
	mu                sync.Mutex
	requestsByIP      map[string][]time.Time
	recentOrderByHash map[string]time.Time
}

func newGuestOrderGuard() *guestOrderGuard {
	return &guestOrderGuard{
		requestsByIP:      make(map[string][]time.Time),
		recentOrderByHash: make(map[string]time.Time),
	}
}

func (g *guestOrderGuard) Check(c *gin.Context, dto domain.CreateOrderDTO) error {
	ip := requestIP(c)
	if ip == "" {
		ip = "unknown"
	}

	now := time.Now()
	orderHash := buildGuestOrderHash(dto)

	g.mu.Lock()
	defer g.mu.Unlock()

	g.cleanup(now)

	recentRequests := append([]time.Time(nil), g.requestsByIP[ip]...)
	if len(recentRequests) >= guestOrderMaxPerIP {
		return fmt.Errorf("too many order attempts from one IP, please try again later")
	}

	if submittedAt, exists := g.recentOrderByHash[orderHash]; exists && now.Sub(submittedAt) < guestOrderDuplicateWindow {
		return fmt.Errorf("a similar order was already submitted recently, please wait before retrying")
	}

	g.requestsByIP[ip] = append(recentRequests, now)
	g.recentOrderByHash[orderHash] = now
	return nil
}

func (g *guestOrderGuard) cleanup(now time.Time) {
	for ip, timestamps := range g.requestsByIP {
		filtered := timestamps[:0]
		for _, ts := range timestamps {
			if now.Sub(ts) < guestOrderRateWindow {
				filtered = append(filtered, ts)
			}
		}
		if len(filtered) == 0 {
			delete(g.requestsByIP, ip)
			continue
		}
		g.requestsByIP[ip] = filtered
	}

	for hash, ts := range g.recentOrderByHash {
		if now.Sub(ts) >= guestOrderDuplicateWindow {
			delete(g.recentOrderByHash, hash)
		}
	}
}

func buildGuestOrderHash(dto domain.CreateOrderDTO) string {
	normalizedPhone := normalizePhone(dto.CustomerPhone)
	items := make([]string, 0, len(dto.Items))
	for _, item := range dto.Items {
		items = append(items, item.LotID.String()+":"+fmt.Sprintf("%d", item.Quantity))
	}
	sort.Strings(items)
	return normalizedPhone + "|" + strings.Join(items, ",")
}

func normalizePhone(phone string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", "+", "")
	return replacer.Replace(strings.TrimSpace(strings.ToLower(phone)))
}

func requestIP(c *gin.Context) string {
	if ip := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if forwardedFor := strings.TrimSpace(c.GetHeader("X-Forwarded-For")); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if ip := strings.TrimSpace(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}
	return c.ClientIP()
}
