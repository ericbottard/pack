package pack_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/buildpack/pack"
	"github.com/buildpack/pack/config"
	"github.com/buildpack/pack/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRebase(t *testing.T) {
	spec.Run(t, "rebase", testRebase, spec.Parallel(), spec.Report(report.Terminal{}))
}

//go:generate mockgen -package mocks -destination mocks/writablestore.go github.com/buildpack/pack WritableStore

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
				_, err := factory.RebaseConfigFromFlags(pack.RebaseFlags{})
				assertNil(t, err)
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
