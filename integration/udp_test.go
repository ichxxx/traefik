package integration

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type UDPSuite struct{ BaseSuite }

func TestUDPSuite(t *testing.T) {
	suite.Run(t, new(UDPSuite))
}

func (s *UDPSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("udp")
	s.composeUp()
}

func (s *UDPSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func guessWhoUDP(addr string) (string, error) {
	var conn net.Conn
	var err error

	udpAddr, err2 := net.ResolveUDPAddr("udp", addr)
	if err2 != nil {
		return "", err2
	}

	conn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return "", err
	}

	_, err = conn.Write([]byte("WHO"))
	if err != nil {
		return "", err
	}

	outCh := make(chan []byte)
	errCh := make(chan error)
	go func() {
		out := make([]byte, 2048)
		n, err := conn.Read(out)
		if err != nil {
			errCh <- err
			return
		}
		outCh <- out[:n]
	}()

	select {
	case out := <-outCh:
		return string(out), nil
	case err := <-errCh:
		return "", err
	case <-time.After(5 * time.Second):
		return "", errors.New("timeout")
	}
}

func (s *UDPSuite) TestWRR() {
	file := s.adaptFile("fixtures/udp/wrr.toml", struct {
		WhoamiAIP string
		WhoamiBIP string
		WhoamiCIP string
		WhoamiDIP string
	}{
		WhoamiAIP: s.getComposeServiceIP("whoami-a"),
		WhoamiBIP: s.getComposeServiceIP("whoami-b"),
		WhoamiCIP: s.getComposeServiceIP("whoami-c"),
		WhoamiDIP: s.getComposeServiceIP("whoami-d"),
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("whoami-a"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8093/who", 5*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	stop := make(chan struct{})
	go func() {
		call := map[string]int{}
		for i := 0; i < 8; i++ {
			out, err := guessWhoUDP("127.0.0.1:8093")
			require.NoError(s.T(), err)
			switch {
			case strings.Contains(out, "whoami-a"):
				call["whoami-a"]++
			case strings.Contains(out, "whoami-b"):
				call["whoami-b"]++
			case strings.Contains(out, "whoami-c"):
				call["whoami-c"]++
			default:
				call["unknown"]++
			}
		}
		assert.EqualValues(s.T(), map[string]int{"whoami-a": 3, "whoami-b": 2, "whoami-c": 3}, call)
		close(stop)
	}()

	select {
	case <-stop:
	case <-time.Tick(5 * time.Second):
		log.Info().Msg("Timeout")
	}
}

func (s *UDPSuite) TestMiddlewareAllowList() {
	file := s.adaptFile("fixtures/udp/ip-allowlist.toml", struct {
		WhoamiA string
		WhoamiB string
	}{
		WhoamiA: s.getComposeServiceIP("whoami-a"),
		WhoamiB: s.getComposeServiceIP("whoami-b"),
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	stop := make(chan struct{})
	go func() {
		_, err := guessWhoUDP("127.0.0.1:8093")
		assert.EqualError(s.T(), err, "timeout")

		out, err := guessWhoUDP("127.0.0.1:8094")
		require.NoError(s.T(), err)
		assert.Contains(s.T(), out, "whoami-b")
		close(stop)
	}()

	select {
	case <-stop:
	case <-time.Tick(10 * time.Second):
		log.Info().Msg("Timeout")
	}
}
