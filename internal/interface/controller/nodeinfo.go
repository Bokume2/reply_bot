package controller

import (
	"sync"

	"github.com/Bokume2/reply_bot"
	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
	"github.com/writeas/go-nodeinfo"

	"github.com/labstack/echo/v5"
)

const NodeInfoPath = "/nodeinfo/2.0"

var (
	nis     *nodeinfo.Service
	nicOnce sync.Once
)

type NodeInfoController struct{}

func NewNodeInfoController() *NodeInfoController {
	return &NodeInfoController{}
}

func (nic NodeInfoController) GetNodeInfoDiscover(c *echo.Context) error {
	niService().NodeInfoDiscover(c.Response(), c.Request())
	return nil
}

func (nic NodeInfoController) GetNodeInfoContent(c *echo.Context) error {
	niService().NodeInfo(c.Response(), c.Request())
	return nil
}

func niService() *nodeinfo.Service {
	nicOnce.Do(func() {
		niCfg := nodeinfo.Config{
			BaseURL: config.LocalOrigin().String(),
			InfoURL: NodeInfoPath,
			Metadata: nodeinfo.Metadata{
				Software: nodeinfo.SoftwareMeta{
					GitHub: "https://github.com/Bokume2/reply_bot",
				},
			},
			Protocols: []nodeinfo.NodeProtocol{
				nodeinfo.ProtocolActivityPub,
			},
			Services: nodeinfo.Services{
				Inbound:  []nodeinfo.NodeService{},
				Outbound: []nodeinfo.NodeService{},
			},
			Software: nodeinfo.SoftwareInfo{
				Name:    "reply-bot",
				Version: reply_bot.Version(),
			},
		}

		nis = nodeinfo.NewService(niCfg, niResolver{})
	})

	return nis
}

type niResolver struct{}

func (r niResolver) IsOpenRegistration() (bool, error) {
	return false, nil
}

func (r niResolver) Usage() (nodeinfo.Usage, error) {
	return nodeinfo.Usage{
		Users: nodeinfo.UsageUsers{
			Total: 1,
		},
	}, nil
}
