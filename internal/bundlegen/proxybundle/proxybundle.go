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

package proxybundle

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"internal/clilog"

	"internal/bundlegen"
	genapi "internal/bundlegen"
	apiproxy "internal/bundlegen/apiproxydef"
	policies "internal/bundlegen/policies"
	proxies "internal/bundlegen/proxies"
	"internal/bundlegen/targets"
	target "internal/bundlegen/targets"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var rootDir = "apiproxy"

func GenerateAPIProxyBundleFromOAS(name string,
	content string,
	fileName string,
	skipPolicy bool,
	addCORS bool,
	oasGoogleAcessTokenScopeLiteral string,
	oasGoogleIdTokenAudLiteral string,
	oasGoogleIdTokenAudRef string,
	oasTargetUrlRef string,
	targetUrl string,
) (err error) {
	var apiProxyData, proxyEndpointData, targetEndpointData string
	const resourceType = "oas"

	tmpDir, err := os.MkdirTemp("", "proxy")
	if err != nil {
		return err
	}

	rootDir = path.Join(tmpDir, rootDir)

	if err = os.Mkdir(rootDir, os.ModePerm); err != nil {
		return err
	}

	// write API Proxy file
	if apiProxyData, err = apiproxy.GetAPIProxy(); err != nil {
		return err
	}

	err = writeXMLData(rootDir+string(os.PathSeparator)+name+".xml", apiProxyData)
	if err != nil {
		return err
	}

	proxiesDirPath := rootDir + string(os.PathSeparator) + "proxies"
	policiesDirPath := rootDir + string(os.PathSeparator) + "policies"
	targetDirPath := rootDir + string(os.PathSeparator) + "targets"
	resDirPath := rootDir + string(os.PathSeparator) + "resources" + string(os.PathSeparator) + resourceType //"oas"

	if err = os.Mkdir(proxiesDirPath, os.ModePerm); err != nil {
		return err
	}

	if proxyEndpointData, err = proxies.GetProxyEndpoint(); err != nil {
		return err
	}

	err = writeXMLData(proxiesDirPath+string(os.PathSeparator)+"default.xml", proxyEndpointData)
	if err != nil {
		return err
	}

	if err = os.Mkdir(targetDirPath, os.ModePerm); err != nil {
		return err
	}

	for _, targetEndpoint := range targets.TargetEndpoints {
		if targetEndpointData, err = target.GetTargetEndpoint(targetEndpoint); err != nil {
			return err
		}

		if err = writeXMLData(targetDirPath+string(os.PathSeparator)+targetEndpoint.Name+".xml", targetEndpointData); err != nil {
			return err
		}
	}

	if !skipPolicy {
		if err = os.MkdirAll(resDirPath, os.ModePerm); err != nil {
			return err
		}
		if err = writeXMLData(resDirPath+string(os.PathSeparator)+fileName, content); err != nil {
			return err
		}
	}

	if err = os.Mkdir(policiesDirPath, os.ModePerm); err != nil {
		return err
	}

	// add set target url
	if targetUrl == "" {
		if genapi.GenerateSetTargetPolicy() {
			if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Set-Target-1.xml",
				policies.AddSetTargetEndpointRef(oasTargetUrlRef)); err != nil {
				return err
			}
		}
	}

	// add security policies
	for _, securityScheme := range genapi.GetSecuritySchemesList() {
		if securityScheme.APIKeyPolicy.APIKeyPolicyEnabled {
			if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Verify-API-Key-"+securityScheme.SchemeName+".xml",
				policies.AddVerifyApiKeyPolicy(securityScheme.APIKeyPolicy.APIKeyLocation,
					securityScheme.SchemeName,
					securityScheme.APIKeyPolicy.APIKeyName)); err != nil {
				return err
			}
		}
		if securityScheme.OAuthPolicy.OAuthPolicyEnabled {
			if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"OAuth-v20-1.xml",
				policies.AddOAuth2Policy(securityScheme.OAuthPolicy.Scope)); err != nil {
				return err
			}
		}
	}

	// add quota policies
	for quotaPolicyName, quotaPolicyContent := range genapi.GetQuotaPolicies() {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Quota-"+quotaPolicyName+".xml", quotaPolicyContent); err != nil {
			return err
		}
	}

	// add spike arrest policies
	for spikeArrestPolicyName, spikeArrestPolicyContent := range genapi.GetSpikeArrestPolicies() {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Spike-Arrest-"+spikeArrestPolicyName+".xml", spikeArrestPolicyContent); err != nil {
			return err
		}
	}

	if !skipPolicy {
		// add oas policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"OpenAPI-Spec-Validation-1.xml",
			policies.AddOpenAPIValidatePolicy(fileName)); err != nil {
			return err
		}
	}

	if addCORS {
		// add cors policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Add-CORS.xml", policies.AddCORSPolicy()); err != nil {
			return err
		}
	}

	if err = archiveBundle(rootDir, name+".zip"); err != nil {
		return err
	}

	defer os.RemoveAll(rootDir) // clean up
	return nil
}

func GenerateAPIProxyBundleFromGQL(name string,
	content string,
	fileName string,
	action string,
	location string,
	keyName string,
	skipPolicy bool,
	addCORS bool,
	targetUrlRef string,
	targetUrl string,
) (err error) {
	var apiProxyData, proxyEndpointData, targetEndpointData string
	const resourceType = "graphql"

	tmpDir, err := os.MkdirTemp("", "proxy")
	if err != nil {
		return err
	}

	rootDir = path.Join(tmpDir, rootDir)

	if err = os.Mkdir(rootDir, os.ModePerm); err != nil {
		return err
	}

	// write API Proxy file
	if apiProxyData, err = apiproxy.GetAPIProxy(); err != nil {
		return err
	}

	err = writeXMLData(rootDir+string(os.PathSeparator)+name+".xml", apiProxyData)
	if err != nil {
		return err
	}

	proxiesDirPath := rootDir + string(os.PathSeparator) + "proxies"
	policiesDirPath := rootDir + string(os.PathSeparator) + "policies"
	targetDirPath := rootDir + string(os.PathSeparator) + "targets"
	resDirPath := rootDir + string(os.PathSeparator) + "resources" + string(os.PathSeparator) + resourceType //"graphql"

	if err = os.Mkdir(proxiesDirPath, os.ModePerm); err != nil {
		return err
	}

	if proxyEndpointData, err = proxies.GetProxyEndpoint(); err != nil {
		return err
	}

	err = writeXMLData(proxiesDirPath+string(os.PathSeparator)+"default.xml", proxyEndpointData)
	if err != nil {
		return err
	}

	if err = os.Mkdir(targetDirPath, os.ModePerm); err != nil {
		return err
	}

	for _, targetEndpoint := range targets.TargetEndpoints {
		if targetEndpointData, err = target.GetTargetEndpoint(targetEndpoint); err != nil {
			return err
		}

		if err = writeXMLData(targetDirPath+string(os.PathSeparator)+targetEndpoint.Name+".xml", targetEndpointData); err != nil {
			return err
		}
	}

	if !skipPolicy {
		if err = os.MkdirAll(resDirPath, os.ModePerm); err != nil {
			return err
		}
		if err = writeXMLData(resDirPath+string(os.PathSeparator)+fileName, content); err != nil {
			return err
		}
	}

	if err = os.Mkdir(policiesDirPath, os.ModePerm); err != nil {
		return err
	}

	if targetUrlRef != "" {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Set-Target-1.xml",
			policies.AddSetTargetEndpointRef(targetUrlRef)); err != nil {
			return err
		}
	}

	if !skipPolicy {
		// add gql policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Validate-"+name+"-Schema.xml",
			policies.AddGraphQLPolicy(name, action, fileName)); err != nil {
			return err
		}
	}

	if keyName != "" {
		// add verifyapi key policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Verify-API-Key-"+name+".xml",
			policies.AddVerifyApiKeyPolicy(location, name, keyName)); err != nil {
			return err
		}
	}

	if addCORS {
		// add cors policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Add-CORS.xml", policies.AddCORSPolicy()); err != nil {
			return err
		}
	}

	if err = archiveBundle(rootDir, name+".zip"); err != nil {
		return err
	}

	defer os.RemoveAll(rootDir) // clean up
	return nil
}

func GenerateIntegrationAPIProxyBundle(name string, integration string, apitrigger string, skipPolicy bool) (err error) {
	var apiProxyData, proxyEndpointData, integrationEndpointData string

	tmpDir, err := os.MkdirTemp("", "proxy")
	if err != nil {
		return err
	}

	rootDir = path.Join(tmpDir, rootDir)

	if err = os.Mkdir(rootDir, os.ModePerm); err != nil {
		return err
	}

	// write API Proxy file
	if apiProxyData, err = apiproxy.GetAPIProxy(); err != nil {
		return err
	}

	err = writeXMLData(rootDir+string(os.PathSeparator)+name+".xml", apiProxyData)
	if err != nil {
		return err
	}

	proxiesDirPath := rootDir + string(os.PathSeparator) + "proxies"
	policiesDirPath := rootDir + string(os.PathSeparator) + "policies"
	integrationDirPath := rootDir + string(os.PathSeparator) + "integration-endpoints"

	if err = os.Mkdir(proxiesDirPath, os.ModePerm); err != nil {
		return err
	}

	if proxyEndpointData, err = proxies.GetProxyEndpoint(); err != nil {
		return err
	}

	err = writeXMLData(proxiesDirPath+string(os.PathSeparator)+"default.xml", proxyEndpointData)
	if err != nil {
		return err
	}

	if err = os.Mkdir(integrationDirPath, os.ModePerm); err != nil {
		return err
	}

	integrationEndpointData = target.GetIntegrationEndpoint()
	if err = writeXMLData(integrationDirPath+string(os.PathSeparator)+"default.xml", integrationEndpointData); err != nil {
		return err
	}

	if err = os.Mkdir(policiesDirPath, os.ModePerm); err != nil {
		return err
	}

	// add set integration request policy
	if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"set-integration-request.xml",
		policies.AddSetIntegrationRequestPolicy(integration, apitrigger)); err != nil {
		return err
	}

	if err = archiveBundle(rootDir, name+".zip"); err != nil {
		return err
	}

	defer os.RemoveAll(rootDir) // clean up
	return nil
}

func GenerateAPIProxyBundleFromSwagger(name string,
	skipPolicy bool,
	addCORS bool,
) (err error) {
	var apiProxyData, proxyEndpointData, targetEndpointData string

	tmpDir, err := os.MkdirTemp("", "proxy")
	if err != nil {
		return err
	}

	rootDir = path.Join(tmpDir, rootDir)

	if name == "" {
		name = bundlegen.GetGoogleApiName()
	}

	if err = os.Mkdir(rootDir, os.ModePerm); err != nil {
		return err
	}

	// write API Proxy file
	if apiProxyData, err = apiproxy.GetAPIProxy(); err != nil {
		return err
	}

	err = writeXMLData(rootDir+string(os.PathSeparator)+name+".xml", apiProxyData)
	if err != nil {
		return err
	}

	proxiesDirPath := rootDir + string(os.PathSeparator) + "proxies"
	policiesDirPath := rootDir + string(os.PathSeparator) + "policies"
	targetDirPath := rootDir + string(os.PathSeparator) + "targets"

	if err = os.Mkdir(proxiesDirPath, os.ModePerm); err != nil {
		return err
	}

	if proxyEndpointData, err = proxies.GetProxyEndpoint(); err != nil {
		return err
	}

	err = writeXMLData(proxiesDirPath+string(os.PathSeparator)+"default.xml", proxyEndpointData)
	if err != nil {
		return err
	}

	if err = os.Mkdir(targetDirPath, os.ModePerm); err != nil {
		return err
	}

	for _, targetEndpoint := range targets.TargetEndpoints {
		if targetEndpointData, err = target.GetTargetEndpoint(targetEndpoint); err != nil {
			return err
		}

		if err = writeXMLData(targetDirPath+string(os.PathSeparator)+targetEndpoint.Name+".xml", targetEndpointData); err != nil {
			return err
		}
	}

	if err = os.Mkdir(policiesDirPath, os.ModePerm); err != nil {
		return err
	}

	// add AM policies
	for amPolicyName, amPolicyContent := range genapi.GetAMPolicies() {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"AM-"+amPolicyName+".xml", amPolicyContent); err != nil {
			return err
		}
	}

	// add security policies
	for _, securityScheme := range genapi.GetSecuritySchemesList() {
		if securityScheme.JWTPolicy.JWTPolicyEnabled {
			if len(securityScheme.JWTPolicy.Location) > 0 {
				var headerName, headerValue, queryName string
				for locationKey, locationValue := range securityScheme.JWTPolicy.Location {
					if locationKey == "query" {
						queryName = locationValue
					} else if locationKey == "header" {
						headerName = locationValue
					} else if locationKey == "value_prefix" {
						headerValue = locationValue
					}
				}
				if queryName != "" {
					if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"ExtractJWT-"+securityScheme.SchemeName+".xml",
						policies.AddExtractJwtQueryPolicy("ExtractJWT-"+securityScheme.SchemeName, queryName)); err != nil {
						return err
					}
				} else {
					if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"ExtractJWT-"+securityScheme.SchemeName+".xml",
						policies.AddExtractJwtHeaderPolicy("ExtractJWT-"+securityScheme.SchemeName, headerName, headerValue)); err != nil {
						return err
					}
				}
			}
			if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"VerifyJWT-"+securityScheme.SchemeName+".xml",
				policies.AddVerifyJWTPolicy("VerifyJWT-"+securityScheme.SchemeName,
					securityScheme.JWTPolicy.JwkUri,
					securityScheme.JWTPolicy.Issuer,
					securityScheme.JWTPolicy.Audience,
					securityScheme.JWTPolicy.Source)); err != nil {
				return err
			}
		}
		if securityScheme.APIKeyPolicy.APIKeyPolicyEnabled {
			if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Verify-API-Key-"+securityScheme.SchemeName+".xml",
				policies.AddVerifyApiKeyPolicy(securityScheme.APIKeyPolicy.APIKeyLocation,
					securityScheme.SchemeName,
					securityScheme.APIKeyPolicy.APIKeyName)); err != nil {
				return err
			}
		}
	}

	// add quota policies
	for quotaPolicyName, quotaPolicyContent := range genapi.GetQuotaPolicies() {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Quota-"+quotaPolicyName+".xml", quotaPolicyContent); err != nil {
			return err
		}
	}

	if allow := genapi.GetAllowDefinition(); allow == "configured" {
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Raise-Fault-Unknown-Request.xml", policies.AddRaiseFaultPolicy()); err != nil {
			return err
		}
	}

	if addCORS {
		// add cors policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Add-CORS.xml", policies.AddCORSPolicy()); err != nil {
			return err
		}
	}

	if policies.IsCopyAuthEnabled() {
		// add AM policy
		if err = writeXMLData(policiesDirPath+string(os.PathSeparator)+"Copy-Auth-Var.xml", policies.AddCopyAuthHeaderPolicy()); err != nil {
			return err
		}
	}

	if err = archiveBundle(rootDir, name+".zip"); err != nil {
		return err
	}

	defer os.RemoveAll(rootDir) // clean up

	return err
}

func writeXMLData(fileName string, data string) error {
	fileWriter, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = fileWriter.WriteString(data)
	if err != nil {
		return err
	}

	fileWriter.Close()
	return nil
}

func GenerateArchiveBundle(pathToZip, destinationPath string) error {
	return archiveBundle(pathToZip, destinationPath)
}

func archiveBundle(pathToZip, destinationPath string) (err error) {
	var destinationFile *os.File

	pathSep := `/` // For archives/zip the path separator is always /

	if destinationFile, err = os.Create(destinationPath); err != nil {
		return err
	}

	myZip := zip.NewWriter(destinationFile)
	err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			relPath := filepath.ToSlash(strings.TrimPrefix(filePath, filepath.Dir(pathToZip)))
			zipEntry := strings.TrimPrefix(relPath, pathSep) + pathSep
			_, err = myZip.Create(zipEntry)
			return err
		}
		if err != nil {
			return err
		}
		relPath := filepath.ToSlash(strings.TrimPrefix(filePath, filepath.Dir(pathToZip)))
		zipEntry := strings.TrimPrefix(relPath, pathSep)
		zipFile, err := myZip.Create(zipEntry)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = destinationFile.Close()
		return err
	}
	if err = myZip.Close(); err != nil {
		return err
	}
	if err = destinationFile.Close(); err != nil {
		return err
	}
	return nil
}

func GitHubImportBundle(owner string, repo string, repopath string) (err error) {
	// clean up any files or folders
	CleanUp()
	os.RemoveAll(rootDir)

	var client *github.Client

	//
	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()

	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	// 1. download the proxy
	if err := downloadProxyFromRepo(client, ctx, owner, repo, repopath); err != nil {
		return err
	}

	if client != nil {
		clilog.Info.Println("")
	}

	// 2. compress the proxy folder
	curDir, _ := os.Getwd()
	if err := archiveBundle(path.Join(curDir, rootDir), path.Join(curDir, rootDir+".zip")); err != nil {
		return err
	}

	defer os.RemoveAll(rootDir) // clean up
	return err
}

func CleanUp() {
	if _, err := os.Stat(rootDir + ".zip"); err == nil {
		_ = os.Remove(rootDir + ".zip")
	}
}

func downloadProxyFromRepo(client *github.Client, ctx context.Context, owner string, repo string, repopath string) (err error) {
	var fileContent *github.RepositoryContent
	var directoryContents []*github.RepositoryContent

	if fileContent, directoryContents, _, err = client.Repositories.GetContents(ctx, owner, repo, repopath, nil); err != nil {
		return err
	}

	if fileContent != nil {
		if err = downloadResource(*fileContent.Path, *fileContent.DownloadURL); err != nil {
			return err
		}
	}

	if len(directoryContents) > 0 {
		for _, directoryContent := range directoryContents {
			if *directoryContent.Type == "dir" {
				if err = downloadProxyFromRepo(client, ctx, owner, repo, path.Join(repopath, *directoryContent.Name)); err != nil {
					return err
				}
			} else if *directoryContent.Type == "file" {
				if err = downloadResource(*directoryContent.Path, *directoryContent.DownloadURL); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getApiProxyFolder(repoPath string) (apiProxyFolder string, apiProxyFile string) {
	re := regexp.MustCompile(`(\S*)?(\/?)apiproxy`)

	apiProxyFileBytes := re.ReplaceAll([]byte(repoPath), []byte(rootDir))
	apiProxyFile = string(apiProxyFileBytes)

	apiProxyFolder = filepath.Dir(apiProxyFile)
	return apiProxyFolder, apiProxyFile
}

// downloadResource method is used to download resources, proxy bundles, sharedflows
func downloadResource(repoPath string, url string) (err error) {
	var apiproxyFolder, apiproxyFile string

	if apiproxyFolder, apiproxyFile = getApiProxyFolder(repoPath); err != nil {
		return err
	}

	_ = os.MkdirAll(apiproxyFolder, 0o755)

	out, err := os.Create(apiproxyFile)
	if err != nil {
		clilog.Debug.Println("error creating file: ", err)
		return err
	}
	defer out.Close()

	client := &http.Client{}

	clilog.Debug.Println("Connecting to : ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		clilog.Debug.Println("error in client: ", err)
		return err
	}

	resp, err := client.Do(req)

	if err != nil {
		clilog.Debug.Println("error connecting: ", err)
		return err
	} else if resp.StatusCode > 299 {
		clilog.Debug.Println("error in response: ", resp.Body)
		return errors.New("error in response")
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp == nil {
		clilog.Debug.Println("error in response: Response was null")
		return fmt.Errorf("error in response: Response was null")
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		clilog.Debug.Println("error writing response to file: ", err)
		return err
	}

	clilog.Debug.Println("Resource " + apiproxyFolder + " completed")
	return nil
}
