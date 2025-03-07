// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package developers

import (
	"internal/apiclient"

	"internal/client/developers"

	"github.com/spf13/cobra"
)

// ExportSubCmd to export developer
var ExportSubCmd = &cobra.Command{
	Use:   "export",
	Short: "Export Developer subscriptions to a file",
	Long:  "Export Developer subscriptions to a file",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		return apiclient.SetApigeeOrg(org)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		exportFileName := "subscription_" + email + ".json"

		respBody, err := developers.ExportSubscriptions(email)
		if err != nil {
			return err
		}

		return apiclient.WriteByteArrayToFile(exportFileName, false, respBody)
	},
}

func init() {
	ExportSubCmd.Flags().StringVarP(&email, "email", "n",
		"", "The developer's email")

	_ = ExportSubCmd.MarkFlagRequired("email")
}
