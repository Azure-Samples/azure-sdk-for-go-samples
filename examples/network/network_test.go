package network

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

type NetworkSuite struct{}

var _ = chk.Suite(&NetworkSuite{})

var (
	virtualNetworkName string
	subnet1Name        = "subnet" + randname.AdjNoun{}.Generate()
	subnet2Name        = "subnet" + randname.AdjNoun{}.Generate()
	nsgName            = "nsg" + randname.AdjNoun{}.Generate()
	nicName            = "nic" + randname.AdjNoun{}.Generate()
	ipName             = "ip" + randname.AdjNoun{}.Generate()
)

func init() {
	management.GetStartParams()
	flag.StringVar(&virtualNetworkName, "vNetName", "vnet"+randname.AdjNoun{}.Generate(), "Provide a name for the virtual network to be created")
	flag.Parse()
}

func (s *NetworkSuite) TestCreateNIC(c *chk.C) {
	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	c.Check(err, chk.IsNil)
	log.Printf("group: %+v\n", group)

	network, errC := CreateVirtualNetworkAndSubnets(virtualNetworkName, subnet1Name, subnet2Name)
	c.Check(<-errC, chk.IsNil)
	log.Printf("vnet: %+v\n", <-network)

	nsg, errC := CreateNetworkSecurityGroup(nsgName)
	c.Check(<-errC, chk.IsNil)
	log.Printf("network security group: %+v\n", <-nsg)

	ip, errC := CreatePublicIp(ipName)
	c.Check(<-errC, chk.IsNil)
	log.Printf("ip address: %+v\n", <-ip)

	nic, errC := CreateNic(virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	c.Check(<-errC, chk.IsNil)
	log.Printf("nic: %+v\n", <-nic)
}
