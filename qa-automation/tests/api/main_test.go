package api

import (
	"os"
	"testing"

	"moneyflow/qa-automation/helpers"
)

func TestMain(m *testing.M) { os.Exit(helpers.RunMain(m)) }