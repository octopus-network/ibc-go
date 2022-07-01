package grandpa

import (
	"github.com/cosmos/ibc-go/v3/modules/light-clients/10-grandpa/types"
)

// Name returns the IBC client name
func Name() string {
	return types.SubModuleName
}
