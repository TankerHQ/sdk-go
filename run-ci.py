import argparse
import os
import sys
import json

from path import Path

import tankerci
import tankerci.conan
import tankerci.cpp
import cli_ui as ui

PROFILE_OS_ARCHS = {
    "default": ["linux", "amd64"],
    "gcc8": ["linux", "amd64"],
    "macos": ["darwin", "amd64"],
    "mingw32": ["windows", "386"],
}


def install_tanker_native(profile: str, install_folder: Path, use_tanker: str) -> None:
    cwd = Path.getcwd()
    conanfile = cwd / "conanfile-local.txt"

    install_args = []
    if use_tanker == "deployed":
        conanfile = cwd / "conanfile-deployed.txt"
    elif use_tanker == "local":
        tankerci.conan.export(
            src_path=Path.getcwd().parent / "sdk-native", ref_or_channel="tanker/dev"
        )
    elif use_tanker == "same-as-branch":
        workspace = tankerci.git.prepare_sources(repos=["sdk-native", "sdk-go"])
        tankerci.conan.export(
            src_path=workspace / "sdk-native", ref_or_channel="tanker/dev"
        )
    # fmt: off
    tankerci.conan.run(
        "install", conanfile,
        "--update",
        "--profile", profile,
        "--install-folder", install_folder,
        "--generator", "json",
        *install_args,
    )
    # fmt: on


def get_deps_link_flags(install_path: Path) -> str:
    json_file = install_path / "conanbuildinfo.json"
    conan_info = json.loads(json_file.text())
    deps = []
    for dep in conan_info["dependencies"]:
        deps += dep["libs"]
        deps += dep["system_libs"]
    return " ".join([f"-l{d}" for d in deps])


def generate_cgo_file(install_path: Path, go_os: str, go_arch: str) -> None:
    link_flags = get_deps_link_flags(install_path)
    if go_os in ("linux", "windows"):
        link_flags += " -static-libstdc++ -static-libgcc"
    template_file = Path.getcwd() / "cgo_template.go.in"
    # having go_os and go_arch in the filename acts as an implicit build rule
    # e.g. only build cgo_linux_amd64.go on Linux amd64
    dst_file = Path.getcwd() / "core" / f"cgo_{go_os}_{go_arch}.go"
    ui.info_1(f"Generating {dst_file}")
    template_file.copy(dst_file)
    content = dst_file.text()
    content = content.replace("{{GO_OS}}", go_os)
    content = content.replace("{{GO_ARCH}}", go_arch)
    content = content.replace("{{CONAN_LIBS}}", link_flags)
    with open(dst_file, mode="w") as f:
        f.write(content)


def install_deps(profile: str, use_tanker: str) -> None:
    profile_prefix = profile.split("-")[0]
    go_os, go_arch = PROFILE_OS_ARCHS[profile_prefix]
    deps_install_path = Path.getcwd() / "core/ctanker" / f"{go_os}-{go_arch}"
    deps_install_path.rmtree_p()

    install_tanker_native(profile, deps_install_path, use_tanker)
    generate_cgo_file(deps_install_path, go_os, go_arch)


def build_and_check() -> None:
    # -v shows the logs as they appear, even if tests wlll succeed
    # -ginkgo.v shows the name of each test as it starts
    # -count=1 forces the tests to run instead of showing a cached result
    tankerci.run("go", "test", "./...", "-v", "-ginkgo.v", "-count=1")


def make_bump_commit(version: str):
    tankerci.bump.bump_files(version)
    cwd = Path.getcwd()
    tankerci.git.run(cwd, "add", "--update")
    cgo_sources = (cwd / "core").files("cgo_*.go")
    for cgo_source in cgo_sources:
        tankerci.git.run(cwd, "add", "--force", cgo_source)
    ctanker_files = (cwd / "core/ctanker").walkfiles()
    for ctanker_file in ctanker_files:
        tankerci.git.run(cwd, "add", "--force", ctanker_file)
    tankerci.git.run(
        cwd, "commit", "--message", f"add binary files for version v{version}"
    )


def deploy(*, version: str) -> None:
    cwd = Path.getcwd()
    tag = "v" + version
    make_bump_commit(version)
    tankerci.git.run(cwd, "tag", tag)
    github_url = "git@github.com:TankerHQ/sdk-go"
    tankerci.git.run(cwd, "push", github_url, f"{tag}:{tag}")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--isolate-conan-user-home",
        action="store_true",
        dest="home_isolation",
        default=False,
    )
    subparsers = parser.add_subparsers(title="subcommands", dest="command")

    install_deps_parser = subparsers.add_parser("install-deps")
    install_deps_parser.add_argument(
        "--use-tanker", choices=["deployed", "local", "same-as-branch"], default="local"
    )
    install_deps_parser.add_argument("--profile", required=True)

    subparsers.add_parser("build-and-test")

    subparsers.add_parser("mirror")

    deploy_parser = subparsers.add_parser("deploy")
    deploy_parser.add_argument("--version", required=True)

    args = parser.parse_args()
    if args.home_isolation:
        tankerci.conan.set_home_isolation()
        tankerci.conan.update_config()

    if args.command == "install-deps":
        install_deps(args.profile, args.use_tanker)
    elif args.command == "build-and-test":
        build_and_check()
    elif args.command == "deploy":
        deploy(version=args.version)
    elif args.command == "mirror":
        tankerci.git.mirror(github_url="git@github.com:TankerHQ/sdk-go")
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()
