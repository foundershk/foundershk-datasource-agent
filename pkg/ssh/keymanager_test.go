
package ssh_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/go-kit/log"

	"github.com/grafana/pdc-agent/pkg/pdc"
	"github.com/grafana/pdc-agent/pkg/ssh"
	"github.com/mikesmitty/edkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	knownHosts   = `known hosts`
	expectedCert = `
-----BEGIN CERTIFICATE-----
c3NoLWVkMjU1MTktY2VydC12MDFAb3BlbnNzaC5jb20gQUFBQUlITnphQzFsWkRJ
MU5URTVMV05sY25RdGRqQXhRRzl3Wlc1emMyZ3VZMjl0QUFBQUlESlIvSnNPT1Ev
UWlkdGhOVWZ3aUZoM0tDSHcySXpGaHI1dVNmOWJVR1pUQUFBQUlFMS9MRHBGd0Fl
bit6WFZNcTZuZmpBaEFtL1NpM3ZpaFJjd3ZrdG1YQUtuQUFBQUFBQUFBQUFBQUFB
Q0FBQUFBemN3TlFBQUFINEFBQUE3Y0hKcGRtRjBaUzFrWVhSaGMyOTFjbU5sTFdO
dmJtNWxZM1F1YUc5emRHVmtMV2R5WVdaaGJtRXVjM1pqTG1Oc2RYTjBaWEl1Ykc5
allXd0FBQUE3Y0hKcGRtRjBaUzFrWVhSaGMyOTFjbU5sTFdOdmJtNWxZM1F0WkdW
MkxYVnpMV05sYm5SeVlXd3RNQzVuY21GbVlXNWhMV1JsZGk1dVpYUUFBQUFBWkZP
SExBQUFBQUJrVTVVOEFBQUFBQUFBQUFBQUFBQUFBQUFDRndBQUFBZHpjMmd0Y25O
aEFBQUFBd0VBQVFBQUFnRUE5R0MzZUVjREpzYnFMQnVnMWMvQmVsUW5uNEdGYWxP
KzdJV2ZwdmU2YU9oYi8xVGRnNnVMMkRjRnRYMTlINGdycU1FV1paV0lvNHZQdHV3
UGZHQ3Rod000cWY2ZFNocUpCcC9KZDg2aENwOENTRldDZFBQNVpVWVB3RHpsNStE
ZG9zOExYVEF1czZXSWxxcGliRmJXS05NZkNTbld5M3J3UHRKeTEvbFhwT0FKenFE
VC80SWdhZFNDM0MyUFo2L1lpUzN2anJWazdFS0VKclc2Yk9oQzI3TGcybkZNVzgw
WEt5L0FsVktGa0k2OFV6Rll4QzMxbTd0VzkxOWNTOS9Gc1pFQWd4ZFdJU1VUVlg4
UW5zbHdzRUN4OUlhNmxKbU5RQ1lMU3Q1d2NaeVloOFV0T21UbDFrZjlRdGhjcXRv
Z3UybmhXRHRsWlp5cVpRS0tYaUJaRzl5YTl2WVZYdmUzbzcvUGJqNklHbFdybkFZ
ZVB4YSs4ZzdFNmY2aFMwQ3lmZExEb1BweFJFYTlzdGxFRjk2am00bC9zcUUwTCta
OVRjb0FzNTI5b0xQMkFkRStzK2xiWHR1ZlJjNHh4cWJJSW04TGlVY0pEa0NYZ0V3
MnlpK3crTFNaMUhMRGFXelVkVzVFcmgvZC9qbXV6elZyZWNaL0p3clFEem5KNFp5
VzJXUEtpTmY1bExLYkhyR2I4aFpoUEphRFVNOTlJMkVNbmNlbDNLOFlkYjl2YTFP
ZnB2TWI5SjNpcVlmTEs4dm4rSEJZNGE5eXhIWGcwNEZwV3VtR2pvaTBINkJkelFL
TkpNcUNFNVBOS0RicU1NeDc4cjRGUVJtNmlGaTdvdVRJUGRsU3FCdmt6ellIaXZQ
UFJCMGFUbWV4OHJNMFBtMXNrSnNMelpWc1Bsa0FBQUlVQUFBQURISnpZUzF6YUdF
eUxUVXhNZ0FBQWdDRkdyUTZVNHFHSVJXZE1rYlBIRCt4NDRaNFhsR08vTUhVemxP
SEtOM0gyRVIzWkxpWFJHazFmclhKU1Y5enhmb1lxOWY2TXdETU85QnZsMnJValRy
bGtwdTByaWE1cjYvSVYrZ3F2OHJXMHpNSUxkeWUyZnBqRXhlT3BReFdqU0pQeVI1
Q0NtWFFtRlkwblEwK1dNQjNiNGQ4ZHNNMGcxakc3aGhhdkk2UUdsa2MvUmpJckg1
QVZ3Z0I3dS80a2hUZE5aS3V1OE1KTkptNWprTkhUaEo2ekNLVi80SXl0dnl1MXpv
cGUxemdBTnF2K1NHd2lIS0FXUzh4N0podG9QMWhOTFpKSHRKOGVteHVjVitlRXZZ
STdQcjFZVzdkc0VUamhDQkUrZUpXd0ZBamYydUJtb1JGcEU1TzhHekg0aW91eEsw
VDJ1OTNSK09ycnNNSTlyS215bk5OcVZGcXd3VU0rUU9Sa0tIbFRoblo0K29zQ2o4
ejdzM3RnYUh4c1FkRW1mNFFEZ0ZBWnVlejlnLzJTYSsxeUhvNklURUs5Q1ZYOVJz
aTFTdElFKzVxWjF0alFjTUtqbDZ0OVU4RGhwdXFKaW5WQnBiN3NjYkVmRVlNcXR6
bTdaRzZBVmlGSm9vMjRMYkJxMi9MdEFwYUFpVU51c2ZYSUt1aTZhUlRuNlhyb0NN
WkFZRERrbkJsS0EwOC9IbHZJYko2VEZ3T2VFbzVtTjhKN3hhSUZ4Zk9PZUNQdFho
RnVYTEQrSmlyOEhuZWZyLzVVOTJjQ0dCS1VGOURYSDhQc1RYR1QxWWNQMkpGRXZL
QW1RbmNCaFJzZE4rblR0WjJ3T2NNaFpyTkpkbFdoWHlrNUNvcnYxTXhiZVBPTUFK
azl0ZGNvOFFqN0pIcFR0WnFBRm12c1E9PQo=
-----END CERTIFICATE-----
`
)

// Contains a KeyManager that can be used for testing
// and the values used to create it.
type testKeyManagerOutput struct {
	pdcCfg pdc.Config
	sshCfg *ssh.Config
	km     *ssh.KeyManager
}

// Instantiates and returns a KeyManager that can be used for testing.
func testKeyManager(t *testing.T) testKeyManagerOutput {
	t.Helper()

	// create default configs
	pdcCfg := pdc.Config{HostedGrafanaID: "1"}
	sshCfg := ssh.DefaultConfig()
	sshCfg.PDC = pdcCfg

	sshCfg.KeyFile = path.Join(t.TempDir(), "testkey")

	url, _ := mockPDC(t, http.MethodPost, "/pdc/api/v1/sign-public-key", http.StatusOK)
	pdcCfg.URL = url

	logger := log.NewNopLogger()

	client, err := pdc.NewClient(&pdcCfg, logger)
	require.Nil(t, err)

	return testKeyManagerOutput{
		pdcCfg: pdcCfg,
		sshCfg: sshCfg,
		km:     ssh.NewKeyManager(sshCfg, logger, client),
	}
}

func TestKeyManager_CreateKeys(t *testing.T) {
	t.Parallel()

	t.Run("ssh key pairs are reused by default, a new ssh pair is not created each time", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		sut := testKeyManager(t)

		// The first call to CreateKeys will create a new ssh pair.
		assert.NoError(t, sut.km.CreateKeys(ctx))

		// Read the private key that was just created.
		key1, err := os.ReadFile(sut.sshCfg.KeyFile)
		assert.NotEmpty(t, key1)
		assert.NoError(t, err)

		// The second call to CreateKeys will see that a ssh pair already exists
		// and it'll not create a new one.
		assert.NoError(t, sut.km.CreateKeys(ctx))

		// Read the key again, it should be the same key we read before.
		key2, err := os.ReadFile(sut.sshCfg.KeyFile)
		assert.NoError(t, err)
		assert.NotEmpty(t, key2)

		assert.Equal(t, key1, key2)
	})

	t.Run("a flag can be used to force a new ssh pair to be generated, should generate a new ssh key pair even if a key pair already exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		sut := testKeyManager(t)

		// Force the creation of a new ssh key pair.
		sut.sshCfg.ForceKeyFileOverwrite = true

		// The first call to CreateKeys will create a new ssh pair.
		assert.NoError(t, sut.km.CreateKeys(ctx))

		// Read the private key that was just created.
		key1, err := os.ReadFile(sut.sshCfg.KeyFile)
		assert.NotEmpty(t, key1)
		assert.NoError(t, err)

		// The second call to CreateKeys will create a new ssh key pair even though a key pair already exists.
		assert.NoError(t, sut.km.CreateKeys(ctx))

		// Read the private key that was just created.
		key2, err := os.ReadFile(sut.sshCfg.KeyFile)
		assert.NotEmpty(t, key2)
		assert.NoError(t, err)

		// A new key should have been generated.
		assert.NotEqual(t, key1, key2)
	})