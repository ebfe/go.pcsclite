package scard_test

import (
	"fmt"
	"github.com/ebfe/go.pcsclite/scard"
	"os"
)

func die(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func waitUntilCardPresent(ctx *scard.Context, readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))
	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.STATE_UNAWARE
	}

	for {
		for i := range rs {
			if rs[i].EventState&scard.STATE_PRESENT != 0 {
				return i, nil
			}
		}
		err := ctx.GetStatusChange(rs, scard.INFINITE)
		if err != nil {
			return -1, err
		}
	}

	panic("unreachable")
}

func Example() {

	// Establish a context
	ctx, err := scard.EstablishContext()
	if err != nil {
		die(err)
	}
	defer ctx.Release()

	// List available readers
	readers, err := ctx.ListReaders()
	if err != nil {
		die(err)
	}

	fmt.Printf("Found %d readers:\n", len(readers))
	for i, reader := range readers {
		fmt.Printf("[%d] %s\n", i, reader)
	}

	if len(readers) > 0 {

		fmt.Println("Waiting for a Card")
		index, err := waitUntilCardPresent(ctx, readers)
		if err != nil {
			die(err)
		}

		// Connect to card
		fmt.Println("Connecting to card in ", readers[index])
		card, err := ctx.Connect(readers[index], scard.SHARE_EXCLUSIVE, scard.PROTOCOL_ANY)
		if err != nil {
			die(err)
		}
		defer card.Disconnect(scard.RESET_CARD)

		fmt.Println("Card status:")
		status, err := card.Status()
		if err != nil {
			die(err)
		}

		fmt.Printf("\treader: %s\n\tstate: %x\n\tactive protocol: %x\n\tatr: % x\n",
			status.Reader, status.State, status.ActiveProtocol, status.ATR)

		var cmd = []byte{0x00, 0xa4, 0x00, 0x0c, 0x02, 0x3f, 0x00} // SELECT MF

		fmt.Println("Transmit:")
		fmt.Printf("\tc-apdu: % x\n", cmd)
		rsp, err := card.Transmit(cmd)
		if err != nil {
			die(err)
		}
		fmt.Printf("\tr-apdu: % x\n", rsp)
	}
}
