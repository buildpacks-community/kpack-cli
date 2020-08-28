package main

import (
	"fmt"
	"github.com/pivotal/build-service-cli/pkg/diff/lib"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
)

func main() {
	fmt.Print("\u001B[0;36m===> PREPARE\u001B[0m\n")
	fmt.Print("Build Info:\n------------------------------------------\n")

	commit().Render(os.Stdout, "COMMIT", true)
	config().Render(os.Stdout, "CONFIG", true)
	buildpacks().Render(os.Stdout, "BUILDPACKS", true)
	stack().Render(os.Stdout, "STACK", true)

	printbuild()
}

func commit() lib.Diff {
	return lib.Diff{
		Before: Commit{
			Revision: "67fcb7f2ed698a79cd2161f9909b8f1b66b8f345",
		},
		After: Commit{
			Revision: "0eccc6c2f01d9f055087ebbf03526ed0623e014a",
			//Revision: "84a63a65172b6c5a538a93946c8695a6d002663c",
		},
	}
}

func stack() lib.Diff {
	return lib.Diff{
		Before: Stack{
			RunImage: "index.docker.io/paketobuildpacks/run@sha256:faa51b22cdafcb971c4923b6df61f7624a3338cca9d578e569578ebc196a3813",
		},
		After: Stack{
			RunImage: "index.docker.io/paketobuildpacks/run@sha256:1b402a7bbf8f5ce530407d3d2354cd09de755e03a0a9cd926236f637bfaa0871",
		},
	}
}

func buildpacks() lib.Diff {
	return lib.Diff{
		Before: []Buildpack{
			{
				ID:      "paketo/bellsoft-liberica",
				Version: "1.2.3",
			},
			{
				ID:      "paketo/spring-boot",
				Version: "2.0.3",
			},
		},
		After: []Buildpack{
			{
				ID:      "paketo/bellsoft-liberica",
				Version: "1.2.4",
			},
			{
				ID:      "paketo/spring-boot",
				Version: "2.0.3",
			},
		},
	}
}

func config() lib.Diff {
	parse := resource.MustParse("1M")
	parse1 := resource.MustParse("12M")
	diff := lib.Diff{
		Before: Config{
			Source: v1alpha1.SourceConfig{
				Git: &v1alpha1.Git{
					URL:      "https://github.com/pivotal/kpack",
					Revision: "master",
				},
			},
			CacheSize: &parse,
			Build: v1alpha1.ImageBuild{
				Bindings: nil,
				Env: []corev1.EnvVar{
					{
						Name:  "BP_JAVA_VERSION",
						Value: "11.*",
					},
				},
				Resources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
		},
		After: Config{
			Source: v1alpha1.SourceConfig{
				Git: &v1alpha1.Git{
					URL:      "https://github.com/pivotal/kpack",
					Revision: "newbranch",
				},
			},
			CacheSize: &parse1,
			Build: v1alpha1.ImageBuild{
				Bindings: nil,
				Env: []corev1.EnvVar{
					{
						Name:  "BP_JAVA_VERSION",
						Value: "14.*",
					},
				},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu": parse,
					},
					Requests: nil,
				},
			},
		},
	}
	return diff
}

type Config struct {
	Source    v1alpha1.SourceConfig
	CacheSize *resource.Quantity
	Build     v1alpha1.ImageBuild
}

type Commit struct {
	Revision string
}

type Stack struct {
	RunImage string
}

type Buildpack struct {
	ID      string
	Version string
}

func printbuild() {
	fmt.Print(`------------------------------------------
[0mLoading secret for "https://index.docker.io/v1/" from secret "dockerhub" at location "/var/build-secrets/dockerhub"
Loading secrets for "git@github.com" from secret "blah"
Successfully cloned "https://github.com/buildpack/sample-java-app.git" @ "0eccc6c2f01d9f055087ebbf03526ed0623e014a" in path "/workspace"
[0;36m===> DETECT
[0mpaketo-buildpacks/bellsoft-liberica 2.7.4
paketo-buildpacks/maven             1.4.4
paketo-buildpacks/executable-jar    1.2.6
paketo-buildpacks/apache-tomcat     1.3.0
paketo-buildpacks/dist-zip          1.3.4
paketo-buildpacks/spring-boot       1.5.6
[0;36m===> ANALYZE
[0mRestoring metadata for "paketo-buildpacks/bellsoft-liberica:security-providers-configurer" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:class-counter" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:java-security-properties" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:jre" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:jvmkill" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:link-local-dns" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:memory-calculator" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:openssl-security-provider" from app image
Restoring metadata for "paketo-buildpacks/bellsoft-liberica:jdk" from cache
Restoring metadata for "paketo-buildpacks/maven:cache" from cache
Restoring metadata for "paketo-buildpacks/maven:application" from cache
Restoring metadata for "paketo-buildpacks/executable-jar:class-path" from app image
[0;36m===> RESTORE
[0mRestoring data for "paketo-buildpacks/bellsoft-liberica:jdk" from cache
Restoring data for "paketo-buildpacks/maven:application" from cache
Restoring data for "paketo-buildpacks/maven:cache" from cache
[0;36m===> BUILD
[0m[34m[0m
[34m[1mPaketo BellSoft Liberica Buildpack[0m[34m 2.7.4[0m
  [34;2;3mhttps://github.com/paketo-buildpacks/bellsoft-liberica[0m
  [2mBuild Configuration:[0m
[2m    $BP_JVM_VERSION              11.* (default)            the Java version[0m
  [2mLaunch Configuration:[0m
[2m    $BPL_JVM_HEAD_ROOM           0 (default)               the headroom in memory calculation[0m
[2m    $BPL_JVM_LOADED_CLASS_COUNT  35% of classes (default)  the number of loaded classes in memory calculation[0m
[2m    $BPL_JVM_THREAD_COUNT        250 (default)             the number of threads in memory calculation[0m
  [34mBellSoft Liberica JDK 11.0.7[0m: [32mReusing[0m cached layer
  [34mBellSoft Liberica JRE 11.0.7[0m: [32mReusing[0m cached layer
  [34mMemory Calculator 4.0.0[0m: [32mReusing[0m cached layer
  [34mClass Counter[0m: [32mReusing[0m cached layer
  [34mJVMKill Agent 1.16.0[0m: [32mReusing[0m cached layer
  [34mLink-Local DNS[0m: [32mReusing[0m cached layer
  [34mJava Security Properties[0m: [32mReusing[0m cached layer
  [34mSecurity Providers Configurer[0m: [32mReusing[0m cached layer
  [34mOpenSSL Certificate Loader[0m: [32mReusing[0m cached layer
[34m[0m
[34m[1mPaketo Maven Buildpack[0m[34m 1.4.4[0m
  [34;2;3mhttps://github.com/paketo-buildpacks/maven[0m
  [2mBuild Configuration:[0m
[2m    $BP_MAVEN_BUILD_ARGUMENTS  -Dmaven.test.skip=true package (default)  the arguments to pass to Maven[0m
[2m    $BP_MAVEN_BUILT_ARTIFACT   target/*.[jw]ar (default)                 the built application artifact explicitly.  Supersedes $BP_MAVEN_BUILT_MODULE[0m
[2m    $BP_MAVEN_BUILT_MODULE                                               the module to find application artifact in[0m
[2m    Creating cache directory /home/cnb/.m2[0m
  [34mCompiled Application[0m: [33mContributing[0m to layer
[2m    Executing mvnw -Dmaven.test.skip=true package[0m
[[1;34mINFO[m] Scanning for projects...
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--------------------< [0;36mio.buildpacks.example:sample[0;1m >--------------------[m
[[1;34mINFO[m] [1mBuilding sample 0.0.1-SNAPSHOT[m
[[1;34mINFO[m] [1m--------------------------------[ jar ]---------------------------------[m
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--- [0;32mmaven-resources-plugin:3.1.0:resources[m [1m(default-resources)[m @ [36msample[0;1m ---[m
[[1;34mINFO[m] Using 'UTF-8' encoding to copy filtered resources.
[[1;34mINFO[m] Copying 1 resource
[[1;34mINFO[m] Copying 4 resources
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--- [0;32mmaven-compiler-plugin:3.8.0:compile[m [1m(default-compile)[m @ [36msample[0;1m ---[m
[[1;34mINFO[m] Changes detected - recompiling the module!
[[1;34mINFO[m] Compiling 1 source file to /workspace/target/classes
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--- [0;32mmaven-resources-plugin:3.1.0:testResources[m [1m(default-testResources)[m @ [36msample[0;1m ---[m
[[1;34mINFO[m] Not copying test resources
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--- [0;32mmaven-compiler-plugin:3.8.0:testCompile[m [1m(default-testCompile)[m @ [36msample[0;1m ---[m
[[1;34mINFO[m] Not compiling test sources
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--- [0;32mmaven-surefire-plugin:2.22.1:test[m [1m(default-test)[m @ [36msample[0;1m ---[m
[[1;34mINFO[m] Tests are skipped.
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--- [0;32mmaven-jar-plugin:3.1.1:jar[m [1m(default-jar)[m @ [36msample[0;1m ---[m
[[1;34mINFO[m] Building jar: /workspace/target/sample-0.0.1-SNAPSHOT.jar
[[1;34mINFO[m] 
[[1;34mINFO[m] [1m--- [0;32mspring-boot-maven-plugin:2.1.3.RELEASE:repackage[m [1m(repackage)[m @ [36msample[0;1m ---[m
[[1;34mINFO[m] Replacing main artifact with repackaged archive
[[1;34mINFO[m] [1m------------------------------------------------------------------------[m
[[1;34mINFO[m] [1;32mBUILD SUCCESS[m
[[1;34mINFO[m] [1m------------------------------------------------------------------------[m
[[1;34mINFO[m] Total time:  3.675 s
[[1;34mINFO[m] Finished at: 2020-08-28T16:21:50Z
[[1;34mINFO[m] [1m------------------------------------------------------------------------[m
  Removing source code
[34m[0m
[34m[1mPaketo Executable JAR Buildpack[0m[34m 1.2.6[0m
  [34;2;3mhttps://github.com/paketo-buildpacks/executable-jar[0m
  Process types:
    [36mexecutable-jar[0m: java -cp "${CLASSPATH}" ${JAVA_OPTS} org.springframework.boot.loader.JarLauncher
    [36mtask[0m:           java -cp "${CLASSPATH}" ${JAVA_OPTS} org.springframework.boot.loader.JarLauncher
    [36mweb[0m:            java -cp "${CLASSPATH}" ${JAVA_OPTS} org.springframework.boot.loader.JarLauncher
[34m[0m
[34m[1mPaketo Spring Boot Buildpack[0m[34m 1.5.6[0m
  [34;2;3mhttps://github.com/paketo-buildpacks/spring-boot[0m
  Image labels:
    org.opencontainers.image.title
    org.opencontainers.image.version
    org.springframework.boot.spring-configuration-metadata.json
    org.springframework.boot.version
[0;36m===> EXPORT
[0mReusing layers from image 'index.docker.io/kpack/test@sha256:f61ca286223a29f3f1f981bb9778c5a4d1ccf4a3f285741800b94e16226975cb'
Reusing layer 'launcher'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:class-counter'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:java-security-properties'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:jre'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:jvmkill'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:link-local-dns'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:memory-calculator'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:openssl-security-provider'
Reusing layer 'paketo-buildpacks/bellsoft-liberica:security-providers-configurer'
Reusing layer 'paketo-buildpacks/executable-jar:class-path'
Reusing 1/1 app layer(s)
Adding layer 'config'
*** Images (sha256:28d30b0c4210eceaf89b2bf482f46e52de9a1b37a8357d6dcf2e57604e43ecb4):
      kpack/test
      index.docker.io/kpack/test:b12.20200828.162107
Reusing cache layer 'paketo-buildpacks/bellsoft-liberica:jdk'
Adding cache layer 'paketo-buildpacks/maven:application'
Reusing cache layer 'paketo-buildpacks/maven:cache'
[0;36m===> COMPLETION
[0mBuild successful

`)
}
