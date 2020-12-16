import argparse
from pathlib import Path
import shutil
import sys
from typing import List, Optional


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
    lib_names: List[str],
    installed_include_paths: List[Path],
    go_os: str,
    go_arch: str,
) -> None:
    libs = [str(x) for x in installed_lib_paths]
    libs.extend([f"-l{lib}" for lib in lib_names])
    if go_os in ("linux", "windows"):
        libs.extend(["-static-libstdc++", "-static-libgcc"])
    template_file = Path.cwd() / "cgo_template.go.in"
    # having go_os and go_arch in the filename acts as an implicit build rule
    # e.g. only build cgo_linux_amd64.go on Linux amd64
    dst_file = Path.cwd() / "core" / f"cgo_{go_os}_{go_arch}.go"
    ui.info_1(f"Generating {dst_file}")
    shutil.copy(template_file, dst_file)
    content = dst_file.read_text()
    content = content.replace(
        "{{INCLUDEDIRS}}", " ".join([f"-I{dir}" for dir in installed_include_paths]),
    )
    content = content.replace("{{LIBS}}", " ".join(libs))
    with open(dst_file, mode="w") as f:
        f.write(content)
    ui.info_2("Generated", dst_file)


def copy_deps(deps_info: DepsConfig, dest_path: Path) -> None:
    if dest_path.exists():
        shutil.rmtree(dest_path)
    ui.info_1(f"creating {dest_path}")
    dest_path.mkdir(parents=True, exist_ok=True)
    dest_lib_path = dest_path / "lib"
    dest_lib_path.mkdir(parents=True, exist_ok=True)
    dest_include_path = dest_path / "include"
    dest_include_path.mkdir(parents=True, exist_ok=True)
    for include_dir in deps_info["tanker"].include_dirs:
        include_path = Path(include_dir)
        ui.info_1(f"copying {include_dir} -> {dest_include_path}")
        for header in include_path.glob("**/*"):
            if header.is_dir():
                continue
            rel_dir = header.parent.relative_to(include_dir)
            header_dest_dir = dest_include_path / rel_dir
            header_dest_dir.mkdir(parents=True, exist_ok=True)
            shutil.copy(header, header_dest_dir)
    for source_lib in deps_info.all_lib_paths():
        ui.info_1(f"copying {source_lib} -> {dest_lib_path}")
        shutil.copy(source_lib, dest_lib_path)


def prepare(
    profile: str,
    tanker_source: TankerSource,
    update: bool,
    tanker_ref: Optional[str] = None,
) -> None:
    profile_prefix = profile.split("-")[0]
    go_os, go_arch = PROFILE_OS_ARCHS[profile_prefix]
    conan_out = Path.cwd() / "conan"
    if tanker_source == TankerSource.DEPLOYED and not tanker_ref:
        tanker_ref = "tanker/latest-stable@"

    tankerci.conan.install_tanker_source(
        tanker_source,
        output_path=conan_out,
        profiles=[profile],
        update=update,
        tanker_deployed_ref=tanker_ref,
    )
    conan_path = [x for x in conan_out.iterdir() if x.is_dir()][0]
    deps_info = DepsConfig(conan_path)
    go_install = Path("ctanker") / f"{go_os}-{go_arch}"
    install_path = Path.cwd() / "core" / go_install

    if tanker_source == TankerSource.DEPLOYED:
        copy_deps(deps_info, install_path)
        installed_include_path = install_path / "include"
        ui.info_1(f"cleaning {installed_include_path}")
        tanker_headers = installed_include_path / "Tanker"
        helpers_headers = installed_include_path / "Helpers"
        if tanker_headers.exists():
            shutil.rmtree(tanker_headers)
        if helpers_headers.exists():
            shutil.rmtree(helpers_headers)
        generate_cgo_file(
            [Path("-L${SRCDIR}") / go_install / "lib"],
            list(deps_info.all_libs()),
            [go_install / "include"],
            go_os,
            go_arch,
        )
    else:
        generate_cgo_file(
            list(deps_info.all_lib_paths()),
            list(deps_info.all_system_libs()),
            deps_info["tanker"].include_dirs,
            go_os,
            go_arch,
        )


def build_and_test(
    profile: str, tanker_source: TankerSource, tanker_ref: Optional[str] = None
) -> None:
    prepare(profile, tanker_source, False, tanker_ref)
    # -v shows the logs as they appear, even if tests wlll succeed
    # -ginkgo.v shows the name of each test as it starts
    # -count=1 forces the tests to run instead of showing a cached result
    tankerci.run("go", "test", "./...", "-v", "-ginkgo.v", "-count=1")


def make_bump_commit(version: str):
    tankerci.bump.bump_files(version)
    cwd = Path.cwd()
    tankerci.git.run(cwd, "add", "--update")
    cgo_sources = (cwd / "core").glob("cgo_*.go")
    for cgo_source in cgo_sources:
        tankerci.git.run(cwd, "add", "--force", str(cgo_source))
    ctanker_files = (cwd / "core/ctanker").glob("**/*")
    for ctanker_file in ctanker_files:
        tankerci.git.run(cwd, "add", "--force", str(ctanker_file))
    tankerci.git.run(
        cwd, "commit", "--message", f"add binary files for version v{version}"
    )


def deploy(*, version: str) -> None:
    cwd = Path.cwd()
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
    prepare_parser.add_argument("--tanker-ref")
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
    build_parser.add_argument("--tanker-ref")

    subparsers.add_parser("mirror")

    deploy_parser = subparsers.add_parser("deploy")
    deploy_parser.add_argument("--version", required=True)

    args = parser.parse_args()
    if args.home_isolation:
        tankerci.conan.set_home_isolation()
        tankerci.conan.update_config()

    if args.command == "prepare":
        prepare(
            args.profile, args.tanker_source, args.update, tanker_ref=args.tanker_ref
        )
    elif args.command == "build-and-test":
        build_and_test(args.profile, args.tanker_source, tanker_ref=args.tanker_ref)
    elif args.command == "deploy":
        deploy(version=args.version)
    elif args.command == "mirror":
        tankerci.git.mirror(github_url="git@github.com:TankerHQ/sdk-go")
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()
