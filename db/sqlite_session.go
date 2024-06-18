package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"gitea.com/go-chi/session"

	// _ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// SqliteStore represents a sqlite session store implementation.
type SqliteStore struct {
	c    *sql.DB
	sid  string
	lock sync.RWMutex
	data map[interface{}]interface{}
}

// NewSqliteStore creates and returns a sqlite session store.
func NewSqliteStore(c *sql.DB, sid string, kv map[interface{}]interface{}) *SqliteStore {
	return &SqliteStore{
		c:    c,
		sid:  sid,
		data: kv,
	}
}

// Set sets value to given key in session.
func (s *SqliteStore) Set(key, value interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = value
	return nil
}

// Get gets value by given key in session.
func (s *SqliteStore) Get(key interface{}) interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data[key]
}

// Delete delete a key from session.
func (s *SqliteStore) Delete(key interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
	return nil
}

// ID returns current session ID.
func (s *SqliteStore) ID() string {
	return s.sid
}

// save sqlite session values to database.
// must call this method to save values to database.
func (s *SqliteStore) Release() error {
	// Skip encoding if the data is empty
	if len(s.data) == 0 {
		return nil
	}

	data, err := session.EncodeGob(s.data)
	if err != nil {
		return err
	}

	_, err = s.c.Exec("UPDATE session SET data=$1, expiry=$2 WHERE key=$3",
		data, time.Now().Unix(), s.sid)
	return err
}

// Flush deletes all session data.
func (s *SqliteStore) Flush() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = make(map[interface{}]interface{})
	return nil
}

// sqliteProvider represents a sqlite session provider implementation.
type sqliteProvider struct {
	c           *sql.DB
	maxlifetime int64
}

// Init initializes sqlite session provider.
// connStr: /path/to/db
func (p *sqliteProvider) Init(maxlifetime int64, connStr string) (err error) {
	p.maxlifetime = maxlifetime

	p.c, err = sql.Open("sqlite3", connStr)
	if err != nil {
		return err
	}
	return p.c.Ping()
}

// Read returns raw session store by session ID.
func (p *sqliteProvider) Read(sid string) (session.RawStore, error) {
	now := time.Now().Unix()
	var data []byte
	expiry := now
	err := p.c.QueryRow("SELECT data, expiry FROM session WHERE key=$1", sid).Scan(&data, &expiry)
	if err == sql.ErrNoRows {
		_, err = p.c.Exec("INSERT INTO session(key,data,expiry) VALUES($1,$2,$3)",
			sid, "", now)
	}
	if err != nil {
		return nil, err
	}

	var kv map[interface{}]interface{}
	if len(data) == 0 || expiry+p.maxlifetime <= now {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = session.DecodeGob(data)
		if err != nil {
			return nil, err
		}
	}

	return NewSqliteStore(p.c, sid, kv), nil
}

// Exist returns true if session with given ID exists.
func (p *sqliteProvider) Exist(sid string) bool {
	var data []byte
	err := p.c.QueryRow("SELECT data FROM session WHERE key=$1", sid).Scan(&data)
	if err != nil && err != sql.ErrNoRows {
		panic("session/sqlite: error checking existence: " + err.Error())
	}
	return err != sql.ErrNoRows
}

// Destroy deletes a session by session ID.
func (p *sqliteProvider) Destroy(sid string) error {
	_, err := p.c.Exec("DELETE FROM session WHERE key=$1", sid)
	return err
}

// Regenerate regenerates a session store from old session ID to new one.
func (p *sqliteProvider) Regenerate(oldsid, sid string) (_ session.RawStore, err error) {
	if p.Exist(sid) {
		return nil, fmt.Errorf("new sid '%s' already exists", sid)
	}

	if !p.Exist(oldsid) {
		if _, err = p.c.Exec("INSERT INTO session(key,data,expiry) VALUES($1,$2,$3)",
			oldsid, "", time.Now().Unix()); err != nil {
			return nil, err
		}
	}

	if _, err = p.c.Exec("UPDATE session SET key=$1 WHERE key=$2", sid, oldsid); err != nil {
		return nil, err
	}

	return p.Read(sid)
}

// Count counts and returns number of sessions.
func (p *sqliteProvider) Count() (total int) {
	if err := p.c.QueryRow("SELECT COUNT(*) AS NUM FROM session").Scan(&total); err != nil {
		panic("session/sqlite: error counting records: " + err.Error())
	}
	return total
}

// GC calls GC to clean expired sessions.
func (p *sqliteProvider) GC() {
	if _, err := p.c.Exec("DELETE FROM session WHERE strftime('%s', 'now') - expiry > ?", p.maxlifetime); err != nil {
		log.Printf("session/sqlite: error garbage collecting: %v", err)
	}
}

func init() {
	session.Register("sqlite3", &sqliteProvider{})
}
