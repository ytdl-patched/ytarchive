try:
    from setuptools import Command, setup
except ImportError:
    from distutils.core import Command, setup
from distutils.spawn import find_executable

from novaytarchive.util import derive_binname, golang_target

# we query the system or env for cross-compiling
go_target = golang_target(True, True)
binname = derive_binname(*go_target)

def du_spawn(cmd, dry_run=0, env=None):
    """ Modified version of distutils.spawn.spawn that accepts env """

    import subprocess

    from distutils.errors import DistutilsExecError
    from distutils.debug import DEBUG
    from distutils import log

    # cmd is documented as a list, but just in case some code passes a tuple
    # in, protect our %-formatting code against horrible death
    cmd = list(cmd)

    log.info(' '.join(cmd))
    if dry_run:
        return

    executable = find_executable(cmd[0])
    if executable is not None:
        cmd[0] = executable

    try:
        proc = subprocess.Popen(cmd, env=env)
        proc.wait()
        exitcode = proc.returncode
    except OSError as exc:
        if not DEBUG:
            cmd = cmd[0]
        raise DistutilsExecError(
            "command %r failed: %s" % (cmd, exc.args[-1])) from exc

    if exitcode:
        if not DEBUG:
            cmd = cmd[0]
        raise DistutilsExecError(
              "command %r failed with exit code %s" % (cmd, exitcode))

class BuildGolangPart(Command):
    description = 'Build the golang counterpart of this module'
    user_options = []

    def initialize_options(self):
        pass

    def finalize_options(self):
        pass

    def run(self):
        du_spawn([
            "go", "build",
            "-buildmode=c-shared", "-o",
            binname],
        dry_run=self.dry_run,
        env={
            "GOOS": go_target[0],
            "GOARCH": go_target[1],
            "CGO_ENABLED": "1",
        })

setup(
    name='nova-ytarchive',
    version='0.1.0',
    maintainer='Lesmiscore',
    maintainer_email='nao20010128@gmail.com',
    description='Golang version of ytarchive to be called from Python',
    url='https://github.com/ytdl-patched/ytarchive',
    packages=['novaytarchive'],
    project_urls={
        'Source': 'https://github.com/ytdl-patched/ytarchive',
    },
    classifiers=[
        'Development Status :: 4 - Beta',
        'Programming Language :: Python',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'License :: OSI Approved :: MIT License',
    ],
    python_requires='>=3.6',
    data_files=[binname],

    cmdclass={'BuildGolangPart': BuildGolangPart},
)