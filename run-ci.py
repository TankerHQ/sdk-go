import argparse
import sys
import json

from typing import Any, List

from path import Path

import tankerci
from tankerci.conan import TankerSource
from tankerci.build_info import DepsConfig
import tankerci.conan
import tankerci.cpp
import cli_ui as ui

PROFILE_OS_ARCHS = {
    "default": ["linux", "amd64"],
    "gcc8": ["linux", "amd64"],
    "macos": ["darwin", "amd64"],
    "mingw32": ["windows", "386"],
}


def generate_cgo_file(
    installed_lib_paths: List[Path],
    system_libs: List[str],
    installed_include_paths: List[Path],
    go_os: str,
    go_arch: str,
) -> None:
    libs = installed_lib_paths
    libs.extend([f"-l{lib}" for lib in system_libs])
    if go_os in ("linux", "windows"):
        libs.extend(["-static-libstdc++", "-static-libgcc"])
    template_file = Path.getcwd() / "cgo_template.go.in"
    # having go_os and go_arch in the filename acts as an implicit build rule
    # e.g. only build cgo_linux_amd64.go on Linux amd64
    dst_file = Path.getcwd() / "core" / f"cgo_{go_os}_{go_arch}.go"
    ui.info_1(f"Generating {dst_file}")
    template_file.copy(dst_file)
    content = dst_file.text()
    content = content.replace(
        "{{INCLUDEDIRS}}", " ".join([f"-I{dir}" for dir in installed_include_paths]),
    )
    content = content.replace("{{LIBS}}", " ".join(libs))
    with open(dst_file, mode="w") as f:
        f.write(content)


def copy_deps(deps_info: DepsConfig, dest_path: Path) -> None:
    dest_path.rmtree_p()
    ui.info_1(f"creating {dest_path}")
    dest_path.makedirs_p()
    dest_lib_path = dest_path / "lib"
    dest_lib_path.makedirs_p()
    dest_include_path = dest_path / "include"
    dest_include_path.makedirs_p()
    lib_paths: List[Path] = []
    for include_dir in deps_info["tanker"].include_dirs:
        ui.info_1(f"copying {include_dir} -> {dest_include_path}")
        Path(include_dir).merge_tree(dest_include_path)
    for source_lib in deps_info.all_lib_paths():
        ui.info_1(f"copying {source_lib} -> {dest_lib_path}")
        lib_paths.append(source_lib.copy2(dest_lib_path))
    return lib_paths


def prepare(profile: str, tanker_source: TankerSource, update: bool) -> None:
    profile_prefix = profile.split("-")[0]
    go_os, go_arch = PROFILE_OS_ARCHS[profile_prefix]
    conan_out = Path.getcwd() / "conan"

    tankerci.conan.install_tanker_source(
        tanker_source, output_path=conan_out, profiles=[profile], update=update
    )
    conan_path = conan_out.dirs()[0]
    deps_info = DepsConfig(conan_path)
    install_path = Path.getcwd() / "core" / "ctanker" / f"{go_os}-{go_arch}"
    if tanker_source == TankerSource.DEPLOYED:
        installed_lib_paths = copy_deps(deps_info, install_path)
        installed_include_path = install_path / "include"
        ui.info_1(f"cleaning {installed_include_path}")
        with installed_include_path:
            Path("Tanker").rmtree_p()
            Path("Helpers").rmtree_p()
        generate_cgo_file(
            installed_lib_paths,
            deps_info.all_system_libs(),
            [installed_include_path],
            go_os,
            go_arch,
        )
    else:
        print(deps_info["tanker"].include_dirs)
        generate_cgo_file(
            deps_info.all_lib_paths(),
            deps_info.all_system_libs(),
            deps_info["tanker"].include_dirs,
            go_os,
            go_arch,
        )


def build_and_test(profile: str, tanker_source: TankerSource) -> None:
    prepare(profile, tanker_source, False)
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

    prepare_parser = subparsers.add_parser("prepare")
    prepare_parser.add_argument(
        "--use-tanker",
        type=TankerSource,
        default=TankerSource.EDITABLE,
        dest="tanker_source",
    )
    prepare_parser.add_argument("--profile", required=True)
    prepare_parser.add_argument(
        "--update", action="store_true", default=False, dest="update"
    )

    build_parser = subparsers.add_parser("build-and-test")
    build_parser.add_argument(
        "--use-tanker",
        type=TankerSource,
        default=TankerSource.EDITABLE,
        dest="tanker_source",
    )
    build_parser.add_argument("--profile", default="default", required=True)

    subparsers.add_parser("mirror")

    deploy_parser = subparsers.add_parser("deploy")
    deploy_parser.add_argument("--version", required=True)

    args = parser.parse_args()
    if args.home_isolation:
        tankerci.conan.set_home_isolation()
        tankerci.conan.update_config()

    if args.command == "prepare":
        prepare(args.profile, args.tanker_source, args.update)
    elif args.command == "build-and-test":
        build_and_test(args.profile, args.tanker_source)
    elif args.command == "deploy":
        deploy(version=args.version)
    elif args.command == "mirror":
        tankerci.git.mirror(github_url="git@github.com:TankerHQ/sdk-go")
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()
