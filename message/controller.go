package message

import "github.com/VictorAnnell/kandidat-backend/rediscli"

type Controller struct {
	r *rediscli.Redis
}

func NewController(r *rediscli.Redis) *Controller {
	return &Controller{
		r: r,
	}
}
