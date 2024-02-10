package duplocloudtest

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCheckDestroy(resourceAddress string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, ok := state.RootModule().Resources[resourceAddress]
		if ok {
			return fmt.Errorf("Not destroyed in TF: %s", resourceAddress)
		}

		return nil
	}
}
