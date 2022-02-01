// Copyright 2021 Google LLC
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

package apis

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/srinandan/apigeecli/apiclient"
	proxybundle "github.com/srinandan/apigeecli/bundlegen/proxybundle"
	"github.com/srinandan/apigeecli/client/apis"
)

var GhCreateCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh"},
	Short:   "Creates an API proxy from a GitHub repo",
	Long:    "Creates an API proxy from a GitHub repo",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if os.Getenv("GITHUB_TOKEN") == "" {
			return fmt.Errorf("Github access token must be set with this feature")
		}
		//(\w+)?\/apiproxy$
		re := regexp.MustCompile(`(\w+)?\/apiproxy$`)
		if ok := re.Match([]byte(ghPath)); !ok {
			return fmt.Errorf("Github path must end with /apiproxy")
		}

		return apiclient.SetApigeeOrg(org)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = proxybundle.GitHubImportBundle(ghOwner, ghRepo, ghPath); err != nil {
			proxybundle.CleanUp()
			return err
		}
		_, err = apis.CreateProxy(name, bundleName)
		proxybundle.CleanUp()
		return err
	},
}

const bundleName = "apiproxy.zip"

var ghOwner, ghRepo, ghPath string

func init() {
	GhCreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "API Proxy name")
	GhCreateCmd.Flags().StringVarP(&ghOwner, "owner", "u",
		"", "The github organization or username. ex: In https://github.com/srinandan, srinandan is the user")
	GhCreateCmd.Flags().StringVarP(&ghRepo, "repo", "r",
		"", "The github repo name. ex: https://github.com/srinandan/sample-apps, sample-apps is the repo")
	GhCreateCmd.Flags().StringVarP(&ghPath, "proxy-path", "p",
		"", "The path in the repo to the apiproxy folder. ex: my-repo/apiproxy")

	_ = GhCreateCmd.MarkFlagRequired("name")
	_ = GhCreateCmd.MarkFlagRequired("owner")
	_ = GhCreateCmd.MarkFlagRequired("repo")
	_ = GhCreateCmd.MarkFlagRequired("proxy-path")

}
