package authserver

func (s *server) routes() {
	s.router.HandleFunc("/authenticate", s.handleAuthentication())
}
