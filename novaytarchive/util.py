import platform
import os
import shlex

from subprocess import Popen, PIPE
from typing import Tuple


def golang_target(from_env=False, with_go=False) -> Tuple[str, str]:
    """ (try to) Get GOOS and GOARCH from running Python interpreter """
    if from_env:
        try:
            return os.environ["GOOS"], os.environ["GOARCH"]
        except KeyboardInterrupt:
            pass

    if with_go:
        newenv = dict(os.environ)
        newenv.pop("GOOS", None)
        newenv.pop("GOARCH", None)
        try:
            with Popen(["go", "tool", "dist", "env"], stdout=PIPE, text=True, env=newenv) as p:
                lex = shlex.shlex(p.stdout, posix=True)
                lex.whitespace_split = True
                envs = dict(x.split("=", 1) for x in lex)
            return envs["GOOS"], envs["GOARCH"]
        except (OSError, KeyError):
            pass
    # run "go tool dist list" to get all combinations of GOOS and GOARCH
    # ref. https://wiki.debian.org/Multiarch/Tuples
    # ref. https://stackoverflow.com/questions/45125516/possible-values-for-uname-m
    # Linux x86_64: uname_result(system='Linux', node=..., release=..., version=..., machine='x86_64')
    # Linux aarch64: uname_result(system='Linux', node=..., release=..., version=..., machine='aarch64', processor='aarch64')
    # Windows x86_64: uname_result(system='Windows', node=..., release=..., version=..., machine='AMD64', processor='Intel64 Family 6 Model 126 Stepping 5, GenuineIntel')
    # Windows ARM: system=Windows, machine=ARM64? (ref. https://docs.microsoft.com/en-us/windows/win32/winprog64/wow64-implementation-details?redirectedfrom=MSDN )
    un = platform.uname()
    goos = un.system.lower()
    goarch = un.machine.lower()
    goarch = {
        'x86_64': 'amd64',  # Linux
        'ia64': 'amd64',  # Windows/Linux IA64
        'i386': '386',  # Linux
        'x86': '386',  # Windows/Linux
        'aarch64': 'arm64',  # Linux
    }.get(goarch, goarch)

    return goos, goarch

def derive_binname(goos: str, goarch: str) -> str:
    """ Generates binary name matching the given system """
    ext = ".so"
    if goos == "windows":
        ext = ".dll"
    elif goos == "darwin":
        ext = ".dylib"
    return f"_ytarchive_{goos}_{goarch}{ext}"
