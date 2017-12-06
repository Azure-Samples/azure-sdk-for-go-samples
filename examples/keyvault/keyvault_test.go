package keyvault

import (
	"flag"
	"log"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/examples/resources"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/marstr/randname"
	chk "gopkg.in/check.v1"
)

func Test(t *testing.T) { chk.TestingT(t) }

type KeyVaultSuite struct{}

var _ = chk.Suite(&KeyVaultSuite{})

var (
	keyValutName string
)

func init() {
	management.GetStartParams()
	flag.StringVar(&keyValutName, "keyValutName", "valut"+randname.AdjNoun{}.Generate(), "Provide a name for the keyvault to be created")
	flag.Parse()
}
func (s *KeyVaultSuite) TestSetVaultPermissions(c *chk.C) {
	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	c.Check(err, chk.IsNil)
	log.Printf("created group: %v\n", group)

	v, err := CreateVault(keyValutName)
	c.Check(err, chk.IsNil)
	log.Printf("created vault: %v\n", v)

	v, err = SetVaultPermissions(keyValutName)
	c.Check(err, chk.IsNil)
	log.Printf("set vault permissions: %v\n", v)
}
