package middleware

import (
	"sync"
	"time"
)

const MaxBlacklistSize = 10000 // Maximum entries

type blacklistEntry struct {
	expiry   time.Time
	accessed time.Time
}

type LRUBlacklist struct {
	mu          sync.RWMutex
	entries     map[string]blacklistEntry
	accessQueue []string // LRU queue order
}

var blacklistInstance *LRUBlacklist

// InitTokenBlacklist 启动后台清理任务
func InitTokenBlacklist() {
	blacklistInstance = &LRUBlacklist{
		entries:     make(map[string]blacklistEntry),
		accessQueue: make([]string, 0),
	}
	go cleanupExpiredTokens()
}

// AddToBlacklist adds token to blacklist with eviction
func AddToBlacklist(token string, expiration time.Duration) {
	blacklistInstance.mu.Lock()
	defer blacklistInstance.mu.Unlock()
	
	now := time.Now()
	newExpiry := now.Add(expiration)
	
	// Check if token already exists
	if _, exists := blacklistInstance.entries[token]; exists {
		blacklistInstance.entries[token] = blacklistEntry{expiry: newExpiry, accessed: now}
		return
	}
	
	// Evict oldest entries if at capacity
	for len(blacklistInstance.entries) >= MaxBlacklistSize {
		if len(blacklistInstance.accessQueue) > 0 {
			oldest := blacklistInstance.accessQueue[0]
			blacklistInstance.accessQueue = blacklistInstance.accessQueue[1:]
			delete(blacklistInstance.entries, oldest)
		} else {
			break
		}
	}
	
	blacklistInstance.entries[token] = blacklistEntry{expiry: newExpiry, accessed: now}
	blacklistInstance.accessQueue = append(blacklistInstance.accessQueue, token)
}

// IsBlacklisted checks if token is blacklisted and not expired
func IsBlacklisted(token string) bool {
	blacklistInstance.mu.RLock()
	defer blacklistInstance.mu.RUnlock()
	
	entry, exists := blacklistInstance.entries[token]
	if !exists {
		return false
	}
	
	// Update access time for LRU
	entry.accessed = time.Now()
	blacklistInstance.mu.RUnlock()
	
	blacklistInstance.mu.Lock()
	blacklistInstance.entries[token] = entry
	blacklistInstance.mu.Unlock()
	
	return time.Now().After(entry.expiry)
}

// cleanupExpiredTokens 定期清理过期 token (collect then delete pattern)
func cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		var toDelete []string
		
		// Collect expired tokens (holding lock briefly)
		blacklistInstance.mu.Lock()
		now := time.Now()
		for token, entry := range blacklistInstance.entries {
			if now.After(entry.expiry) {
				toDelete = append(toDelete, token)
			}
		}
		blacklistInstance.mu.Unlock()
		
		// Delete outside main lock section
		if len(toDelete) > 0 {
			blacklistInstance.mu.Lock()
			for _, token := range toDelete {
				delete(blacklistInstance.entries, token)
			}
			blacklistInstance.mu.Unlock()
		}
	}
}
