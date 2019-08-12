import os

from setuptools import Extension
from setuptools import setup

if not os.path.exists('vendor/github.com/moby/buildkit/frontend'):
    print('moby checkout is missing!')
    print('Run `git submodule update --init`')
    exit(1)

setup(
    ext_modules=[Extension('dockerfile', ['pylib/main.go'])],
    build_golang={'root': 'github.com/asottile/dockerfile'},
)
