package pack_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/buildpack/pack"
	"github.com/buildpack/pack/config"
	"github.com/buildpack/pack/mocks"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/golang/mock/gomock"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRebase(t *testing.T) {
	spec.Run(t, "rebase", testRebase, spec.Parallel(), spec.Report(report.Terminal{}))
}

//go:generate mockgen -package mocks -destination mocks/writablestore.go github.com/buildpack/pack WritableStore
//go:generate mockgen -package mocks -destination mocks/layer.go github.com/google/go-containerregistry/pkg/v1 Layer

func testRebase(t *testing.T, when spec.G, it spec.S) {
	when("#RebaseFactory", func() {
		var (
			mockController *gomock.Controller
			mockDocker     *mocks.MockDocker
			mockImages     *mocks.MockImages
			factory        pack.RebaseFactory
		)
		it.Before(func() {
			mockController = gomock.NewController(t)
			mockDocker = mocks.NewMockDocker(mockController)
			mockImages = mocks.NewMockImages(mockController)

			factory = pack.RebaseFactory{
				Docker: mockDocker,
				Log:    log.New(ioutil.Discard, "", log.LstdFlags),
				Config: &config.Config{
					DefaultStackID: "some.default.stack",
					Stacks: []config.Stack{
						{
							ID:          "some.default.stack",
							BuildImages: []string{"default/build", "registry.com/build/image"},
							RunImages:   []string{"default/run"},
						},
						{
							ID:          "some.other.stack",
							BuildImages: []string{"other/build"},
							RunImages:   []string{"other/run"},
						},
					},
				},
				Images: mockImages,
			}

			// output, err := exec.Command("docker", "pull", "packs/build").CombinedOutput()
			// if err != nil {
			// 	t.Fatalf("Failed to pull the base image in test setup: %s: %s", output, err)
			// }
		})

		it.After(func() {
			mockController.Finish()
		})

		when("#RebaseConfigFromFlags", func() {
			it("XXXX", func() {
				mockRepoStore := mocks.NewMockStore(mockController)
				mockRepoImage := mocks.NewMockImage(mockController)
				mockBaseImage := mocks.NewMockImage(mockController)

				mockDocker.EXPECT().PullImage("default/build")
				mockImages.EXPECT().ReadImage("default/build", true).Return(mockBaseImage, nil)

				mockDocker.EXPECT().PullImage("myorg/myrepo")
				mockImages.EXPECT().RepoStore("myorg/myrepo", true).Return(mockRepoStore, nil)
				mockImages.EXPECT().ReadImage("myorg/myrepo", true).Return(mockRepoImage, nil)
				mockDocker.EXPECT().ImageInspectWithRaw(gomock.Any(), "myorg/myrepo").Return(dockertypes.ImageInspect{
					Config: &dockercontainer.Config{
						Labels: map[string]string{
							"io.buildpacks.stack.id":           "some.default.stack",
							"io.buildpacks.lifecycle.metadata": `{"runimage":{"sha":"sha256:abcdef"}}`,
						},
					},
				}, nil, nil).AnyTimes()

				cfg, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
					RepoName: "myorg/myrepo",
					Publish:  false,
					NoPull:   false,
				})
				assertNil(t, err)

				assertEq(t, cfg.RepoName, "myorg/myrepo")
				assertEq(t, cfg.Publish, false)
				assertSame(t, cfg.Repo, mockRepoStore)
				assertSame(t, cfg.RepoImage, mockRepoImage)
				assertNotNil(t, cfg.OldBase)
				assertSame(t, cfg.NewBase, mockBaseImage)

				layer1 := mocks.NewMockLayer(mockController)
				layer1.EXPECT().DiffID().Return(v1.Hash{Algorithm: "sha256", Hex: "12345"}, nil)
				layer2 := mocks.NewMockLayer(mockController)
				layer2.EXPECT().DiffID().Return(v1.Hash{Algorithm: "sha256", Hex: "abcdef"}, nil)
				layer3 := mocks.NewMockLayer(mockController)

				mockRepoImage.EXPECT().Layers().Return([]v1.Layer{layer1, layer2, layer3}, nil)

				oldBaseLayers, err := cfg.OldBase.Layers()
				assertNil(t, err)
				assertEq(t, len(oldBaseLayers), 2)
				assertSame(t, oldBaseLayers[0], layer1)
				assertSame(t, oldBaseLayers[1], layer2)
			})

			// 	it("uses default stack build image as base image", func() {
			// 		mockBaseImage := mocks.NewMockImage(mockController)
			// 		mockImageStore := mocks.NewMockStore(mockController)
			// 		mockDocker.EXPECT().PullImage("default/build")
			// 		mockImages.EXPECT().ReadImage("default/build", true).Return(mockBaseImage, nil)
			// 		mockImages.EXPECT().RepoStore("some/image", true).Return(mockImageStore, nil)
			//
			// 		config, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 			RepoName:        "some/image",
			// 			RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 		})
			// 		if err != nil {
			// 			t.Fatalf("error creating builder config: %s", err)
			// 		}
			// 		assertSameInstance(t, config.BaseImage, mockBaseImage)
			// 		assertSameInstance(t, config.Repo, mockImageStore)
			// 		checkBuildpacks(t, config.Buildpacks)
			// 		checkGroups(t, config.Groups)
			// 		assertEq(t, config.RebaseDir, "testdata")
			// 	})
			//
			// 	it("select the build image with matching registry", func() {
			// 		mockBaseImage := mocks.NewMockImage(mockController)
			// 		mockImageStore := mocks.NewMockStore(mockController)
			// 		mockDocker.EXPECT().PullImage("registry.com/build/image")
			// 		mockImages.EXPECT().ReadImage("registry.com/build/image", true).Return(mockBaseImage, nil)
			// 		mockImages.EXPECT().RepoStore("registry.com/some/image", true).Return(mockImageStore, nil)
			//
			// 		config, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 			RepoName:        "registry.com/some/image",
			// 			RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 		})
			// 		if err != nil {
			// 			t.Fatalf("error creating builder config: %s", err)
			// 		}
			// 		assertSameInstance(t, config.BaseImage, mockBaseImage)
			// 		assertSameInstance(t, config.Repo, mockImageStore)
			// 		checkBuildpacks(t, config.Buildpacks)
			// 		checkGroups(t, config.Groups)
			// 		assertEq(t, config.RebaseDir, "testdata")
			// 	})
			//
			// 	it("doesn't pull base a new image when --no-pull flag is provided", func() {
			// 		mockBaseImage := mocks.NewMockImage(mockController)
			// 		mockImageStore := mocks.NewMockStore(mockController)
			// 		mockImages.EXPECT().ReadImage("default/build", true).Return(mockBaseImage, nil)
			// 		mockImages.EXPECT().RepoStore("some/image", true).Return(mockImageStore, nil)
			//
			// 		config, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 			RepoName:        "some/image",
			// 			RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 			NoPull:          true,
			// 		})
			// 		if err != nil {
			// 			t.Fatalf("error creating builder config: %s", err)
			// 		}
			// 		assertSameInstance(t, config.BaseImage, mockBaseImage)
			// 		assertSameInstance(t, config.Repo, mockImageStore)
			// 		checkBuildpacks(t, config.Buildpacks)
			// 		checkGroups(t, config.Groups)
			// 		assertEq(t, config.RebaseDir, "testdata")
			// 	})
			//
			// 	it("fails if the base image cannot be found", func() {
			// 		mockImages.EXPECT().ReadImage("default/build", true).Return(nil, nil)
			//
			// 		_, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 			RepoName:        "some/image",
			// 			RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 			NoPull:          true,
			// 		})
			// 		if err == nil {
			// 			t.Fatalf("Expected error when base image is missing from daemon")
			// 		}
			// 	})
			//
			// 	it("fails if the base image cannot be pulled", func() {
			// 		mockDocker.EXPECT().PullImage("default/build").Return(fmt.Errorf("some-error"))
			//
			// 		_, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 			RepoName:        "some/image",
			// 			RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 		})
			// 		if err == nil {
			// 			t.Fatalf("Expected error when base image is missing from daemon")
			// 		}
			// 	})
			//
			// 	it("fails if there is no build image for the stack", func() {
			// 		factory.Config = &config.Config{
			// 			DefaultStackID: "some.bad.stack",
			// 			Stacks: []config.Stack{
			// 				{
			// 					ID: "some.bad.stack",
			// 				},
			// 			},
			// 		}
			// 		_, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 			RepoName:        "some/image",
			// 			RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 			NoPull:          true,
			// 		})
			// 		assertError(t, err, `Invalid stack: stack "some.bad.stack" requies at least one build image`)
			// 	})
			//
			// 	it("uses the build image that matches the repoName registry", func() {})
			//
			// 	when("-s flag is provided", func() {
			// 		it("used the build image from the selected stack", func() {
			// 			mockBaseImage := mocks.NewMockImage(mockController)
			// 			mockImageStore := mocks.NewMockStore(mockController)
			// 			mockDocker.EXPECT().PullImage("other/build")
			// 			mockImages.EXPECT().ReadImage("other/build", true).Return(mockBaseImage, nil)
			// 			mockImages.EXPECT().RepoStore("some/image", true).Return(mockImageStore, nil)
			//
			// 			config, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 				RepoName:        "some/image",
			// 				RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 				StackID:         "some.other.stack",
			// 			})
			// 			if err != nil {
			// 				t.Fatalf("error creating builder config: %s", err)
			// 			}
			// 			assertSameInstance(t, config.BaseImage, mockBaseImage)
			// 			assertSameInstance(t, config.Repo, mockImageStore)
			// 			checkBuildpacks(t, config.Buildpacks)
			// 			checkGroups(t, config.Groups)
			// 		})
			//
			// 		it("fails if the provided stack id does not exist", func() {
			// 			_, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 				RepoName:        "some/image",
			// 				RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 				NoPull:          true,
			// 				StackID:         "some.missing.stack",
			// 			})
			// 			assertError(t, err, `Missing stack: stack with id "some.missing.stack" not found in pack config.toml`)
			// 		})
			// 	})
			//
			// 	when("--publish is passed", func() {
			// 		it("uses a registry store and doesn't pull base image", func() {
			// 			mockBaseImage := mocks.NewMockImage(mockController)
			// 			mockImageStore := mocks.NewMockStore(mockController)
			// 			mockImages.EXPECT().ReadImage("default/build", false).Return(mockBaseImage, nil)
			// 			mockImages.EXPECT().RepoStore("some/image", false).Return(mockImageStore, nil)
			//
			// 			config, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{
			// 				RepoName:        "some/image",
			// 				RebaseTomlPath: filepath.Join("testdata", "builder.toml"),
			// 				Publish:         true,
			// 			})
			// 			if err != nil {
			// 				t.Fatalf("error creating builder config: %s", err)
			// 			}
			// 			assertSameInstance(t, config.BaseImage, mockBaseImage)
			// 			assertSameInstance(t, config.Repo, mockImageStore)
			// 			checkBuildpacks(t, config.Buildpacks)
			// 			checkGroups(t, config.Groups)
			// 			assertEq(t, config.RebaseDir, "testdata")
			// 		})
			// 	})
			// })
		})
	})
}
