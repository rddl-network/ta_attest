package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rddl-network/go-utils/logger"
	"github.com/rddl-network/ta_attest/config"
	"github.com/syndtr/goleveldb/leveldb"
)

type TAAService struct {
	cfg    *config.Config
	router *gin.Engine
	db     *leveldb.DB
	pmc    IPlanetmintClient
	logger logger.AppLogger
}

func NewTrustAnchorAttestationService(cfg *config.Config, db *leveldb.DB, pmc IPlanetmintClient) *TAAService {
	service := &TAAService{
		db:     db,
		cfg:    cfg,
		pmc:    pmc,
		logger: logger.GetLogger(cfg.LogLevel),
	}

	gin.SetMode(gin.ReleaseMode)
	service.router = gin.New()
	service.router.POST("/create-account", service.createAccount)
	if service.cfg.TestnetMode {
		service.router.POST("/register/:pubkey", service.postPubKey)
	}

	return service
}

func (s *TAAService) Run() (err error) {
	err = s.startWebService()
	if err != nil {
		fmt.Print(err.Error())
	}
	return err
}

func (s *TAAService) startWebService() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.ServiceBind, s.cfg.ServicePort)
	err := s.router.Run(addr)

	return err
}
