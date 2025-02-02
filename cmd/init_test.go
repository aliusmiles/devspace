package cmd

import (
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

/*import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cloudpkg "github.com/loft-sh/devspace/pkg/devspace/cloud"
	cloudconfig "github.com/loft-sh/devspace/pkg/devspace/cloud/config"
	cloudlatest "github.com/loft-sh/devspace/pkg/devspace/cloud/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/devspace/config/loader"
	"github.com/loft-sh/devspace/pkg/devspace/config/constants"
	"github.com/loft-sh/devspace/pkg/devspace/config/generated"
	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/devspace/docker"
	"github.com/loft-sh/devspace/pkg/util/fsutil"
	"github.com/loft-sh/devspace/pkg/util/kubeconfig"
	"github.com/loft-sh/devspace/pkg/util/log"
	"github.com/loft-sh/devspace/pkg/util/ptr"
	"github.com/loft-sh/devspace/pkg/util/survey"
	dockertypes "github.com/docker/docker/api/types"
	"k8s.io/client-go/tools/clientcmd"

	"gopkg.in/yaml.v2"
	"gotest.tools/assert"
)

type initTestCase struct {
	name string

	fakeConfig       *latest.Config
	fakeKubeConfig   clientcmd.ClientConfig
	fakeDockerClient docker.ClientInterface
	files            map[string]interface{}
	graphQLResponses []interface{}
	providerList     []*cloudlatest.Provider
	answers          []string

	reconfigureFlag bool
	dockerfileFlag  string
	contextFlag     string

	expectedErr    string
	expectedConfig *latest.Config
}

func TestInit(t *testing.T) {
	t.Skip("Errors")
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}

	wdBackup, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current working directory: %v", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("Error changing working directory: %v", err)
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		//Delete temp folder
		err = os.Chdir(wdBackup)
		if err != nil {
			t.Fatalf("Error changing dir back: %v", err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatalf("Error removing dir: %v", err)
		}
	}()

	testCases := []initTestCase{
		initTestCase{
			name: "Don't reconfigure the existing config",
			files: map[string]interface{}{
				constants.DefaultConfigPath: latest.Config{
					Version: latest.Version,
				},
			},
			expectedConfig: &latest.Config{
				Version: latest.Version,
				Dev:     &latest.DevConfig{},
			},
		},
		initTestCase{
			name:    "Init with helm chart",
			answers: []string{enterHelmChartOption, "someChart"},
			expectedConfig: &latest.Config{
				Version: latest.Version,
				Deployments: []*latest.DeploymentConfig{
					&latest.DeploymentConfig{
						Name: filepath.Base(dir),
						Helm: &latest.HelmConfig{
							Chart: &latest.ChartConfig{
								Name: "someChart",
							},
						},
					},
				},
				Dev: &latest.DevConfig{},
			},
		},
		initTestCase{
			name: "Init with manifests",
			files: map[string]interface{}{
				filepath.Join(gitIgnoreFile, "someFile"): "",
			},
			answers: []string{enterManifestsOption, "myManifest"},
			expectedConfig: &latest.Config{
				Version: latest.Version,
				Deployments: []*latest.DeploymentConfig{
					&latest.DeploymentConfig{
						Name: filepath.Base(dir),
						Kubectl: &latest.KubectlConfig{
							Manifests: []string{"myManifest"},
						},
					},
				},
				Dev: &latest.DevConfig{},
			},
		},
		initTestCase{
			name: "Init with existing image",
			files: map[string]interface{}{
				gitIgnoreFile: "",
			},
			answers: []string{useExistingImageOption, "someImage", "1000", "1234"},
			expectedConfig: &latest.Config{
				Version: latest.Version,
				Images: map[string]*latest.ImageConfig{
					"default": &latest.ImageConfig{
						Image:            "someImage",
						Tag:              "latest",
						CreatePullSecret: ptr.Bool(true),
						Build: &latest.BuildConfig{
							Disabled: ptr.Bool(true),
						},
					},
				},
				Deployments: []*latest.DeploymentConfig{
					&latest.DeploymentConfig{
						Name: filepath.Base(dir),
						Helm: &latest.HelmConfig{
							ComponentChart: ptr.Bool(true),
							Values: map[interface{}]interface{}{
								"containers": []*latest.ContainerConfig{
									{
										Image: "someImage",
									},
								},
								"service": &latest.ServiceConfig{
									Ports: []*latest.ServicePortConfig{
										{
											Port: ptr.Int(1000),
										},
									},
								},
							},
						},
					},
				},
				Dev: &latest.DevConfig{
					Ports: []*latest.PortForwardingConfig{
						&latest.PortForwardingConfig{
							ImageName: "default",
							PortMappings: []*latest.PortMapping{
								&latest.PortMapping{
									LocalPort:  ptr.Int(1234),
									RemotePort: ptr.Int(1000),
								},
							},
						},
					},
					Open: []*latest.OpenConfig{
						&latest.OpenConfig{
							URL: "http://localhost:1234",
						},
					},
				},
			},
		},
		initTestCase{
			name: "Entered existing Dockerfile",
			files: map[string]interface{}{
				"aDockerfile": "",
			},
			fakeDockerClient: &docker.FakeClient{
				AuthConfig: &dockertypes.AuthConfig{
					Username: "user",
					Password: "pass",
				},
			},
			answers: []string{enterDockerfileOption, "aDockerfile", "Use hub.docker.com => you are logged in as user"},
			expectedConfig: &latest.Config{
				Version: latest.Version,
				Images: map[string]*latest.ImageConfig{
					"default": &latest.ImageConfig{
						Image:      "",
						Dockerfile: "aDockerfile",
					},
				},
				Deployments: []*latest.DeploymentConfig{
					&latest.DeploymentConfig{
						Name: filepath.Base(dir),
						Helm: &latest.HelmConfig{
							ComponentChart: ptr.Bool(true),
							Values: map[interface{}]interface{}{
								"containers": []interface{}{
									struct{}{},
								},
							},
						},
					},
				},
				Dev: &latest.DevConfig{
					Sync: []*latest.SyncConfig{
						&latest.SyncConfig{
							ImageName:    "default",
							ExcludePaths: []string{"devspace.yaml"},
						},
					},
				},
			},
		},
	}

	log.OverrideRuntimeErrorHandler(true)
	log.SetInstance(&log.DiscardLogger{PanicOnExit: true})

	for _, testCase := range testCases {
		testInit(t, testCase)
	}
}

func testInit(t *testing.T, testCase initTestCase) {
	defer func() {
		for path := range testCase.files {
			removeTask := strings.Split(path, "/")[0]
			err := os.RemoveAll(removeTask)
			assert.NilError(t, err, "Error cleaning up folder in testCase %s", testCase.name)
		}
		err := os.RemoveAll(log.Logdir)
		assert.NilError(t, err, "Error cleaning up folder in testCase %s", testCase.name)
	}()

	cloudpkg.DefaultGraphqlClient = &customGraphqlClient{
		responses: testCase.graphQLResponses,
	}

	for _, answer := range testCase.answers {
		survey.SetNextAnswer(answer)
	}

	providerConfig, err := cloudconfig.Load()
	assert.NilError(t, err, "Error getting provider config in testCase %s", testCase.name)
	providerConfig.Providers = testCase.providerList

	loader.SetFakeConfig(testCase.fakeConfig)
	loader.ResetConfig()
	generated.ResetConfig()
	kubeconfig.SetFakeConfig(testCase.fakeKubeConfig)
	docker.SetFakeClient(testCase.fakeDockerClient)

	for path, content := range testCase.files {
		asYAML, err := yaml.Marshal(content)
		assert.NilError(t, err, "Error parsing config to yaml in testCase %s", testCase.name)
		err = fsutil.WriteToFile(asYAML, path)
		assert.NilError(t, err, "Error writing file in testCase %s", testCase.name)
	}

	err = (&InitCmd{
		Reconfigure: testCase.reconfigureFlag,
		Dockerfile:  testCase.dockerfileFlag,
		Context:     testCase.contextFlag,
	}).Run(nil, []string{})

	if testCase.expectedErr == "" {
		assert.NilError(t, err, "Unexpected error in testCase %s.", testCase.name)

		config, err := loader.GetConfig(nil)
		assert.NilError(t, err, "Error getting config after init call in testCase %s.", testCase.name)
		configYaml, err := yaml.Marshal(config)
		assert.NilError(t, err, "Error parsing config to yaml after init call in testCase %s.", testCase.name)
		expectedConfigYaml, err := yaml.Marshal(testCase.expectedConfig)
		assert.NilError(t, err, "Error parsing expected config to yaml after init call in testCase %s.", testCase.name)
		assert.Equal(t, string(configYaml), string(expectedConfigYaml), "Initialized config is wrong in testCase %s.", testCase.name)
	} else {
		assert.Error(t, err, testCase.expectedErr, "Wrong or no error in testCase %s.", testCase.name)
	}

	err = filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		os.RemoveAll(path)
		return nil
	})
	assert.NilError(t, err, "Error cleaning up in testCase %s", testCase.name)
}*/

type parseImagesTestCase struct {
	name      string
	manifests string
	expected  []string
}

func TestParseImages(t *testing.T) {
	testCases := []parseImagesTestCase{
		{
			name: `Single`,
			manifests: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "new"
  labels:
    "app.kubernetes.io/name": "devspace-app"
    "app.kubernetes.io/component": "test"
    "app.kubernetes.io/managed-by": "Helm"
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      "app.kubernetes.io/name": "devspace-app"
      "app.kubernetes.io/component": "test"
      "app.kubernetes.io/managed-by": "Helm"
  template:
    metadata:
      labels:
        "app.kubernetes.io/name": "devspace-app"
        "app.kubernetes.io/component": "test"
        "app.kubernetes.io/managed-by": "Helm"
    spec:
      containers:
        - image: "username/app"
          name: "container-0"
`,
			expected: []string{
				"username/app",
			},
		},
		{
			name: `Multiple`,
			manifests: `
---
# Source: my-app/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: php
  labels:
    release: "test-helm"
spec:
  ports:
  - port: 80
    protocol: TCP
  selector:
    release: "test-helm"
---
# Source: my-app/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-helm
  labels:
    release: "test-helm"
spec:
  replicas: 1
  selector:
    matchLabels:
      release: "test-helm"
  template:
    metadata:
      annotations:
        revision: "1"
      labels:
        release: "test-helm"
    spec:
      containers:
      - name: default
        image: "php"
`,
			expected: []string{
				"php",
			},
		},
	}

	for _, testCase := range testCases {
		manifests := testCase.manifests

		actual, err := parseImages(manifests)
		assert.NilError(
			t,
			err,
			"Unexpected error in test case %s",
			testCase.name,
		)

		expected := testCase.expected
		assert.Assert(
			t,
			cmp.DeepEqual(expected, actual),
			"Unexpected values in test case %s",
			testCase.name,
		)
	}
}
