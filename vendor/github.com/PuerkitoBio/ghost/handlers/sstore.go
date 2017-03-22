package handlers

import (
	"sync"
	"time"
)

// SessionStore interface, must be implemented by any store to be used
// for session storage.
type SessionStore interface {
	Get(id string) (*Session, error) // Get the session from the store
	Set(sess *Session) error         // Save the session in the store
	Delete(id string) error          // Delete the session from the store
	Clear() error                    // Delete all sessions from the store
	Len() int                        // Get the number of sessions in the store
}

// In-memory implementation of a session store. Not recommended for production
// use.
type MemoryStore struct {
	l    sync.RWMutex
	m    map[string]*Session
	capc int
}

// Create a new memory store.
func NewMemoryStore(capc int) *MemoryStore {
	m := &MemoryStore{}
	m.capc = capc
	m.newMap()
	return m
}

// Get the number of sessions saved in the store.
func (this *MemoryStore) Len() int {
	return len(this.m)
}

// Get the requested session from the store.
func (this *MemoryStore) Get(id string) (*Session, error) {
	this.l.RLock()
	defer this.l.RUnlock()
	return this.m[id], nil
}

// Save the session to the store.
func (this *MemoryStore) Set(sess *Session) error {
	this.l.Lock()
	defer this.l.Unlock()
	this.m[sess.ID()] = sess
	if sess.IsNew() {
		// Since the memory store doesn't marshal to a string without the isNew, if it is left
		// to true, it will stay true forever.
		sess.isNew = false
		// Expire in the given time. If the maxAge is 0 (which means browser-session lifetime),
		// expire in a reasonable delay, 2 days. The weird case of a negative maxAge will
		// cause the immediate Delete call.
		wait := sess.MaxAge()
		if wait == 0 {
			wait = 2 * 24 * time.Hour
		}
		go func() {
			// Clear the session after the specified delay
			<-time.After(wait)
			this.Delete(sess.ID())
		}()
	}
	return nil
}

// Delete the specified session ID from the store.
func (this *MemoryStore) Delete(id string) error {
	this.l.Lock()
	defer this.l.Unlock()
	delete(this.m, id)
	return nil
}

// Clear all sessions from the store.
func (this *MemoryStore) Clear() error {
	this.l.Lock()
	defer this.l.Unlock()
	this.newMap()
	return nil
}

// Re-create the internal map, dropping all existing sessions.
func (this *MemoryStore) newMap() {
	this.m = make(map[string]*Session, this.capc)
}
