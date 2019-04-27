package rpc

import (
	"github.com/gofrs/uuid"
	"time"
)

type SessionData struct {
	expiration time.Time
	loginAddr  string
}

func (s *Server) addSession(session SessionData) string {
	u, _ := uuid.NewV4()
	token := u.String()

	s.sessionLock.Lock()
	s.sessions[token] = session
	s.sessionLock.Unlock()

	return token
}

func (s *Server) removeSession(token string) {
	s.sessionLock.Lock()
	delete(s.sessions, token)
	s.sessionLock.Unlock()
}

func (s *Server) checkSession(token string) bool {
	s.sessionLock.Lock()
	session, ok := s.sessions[token]
	s.sessionLock.Unlock()

	if !ok {
		return false
	}

	if time.Now().After(session.expiration) {
		s.removeSession(token)
		return false
	}

	return true
}
