package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rddl-network/ta_attest/config"
	"github.com/syndtr/goleveldb/leveldb"
)

type TAAService struct {
	cfg             *config.Config
	router          *gin.Engine
	db              *leveldb.DB
	firmwareESP32   []byte
	firmwareESP32C3 []byte
}

func NewTrustAnchorAttestationService(cfg *config.Config) *TAAService {
	libConfig.SetChainID(cfg.PlanetmintChainID)
	service := &TAAService{}
	service.cfg = cfg

	service.db = initDB(cfg.DBPath)
	gin.SetMode(gin.ReleaseMode)
	service.router = gin.New()
	service.router.GET("/firmware/:mcu", service.getFirmware)
	if service.cfg.TestnetMode {
		service.router.POST("/register/:pubkey", service.postPubKey)
	}

	return service
}

func (s *TAAService) Run() (err error) {
	s.loadFirmwares()
	err = s.startWebService()
	if err != nil {
		fmt.Print(err.Error())
	}
	return err
}

func (s *TAAService) loadFirmwares() {
	s.firmwareESP32 = loadFirmware(s.cfg.FirmwareESP32)
	s.firmwareESP32C3 = loadFirmware(s.cfg.FirmwareESP32C3)
}

func (s *TAAService) startWebService() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.ServiceBind, s.cfg.ServicePort)
	err := s.router.Run(addr)

	return err
}
