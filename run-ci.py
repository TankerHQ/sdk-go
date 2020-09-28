import argparse
import sys
import json

from typing import Any, List, Dict

from path import Path

import tankerci
from tankerci.conan import TankerSource
import tankerci.conan
import tankerci.cpp
import cli_ui as ui

PROFILE_OS_ARCHS = {
    "default": ["linux", "amd64"],
    "gcc8": ["linux", "amd64"],
    "macos": ["darwin", "amd64"],
    "mingw32": ["windows", "386"],
}


def get_deps_infos(install_path: Path) -> Any:
    libs = []
    libdirs = []
    includedirs = []
    json_file = install_path / "conanbuildinfo.json"
    conan_info = json.loads(json_file.text())
    for dep in conan_info["dependencies"]:
        libs += dep["libs"]
        libs += dep["system_libs"]
        if len(dep["lib_paths"]):
            libdirs += dep["lib_paths"]
        if dep["name"] == "tanker" and len(dep["include_paths"]):
            includedirs += dep["include_paths"]

    return {"libs": libs, "libdirs": libdirs, "includedirs": includedirs}


def generate_cgo_file(deps_infos: Any, go_os: str, go_arch: str) -> None:
    libs = " ".join([f"-l{lib}" for lib in deps_infos["libs"]])
    if go_os in ("linux", "windows"):
        libs += " -static-libstdc++ -static-libgcc"
    template_file = Path.getcwd() / "cgo_template.go.in"
    # having go_os and go_arch in the filename acts as an implicit build rule
    # e.g. only build cgo_linux_amd64.go on Linux amd64
    dst_file = Path.getcwd() / "core" / f"cgo_{go_os}_{go_arch}.go"
    ui.info_1(f"Generating {dst_file}")
    template_file.copy(dst_file)
    content = dst_file.text()
    content = content.replace(
        "{{INCLUDEDIRS}}", " ".join([f"-I{dir}" for dir in deps_infos["includedirs"]])
    )
    content = content.replace(
        "{{LIBDIRS}}", " ".join([f"-L{dir}" for dir in deps_infos["libdirs"]])
    )
    content = content.replace("{{LIBS}}", libs)
    with open(dst_file, mode="w") as f:
        f.write(content)


def platform_libnames(name: str) -> str:
    if sys.platform == "win32":
        yield f"{name}.lib"
        yield f"{name}.dll"
    elif sys.platform == "linux":
        yield f"lib{name}.a"
        yield f"lib{name}.so"
    elif sys.platform == "darwin":
        yield f"lib{name}.a"
        yield f"lib{name}.dylib"


def find_libs(names: List[str], lib_paths: List[str]) -> Path:
    for name in names:
        for lib_path in lib_paths:
            for lib in platform_libnames(name):
                candidate = Path(lib_path) / lib
                if candidate.exists():
                    yield candidate


def find_all_dep_libs(libs: List[str], lib_paths: List[str]) -> Path:
    for lib_path in find_libs(libs, lib_paths):
        yield lib_path


def copy_deps(deps_infos: Any, dest_path: Path) -> None:
    dest_path.rmtree_p()
    ui.info_1(f"creating {dest_path}")
    dest_path.mkdir_p()
    dest_lib_path = dest_path / "lib"
    dest_lib_path.mkdir_p()
    dest_include_path = dest_path / "include"
    dest_include_path.mkdir_p()
    for source_lib in find_all_dep_libs(deps_infos["libs"], deps_infos["libdirs"]):
        ui.info_1(f"copying {source_lib} -> {dest_lib_path}")
        source_lib.copy2(dest_lib_path)
    for include_dir in deps_infos["includedirs"]:
        ui.info_1(f"copying {include_dir} -> {dest_include_path}")
        Path(include_dir).merge_tree(dest_include_path)


def prepare(profile: str, tanker_source: TankerSource, update: bool) -> None:
    profile_prefix = profile.split("-")[0]
    go_os, go_arch = PROFILE_OS_ARCHS[profile_prefix]
    conan_out = Path.getcwd() / "conan"

    tankerci.conan.install_tanker_source(
        tanker_source, output_path=conan_out, profiles=[profile], update=update
    )
    conan_path = conan_out.dirs()[0]
    deps_infos = get_deps_infos(conan_path)
    install_path = Path.getcwd() / "core" / "ctanker" / f"{go_os}-{go_arch}"
    if tanker_source == TankerSource.DEPLOYED:
        copy_deps(deps_infos, install_path)
        deps_infos["libdirs"] = [install_path / "lib"]
        deps_infos["includedirs"] = [install_path / "include"]
    generate_cgo_file(deps_infos, go_os, go_arch)


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

    subparsers.add_parser("build-and-test")

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
