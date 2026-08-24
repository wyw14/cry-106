//go:build concurrencycheck
// +build concurrencycheck

package main

import (
	"fmt"

	"github.com/wyw14/cry-106/internal/interlock"
	"github.com/wyw14/cry-106/internal/operation"
	"github.com/wyw14/cry-106/internal/ventilation"
)

func main() {
	arbiter := interlock.NewFanArbiter()
	owners := operation.NewRegistry()
	commander := operation.NewCommander()
	svc := ventilation.NewService(arbiter, owners, commander)

	const group = "FG1"

	// V1 -> V2 switch starts first.
	transferOwner, err := svc.Transfer(group, "V2")
	fmt.Printf("transfer: operation=%s err=%v\n", transferOwner.OperationID, err)

	// While the switch holds the group, gas over-limit triggers reversal.
	reverseOwner, err := svc.Reverse(group)
	fmt.Printf("reverse:  operation=%s err=%v\n", reverseOwner.OperationID, err)

	cmds := commander.Commands()
	fmt.Printf("commands issued to fan group %s: %d\n", group, len(cmds))
	for _, c := range cmds {
		fmt.Printf("  - target=%s owner_gen=%d\n", c.Target, c.OwnerGeneration)
	}

	if err != nil {
		fmt.Println("RESULT: reversal correctly rejected while switch owns the group (no conflicting commands)")
	} else {
		fmt.Println("RESULT: BUG STILL PRESENT - reversal accepted alongside switch")
	}
}
