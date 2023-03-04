package types

import (
	"os"

	// log "github.com/go-kit/log"
	"github.com/tendermint/tendermint/libs/log"
)


// var logger = log.Logger.With("light-client/10-grandpa/client_state")
var Logger = log.NewTMLogger(os.Stderr)