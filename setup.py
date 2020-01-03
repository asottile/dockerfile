import os
import sys

from setuptools import Extension
from setuptools import setup

if not os.path.exists('vendor/github.com/moby/buildkit/frontend'):
    sys.exit('moby checkout is missing!\n'
             'Run `git submodule update --init`')

setup(
    ext_modules=[Extension('dockerfile', ['pylib/main.go'])],
    build_golang={'root': 'github.com/asottile/dockerfile'},
)
