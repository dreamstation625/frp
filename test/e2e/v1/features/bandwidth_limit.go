package features

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"

	plugin "github.com/fatedier/frp/pkg/plugin/server"
	"github.com/fatedier/frp/test/e2e/framework"
	"github.com/fatedier/frp/test/e2e/framework/consts"
	"github.com/fatedier/frp/test/e2e/mock/server/streamserver"
	pluginpkg "github.com/fatedier/frp/test/e2e/pkg/plugin"
	"github.com/fatedier/frp/test/e2e/pkg/request"
)

func loadBinaryTestPayload() []byte {
	path := filepath.Join("..", "testdata", "traffic_payload.bin")
	payload, err := os.ReadFile(path)
	framework.ExpectNoError(err)
	return payload
}

func buildAsymmetricPayload(base []byte) []byte {
	extraPercent := rand.IntN(8) + 3
	extraSize := len(base) * extraPercent / 100

	buf := make([]byte, len(base)+extraSize)
	copy(buf, base)
	if _, err := cryptorand.Read(buf[len(base):]); err != nil {
		framework.ExpectNoError(err)
	}

	return buf
}

var _ = ginkgo.Describe("[Feature: Bandwidth Limit]", func() {
	f := framework.NewDefaultFramework()

	ginkgo.It("Proxy Bandwidth Limit by Client", func() {
		serverConf := consts.DefaultServerConfig
		clientConf := consts.DefaultClientConfig
		basePayload := loadBinaryTestPayload()
		uplinkPayload := buildAsymmetricPayload(basePayload)

		localPort := f.AllocPort()
		localServer := streamserver.New(
			streamserver.TCP,
			streamserver.WithBindPort(localPort),
			streamserver.WithRespContent(basePayload),
		)
		f.RunServer("", localServer)

		remotePort := f.AllocPort()
		clientConf += fmt.Sprintf(`
			[[proxies]]
			name = "tcp"
			type = "tcp"
			localPort = %d
			remotePort = %d
			transport.bandwidthLimit = "10KB"
			`, localPort, remotePort)

		f.RunProcesses(serverConf, []string{clientConf})

		start := time.Now()
		framework.NewRequestExpect(f).Port(remotePort).RequestModify(func(r *request.Request) {
			r.Body(uplinkPayload).Timeout(30 * time.Second)
		}).ExpectResp(basePayload).Ensure()

		duration := time.Since(start)
		framework.Logf("request duration: %s", duration.String())

		framework.ExpectTrue(duration.Seconds() > 8, "100Kb with 10KB limit, want > 8 seconds, but got %s", duration.String())
	})

	ginkgo.It("Proxy Bandwidth Limit by Server", func() {
		basePayload := loadBinaryTestPayload()
		uplinkPayload := buildAsymmetricPayload(basePayload)

		// new test plugin server
		newFunc := func() *plugin.Request {
			var r plugin.Request
			r.Content = &plugin.NewProxyContent{}
			return &r
		}
		pluginPort := f.AllocPort()
		handler := func(req *plugin.Request) *plugin.Response {
			var ret plugin.Response
			content := req.Content.(*plugin.NewProxyContent)
			content.BandwidthLimit = "10KB"
			content.BandwidthLimitMode = "server"
			ret.Content = content
			return &ret
		}
		pluginServer := pluginpkg.NewHTTPPluginServer(pluginPort, newFunc, handler, nil)

		f.RunServer("", pluginServer)

		serverConf := consts.DefaultServerConfig + fmt.Sprintf(`
		[[httpPlugins]]
		name = "test"
		addr = "127.0.0.1:%d"
		path = "/handler"
		ops = ["NewProxy"]
		`, pluginPort)
		clientConf := consts.DefaultClientConfig

		localPort := f.AllocPort()
		localServer := streamserver.New(
			streamserver.TCP,
			streamserver.WithBindPort(localPort),
			streamserver.WithRespContent(basePayload),
		)
		f.RunServer("", localServer)

		remotePort := f.AllocPort()
		clientConf += fmt.Sprintf(`
			[[proxies]]
			name = "tcp"
			type = "tcp"
			localPort = %d
			remotePort = %d
			`, localPort, remotePort)

		f.RunProcesses(serverConf, []string{clientConf})

		start := time.Now()
		framework.NewRequestExpect(f).Port(remotePort).RequestModify(func(r *request.Request) {
			r.Body(uplinkPayload).Timeout(30 * time.Second)
		}).ExpectResp(basePayload).Ensure()

		duration := time.Since(start)
		framework.Logf("request duration: %s", duration.String())

		framework.ExpectTrue(duration.Seconds() > 8, "100Kb with 10KB limit, want > 8 seconds, but got %s", duration.String())
	})
})
