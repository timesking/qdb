// Copyright 2015 Reborndb Org. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package service

import (
	"crypto/rand"
	"io"
	"net"
	"sync"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/reborndb/go/atomic2"
	"github.com/reborndb/go/errors2"
	redis "github.com/reborndb/go/redis/resp"
	"github.com/reborndb/go/ring"
	"github.com/reborndb/go/sync2"
	"github.com/timesking/qdb/pkg/store"
)

type Server struct {
	mu     sync.Mutex
	h      *Handler
	closed bool
}

func NewServer(c *Config, s *store.Store) (*Server, error) {
	h, err := newHandler(c, s)
	if err != nil {
		return nil, errors.Trace(err)
	}

	server := &Server{h: h, closed: false}
	return server, nil
}

func (s *Server) Serve() error {
	if s.h == nil {
		return errors.New("empty server handler")
	}

	err := s.h.run()
	return errors.Trace(err)
}

func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	if s.h != nil {
		s.h.close()
	}

	s.closed = true
}

type Session interface {
	DB() uint32
	SetDB(db uint32)
	Store() *store.Store
}

type Handler struct {
	config *Config
	htable map[string]*command

	store *store.Store

	l net.Listener

	// replication sync master address
	masterAddr atomic2.String
	// replication sync from time
	syncSince atomic2.Int64
	// replication sync offset
	syncOffset atomic2.Int64
	// replication sync master run ID
	masterRunID string
	// replication master connection
	master chan *conn
	// replication slaveof reply channel
	slaveofReply chan struct{}
	// replication master connection state
	masterConnState atomic2.String

	signal chan int

	counters struct {
		bgsave          atomic2.Int64
		clients         atomic2.Int64
		clientsAccepted atomic2.Int64
		commands        atomic2.Int64
		commandsFailed  atomic2.Int64
		syncRdbRemains  atomic2.Int64
		syncCacheBytes  atomic2.Int64
		syncTotalBytes  atomic2.Int64
		syncFull        atomic2.Int64
		syncPartialOK   atomic2.Int64
		syncPartialErr  atomic2.Int64
	}

	// 40 bytes, hex random run id for different server
	runID []byte

	bgSaveSem *sync2.Semaphore

	repl struct {
		sync.RWMutex

		// replication backlog buffer
		backlogBuf *ring.Ring

		// global master replication offset
		masterOffset int64

		// replication offset of first byte in the backlog buffer
		backlogOffset int64

		lastSelectDB atomic2.Int64

		slaves map[*conn]chan struct{}
	}

	// conn mutex
	mu sync.Mutex

	// conn map
	conns map[*conn]struct{}
}

func newHandler(c *Config, s *store.Store) (*Handler, error) {
	h := &Handler{
		config:       c,
		master:       make(chan *conn, 0),
		slaveofReply: make(chan struct{}, 1),
		signal:       make(chan int, 0),
		store:        s,
		bgSaveSem:    sync2.NewSemaphore(1),
		conns:        make(map[*conn]struct{}),
	}

	h.runID = make([]byte, 40)
	getRandomHex(h.runID)
	log.Infof("server runid is %s", h.runID)

	l, err := net.Listen("tcp", h.config.Listen)
	if err != nil {
		return nil, errors.Trace(err)
	}
	h.l = l

	if err = h.initReplication(s); err != nil {
		h.close()
		return nil, errors.Trace(err)
	}

	h.htable = globalCommands

	go h.daemonSyncMaster()

	return h, nil
}

func (h *Handler) addConn(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.conns[c] = struct{}{}
}

func (h *Handler) removeConn(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.conns, c)
}

func (h *Handler) closeConns() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for c, _ := range h.conns {
		c.Close()
	}

	h.conns = make(map[*conn]struct{})
}

func (h *Handler) close() {
	h.l.Close()

	h.closeConns()

	h.closeReplication()

	h.store.Close()

	close(h.signal)
}

func (h *Handler) run() error {
	log.Infof("open listen address '%s' and start service", h.l.Addr())

	for {
		if nc, err := h.l.Accept(); err != nil {
			return errors.Trace(err)
		} else {
			h.counters.clientsAccepted.Add(1)
			go func() {
				h.counters.clients.Add(1)
				defer h.counters.clients.Sub(1)

				c := newConn(nc, h, h.config.ConnTimeout)

				log.Infof("new connection: %s", c)
				if err := c.serve(h); err != nil {
					if errors2.ErrorEqual(err, io.EOF) {
						log.Infof("connection lost: %s [io.EOF]", c)
					} else {
						log.Warningf("connection lost: %s, err = %s", c, err)
					}
				} else {
					log.Infof("connection exit: %s", c)
				}
			}()
		}
	}
	return nil
}

func toRespError(err error) (redis.Resp, error) {
	return redis.NewError(err), errors.Trace(err)
}

func toRespErrorf(format string, args ...interface{}) (redis.Resp, error) {
	err := errors.Errorf(format, args...)
	return toRespError(err)
}

func iconvert(args [][]byte) []interface{} {
	iargs := make([]interface{}, len(args))
	for i, v := range args {
		iargs[i] = v
	}
	return iargs
}

func getRandomHex(buf []byte) []byte {
	charsets := "0123456789abcdef"

	rand.Read(buf)

	for i := range buf {
		buf[i] = charsets[buf[i]&0x0F]
	}

	return buf
}
