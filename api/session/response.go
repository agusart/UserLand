package session

import "userland/store/postgres"

type ListSessionResponse struct {
	Session [] postgres.Session `json:"data"`
}
