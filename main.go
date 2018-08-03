package main
import (
	"fmt"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/compute"
)
func main() {
	fmt.Println("running vm test logic")
	compute.CreateResourceGroup()
}