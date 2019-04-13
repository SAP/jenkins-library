from setuptools import setup, find_packages
from codecs import open
from os import path

def get_version():
    with open('version.txt') as ver_file:
        version_str = ver_file.readline().rstrip()
    return version_str


def get_install_requires():
    with open('requirements.txt') as reqs_file:
        reqs = [line.rstrip() for line in reqs_file.readlines()]
    return reqs

setup(name="some-test",
      version=get_version(),
      python_requires='>=3',
      packages=find_packages(exclude=['contrib', 'docs', 'tests*', 'coverage', 'bin']),
      description="test",
      install_requires=get_install_requires(),
     )
