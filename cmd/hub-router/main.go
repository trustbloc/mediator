/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/trustbloc/hub-router/cmd/hub-router/startcmd"
)

func main() {
	cmd := &cobra.Command{
		Use: "hub-router",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	cmd.AddCommand(startcmd.GetStartCmd(&startcmd.HTTPServer{}))

	if err := cmd.Execute(); err != nil {
		log.Fatalf("failed to run hub-router: %s", err.Error())
	}
}
