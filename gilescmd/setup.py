from distutils.core import setup
import shutil
import os

setup(name="gilescmd",
      version="0.2.1",
      description="Command-line utility accompanying the Giles sMAP archiver",
      author="Gabe Fierro",
      author_email="fierro@eecs.berkeley.edu",
      url="http://github.com/gtfierro/giles",
      requires=["argparse", "pandas", "pymongo", "smap"],
      install_requires=["argparse", "pandas", "pymongo", "smap"],
      scripts=["gilescmd"]
)
