package wiree

import (
	h "auth/controller/handler"
	s "auth/controller/service"
	re "auth/controller/repo"

	"github.com/google/wire"

	r "auth/infra/redis"
	p "auth/infra/postgre"	
)

func initializeApp() (*h.Handler, func(), error) {
	wire.Build(
		h.InitHandler,
		s.InitService,
		re.InitRepo,

		r.ProviderCTX,
		r.Init,
		p.ProviderConnStr,
		p.Init,
	)
	return nil, nil, nil
}